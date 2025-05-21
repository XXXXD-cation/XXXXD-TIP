package kafka

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 创建配置等辅助函数
func createTestConfig() Config {
	return Config{
		Brokers:         []string{"localhost:9093"},
		ClientID:        "test-client",
		ConsumerGroupID: "test-group",
	}
}

// 测试配置DSN和String方法
func TestConfigDSNAndString(t *testing.T) {
	tests := []struct {
		name           string
		config         Config
		expectedDSN    string
		expectedString string
	}{
		{
			name: "基本配置",
			config: Config{
				Brokers: []string{"localhost:9093"},
			},
			expectedDSN:    "kafka://localhost:9093",
			expectedString: "kafka://localhost:9093",
		},
		{
			name: "多个Broker",
			config: Config{
				Brokers: []string{"broker1:9093", "broker2:9093"},
			},
			expectedDSN:    "kafka://broker1:9093,broker2:9093",
			expectedString: "kafka://broker1:9093,broker2:9093",
		},
		{
			name: "带认证的配置",
			config: Config{
				Brokers:          []string{"localhost:9093"},
				SecurityProtocol: "sasl_plaintext",
				SASLMechanism:    "plain",
				SASLUsername:     "user",
				SASLPassword:     "password",
			},
			expectedDSN:    "kafka://sasl_plaintext://user:password@localhost:9093",
			expectedString: "kafka://sasl_plaintext://user:***@localhost:9093",
		},
		{
			name: "带消费者组的配置",
			config: Config{
				Brokers:         []string{"localhost:9093"},
				ConsumerGroupID: "test-group",
			},
			expectedDSN:    "kafka://localhost:9093",
			expectedString: "kafka://localhost:9093 (group=test-group)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedDSN, tt.config.DSN())
			assert.Equal(t, tt.expectedString, tt.config.String())
		})
	}
}

// 测试默认选项
func TestDefaultOptions(t *testing.T) {
	options := DefaultOptions()

	// 检查通用选项是否合理
	assert.Greater(t, options.ConnectTimeout, time.Duration(0))
	assert.Greater(t, options.SessionTimeout, time.Duration(0))
	assert.Greater(t, options.ReconnectBackoff, time.Duration(0))
	assert.GreaterOrEqual(t, options.MaxReconnectRetry, 0)

	// 检查生产者选项
	assert.Contains(t, []int{-1, 0, 1}, options.Acks)
	assert.GreaterOrEqual(t, options.RetryMax, 0)
	assert.Greater(t, options.BatchSize, 0)
	assert.GreaterOrEqual(t, options.LingerMS, time.Duration(0))
	assert.Greater(t, options.MaxMessageBytes, 0)

	// 检查消费者选项
	assert.Contains(t, []string{"earliest", "latest"}, options.AutoOffsetReset)
	assert.GreaterOrEqual(t, options.AutoCommitInterval, time.Duration(0))
	assert.Greater(t, options.MaxPollIntervalMS, time.Duration(0))
	assert.GreaterOrEqual(t, options.FetchMaxWaitMS, time.Duration(0))
	assert.Greater(t, options.MaxPartitionFetchBytes, 0)
}

// 创建消息测试
func TestMessageCreation(t *testing.T) {
	topic := "test-topic"
	key := []byte("test-key")
	value := []byte("test-value")
	headers := map[string]string{
		"header1": "value1",
		"header2": "value2",
	}
	timestamp := time.Now()

	msg := &Message{
		Topic:     topic,
		Key:       key,
		Value:     value,
		Headers:   headers,
		Partition: 0,
		Offset:    123,
		Timestamp: timestamp,
	}

	assert.Equal(t, topic, msg.Topic)
	assert.Equal(t, key, msg.Key)
	assert.Equal(t, value, msg.Value)
	assert.Equal(t, headers, msg.Headers)
	assert.Equal(t, 0, msg.Partition)
	assert.Equal(t, int64(123), msg.Offset)
	assert.Equal(t, timestamp, msg.Timestamp)
}

// 创建生产者测试
func TestNewProducer(t *testing.T) {
	config := createTestConfig()
	options := DefaultOptions()

	producer := NewProducer(config, options)
	require.NotNil(t, producer)

	// 确认实现了Producer接口
	_, ok := producer.(Producer)
	assert.True(t, ok)
}

// 创建消费者测试
func TestNewConsumer(t *testing.T) {
	config := createTestConfig()
	options := DefaultOptions()

	consumer := NewConsumer(config, options)
	require.NotNil(t, consumer)

	// 确认实现了Consumer接口
	_, ok := consumer.(Consumer)
	assert.True(t, ok)
}

// 订阅主题测试
func TestConsumerSubscribe(t *testing.T) {
	config := createTestConfig()
	options := DefaultOptions()

	consumer := NewConsumer(config, options)
	require.NotNil(t, consumer)

	// 订阅单个主题
	err := consumer.Subscribe("test-topic")
	assert.NoError(t, err)

	// 订阅多个主题
	err = consumer.Subscribe("topic1", "topic2", "topic3")
	assert.NoError(t, err)

	// 订阅空主题应该返回错误
	err = consumer.Subscribe()
	assert.Error(t, err)
}

// 模拟测试Produce方法
func TestProducerProduce(t *testing.T) {
	// 注意：这是一个模拟测试，不会真正连接到Kafka
	config := createTestConfig()
	options := DefaultOptions()

	// 创建生产者
	producer := &ProducerImpl{
		config:     config,
		options:    options,
		topicCache: make(map[string]struct{}),
		connected:  false, // 设置为false，因为我们没有真正的连接
	}

	// 由于没有实际连接，调用Producer.Produce会失败，这是预期的
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := producer.Produce(ctx, "test-topic", []byte("key"), []byte("value"), nil)
	assert.Error(t, err) // 应该失败，因为没有实际连接到Kafka
}
