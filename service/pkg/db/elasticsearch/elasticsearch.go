// Package elasticsearch 提供了Elasticsearch客户端的封装，基于go-elasticsearch实现
package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/XXXXD-cation/XXXXD-TIP/service/pkg/db"
	"github.com/XXXXD-cation/XXXXD-TIP/service/pkg/log"
	"github.com/cenkalti/backoff/v4"
	"github.com/elastic/go-elasticsearch/v8"
)

// Config Elasticsearch连接配置
type Config struct {
	Addresses  []string          // Elasticsearch服务器地址
	Username   string            // 认证用户名
	Password   string            // 认证密码
	CloudID    string            // Elasticsearch Cloud ID
	APIKey     string            // API Key认证
	CACert     string            // CA证书
	Headers    map[string]string // 自定义请求头
	MaxRetries int               // 最大重试次数
}

// DSN 返回Elasticsearch连接字符串
func (c Config) DSN() string {
	if c.CloudID != "" {
		auth := ""
		if c.Username != "" && c.Password != "" {
			auth = fmt.Sprintf("%s:%s@", c.Username, c.Password)
		}
		return fmt.Sprintf("es+cloud://%s%s", auth, c.CloudID)
	}

	if len(c.Addresses) == 0 {
		return "es://"
	}

	auth := ""
	if c.Username != "" && c.Password != "" {
		auth = fmt.Sprintf("%s:%s@", c.Username, c.Password)
	}

	addrs := "["
	for i, addr := range c.Addresses {
		if i > 0 {
			addrs += ", "
		}
		addrs += addr
	}
	addrs += "]"
	return fmt.Sprintf("es://%s%s", auth, addrs)
}

// String 返回连接信息的字符串表示（去除敏感信息）
func (c Config) String() string {
	if c.CloudID != "" {
		auth := ""
		if c.Username != "" {
			auth = fmt.Sprintf("%s:***@", c.Username)
		}
		return fmt.Sprintf("es+cloud://%s%s", auth, c.CloudID)
	}

	if len(c.Addresses) == 0 {
		return "es://"
	}

	auth := ""
	if c.Username != "" {
		auth = fmt.Sprintf("%s:***@", c.Username)
	}

	addrs := "["
	for i, addr := range c.Addresses {
		if i > 0 {
			addrs += ", "
		}
		addrs += addr
	}
	addrs += "]"
	return fmt.Sprintf("es://%s%s", auth, addrs)
}

// ensure Config implements ConnectionInfo interface
var _ db.ConnectionInfo = (*Config)(nil)

// ClientStats 表示Elasticsearch客户端统计信息
type ClientStats struct {
	Connections int `json:"connections"`
	Idle        int `json:"idle"`
	InUse       int `json:"in_use"`
}

// Client 表示Elasticsearch客户端
type Client struct {
	config    Config
	opts      db.Options
	client    *elasticsearch.Client
	connected bool
}

// ensure Client implements db.Client interface
var _ db.Client = (*Client)(nil)

// New 创建一个新的Elasticsearch客户端
func New(config Config, opts db.Options) *Client {
	return &Client{
		config: config,
		opts:   opts,
	}
}

