// Package kafka 提供了Kafka客户端的封装，包括生产者和消费者接口及实现
package kafka

import (
	"context"
	"time"
)

// Message 表示一条Kafka消息
type Message struct {
	Topic     string
	Key       []byte
	Value     []byte
	Headers   map[string]string
	Partition int
	Offset    int64
	Timestamp time.Time
}

// Producer 定义消息生产者接口
type Producer interface {
	// Produce 发送消息到Kafka
	Produce(ctx context.Context, topic string, key, value []byte, headers map[string]string) error

	// ProduceMessage 发送完整的消息对象
	ProduceMessage(ctx context.Context, msg *Message) error

	// Flush 确保所有消息已发送
	Flush(timeout time.Duration) error

	// Close 关闭生产者连接
	Close() error
}

// Consumer 定义消息消费者接口
type Consumer interface {
	// Subscribe 订阅指定主题
	Subscribe(topics ...string) error

	// Consume 开始消费消息，处理函数会被用于处理每条接收的消息
	Consume(ctx context.Context, handler func(msg *Message) error) error

	// Commit 手动提交指定消息的偏移量
	Commit(ctx context.Context, msg *Message) error

	// Close 关闭消费者连接
	Close() error

	// Pause 暂停指定主题的消费
	Pause(topics ...string) error

	// Resume 恢复指定主题的消费
	Resume(topics ...string) error
}

// Options Kafka客户端配置选项
type Options struct {
	// 通用选项
	ConnectTimeout    time.Duration // 连接超时
	SessionTimeout    time.Duration // 会话超时
	ReconnectBackoff  time.Duration // 重连退避间隔
	MaxReconnectRetry int           // 最大重试次数

	// Producer选项
	Acks             int           // 确认模式：0=不等待确认，1=等待leader确认，-1=等待所有副本确认
	RetryMax         int           // 最大重试次数
	BatchSize        int           // 批处理大小
	LingerMS         time.Duration // 发送前等待时间
	CompressionType  string        // 压缩类型：none, gzip, snappy, lz4
	MaxMessageBytes  int           // 最大消息大小
	EnableIdempotent bool          // 启用幂等性

	// Consumer选项
	GroupID                string        // 消费组ID
	AutoOffsetReset        string        // 自动偏移量重置：earliest, latest
	EnableAutoCommit       bool          // 启用自动提交
	AutoCommitInterval     time.Duration // 自动提交间隔
	MaxPollIntervalMS      time.Duration // 最大轮询间隔
	FetchMaxWaitMS         time.Duration // 最大等待时间
	MaxPartitionFetchBytes int           // 每个分区最大获取字节数
}

// DefaultOptions 返回默认配置选项
func DefaultOptions() Options {
	return Options{
		ConnectTimeout:    10 * time.Second,
		SessionTimeout:    30 * time.Second,
		ReconnectBackoff:  100 * time.Millisecond,
		MaxReconnectRetry: 5,

		// Producer默认选项
		Acks:             -1,
		RetryMax:         3,
		BatchSize:        16384,
		LingerMS:         100 * time.Millisecond,
		CompressionType:  "snappy",
		MaxMessageBytes:  1000000,
		EnableIdempotent: true,

		// Consumer默认选项
		AutoOffsetReset:        "latest",
		EnableAutoCommit:       true,
		AutoCommitInterval:     5 * time.Second,
		MaxPollIntervalMS:      300 * time.Second,
		FetchMaxWaitMS:         500 * time.Millisecond,
		MaxPartitionFetchBytes: 1048576,
	}
}
