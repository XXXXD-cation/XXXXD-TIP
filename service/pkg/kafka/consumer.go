// Package kafka 提供了Kafka客户端的封装
package kafka

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/XXXXD-cation/XXXXD-TIP/service/pkg/log"
	// backoff包在当前消费者实现中未使用，移除导入
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/segmentio/kafka-go/sasl/scram"
)

// ConsumerImpl 基于segmentio/kafka-go的Kafka消费者实现
type ConsumerImpl struct {
	config       Config
	options      Options
	reader       *kafka.Reader
	topics       []string
	connected    bool
	mu           sync.Mutex
	pausedTopics map[string]struct{}
	cancelFunc   context.CancelFunc
}

// NewConsumer 创建一个新的Kafka消费者
func NewConsumer(config Config, options Options) Consumer {
	if config.ConsumerGroupID == "" {
		config.ConsumerGroupID = config.ClientID + "-group"
	}

	return &ConsumerImpl{
		config:       config,
		options:      options,
		pausedTopics: make(map[string]struct{}),
	}
}

// Subscribe 订阅指定主题
func (c *ConsumerImpl) Subscribe(topics ...string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 如果已经连接，不允许更改订阅
	if c.connected {
		return errors.New("已连接的消费者不能改变订阅")
	}

	if len(topics) == 0 {
		return errors.New("至少需要订阅一个主题")
	}

	c.topics = make([]string, len(topics))
	copy(c.topics, topics)
	return nil
}

// setup 初始化Kafka消费者
func (c *ConsumerImpl) setup() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return nil
	}

	if len(c.topics) == 0 {
		return errors.New("未订阅任何主题")
	}

	// 创建消费者配置
	// 如果有多个主题，需要记录日志，并只使用第一个主题
	if len(c.topics) > 1 {
		log.Warn().
			Strs("topics", c.topics).
			Msg("kafka-go Reader只支持一个主题，将只使用第一个主题")
	}

	readerConfig := kafka.ReaderConfig{
		Brokers:               c.config.Brokers,
		GroupID:               c.config.ConsumerGroupID,
		Topic:                 c.topics[0], // kafka-go只支持单个Topic
		MinBytes:              1e3,         // 1KB
		MaxBytes:              c.options.MaxPartitionFetchBytes,
		MaxWait:               c.options.FetchMaxWaitMS,
		ReadBackoffMin:        c.options.ReconnectBackoff / 2,
		ReadBackoffMax:        c.options.ReconnectBackoff * 2,
		CommitInterval:        c.options.AutoCommitInterval,
		StartOffset:           kafka.FirstOffset, // 默认从最旧的消息开始
		ReadLagInterval:       0,                 // 不跟踪滞后
		HeartbeatInterval:     5 * time.Second,
		SessionTimeout:        c.options.SessionTimeout,
		RebalanceTimeout:      c.options.SessionTimeout * 2,
		RetentionTime:         24 * time.Hour, // 消息保留时间
		WatchPartitionChanges: true,
		OffsetOutOfRangeError: true,
		Logger:                nil, // 使用默认logger
	}

	// 设置自动偏移量重置策略
	switch c.options.AutoOffsetReset {
	case "earliest":
		readerConfig.StartOffset = kafka.FirstOffset
	case "latest":
		readerConfig.StartOffset = kafka.LastOffset
	default:
		readerConfig.StartOffset = kafka.LastOffset
	}

	// 设置安全认证
	if c.config.SecurityProtocol != "" && c.config.SecurityProtocol != "plaintext" {
		var mechanism sasl.Mechanism
		var err error

		switch c.config.SASLMechanism {
		case "plain":
			mechanism = plain.Mechanism{
				Username: c.config.SASLUsername,
				Password: c.config.SASLPassword,
			}
		case "scram-sha-256":
			mechanism, err = scram.Mechanism(scram.SHA256, c.config.SASLUsername, c.config.SASLPassword)
			if err != nil {
				return fmt.Errorf("配置SASL SCRAM-SHA-256认证失败: %w", err)
			}
		case "scram-sha-512":
			mechanism, err = scram.Mechanism(scram.SHA512, c.config.SASLUsername, c.config.SASLPassword)
			if err != nil {
				return fmt.Errorf("配置SASL SCRAM-SHA-512认证失败: %w", err)
			}
		default:
			return fmt.Errorf("不支持的SASL机制: %s", c.config.SASLMechanism)
		}

		// 设置SASL认证
		dialer := &kafka.Dialer{
			SASLMechanism: mechanism,
		}

		// 如果使用SSL
		if c.config.SecurityProtocol == "sasl_ssl" {
			dialer.TLS = nil // 默认TLS配置
		}

		readerConfig.Dialer = dialer
	}

	// 注意：kafka-go的Reader不支持设置ClientID
	// 移除ClientID设置
	// 如果需要使用多个主题，则需要为每个主题创建一个Reader，或使用ConsumerGroup

	// 创建Reader
	c.reader = kafka.NewReader(readerConfig)
	c.connected = true
	log.Info().
		Str("dsn", c.config.String()).
		Strs("topics", c.topics).
		Msg("Kafka消费者已初始化")

	return nil
}

