// Package postgres 提供了PostgreSQL客户端的封装，基于GORM实现
package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/XXXXD-cation/XXXXD-TIP/service/pkg/db"
	"github.com/XXXXD-cation/XXXXD-TIP/service/pkg/log"
	"github.com/cenkalti/backoff/v4"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Config PostgreSQL连接配置
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
	SSLMode  string
	TimeZone string
	Schema   string
	Options  map[string]string
}

// DSN 返回PostgreSQL连接字符串
func (c Config) DSN() string {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s",
		c.Host, c.Port, c.Username, c.Password, c.Database)

	if c.SSLMode != "" {
		dsn += fmt.Sprintf(" sslmode=%s", c.SSLMode)
	}

	if c.TimeZone != "" {
		dsn += fmt.Sprintf(" TimeZone=%s", c.TimeZone)
	}

	if c.Schema != "" {
		dsn += fmt.Sprintf(" search_path=%s", c.Schema)
	}

	for k, v := range c.Options {
		dsn += fmt.Sprintf(" %s=%s", k, v)
	}

	return dsn
}

// String 返回连接信息的字符串表示（去除敏感信息）
func (c Config) String() string {
	return fmt.Sprintf("postgres://%s:***@%s:%d/%s", c.Username, c.Host, c.Port, c.Database)
}

// ensure Config implements ConnectionInfo interface
var _ db.ConnectionInfo = (*Config)(nil)

// Client 表示PostgreSQL客户端
type Client struct {
	config Config
	opts   db.Options
	db     *gorm.DB
	sqlDB  gorm.ConnPool
}

// 确保Client实现db.Client接口
var _ db.Client = (*Client)(nil)

// Stats 表示数据库连接统计信息
type Stats struct {
	OpenConnections   int
	InUse             int
	Idle              int
	WaitCount         int64
	WaitDuration      time.Duration
	MaxIdleClosed     int64
	MaxLifetimeClosed int64
}

// New 创建一个新的PostgreSQL客户端
func New(config Config, opts db.Options) *Client {
	return &Client{
		config: config,
		opts:   opts,
	}
}

// DB 返回底层的GORM数据库实例
func (c *Client) DB() *gorm.DB {
	return c.db
}

// Connect 连接到PostgreSQL数据库
func (c *Client) Connect(ctx context.Context) error {
	// 创建带有超时的上下文
	ctx, cancel := context.WithTimeout(ctx, c.opts.Timeout)
	defer cancel()

	// 设置GORM日志配置
	gormLogger := logger.New(
		&logAdapter{},
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	// 配置GORM
	gormConfig := &gorm.Config{
		Logger: gormLogger,
	}

	// 使用指数退避进行重试连接
	var db *gorm.DB
	var err error

	operation := func() error {
		select {
		case <-ctx.Done():
			return backoff.Permanent(ctx.Err())
		default:
			tmpDb, tmpErr := gorm.Open(postgres.Open(c.config.DSN()), gormConfig)
			if tmpErr != nil {
				log.Error().Err(tmpErr).Str("dsn", c.config.String()).Msg("连接PostgreSQL失败")
				return tmpErr
			}
			db = tmpDb
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
	err = backoff.Retry(operation, retryBackOff)
	if err != nil {
		return fmt.Errorf("连接PostgreSQL失败: %w", err)
	}

	// 获取底层SQL DB连接池
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取SQL DB失败: %w", err)
	}

	// 配置连接池
	sqlDB.SetMaxOpenConns(c.opts.MaxOpenConns)
	sqlDB.SetMaxIdleConns(c.opts.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(c.opts.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(c.opts.ConnMaxIdleTime)

	c.db = db
	c.sqlDB = sqlDB

	log.Info().Str("dsn", c.config.String()).Msg("已连接到PostgreSQL")
	return nil
}

// Close 关闭数据库连接
func (c *Client) Close() error {
	if c.db == nil {
		return nil
	}

	sqlDB, err := c.db.DB()
	if err != nil {
		return fmt.Errorf("获取SQL DB失败: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("关闭PostgreSQL连接失败: %w", err)
	}

	log.Info().Str("dsn", c.config.String()).Msg("已关闭PostgreSQL连接")
	return nil
}

// Ping 检查数据库连接是否有效
func (c *Client) Ping(ctx context.Context) error {
	if c.db == nil {
		return fmt.Errorf("数据库未连接")
	}

	sqlDB, err := c.db.DB()
	if err != nil {
		return fmt.Errorf("获取SQL DB失败: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, c.opts.Timeout)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("ping PostgreSQL失败: %w", err)
	}

	return nil
}

// Stats 返回连接统计信息
func (c *Client) Stats() interface{} {
	if c.db == nil {
		return nil
	}

	sqlDB, err := c.db.DB()
	if err != nil {
		log.Error().Err(err).Msg("获取SQL DB统计信息失败")
		return nil
	}

	stats := sqlDB.Stats()
	return Stats{
		OpenConnections:   stats.OpenConnections,
		InUse:             stats.InUse,
		Idle:              stats.Idle,
		WaitCount:         stats.WaitCount,
		WaitDuration:      stats.WaitDuration,
		MaxIdleClosed:     stats.MaxIdleClosed,
		MaxLifetimeClosed: stats.MaxLifetimeClosed,
	}
}

// AutoMigrate 自动迁移模型到数据库
func (c *Client) AutoMigrate(models ...interface{}) error {
	if c.db == nil {
		return fmt.Errorf("数据库未连接")
	}

	log.Info().Msg("开始自动迁移模型")
	return c.db.AutoMigrate(models...)
}

// Transaction 在事务中执行函数
func (c *Client) Transaction(fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) error {
	if c.db == nil {
		return fmt.Errorf("数据库未连接")
	}

	return c.db.Transaction(fc, opts...)
}

// logAdapter GORM日志适配器
type logAdapter struct{}

// Printf 实现GORM日志接口
func (l *logAdapter) Printf(format string, args ...interface{}) {
	log.Debug().Msgf(format, args...)
}
