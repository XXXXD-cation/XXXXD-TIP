// Package httpclient 提供了HTTP客户端的封装，支持超时控制、重试、跟踪和日志记录
// 用于服务间调用和第三方API交互
package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/XXXXD-cation/XXXXD-TIP/service/pkg/errors"
	"github.com/XXXXD-cation/XXXXD-TIP/service/pkg/log"
)

// Client 是HTTP客户端的封装
type Client struct {
	// 底层HTTP客户端
	httpClient *http.Client
	// 默认请求选项
	options *Options
}

// Options 配置HTTP客户端的选项
type Options struct {
	// 基础URL，如果设置了，将与请求路径拼接
	BaseURL string
	// 默认请求头
	Headers map[string]string
	// 连接超时时间
	ConnectionTimeout time.Duration
	// 请求超时时间
	RequestTimeout time.Duration
	// 是否自动重试
	Retry bool
	// 最大重试次数
	MaxRetries int
	// 重试间隔基础时间
	RetryInterval time.Duration
	// 是否启用日志记录
	EnableLogging bool
}

// DefaultOptions 返回默认选项
func DefaultOptions() *Options {
	return &Options{
		Headers: map[string]string{
			"Content-Type": "application/json",
			"Accept":       "application/json",
		},
		ConnectionTimeout: 10 * time.Second,
		RequestTimeout:    30 * time.Second,
		Retry:             true,
		MaxRetries:        3,
		RetryInterval:     500 * time.Millisecond,
		EnableLogging:     true,
	}
}

// New 创建一个新的HTTP客户端
func New(options *Options) *Client {
	if options == nil {
		options = DefaultOptions()
	}

	// 创建Transport
	transport := &http.Transport{
		DisableKeepAlives:     false,
		DisableCompression:    false,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		MaxConnsPerHost:       100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   options.ConnectionTimeout,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: options.ConnectionTimeout,
	}

	// 创建HTTP客户端
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   options.RequestTimeout,
	}

	return &Client{
		httpClient: httpClient,
		options:    options,
	}
}

// RequestOptions 定义单个请求的选项
type RequestOptions struct {
	// 请求头
	Headers map[string]string
	// 查询参数
	QueryParams map[string]string
	// 是否重试
	Retry bool
	// 最大重试次数（仅当Retry为true时有效）
	MaxRetries int
}

// Response 表示HTTP响应
type Response struct {
	// 状态码
	StatusCode int
	// 响应体
	Body []byte
	// 响应头
	Headers http.Header
}

// Get 发送GET请求
func (c *Client) Get(ctx context.Context, url string, options *RequestOptions) (*Response, error) {
	return c.Request(ctx, http.MethodGet, url, nil, options)
}

// Post 发送POST请求
func (c *Client) Post(ctx context.Context, url string, body interface{}, options *RequestOptions) (*Response, error) {
	return c.Request(ctx, http.MethodPost, url, body, options)
}

// Put 发送PUT请求
func (c *Client) Put(ctx context.Context, url string, body interface{}, options *RequestOptions) (*Response, error) {
	return c.Request(ctx, http.MethodPut, url, body, options)
}

// Delete 发送DELETE请求
func (c *Client) Delete(ctx context.Context, url string, options *RequestOptions) (*Response, error) {
	return c.Request(ctx, http.MethodDelete, url, nil, options)
}

// Patch 发送PATCH请求
func (c *Client) Patch(ctx context.Context, url string, body interface{}, options *RequestOptions) (*Response, error) {
	return c.Request(ctx, http.MethodPatch, url, body, options)
}

