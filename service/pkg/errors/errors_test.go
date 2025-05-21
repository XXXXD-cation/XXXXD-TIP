package errors

import (
	"fmt"
	"net/http"
	"testing"
)

// TestErrorCreation 测试错误创建
func TestErrorCreation(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		status   int
		message  string
		expected string
	}{
		{
			name:     "基本错误",
			code:     1001,
			status:   400,
			message:  "参数错误",
			expected: "[1001] 参数错误",
		},
		{
			name:     "带代码的错误",
			code:     2001,
			status:   401,
			message:  "未授权访问",
			expected: "[2001] 未授权访问",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := New(tt.code, tt.status, tt.message)
			if err.Error() != tt.expected {
				t.Errorf("Error() = %v, 期望 %v", err.Error(), tt.expected)
			}
			if err.Code != tt.code {
				t.Errorf("err.Code = %v, 期望 %v", err.Code, tt.code)
			}
			if err.Status != tt.status {
				t.Errorf("err.Status = %v, 期望 %v", err.Status, tt.status)
			}
			if err.Message != tt.message {
				t.Errorf("err.Message = %v, 期望 %v", err.Message, tt.message)
			}
		})
	}
}

// TestErrorWrapping 测试错误包装
func TestErrorWrapping(t *testing.T) {
	baseErr := fmt.Errorf("原始错误")
	wrappedErr := Wrap(baseErr, 7001, 500, "包装的错误")

	if wrappedErr.Error() != "[7001] 包装的错误: 原始错误" {
		t.Errorf("错误消息格式不正确，得到: %s", wrappedErr.Error())
	}

	if unwrapped := wrappedErr.Unwrap(); unwrapped != baseErr {
		t.Errorf("Unwrap() = %v, 期望 %v", unwrapped, baseErr)
	}

	if wrappedErr.Internal != baseErr {
		t.Errorf("wrappedErr.Internal = %v, 期望 %v", wrappedErr.Internal, baseErr)
	}
}

// TestWithDetail 测试添加详细信息
func TestWithDetail(t *testing.T) {
	err := New(1001, 400, "参数错误")
	err = err.WithDetail("用户名不能为空")

	if err.Detail != "用户名不能为空" {
		t.Errorf("err.Detail = %v, 期望 '用户名不能为空'", err.Detail)
	}
}

// TestWithInternal 测试添加内部错误
func TestWithInternal(t *testing.T) {
	internalErr := fmt.Errorf("内部错误")
	err := New(7001, 500, "服务器错误")
	err = err.WithInternal(internalErr)

	if err.Internal != internalErr {
		t.Errorf("err.Internal = %v, 期望 %v", err.Internal, internalErr)
	}

	if err.Error() != "[7001] 服务器错误: 内部错误" {
		t.Errorf("错误消息格式不正确，得到: %s", err.Error())
	}
}

// TestPredefinedErrors 测试预定义错误类型
func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        *Error
		code       int
		statusCode int
	}{
		{
			name:       "BadRequest",
			err:        BadRequest("无效参数"),
			code:       CodeInvalid,
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "Unauthorized",
			err:        Unauthorized("认证失败"),
			code:       CodeUnauthorized,
			statusCode: http.StatusUnauthorized,
		},
		{
			name:       "Forbidden",
			err:        Forbidden("权限不足"),
			code:       CodeForbidden,
			statusCode: http.StatusForbidden,
		},
		{
			name:       "NotFound",
			err:        NotFound("资源不存在"),
			code:       CodeNotFound,
			statusCode: http.StatusNotFound,
		},
		{
			name:       "Conflict",
			err:        Conflict("资源冲突"),
			code:       CodeAlreadyExists,
			statusCode: http.StatusConflict,
		},
		{
			name:       "InternalServerError",
			err:        InternalServerError("服务器错误"),
			code:       CodeInternalServer,
			statusCode: http.StatusInternalServerError,
		},
		{
			name:       "ServiceUnavailable",
			err:        ServiceUnavailable("服务不可用"),
			code:       CodeUnavailable,
			statusCode: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Code != tt.code {
				t.Errorf("err.Code = %v, 期望 %v", tt.err.Code, tt.code)
			}
			if tt.err.Status != tt.statusCode {
				t.Errorf("err.Status = %v, 期望 %v", tt.err.Status, tt.statusCode)
			}
		})
	}
}

