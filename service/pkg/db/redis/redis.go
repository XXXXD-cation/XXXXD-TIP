// Package redis 提供了Redis客户端的封装，基于go-redis/redis实现
package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/XXXXD-cation/XXXXD-TIP/service/pkg/db"
	"github.com/XXXXD-cation/XXXXD-TIP/service/pkg/log"
	"github.com/cenkalti/backoff/v4"
	"github.com/redis/go-redis/v9"
)

// Config Redis连接配置
type Config struct {
	// 单节点配置
	Addr     string
	Password string
	DB       int

	// 集群配置
	Addrs []string

	// 哨兵配置
	MasterName       string
	SentinelAddrs    []string
	SentinelPassword string

	// 通用配置
	Username     string
	MaxRetries   int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	// 是否使用集群模式
	UseCluster bool
	// 是否使用哨兵模式
	UseSentinel bool
}

// DSN 返回Redis连接字符串
func (c Config) DSN() string {
	if c.UseCluster {
		addrs := "["
		for i, addr := range c.Addrs {
			if i > 0 {
				addrs += ", "
			}
			addrs += addr
		}
		addrs += "]"
		return fmt.Sprintf("redis-cluster://%s", addrs)
	}

	if c.UseSentinel {
		addrs := "["
		for i, addr := range c.SentinelAddrs {
			if i > 0 {
				addrs += ", "
			}
			addrs += addr
		}
		addrs += "]"
		return fmt.Sprintf("redis-sentinel://%s/%s", addrs, c.MasterName)
	}

	auth := ""
	if c.Username != "" {
		auth = c.Username
		if c.Password != "" {
			auth += ":" + c.Password
		}
		auth += "@"
	} else if c.Password != "" {
		auth = ":" + c.Password + "@"
	}

	return fmt.Sprintf("redis://%s%s/%d", auth, c.Addr, c.DB)
}

// String 返回连接信息的字符串表示（去除敏感信息）
func (c Config) String() string {
	if c.UseCluster {
		addrs := "["
		for i, addr := range c.Addrs {
			if i > 0 {
				addrs += ", "
			}
			addrs += addr
		}
		addrs += "]"
		return fmt.Sprintf("redis-cluster://%s", addrs)
	}

	if c.UseSentinel {
		addrs := "["
		for i, addr := range c.SentinelAddrs {
			if i > 0 {
				addrs += ", "
			}
			addrs += addr
		}
		addrs += "]"
		return fmt.Sprintf("redis-sentinel://%s/%s", addrs, c.MasterName)
	}

	auth := ""
	if c.Username != "" {
		auth = c.Username + ":***@"
	} else if c.Password != "" {
		auth = ":***@"
	}

	return fmt.Sprintf("redis://%s%s/%d", auth, c.Addr, c.DB)
}

// ensure Config implements ConnectionInfo interface
var _ db.ConnectionInfo = (*Config)(nil)

// Client 表示Redis客户端
type Client struct {
	config    Config
	opts      db.Options
	client    redis.UniversalClient
	connected bool
}

// ensure Client implements db.Client interface
var _ db.Client = (*Client)(nil)

// New 创建一个新的Redis客户端
func New(config Config, opts db.Options) *Client {
	return &Client{
		config: config,
		opts:   opts,
	}
}

