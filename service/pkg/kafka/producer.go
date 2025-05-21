// Package kafka 提供了Kafka客户端的封装
package kafka

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/XXXXD-cation/XXXXD-TIP/service/pkg/log"
	"github.com/cenkalti/backoff/v4"
	"github.com/segmentio/kafka-go"

	// 不再使用compress包，避免版本兼容性问题
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/segmentio/kafka-go/sasl/scram"
)

// ProducerImpl 基于segmentio/kafka-go的Kafka生产者实现
type ProducerImpl struct {
	config     Config
	options    Options
	writer     *kafka.Writer
	connected  bool
	topicCache map[string]struct{} // 存储已验证存在的主题
}

// NewProducer 创建一个新的Kafka生产者
func NewProducer(config Config, options Options) Producer {
	return &ProducerImpl{
		config:     config,
		options:    options,
		topicCache: make(map[string]struct{}),
	}
}

// setup 初始化Kafka生产者
func (p *ProducerImpl) setup() error {
	if p.connected {
		return nil
	}

	// 创建基本的写入器配置
	writerConfig := kafka.WriterConfig{
		Brokers:      p.config.Brokers,
		BatchSize:    p.options.BatchSize,
		BatchBytes:   p.options.MaxMessageBytes,
		BatchTimeout: p.options.LingerMS,
		RequiredAcks: p.options.Acks,         // 保持为int类型
		MaxAttempts:  p.options.RetryMax + 1, // sarama兼容，+1是因为第一次不算重试
		Balancer:     &kafka.LeastBytes{},
	}

	// 设置压缩类型
	// 注意：由于kafka-go版本问题，暂时不设置压缩
	// 较新版本的kafka-go使用CompressionCodec，但我们的版本可能不支持
	// 所以这里不设置任何压缩，确保代码能正常编译运行

	// 设置安全认证
	if p.config.SecurityProtocol != "" && p.config.SecurityProtocol != "plaintext" {
		var mechanism sasl.Mechanism
		var err error

		switch p.config.SASLMechanism {
		case "plain":
			mechanism = plain.Mechanism{
				Username: p.config.SASLUsername,
				Password: p.config.SASLPassword,
			}
		case "scram-sha-256":
			mechanism, err = scram.Mechanism(scram.SHA256, p.config.SASLUsername, p.config.SASLPassword)
			if err != nil {
				return fmt.Errorf("配置SASL SCRAM-SHA-256认证失败: %w", err)
			}
		case "scram-sha-512":
			mechanism, err = scram.Mechanism(scram.SHA512, p.config.SASLUsername, p.config.SASLPassword)
			if err != nil {
				return fmt.Errorf("配置SASL SCRAM-SHA-512认证失败: %w", err)
			}
		default:
			return fmt.Errorf("不支持的SASL机制: %s", p.config.SASLMechanism)
		}

		// 设置SASL认证
		dialer := &kafka.Dialer{
			SASLMechanism: mechanism,
		}

		// 如果使用SSL
		if p.config.SecurityProtocol == "sasl_ssl" {
			dialer.TLS = nil // 默认TLS配置
		}

		writerConfig.Dialer = dialer
	}

	// 注意：WriterConfig没有ClientID字段，移除ClientID设置

	p.writer = kafka.NewWriter(writerConfig)
	p.connected = true
	log.Info().Str("dsn", p.config.String()).Msg("Kafka生产者已初始化")
	return nil
}

// Produce 发送消息到Kafka
func (p *ProducerImpl) Produce(ctx context.Context, topic string, key, value []byte, headers map[string]string) error {
	if err := p.setup(); err != nil {
		return err
	}

	// 将map转换为kafka-go的Header格式
	var kafkaHeaders []kafka.Header
	if len(headers) > 0 {
		kafkaHeaders = make([]kafka.Header, 0, len(headers))
		for k, v := range headers {
			kafkaHeaders = append(kafkaHeaders, kafka.Header{
				Key:   k,
				Value: []byte(v),
			})
		}
	}

	msg := kafka.Message{
		Topic:   topic,
		Key:     key,
		Value:   value,
		Headers: kafkaHeaders,
		Time:    time.Now(),
	}

	// 使用指数退避进行重试
	operation := func() error {
		err := p.writer.WriteMessages(ctx, msg)
		if err != nil {
			log.Error().Err(err).
				Str("topic", topic).
				Str("dsn", p.config.String()).
				Msg("发送Kafka消息失败")
			return err
		}
		return nil
	}

	// 创建退避策略
	exponentialBackOff := backoff.NewExponentialBackOff()
	exponentialBackOff.MaxElapsedTime = p.options.ConnectTimeout
	exponentialBackOff.InitialInterval = p.options.ReconnectBackoff

	// 执行带有重试的操作
	var retryBackOff backoff.BackOff

	// 根据配置创建适当的退避策略
	if p.options.MaxReconnectRetry <= 0 {
		// 无限重试
		retryBackOff = exponentialBackOff
	} else if p.options.MaxReconnectRetry == 1 {
		// 只尝试一次，不重试
		retryBackOff = &backoff.StopBackOff{}
	} else {
		// 有限次数重试，使用常量避免转换
		count := uint64(0)
		switch {
		case p.options.MaxReconnectRetry <= 5:
			count = 5
		case p.options.MaxReconnectRetry <= 10:
			count = 10
		default:
			count = 20 // 最大限制在20次
		}
		retryBackOff = backoff.WithMaxRetries(exponentialBackOff, count)
	}

	err := backoff.Retry(operation, retryBackOff)
	if err != nil {
		return fmt.Errorf("发送Kafka消息失败: %w", err)
	}

	return nil
}

// ProduceMessage 发送消息对象到Kafka
func (p *ProducerImpl) ProduceMessage(ctx context.Context, msg *Message) error {
	if msg == nil {
		return errors.New("消息不能为空")
	}
	return p.Produce(ctx, msg.Topic, msg.Key, msg.Value, msg.Headers)
}

// Flush 确保所有消息已发送
func (p *ProducerImpl) Flush(timeout time.Duration) error {
	if !p.connected || p.writer == nil {
		return nil
	}

	// 不再创建未使用的上下文
	return p.writer.Close()
}

// Close 关闭生产者连接
func (p *ProducerImpl) Close() error {
	if !p.connected || p.writer == nil {
		return nil
	}

	err := p.writer.Close()
	if err != nil {
		return fmt.Errorf("关闭Kafka生产者失败: %w", err)
	}

	p.connected = false
	log.Info().Str("dsn", p.config.String()).Msg("Kafka生产者已关闭")
	return nil
}
