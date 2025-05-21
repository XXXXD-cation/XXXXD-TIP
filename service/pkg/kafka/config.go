// Package kafka 提供了Kafka客户端的封装
package kafka

import (
	"fmt"
	"strings"

	"github.com/XXXXD-cation/XXXXD-TIP/service/pkg/db"
)

// Config Kafka连接配置
type Config struct {
	Brokers           []string          // Kafka代理服务器地址列表
	SecurityProtocol  string            // 安全协议：plaintext, ssl, sasl_plaintext, sasl_ssl
	SASLMechanism     string            // SASL机制：plain, scram-sha-256, scram-sha-512
	SASLUsername      string            // SASL用户名
	SASLPassword      string            // SASL密码
	ClientID          string            // 客户端ID
	ConsumerGroupID   string            // 消费者组ID
	ConsumerTopics    []string          // 消费者订阅的主题
	ProducerTopics    []string          // 生产者可能会使用的主题
	Properties        map[string]string // 其他配置属性
	ConnectionTimeout int               // 连接超时（毫秒）
	SessionTimeout    int               // 会话超时（毫秒）
}

// DSN 返回Kafka连接字符串
func (c Config) DSN() string {
	brokers := strings.Join(c.Brokers, ",")
	auth := ""

	if c.SecurityProtocol != "" && c.SecurityProtocol != "plaintext" {
		auth = fmt.Sprintf("%s://", c.SecurityProtocol)
		if c.SASLMechanism != "" && c.SASLUsername != "" {
			auth += fmt.Sprintf("%s:%s@", c.SASLUsername, c.SASLPassword)
		}
	}

	return fmt.Sprintf("kafka://%s%s", auth, brokers)
}

// String 返回连接信息的字符串表示（去除敏感信息）
func (c Config) String() string {
	brokers := strings.Join(c.Brokers, ",")
	auth := ""

	if c.SecurityProtocol != "" && c.SecurityProtocol != "plaintext" {
		auth = fmt.Sprintf("%s://", c.SecurityProtocol)
		if c.SASLMechanism != "" && c.SASLUsername != "" {
			auth += fmt.Sprintf("%s:***@", c.SASLUsername)
		}
	}

	groupInfo := ""
	if c.ConsumerGroupID != "" {
		groupInfo = fmt.Sprintf(" (group=%s)", c.ConsumerGroupID)
	}

	return fmt.Sprintf("kafka://%s%s%s", auth, brokers, groupInfo)
}

// ensure Config implements ConnectionInfo interface
var _ db.ConnectionInfo = (*Config)(nil)
