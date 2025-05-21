// Package errors 提供了应用程序的错误处理功能
// 定义了标准错误类型、错误码和错误处理工具函数
package errors

import (
	"fmt"
	"net/http"
)

// Error 定义标准化的错误结构
type Error struct {
	// Code 是自定义错误码，用于唯一标识错误类型
	Code int `json:"code"`
	// Status 是HTTP状态码
	Status int `json:"status"`
	// Message 是面向用户的错误消息
	Message string `json:"message"`
	// Detail 是详细的错误信息，通常只在开发环境展示
	Detail string `json:"detail,omitempty"`
	// Internal 是内部错误，不会暴露给外部
	Internal error `json:"-"`
}

// Error 实现error接口
func (e *Error) Error() string {
	if e.Internal != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Internal)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap 返回内部错误，兼容errors.Unwrap
func (e *Error) Unwrap() error {
	return e.Internal
}

// WithDetail 为错误添加详细信息并返回同一个错误实例
func (e *Error) WithDetail(detail string) *Error {
	e.Detail = detail
	return e
}

// WithInternal 设置内部错误并返回同一个错误实例
func (e *Error) WithInternal(err error) *Error {
	e.Internal = err
	return e
}

// 下面定义一些常用的错误码范围
const (
	// CodeInvalid 表示无效的输入参数错误 (1000-1999)
	CodeInvalid = 1000
	// CodeUnauthorized 表示认证错误 (2000-2999)
	CodeUnauthorized = 2000
	// CodeForbidden 表示授权错误 (3000-3999)
	CodeForbidden = 3000
	// CodeNotFound 表示资源不存在错误 (4000-4999)
	CodeNotFound = 4000
	// CodeAlreadyExists 表示资源已存在错误 (5000-5999)
	CodeAlreadyExists = 5000
	// CodeResourceExhausted 表示资源耗尽错误 (6000-6999)
	CodeResourceExhausted = 6000
	// CodeInternalServer 表示内部服务器错误 (7000-7999)
	CodeInternalServer = 7000
	// CodeUnavailable 表示服务不可用错误 (8000-8999)
	CodeUnavailable = 8000
	// CodeTimeout 表示超时错误 (9000-9999)
	CodeTimeout = 9000
)

// 创建标准错误的工厂函数

// New 创建一个新的错误
func New(code int, status int, message string) *Error {
	return &Error{
		Code:    code,
		Status:  status,
		Message: message,
	}
}

// Wrap 包装一个已有错误
func Wrap(err error, code int, status int, message string) *Error {
	return &Error{
		Code:     code,
		Status:   status,
		Message:  message,
		Internal: err,
	}
}

// 预定义一些常用错误

// BadRequest 创建一个无效请求错误
func BadRequest(message string) *Error {
	return New(CodeInvalid, http.StatusBadRequest, message)
}

// BadRequestWithError 包装错误为无效请求错误
func BadRequestWithError(err error, message string) *Error {
	return Wrap(err, CodeInvalid, http.StatusBadRequest, message)
}

// Unauthorized 创建一个未授权错误
func Unauthorized(message string) *Error {
	if message == "" {
		message = "认证失败"
	}
	return New(CodeUnauthorized, http.StatusUnauthorized, message)
}

// Forbidden 创建一个权限不足错误
func Forbidden(message string) *Error {
	if message == "" {
		message = "权限不足"
	}
	return New(CodeForbidden, http.StatusForbidden, message)
}

// NotFound 创建一个资源不存在错误
func NotFound(message string) *Error {
	if message == "" {
		message = "资源不存在"
	}
	return New(CodeNotFound, http.StatusNotFound, message)
}

// Conflict 创建一个资源冲突错误
func Conflict(message string) *Error {
	return New(CodeAlreadyExists, http.StatusConflict, message)
}

// InternalServerError 创建一个内部服务器错误
func InternalServerError(message string) *Error {
	if message == "" {
		message = "服务器内部错误"
	}
	return New(CodeInternalServer, http.StatusInternalServerError, message)
}

// InternalServerErrorWithError 包装错误为内部服务器错误
func InternalServerErrorWithError(err error, message string) *Error {
	if message == "" {
		message = "服务器内部错误"
	}
	return Wrap(err, CodeInternalServer, http.StatusInternalServerError, message)
}

// ServiceUnavailable 创建一个服务不可用错误
func ServiceUnavailable(message string) *Error {
	if message == "" {
		message = "服务暂不可用"
	}
	return New(CodeUnavailable, http.StatusServiceUnavailable, message)
}

// RequestTimeout 创建一个请求超时错误
func RequestTimeout(message string) *Error {
	if message == "" {
		message = "请求超时"
	}
	return New(CodeTimeout, http.StatusRequestTimeout, message)
}

// ValidationError 创建一个参数验证错误
func ValidationError(message string) *Error {
	return New(CodeInvalid+1, http.StatusBadRequest, message)
}

// DatabaseError 创建一个数据库操作错误
func DatabaseError(err error) *Error {
	return Wrap(err, CodeInternalServer+1, http.StatusInternalServerError, "数据库操作失败")
}

// IsErrorCode 检查错误是否为指定错误码
func IsErrorCode(err error, code int) bool {
	var e *Error
	if err == nil {
		return false
	}

	switch v := err.(type) {
	case *Error:
		e = v
	default:
		return false
	}

	return e.Code == code
}
