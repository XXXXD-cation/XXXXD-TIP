// Package http 提供了HTTP相关的工具函数和类型
// 包括统一的响应格式化、请求处理和中间件等
package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/XXXXD-cation/XXXXD-TIP/service/pkg/errors"
	"github.com/XXXXD-cation/XXXXD-TIP/service/pkg/log"
)

// Response 定义统一的API响应格式
type Response struct {
	// Code 是业务状态码
	Code int `json:"code"`
	// Message 是响应消息
	Message string `json:"message"`
	// Data 是响应数据
	Data interface{} `json:"data,omitempty"`
	// Error 是错误详情，只在非生产环境下返回
	Error string `json:"error,omitempty"`
}

// ResponseOption 是响应选项函数类型
type ResponseOption func(*Response)

// WithData 添加响应数据
func WithData(data interface{}) ResponseOption {
	return func(r *Response) {
		r.Data = data
	}
}

// WithError 添加错误详情
func WithError(err string) ResponseOption {
	return func(r *Response) {
		r.Error = err
	}
}

// NewResponse 创建一个新的响应对象
func NewResponse(code int, message string, opts ...ResponseOption) *Response {
	resp := &Response{
		Code:    code,
		Message: message,
	}

	for _, opt := range opts {
		opt(resp)
	}

	return resp
}

// JSONResponse 是一个通用的JSON响应发送器 (适用于标准net/http)
func JSONResponse(w http.ResponseWriter, statusCode int, resp *Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		// 如果编码失败，记录错误但不能再向客户端发送响应
		log.Error().Err(err).Msg("无法编码JSON响应")
	}
}

// Success 发送成功响应 (适用于标准net/http)
func Success(w http.ResponseWriter, data interface{}) {
	resp := NewResponse(0, "成功", WithData(data))
	JSONResponse(w, http.StatusOK, resp)
}

// Fail 发送错误响应 (适用于标准net/http)
func Fail(ctx context.Context, w http.ResponseWriter, err error) {
	statusCode := http.StatusInternalServerError
	resp := &Response{
		Code:    errors.CodeInternalServer,
		Message: "服务器内部错误",
	}

	// 如果是自定义错误类型
	if appErr, ok := err.(*errors.Error); ok {
		statusCode = appErr.Status
		resp.Code = appErr.Code
		resp.Message = appErr.Message

		// 根据环境添加详细错误信息
		if appErr.Detail != "" {
			resp.Error = appErr.Detail
		}
	} else {
		// 非内部定义的错误，可能是Go标准库错误等
		resp.Error = err.Error()
	}

	// 记录错误日志
	logger := log.FromContext(ctx)
	logger.Error().Int("status", statusCode).
		Int("code", resp.Code).
		Str("message", resp.Message).
		Str("error", resp.Error).
		Msg("请求处理失败")

	JSONResponse(w, statusCode, resp)
}

// 以下是Gin框架特定的响应助手

// GinSuccess 发送Gin框架的成功响应
func GinSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, NewResponse(0, "成功", WithData(data)))
}

// GinFail 发送Gin框架的错误响应
func GinFail(c *gin.Context, err error) {
	statusCode := http.StatusInternalServerError
	resp := &Response{
		Code:    errors.CodeInternalServer,
		Message: "服务器内部错误",
	}

	// 如果是自定义错误类型
	if appErr, ok := err.(*errors.Error); ok {
		statusCode = appErr.Status
		resp.Code = appErr.Code
		resp.Message = appErr.Message

		// 根据环境添加详细错误信息
		if appErr.Detail != "" {
			resp.Error = appErr.Detail
		}
	} else {
		// 非内部定义的错误，可能是Go标准库错误等
		resp.Error = err.Error()
	}

	// 获取上下文中的请求ID和跟踪ID
	requestID, _ := c.Get("request_id")
	traceID, _ := c.Get("trace_id")

	// 记录错误日志
	log.Error().
		Int("status", statusCode).
		Int("code", resp.Code).
		Str("message", resp.Message).
		Str("error", resp.Error).
		Interface("request_id", requestID).
		Interface("trace_id", traceID).
		Str("path", c.Request.URL.Path).
		Str("method", c.Request.Method).
		Msg("请求处理失败")

	c.JSON(statusCode, resp)
}

// GinErrorHandler 是Gin框架的全局错误处理中间件
func GinErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 如果有错误，统一处理
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			GinFail(c, err)
			// 标记请求已处理，防止Gin再次发送响应
			c.Abort()
		}
	}
}
