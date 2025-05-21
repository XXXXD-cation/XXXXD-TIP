// Package db 提供了数据库客户端的封装，支持PostgreSQL、Redis和Elasticsearch
// 简化了连接管理、错误处理和日志记录
package db

import (
	"context"
	"time"
)

// ConnectionInfo 定义连接信息的通用接口
type ConnectionInfo interface {
	// DSN 返回数据源名称或连接字符串
	DSN() string
	// String 返回连接信息的字符串表示（去除敏感信息）
	String() string
}

// Options 定义数据库连接选项
type Options struct {
	// MaxOpenConns 最大打开连接数
	MaxOpenConns int
	// MaxIdleConns 最大空闲连接数
	MaxIdleConns int
	// ConnMaxLifetime 连接最大生存时间
	ConnMaxLifetime time.Duration
	// ConnMaxIdleTime 连接最大空闲时间
	ConnMaxIdleTime time.Duration
	// Timeout 连接超时时间
	Timeout time.Duration
	// HealthCheckInterval 健康检查间隔
	HealthCheckInterval time.Duration
	// RetryAttempts 重试次数
	RetryAttempts int
	// RetryDelay 重试延迟
	RetryDelay time.Duration
}

// DefaultOptions 返回默认选项
func DefaultOptions() Options {
	return Options{
		MaxOpenConns:        100,
		MaxIdleConns:        10,
		ConnMaxLifetime:     time.Hour,
		ConnMaxIdleTime:     time.Minute * 30,
		Timeout:             time.Second * 10,
		HealthCheckInterval: time.Minute,
		RetryAttempts:       3,
		RetryDelay:          time.Second * 2,
	}
}

// Client 定义数据库客户端的通用接口
type Client interface {
	// Connect 连接到数据库
	Connect(ctx context.Context) error
	// Close 关闭连接
	Close() error
	// Ping 检查连接是否可用
	Ping(ctx context.Context) error
	// Stats 返回连接统计信息
	Stats() interface{}
}