// Connect 连接到Redis服务器
func (c *Client) Connect(ctx context.Context) error {
	// 创建带有超时的上下文
	ctx, cancel := context.WithTimeout(ctx, c.opts.Timeout)
	defer cancel()

	// 配置Redis客户端选项
	options := &redis.UniversalOptions{
		Addrs:      []string{c.config.Addr},
		Username:   c.config.Username,
		Password:   c.config.Password,
		DB:         c.config.DB,
		MaxRetries: c.config.MaxRetries,
	}

	// 设置超时
	if c.config.DialTimeout > 0 {
		options.DialTimeout = c.config.DialTimeout
	} else {
		options.DialTimeout = c.opts.Timeout
	}

	if c.config.ReadTimeout > 0 {
		options.ReadTimeout = c.config.ReadTimeout
	}

	if c.config.WriteTimeout > 0 {
		options.WriteTimeout = c.config.WriteTimeout
	}

	// 根据配置选择客户端类型
	if c.config.UseCluster {
		options.Addrs = c.config.Addrs
	} else if c.config.UseSentinel {
		options.MasterName = c.config.MasterName
		options.Addrs = c.config.SentinelAddrs
		options.SentinelPassword = c.config.SentinelPassword
	}

	// 创建通用客户端
	client := redis.NewUniversalClient(options)

	// 使用指数退避进行重试连接
	operation := func() error {
		select {
		case <-ctx.Done():
			return backoff.Permanent(ctx.Err())
		default:
			if err := client.Ping(ctx).Err(); err != nil {
				log.Error().Err(err).Str("dsn", c.config.String()).Msg("连接Redis失败")
				return err
			}
			return nil
		}
	}

	// 创建退避策略
	exponentialBackOff := backoff.NewExponentialBackOff()
	exponentialBackOff.MaxElapsedTime = c.opts.Timeout
	exponentialBackOff.InitialInterval = c.opts.RetryDelay

	// 执行带有重试的操作
	var retryBackOff backoff.BackOff = exponentialBackOff
	if c.opts.RetryAttempts > 0 {
		// 安全转换，避免大整数溢出
		retryBackOff = backoff.WithMaxRetries(exponentialBackOff, uint64(c.opts.RetryAttempts))
	}
	err := backoff.Retry(operation, retryBackOff)
	if err != nil {
		return fmt.Errorf("连接Redis失败: %w", err)
	}

	c.client = client
	c.connected = true

	log.Info().Str("dsn", c.config.String()).Msg("已连接到Redis")
	return nil
}

// Close 关闭Redis连接
func (c *Client) Close() error {
	if c.client == nil {
		return nil
	}

	if err := c.client.Close(); err != nil {
		return fmt.Errorf("关闭Redis连接失败: %w", err)
	}

	c.connected = false
	log.Info().Str("dsn", c.config.String()).Msg("已关闭Redis连接")
	return nil
}

// Ping 检查Redis连接是否有效
func (c *Client) Ping(ctx context.Context) error {
	if c.client == nil {
		return fmt.Errorf("Redis未连接")
	}

	ctx, cancel := context.WithTimeout(ctx, c.opts.Timeout)
	defer cancel()

	if err := c.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("ping Redis失败: %w", err)
	}

	return nil
}

// Stats 返回连接统计信息
func (c *Client) Stats() interface{} {
	if c.client == nil {
		return nil
	}

	stats := c.client.PoolStats()
	return stats
}

// Client 返回底层的Redis客户端
func (c *Client) Client() redis.UniversalClient {
	return c.client
}

// Set 设置键值对
func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if c.client == nil {
		return fmt.Errorf("Redis未连接")
	}

	return c.client.Set(ctx, key, value, expiration).Err()
}

// Get 获取键值
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	if c.client == nil {
		return "", fmt.Errorf("Redis未连接")
	}

	return c.client.Get(ctx, key).Result()
}

// Delete 删除键
func (c *Client) Delete(ctx context.Context, keys ...string) error {
	if c.client == nil {
		return fmt.Errorf("Redis未连接")
	}

	return c.client.Del(ctx, keys...).Err()
}

// Exists 检查键是否存在
func (c *Client) Exists(ctx context.Context, keys ...string) (int64, error) {
	if c.client == nil {
		return 0, fmt.Errorf("Redis未连接")
	}

	return c.client.Exists(ctx, keys...).Result()
}

// Expire 设置键过期时间
func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) error {
	if c.client == nil {
		return fmt.Errorf("Redis未连接")
	}

	return c.client.Expire(ctx, key, expiration).Err()
}