// Connect 连接到Elasticsearch
func (c *Client) Connect(ctx context.Context) error {
	// 创建带有超时的上下文
	ctx, cancel := context.WithTimeout(ctx, c.opts.Timeout)
	defer cancel()

	// 配置Elasticsearch客户端
	esConfig := elasticsearch.Config{
		Addresses: c.config.Addresses,
		Username:  c.config.Username,
		Password:  c.config.Password,
		CloudID:   c.config.CloudID,
		APIKey:    c.config.APIKey,
		Header:    make(http.Header),
	}

	// 设置自定义请求头
	for k, v := range c.config.Headers {
		esConfig.Header.Add(k, v)
	}

	// 设置重试次数
	if c.config.MaxRetries > 0 {
		esConfig.RetryOnStatus = []int{502, 503, 504, 429}
		esConfig.MaxRetries = c.config.MaxRetries
	} else if c.opts.RetryAttempts > 0 {
		esConfig.RetryOnStatus = []int{502, 503, 504, 429}
		esConfig.MaxRetries = c.opts.RetryAttempts
	}

	// 设置超时
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.ResponseHeaderTimeout = c.opts.Timeout
	esConfig.Transport = transport

	// 创建客户端
	client, err := elasticsearch.NewClient(esConfig)
	if err != nil {
		return fmt.Errorf("创建Elasticsearch客户端失败: %w", err)
	}

	// 使用指数退避进行重试连接
	operation := func() error {
		select {
		case <-ctx.Done():
			return backoff.Permanent(ctx.Err())
		default:
			infoRes, infoErr := client.Info()
			if infoErr != nil {
				log.Error().Err(infoErr).Str("dsn", c.config.String()).Msg("连接Elasticsearch失败")
				return infoErr
			}
			defer infoRes.Body.Close()

			if infoRes.IsError() {
				resErr := fmt.Errorf("Elasticsearch错误: %s", infoRes.String())
				log.Error().Err(resErr).Str("dsn", c.config.String()).Msg("连接Elasticsearch失败")
				return resErr
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
	err = backoff.Retry(operation, retryBackOff)
	if err != nil {
		return fmt.Errorf("连接Elasticsearch失败: %w", err)
	}

	c.client = client
	c.connected = true

	log.Info().Str("dsn", c.config.String()).Msg("已连接到Elasticsearch")
	return nil
}

// Close 关闭Elasticsearch连接
func (c *Client) Close() error {
	// Elasticsearch客户端没有显式关闭方法
	if c.client != nil {
		c.connected = false
		log.Info().Str("dsn", c.config.String()).Msg("已关闭Elasticsearch连接")
	}
	return nil
}

// Ping 检查Elasticsearch连接是否有效
func (c *Client) Ping(ctx context.Context) error {
	if c.client == nil {
		return fmt.Errorf("Elasticsearch未连接")
	}

	ctx, cancel := context.WithTimeout(ctx, c.opts.Timeout)
	defer cancel()

	res, err := c.client.Ping(c.client.Ping.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("ping Elasticsearch失败: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("Elasticsearch ping错误: %s", res.String())
	}

	return nil
}

// Stats 返回连接统计信息
func (c *Client) Stats() interface{} {
	return nil // Elasticsearch客户端不提供连接池统计
}

// Client 返回底层的Elasticsearch客户端
func (c *Client) Client() *elasticsearch.Client {
	return c.client
}

// IndexExists 检查索引是否存在
func (c *Client) IndexExists(ctx context.Context, index string) (bool, error) {
	if c.client == nil {
		return false, fmt.Errorf("Elasticsearch未连接")
	}

	res, err := c.client.Indices.Exists([]string{index}, c.client.Indices.Exists.WithContext(ctx))
	if err != nil {
		return false, fmt.Errorf("检查索引失败: %w", err)
	}
	defer res.Body.Close()

	return res.StatusCode == 200, nil
}

// CreateIndex 创建索引
func (c *Client) CreateIndex(ctx context.Context, index string, body map[string]interface{}) error {
	if c.client == nil {
		return fmt.Errorf("Elasticsearch未连接")
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("序列化索引配置失败: %w", err)
	}

	res, err := c.client.Indices.Create(
		index,
		c.client.Indices.Create.WithContext(ctx),
		c.client.Indices.Create.WithBody(strings.NewReader(string(jsonBody))),
	)
	if err != nil {
		return fmt.Errorf("创建索引失败: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("创建索引错误: %s", res.String())
	}

	return nil
}

// DeleteIndex 删除索引
func (c *Client) DeleteIndex(ctx context.Context, index string) error {
	if c.client == nil {
		return fmt.Errorf("Elasticsearch未连接")
	}

	res, err := c.client.Indices.Delete(
		[]string{index},
		c.client.Indices.Delete.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("删除索引失败: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("删除索引错误: %s", res.String())
	}

	return nil
}

// Index 索引文档
func (c *Client) Index(ctx context.Context, index string, id string, document interface{}) error {
	if c.client == nil {
		return fmt.Errorf("Elasticsearch未连接")
	}

	jsonBody, err := json.Marshal(document)
	if err != nil {
		return fmt.Errorf("序列化文档失败: %w", err)
	}

	res, err := c.client.Index(
		index,
		strings.NewReader(string(jsonBody)),
		c.client.Index.WithContext(ctx),
		c.client.Index.WithDocumentID(id),
	)
	if err != nil {
		return fmt.Errorf("索引文档失败: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("索引文档错误: %s", res.String())
	}

	return nil
}

// Get 获取文档
func (c *Client) Get(ctx context.Context, index string, id string) (map[string]interface{}, error) {
	if c.client == nil {
		return nil, fmt.Errorf("Elasticsearch未连接")
	}

	res, err := c.client.Get(
		index,
		id,
		c.client.Get.WithContext(ctx),
	)
	if err != nil {
		return nil, fmt.Errorf("获取文档失败: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			return nil, fmt.Errorf("文档不存在")
		}
		return nil, fmt.Errorf("获取文档错误: %s", res.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析文档失败: %w", err)
	}

	return result, nil
}

// Delete 删除文档
func (c *Client) Delete(ctx context.Context, index string, id string) error {
	if c.client == nil {
		return fmt.Errorf("Elasticsearch未连接")
	}

	res, err := c.client.Delete(
		index,
		id,
		c.client.Delete.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("删除文档失败: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("删除文档错误: %s", res.String())
	}

	return nil
}

// Search 搜索文档
func (c *Client) Search(ctx context.Context, index string, query map[string]interface{}) (map[string]interface{}, error) {
	if c.client == nil {
		return nil, fmt.Errorf("Elasticsearch未连接")
	}

	jsonBody, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("序列化查询失败: %w", err)
	}

	res, err := c.client.Search(
		c.client.Search.WithContext(ctx),
		c.client.Search.WithIndex(index),
		c.client.Search.WithBody(strings.NewReader(string(jsonBody))),
	)
	if err != nil {
		return nil, fmt.Errorf("搜索文档失败: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("搜索文档错误: %s", res.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析搜索结果失败: %w", err)
	}

	return result, nil
}
