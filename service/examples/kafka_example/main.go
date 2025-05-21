package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/XXXXD-cation/XXXXD-TIP/service/pkg/kafka"
)

// IntelligenceItem 威胁情报数据示例
type IntelligenceItem struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Value       string    `json:"value"`
	Source      string    `json:"source"`
	Confidence  float64   `json:"confidence"`
	ValidFrom   time.Time `json:"valid_from"`
	ValidUntil  time.Time `json:"valid_until"`
	Description string    `json:"description"`
	Tags        []string  `json:"tags"`
}

func main() {
	// 创建Kafka配置
	config := kafka.Config{
		Brokers:         []string{"localhost:9093"},
		ClientID:        "tip-example-client",
		ConsumerGroupID: "tip-example-group",
	}

	// 获取默认选项
	options := kafka.DefaultOptions()

	// 创建上下文，用于控制生产者和消费者的生命周期
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 捕获退出信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// 创建生产者
	producer := kafka.NewProducer(config, options)
	defer producer.Close()

	// 创建消费者
	consumer := kafka.NewConsumer(config, options)
	defer consumer.Close()

	// 订阅主题
	if err := consumer.Subscribe("raw-intelligence"); err != nil {
		log.Fatalf("订阅主题失败: %v", err)
	}

	// 异步启动消费者
	go func() {
		err := consumer.Consume(ctx, func(msg *kafka.Message) error {
			// 解析消息内容
			var item IntelligenceItem
			if err := json.Unmarshal(msg.Value, &item); err != nil {
				log.Printf("解析消息失败: %v", err)
				return err
			}

			// 处理消息
			log.Printf("收到威胁情报数据: ID=%s, Type=%s, Value=%s, Confidence=%.2f\n",
				item.ID, item.Type, item.Value, item.Confidence)

			// 发送处理后的消息到另一个主题
			processedItem := item
			processedItem.Description += " [已处理]"

			processedData, err := json.Marshal(processedItem)
			if err != nil {
				log.Printf("序列化处理后的数据失败: %v", err)
				return err
			}

			// 同步处理消息时向下一个主题发送
			err = producer.Produce(ctx, "processed-intelligence", msg.Key, processedData, map[string]string{
				"original_topic":  msg.Topic,
				"processing_time": time.Now().Format(time.RFC3339),
				"processor_id":    "example-processor",
			})

			if err != nil {
				log.Printf("发送处理后的数据失败: %v", err)
				return err
			}

			return nil
		})

		if err != nil && err != context.Canceled {
			log.Printf("消费消息时发生错误: %v", err)
		}
	}()

	// 发送一些示例数据
	go func() {
		for i := 1; i <= 5; i++ {
			// 创建示例威胁情报数据
			item := IntelligenceItem{
				ID:          fmt.Sprintf("ioc-%d", i),
				Type:        "domain",
				Value:       fmt.Sprintf("malicious%d.example.com", i),
				Source:      "example-feed",
				Confidence:  0.8 + float64(i)*0.02,
				ValidFrom:   time.Now(),
				ValidUntil:  time.Now().AddDate(0, 1, 0),
				Description: fmt.Sprintf("Example malicious domain #%d", i),
				Tags:        []string{"malware", "phishing", "example"},
			}

			// 序列化为JSON
			data, err := json.Marshal(item)
			if err != nil {
				log.Printf("序列化数据失败: %v", err)
				continue
			}

			// 发送消息
			err = producer.Produce(ctx, "raw-intelligence", []byte(item.ID), data, map[string]string{
				"content_type": "application/json",
				"source":       "example-producer",
				"timestamp":    time.Now().Format(time.RFC3339),
			})

			if err != nil {
				log.Printf("发送消息失败: %v", err)
			} else {
				log.Printf("已发送威胁情报数据: ID=%s, Type=%s, Value=%s\n", item.ID, item.Type, item.Value)
			}

			time.Sleep(2 * time.Second)
		}
	}()

	// 等待退出信号
	<-sigCh
	log.Println("收到退出信号，正在关闭应用...")

	// 取消上下文，停止消费
	cancel()

	// 等待资源清理
	time.Sleep(500 * time.Millisecond)

	log.Println("应用已关闭")
}