// TestPredefinedErrorsWithDefaultMessages 测试预定义错误的默认消息
func TestPredefinedErrorsWithDefaultMessages(t *testing.T) {
	tests := []struct {
		name           string
		createError    func(string) *Error
		expectedMsg    string
		whenEmptyInput bool
	}{
		{
			name:           "Unauthorized 默认消息",
			createError:    Unauthorized,
			expectedMsg:    "认证失败",
			whenEmptyInput: true,
		},
		{
			name:           "Forbidden 默认消息",
			createError:    Forbidden,
			expectedMsg:    "权限不足",
			whenEmptyInput: true,
		},
		{
			name:           "NotFound 默认消息",
			createError:    NotFound,
			expectedMsg:    "资源不存在",
			whenEmptyInput: true,
		},
		{
			name:           "InternalServerError 默认消息",
			createError:    InternalServerError,
			expectedMsg:    "服务器内部错误",
			whenEmptyInput: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err *Error
			if tt.whenEmptyInput {
				err = tt.createError("")
			} else {
				err = tt.createError(tt.expectedMsg)
			}

			if err.Message != tt.expectedMsg {
				t.Errorf("err.Message = %v, 期望 %v", err.Message, tt.expectedMsg)
			}
		})
	}
}

// TestErrorWithWrapping 测试带包装的预定义错误
func TestErrorWithWrapping(t *testing.T) {
	baseErr := fmt.Errorf("底层错误")
	wrappedErr := InternalServerErrorWithError(baseErr, "服务器错误")

	if wrappedErr.Code != CodeInternalServer {
		t.Errorf("wrappedErr.Code = %v, 期望 %v", wrappedErr.Code, CodeInternalServer)
	}

	if wrappedErr.Status != http.StatusInternalServerError {
		t.Errorf("wrappedErr.Status = %v, 期望 %v", wrappedErr.Status, http.StatusInternalServerError)
	}

	if wrappedErr.Message != "服务器错误" {
		t.Errorf("wrappedErr.Message = %v, 期望 '服务器错误'", wrappedErr.Message)
	}

	if wrappedErr.Internal != baseErr {
		t.Errorf("wrappedErr.Internal = %v, 期望 %v", wrappedErr.Internal, baseErr)
	}
}

// TestIsErrorCode 测试错误码检查
func TestIsErrorCode(t *testing.T) {
	err1 := NotFound("资源不存在")
	err2 := fmt.Errorf("普通错误")
	var err3 error = nil

	if !IsErrorCode(err1, CodeNotFound) {
		t.Error("IsErrorCode(NotFound错误, CodeNotFound) 应该返回 true")
	}

	if IsErrorCode(err1, CodeInvalid) {
		t.Error("IsErrorCode(NotFound错误, CodeInvalid) 应该返回 false")
	}

	if IsErrorCode(err2, CodeNotFound) {
		t.Error("IsErrorCode(普通错误, CodeNotFound) 应该返回 false")
	}

	if IsErrorCode(err3, CodeNotFound) {
		t.Error("IsErrorCode(nil, CodeNotFound) 应该返回 false")
	}
}

// TestDatabaseError 测试数据库错误
func TestDatabaseError(t *testing.T) {
	dbErr := fmt.Errorf("数据库连接失败")
	err := DatabaseError(dbErr)

	if err.Code != CodeInternalServer+1 {
		t.Errorf("err.Code = %v, 期望 %v", err.Code, CodeInternalServer+1)
	}

	if err.Status != http.StatusInternalServerError {
		t.Errorf("err.Status = %v, 期望 %v", err.Status, http.StatusInternalServerError)
	}

	if err.Message != "数据库操作失败" {
		t.Errorf("err.Message = %v, 期望 '数据库操作失败'", err.Message)
	}

	if err.Internal != dbErr {
		t.Errorf("err.Internal = %v, 期望 %v", err.Internal, dbErr)
	}
}

// TestValidationError 测试验证错误
func TestValidationError(t *testing.T) {
	err := ValidationError("用户名不能为空")

	if err.Code != CodeInvalid+1 {
		t.Errorf("err.Code = %v, 期望 %v", err.Code, CodeInvalid+1)
	}

	if err.Status != http.StatusBadRequest {
		t.Errorf("err.Status = %v, 期望 %v", err.Status, http.StatusBadRequest)
	}

	if err.Message != "用户名不能为空" {
		t.Errorf("err.Message = %v, 期望 '用户名不能为空'", err.Message)
	}
}