// Consume 开始消费消息并处理
func (c *ConsumerImpl) Consume(ctx context.Context, handler func(msg *Message) error) error {
	if err := c.setup(); err != nil {
		return err
	}

	// 创建可取消的上下文
	ctx, cancel := context.WithCancel(ctx)
	c.cancelFunc = cancel

	log.Info().
		Str("dsn", c.config.String()).
		Strs("topics", c.topics).
		Msg("开始消费Kafka消息")

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Kafka消费已停止")
			return ctx.Err()
		default:
			kafkaMsg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if errors.Is(err, io.EOF) || errors.Is(err, context.Canceled) {
					return nil
				}

				log.Error().Err(err).Str("dsn", c.config.String()).Msg("读取Kafka消息失败")

				// 使用退避策略重试
				time.Sleep(c.options.ReconnectBackoff)
				continue
			}

			// 检查是否为暂停的主题
			c.mu.Lock()
			_, isPaused := c.pausedTopics[kafkaMsg.Topic]
			c.mu.Unlock()

			if isPaused {
				log.Debug().Str("topic", kafkaMsg.Topic).Msg("跳过暂停的主题消息")
				continue
			}

			// 构建消息对象
			headers := make(map[string]string, len(kafkaMsg.Headers))
			for _, h := range kafkaMsg.Headers {
				headers[h.Key] = string(h.Value)
			}

			msg := &Message{
				Topic:     kafkaMsg.Topic,
				Key:       kafkaMsg.Key,
				Value:     kafkaMsg.Value,
				Headers:   headers,
				Partition: kafkaMsg.Partition,
				Offset:    kafkaMsg.Offset,
				Timestamp: kafkaMsg.Time,
			}

			// 处理消息
			err = handler(msg)
			if err != nil {
				log.Error().
					Err(err).
					Str("topic", msg.Topic).
					Int("partition", msg.Partition).
					Int64("offset", msg.Offset).
					Msg("处理Kafka消息失败")
			}

			// 如果自动提交被禁用且没有发生处理错误，则手动提交
			if !c.options.EnableAutoCommit && err == nil {
				if commitErr := c.Commit(ctx, msg); commitErr != nil {
					log.Error().
						Err(commitErr).
						Str("topic", msg.Topic).
						Int("partition", msg.Partition).
						Int64("offset", msg.Offset).
						Msg("提交Kafka消息偏移量失败")
				}
			}
		}
	}
}

// Commit 提交指定消息的偏移量
func (c *ConsumerImpl) Commit(ctx context.Context, msg *Message) error {
	if c.reader == nil {
		return errors.New("Kafka消费者未连接")
	}

	err := c.reader.CommitMessages(ctx, kafka.Message{
		Topic:     msg.Topic,
		Partition: msg.Partition,
		Offset:    msg.Offset,
	})

	if err != nil {
		return fmt.Errorf("提交偏移量失败: %w", err)
	}

	return nil
}

// Close 关闭消费者连接
func (c *ConsumerImpl) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cancelFunc != nil {
		c.cancelFunc()
	}

	if !c.connected || c.reader == nil {
		return nil
	}

	err := c.reader.Close()
	if err != nil {
		return fmt.Errorf("关闭Kafka消费者失败: %w", err)
	}

	c.connected = false
	log.Info().Str("dsn", c.config.String()).Msg("Kafka消费者已关闭")
	return nil
}

// Pause 暂停指定主题的消费
func (c *ConsumerImpl) Pause(topics ...string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, topic := range topics {
		c.pausedTopics[topic] = struct{}{}
		log.Info().Str("topic", topic).Msg("已暂停主题消费")
	}

	return nil
}

// Resume 恢复指定主题的消费
func (c *ConsumerImpl) Resume(topics ...string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, topic := range topics {
		delete(c.pausedTopics, topic)
		log.Info().Str("topic", topic).Msg("已恢复主题消费")
	}

	return nil
}