// Request 发送HTTP请求
func (c *Client) Request(ctx context.Context, method, url string, body interface{}, options *RequestOptions) (*Response, error) {
	// 如果BaseURL非空，且url不是以http开头，则拼接BaseURL
	if c.options.BaseURL != "" && !isAbsoluteURL(url) {
		url = fmt.Sprintf("%s/%s", trimTrailingSlash(c.options.BaseURL), trimLeadingSlash(url))
	}

	// 准备请求体
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, errors.BadRequestWithError(err, "无法序列化请求体")
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, errors.BadRequestWithError(err, "创建HTTP请求失败")
	}

	// 添加默认请求头
	for k, v := range c.options.Headers {
		req.Header.Set(k, v)
	}

	// 添加自定义请求头
	if options != nil && options.Headers != nil {
		for k, v := range options.Headers {
			req.Header.Set(k, v)
		}
	}

	// 添加查询参数
	if options != nil && options.QueryParams != nil {
		q := req.URL.Query()
		for k, v := range options.QueryParams {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	// 获取请求ID和Trace ID
	var requestID, traceID string
	if reqID, ok := ctx.Value(log.RequestIDKey).(string); ok {
		requestID = reqID
		req.Header.Set("X-Request-ID", requestID)
	}
	if tID, ok := ctx.Value(log.TraceIDKey).(string); ok {
		traceID = tID
		req.Header.Set("X-Trace-ID", traceID)
	}

	// 获取日志记录器
	logger := log.FromContext(ctx).With().
		Str("method", method).
		Str("url", url).
		Str("request_id", requestID).
		Str("trace_id", traceID).
		Logger()

	// 记录请求日志
	if c.options.EnableLogging {
		logger.Debug().
			Interface("headers", req.Header).
			Interface("query_params", req.URL.Query()).
			Msg("发送HTTP请求")
	}

	// 决定是否重试
	shouldRetry := c.options.Retry
	maxRetries := c.options.MaxRetries
	if options != nil {
		if options.Retry {
			shouldRetry = true
			if options.MaxRetries > 0 {
				maxRetries = options.MaxRetries
			}
		} else {
			shouldRetry = false
		}
	}

	// 发送请求，带重试
	var resp *http.Response
	var lastErr error

	// 如果不重试，则最大重试次数设为0
	if !shouldRetry {
		maxRetries = 0
	}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// 如果是重试，则记录重试日志
		if attempt > 0 && c.options.EnableLogging {
			logger.Info().Int("attempt", attempt+1).Msg("HTTP请求重试")
		}

		// 发送请求
		resp, lastErr = c.httpClient.Do(req)
		if lastErr == nil && (resp.StatusCode < 500 || resp.StatusCode == http.StatusNotImplemented) {
			// 非服务器错误，不重试
			break
		}

		// 达到最大重试次数，退出
		if attempt == maxRetries {
			break
		}

		// 计算重试间隔
		retryInterval := c.options.RetryInterval * time.Duration(attempt+1)
		// 等待重试间隔
		select {
		case <-ctx.Done():
			return nil, errors.RequestTimeout("请求被取消或超时")
		case <-time.After(retryInterval):
			// 继续重试
		}
	}

	// 如果所有重试都失败
	if lastErr != nil {
		if c.options.EnableLogging {
			logger.Error().Err(lastErr).Msg("HTTP请求失败")
		}
		return nil, errors.InternalServerErrorWithError(lastErr, "HTTP请求执行失败")
	}

	// 读取响应体
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		if c.options.EnableLogging {
			logger.Error().Err(err).Msg("读取HTTP响应失败")
		}
		return nil, errors.InternalServerErrorWithError(err, "读取HTTP响应失败")
	}

	// 创建响应对象
	response := &Response{
		StatusCode: resp.StatusCode,
		Body:       respBody,
		Headers:    resp.Header,
	}

	// 记录响应日志
	if c.options.EnableLogging {
		logger.Debug().
			Int("status", resp.StatusCode).
			Int("body_size", len(respBody)).
			Interface("headers", resp.Header).
			Msg("收到HTTP响应")
	}

	// 检查状态码，记录非2xx状态码的错误
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if c.options.EnableLogging {
			logger.Warn().
				Int("status", resp.StatusCode).
				Str("body", string(respBody)).
				Msg("HTTP请求返回非成功状态码")
		}
	}

	return response, nil
}

// GetJSON 发送GET请求并解析JSON响应到目标对象
func (c *Client) GetJSON(ctx context.Context, url string, target interface{}, options *RequestOptions) error {
	return c.RequestJSON(ctx, http.MethodGet, url, nil, target, options)
}

// PostJSON 发送POST请求并解析JSON响应到目标对象
func (c *Client) PostJSON(ctx context.Context, url string, body interface{}, target interface{}, options *RequestOptions) error {
	return c.RequestJSON(ctx, http.MethodPost, url, body, target, options)
}

// PutJSON 发送PUT请求并解析JSON响应到目标对象
func (c *Client) PutJSON(ctx context.Context, url string, body interface{}, target interface{}, options *RequestOptions) error {
	return c.RequestJSON(ctx, http.MethodPut, url, body, target, options)
}

// DeleteJSON 发送DELETE请求并解析JSON响应到目标对象
func (c *Client) DeleteJSON(ctx context.Context, url string, target interface{}, options *RequestOptions) error {
	return c.RequestJSON(ctx, http.MethodDelete, url, nil, target, options)
}

// PatchJSON 发送PATCH请求并解析JSON响应到目标对象
func (c *Client) PatchJSON(ctx context.Context, url string, body interface{}, target interface{}, options *RequestOptions) error {
	return c.RequestJSON(ctx, http.MethodPatch, url, body, target, options)
}

// RequestJSON 发送HTTP请求并解析JSON响应到目标对象
func (c *Client) RequestJSON(ctx context.Context, method, url string, body interface{}, target interface{}, options *RequestOptions) error {
	resp, err := c.Request(ctx, method, url, body, options)
	if err != nil {
		return err
	}

	// 检查状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.New(
			errors.CodeInternalServer,
			resp.StatusCode,
			fmt.Sprintf("HTTP请求失败，状态码: %d", resp.StatusCode),
		).WithDetail(string(resp.Body))
	}

	// 解析JSON响应
	if err := json.Unmarshal(resp.Body, target); err != nil {
		return errors.InternalServerErrorWithError(err, "解析JSON响应失败")
	}

	return nil
}

// SetCustomTransport 设置自定义的Transport
func (c *Client) SetCustomTransport(transport http.RoundTripper) {
	c.httpClient.Transport = transport
}

// Close 关闭客户端
func (c *Client) Close() {
	// 目前没有需要关闭的资源，预留接口
}

// 工具函数
func isAbsoluteURL(url string) bool {
	return len(url) > 7 && (url[:7] == "http://" || url[:8] == "https://")
}

func trimTrailingSlash(url string) string {
	if url == "" {
		return url
	}
	if url[len(url)-1] == '/' {
		return url[:len(url)-1]
	}
	return url
}

func trimLeadingSlash(url string) string {
	if url == "" {
		return url
	}
	if url[0] == '/' {
		return url[1:]
	}
	return url
}
