# Kafka客户端封装

本包提供了对Kafka客户端的统一封装，包含生产者和消费者的实现。基于`segmentio/kafka-go`库实现，并提供了以下功能：

## 主要特性

- 统一的配置接口
- 生产者和消费者的高级抽象
- 错误重试和恢复
- 连接池管理
- 安全认证支持（SASL/SSL）
- 结构化日志集成
- 健康检查和监控

## 使用方法

### 创建生产者

```go
// 创建配置
config := kafka.Config{
    Brokers:  []string{"localhost:9093"},
    ClientID: "my-producer",
}

// 使用默认选项
options := kafka.DefaultOptions()

// 创建生产者
producer := kafka.NewProducer(config, options)
defer producer.Close()

// 发送消息
err := producer.Produce(context.Background(), "my-topic", []byte("key"), []byte("value"), nil)
if err != nil {
    log.Fatalf("发送消息失败: %v", err)
}
```

### 创建消费者

```go
// 创建配置
config := kafka.Config{
    Brokers:         []string{"localhost:9093"},
    ClientID:        "my-consumer",
    ConsumerGroupID: "my-group",
}

// 使用默认选项
options := kafka.DefaultOptions()

// 创建消费者
consumer := kafka.NewConsumer(config, options)
defer consumer.Close()

// 订阅主题
if err := consumer.Subscribe("my-topic"); err != nil {
    log.Fatalf("订阅主题失败: %v", err)
}

// 开始消费
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

err := consumer.Consume(ctx, func(msg *kafka.Message) error {
    fmt.Printf("收到消息: topic=%s, key=%s, value=%s\n", 
        msg.Topic, string(msg.Key), string(msg.Value))
    return nil
})

if err != nil && err != context.Canceled {
    log.Fatalf("消费消息失败: %v", err)
}
```

## 高级配置

### 安全认证

```go
// 使用SASL/PLAIN认证
config := kafka.Config{
    Brokers:          []string{"localhost:9093"},
    SecurityProtocol: "sasl_plaintext", // 或 "sasl_ssl" 启用TLS
    SASLMechanism:    "plain",
    SASLUsername:     "username",
    SASLPassword:     "password",
}
```

### 消费者选项

```go
options := kafka.DefaultOptions()

// 修改默认选项
options.AutoOffsetReset = "earliest" // 从最早的消息开始消费
options.EnableAutoCommit = false     // 禁用自动提交
```

### 生产者选项

```go
options := kafka.DefaultOptions()

// 修改默认选项
options.CompressionType = "gzip"     // 使用gzip压缩
options.Acks = 1                     // 只等待leader确认
options.BatchSize = 32768            // 增加批处理大小
```

### 暂停和恢复消费

```go
// 暂停特定主题的消费
consumer.Pause("my-topic")

// 恢复特定主题的消费
consumer.Resume("my-topic")
```

## 完整示例

参见 `service/examples/kafka_example/main.go` 获取完整的示例代码。 