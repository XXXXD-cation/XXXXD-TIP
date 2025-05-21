package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	apperrors "github.com/XXXXD-cation/XXXXD-TIP/service/pkg/errors"
)

// 初始化测试模式
func init() {
	gin.SetMode(gin.TestMode)
}

// TestNewResponse 测试响应对象创建
func TestNewResponse(t *testing.T) {
	// 基本响应
	resp := NewResponse(0, "成功")
	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "成功", resp.Message)
	assert.Nil(t, resp.Data)
	assert.Empty(t, resp.Error)

	// 带数据的响应
	data := map[string]interface{}{"name": "张三", "age": 30}
	resp = NewResponse(0, "成功", WithData(data))
	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "成功", resp.Message)
	assert.Equal(t, data, resp.Data)
	assert.Empty(t, resp.Error)

	// 带错误的响应
	resp = NewResponse(1001, "参数错误", WithError("用户名不能为空"))
	assert.Equal(t, 1001, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	assert.Nil(t, resp.Data)
	assert.Equal(t, "用户名不能为空", resp.Error)

	// 完整响应
	resp = NewResponse(1001, "参数错误", WithData(data), WithError("用户名不能为空"))
	assert.Equal(t, 1001, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	assert.Equal(t, data, resp.Data)
	assert.Equal(t, "用户名不能为空", resp.Error)
}

// TestJSONResponse 测试JSONResponse函数
func TestJSONResponse(t *testing.T) {
	// 创建HTTP响应记录器
	w := httptest.NewRecorder()

	// 创建响应对象
	data := map[string]interface{}{"name": "张三", "age": float64(30)}
	resp := NewResponse(0, "成功", WithData(data))

	// 发送响应
	JSONResponse(w, http.StatusOK, resp)

	// 验证HTTP状态码
	assert.Equal(t, http.StatusOK, w.Code)

	// 验证Content-Type头
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	// 解析响应主体
	var respBody Response
	err := json.Unmarshal(w.Body.Bytes(), &respBody)
	assert.NoError(t, err)

	// 验证响应内容
	assert.Equal(t, 0, respBody.Code)
	assert.Equal(t, "成功", respBody.Message)
	assert.Equal(t, data, respBody.Data)
}

// TestSuccess 测试Success函数
func TestSuccess(t *testing.T) {
	// 创建HTTP响应记录器
	w := httptest.NewRecorder()

	// 准备数据
	data := map[string]interface{}{"name": "张三", "age": float64(30)}

	// 发送成功响应
	Success(w, data)

	// 验证HTTP状态码
	assert.Equal(t, http.StatusOK, w.Code)

	// 解析响应主体
	var respBody Response
	err := json.Unmarshal(w.Body.Bytes(), &respBody)
	assert.NoError(t, err)

	// 验证响应内容
	assert.Equal(t, 0, respBody.Code)
	assert.Equal(t, "成功", respBody.Message)
	assert.Equal(t, data, respBody.Data)
}

// TestFail 测试Fail函数 - 使用自定义错误
func TestFailWithAppError(t *testing.T) {
	// 创建HTTP响应记录器
	w := httptest.NewRecorder()

	// 创建上下文
	ctx := context.Background()

	// 创建自定义错误
	err := apperrors.NotFound("用户不存在").WithDetail("找不到ID为123的用户")

	// 发送错误响应
	Fail(ctx, w, err)

	// 验证HTTP状态码
	assert.Equal(t, http.StatusNotFound, w.Code)

	// 解析响应主体
	var respBody Response
	jsonErr := json.Unmarshal(w.Body.Bytes(), &respBody)
	assert.NoError(t, jsonErr)

	// 验证响应内容
	assert.Equal(t, apperrors.CodeNotFound, respBody.Code)
	assert.Equal(t, "用户不存在", respBody.Message)
	assert.Equal(t, "找不到ID为123的用户", respBody.Error)
}

// TestFail 测试Fail函数 - 使用标准错误
func TestFailWithStandardError(t *testing.T) {
	// 创建HTTP响应记录器
	w := httptest.NewRecorder()

	// 创建上下文
	ctx := context.Background()

	// 创建标准错误
	err := errors.New("发生了错误")

	// 发送错误响应
	Fail(ctx, w, err)

	// 验证HTTP状态码
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// 解析响应主体
	var respBody Response
	jsonErr := json.Unmarshal(w.Body.Bytes(), &respBody)
	assert.NoError(t, jsonErr)

	// 验证响应内容
	assert.Equal(t, apperrors.CodeInternalServer, respBody.Code)
	assert.Equal(t, "服务器内部错误", respBody.Message)
	assert.Equal(t, "发生了错误", respBody.Error)
}

// TestGinSuccess 测试GinSuccess函数
func TestGinSuccess(t *testing.T) {
	// 设置Gin路由
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		data := map[string]interface{}{"name": "张三", "age": 30}
		GinSuccess(c, data)
	})

	// 创建HTTP请求
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// 处理请求
	router.ServeHTTP(w, req)

	// 验证HTTP状态码
	assert.Equal(t, http.StatusOK, w.Code)

	// 解析响应主体
	var respBody Response
	err := json.Unmarshal(w.Body.Bytes(), &respBody)
	assert.NoError(t, err)

	// 验证响应内容
	assert.Equal(t, 0, respBody.Code)
	assert.Equal(t, "成功", respBody.Message)
	assert.NotNil(t, respBody.Data)

	// 验证数据
	dataMap, ok := respBody.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "张三", dataMap["name"])
	assert.Equal(t, float64(30), dataMap["age"]) // JSON数字会被解析为float64
}

// TestGinFail 测试GinFail函数
func TestGinFail(t *testing.T) {
	// 设置Gin路由
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		err := apperrors.NotFound("用户不存在")
		GinFail(c, err)
	})

	// 创建HTTP请求
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// 处理请求
	router.ServeHTTP(w, req)

	// 验证HTTP状态码
	assert.Equal(t, http.StatusNotFound, w.Code)

	// 解析响应主体
	var respBody Response
	err := json.Unmarshal(w.Body.Bytes(), &respBody)
	assert.NoError(t, err)

	// 验证响应内容
	assert.Equal(t, apperrors.CodeNotFound, respBody.Code)
	assert.Equal(t, "用户不存在", respBody.Message)
}

// TestGinErrorHandler 测试Gin错误处理中间件
func TestGinErrorHandler(t *testing.T) {
	// 设置Gin路由
	router := gin.New()
	router.Use(GinErrorHandler())

	router.GET("/test", func(c *gin.Context) {
		// 添加错误
		err := apperrors.BadRequest("参数错误")
		c.Error(err)
	})

	// 创建HTTP请求
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// 处理请求
	router.ServeHTTP(w, req)

	// 验证HTTP状态码
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// 解析响应主体
	var respBody Response
	err := json.Unmarshal(w.Body.Bytes(), &respBody)
	assert.NoError(t, err)

	// 验证响应内容
	assert.Equal(t, apperrors.CodeInvalid, respBody.Code)
	assert.Equal(t, "参数错误", respBody.Message)
}
