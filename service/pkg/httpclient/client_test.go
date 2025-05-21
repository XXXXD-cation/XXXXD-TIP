package httpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/XXXXD-cation/XXXXD-TIP/service/pkg/log"
	"github.com/stretchr/testify/assert"
)

func init() {
	// 配置日志，避免测试时输出过多日志
	log.Setup(log.Config{
		Level:       "error",
		Pretty:      false,
		WithCaller:  true,
		TimeFormat:  time.RFC3339,
		ServiceName: "httpclient-test",
	})
}

type TestResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
	Code    int    `json:"code"`
}

func setupTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/success":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(TestResponse{
				Message: "Success",
				Status:  "ok",
				Code:    200,
			})
		case "/error":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(TestResponse{
				Message: "Internal Server Error",
				Status:  "error",
				Code:    500,
			})
		case "/not-found":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(TestResponse{
				Message: "Not Found",
				Status:  "error",
				Code:    404,
			})
		case "/retry":
			// 第一次返回500，第二次返回200
			counter, ok := r.Context().Value("counter").(int)
			if !ok || counter == 0 {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Server Error"))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(TestResponse{
				Message: "Success after retry",
				Status:  "ok",
				Code:    200,
			})
		case "/timeout":
			// 模拟超时请求
			time.Sleep(2 * time.Second)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Timeout response"))
		case "/echo":
			// 回显请求的内容
			var data map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(fmt.Sprintf("Invalid request body: %v", err)))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(data)
		case "/query":
			// 返回查询参数
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := map[string]string{}
			for key, values := range r.URL.Query() {
				if len(values) > 0 {
					response[key] = values[0]
				}
			}
			json.NewEncoder(w).Encode(response)
		case "/headers":
			// 返回请求头
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := map[string]string{}
			for key, values := range r.Header {
				if len(values) > 0 {
					response[key] = values[0]
				}
			}
			json.NewEncoder(w).Encode(response)
		default:
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Not Found"))
		}
	}))
}

func TestGet(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	client := New(&Options{
		ConnectionTimeout: 1 * time.Second,
		RequestTimeout:    1 * time.Second,
		EnableLogging:     true,
	})

	ctx := context.Background()

	// 测试成功请求
	resp, err := client.Get(ctx, server.URL+"/success", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// 解析响应
	var response TestResponse
	err = json.Unmarshal(resp.Body, &response)
	assert.NoError(t, err)
	assert.Equal(t, "Success", response.Message)
	assert.Equal(t, "ok", response.Status)
	assert.Equal(t, 200, response.Code)

	// 测试失败请求
	resp, err = client.Get(ctx, server.URL+"/error", nil)
	assert.NoError(t, err) // 不会返回错误，只会在状态码中反映
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	// 测试不存在的路径
	resp, err = client.Get(ctx, server.URL+"/not-exist", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetJSON(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	client := New(&Options{
		ConnectionTimeout: 1 * time.Second,
		RequestTimeout:    1 * time.Second,
		EnableLogging:     true,
	})

	ctx := context.Background()

	// 测试成功请求
	var response TestResponse
	err := client.GetJSON(ctx, server.URL+"/success", &response, nil)
	assert.NoError(t, err)
	assert.Equal(t, "Success", response.Message)
	assert.Equal(t, "ok", response.Status)
	assert.Equal(t, 200, response.Code)

	// 测试失败请求
	err = client.GetJSON(ctx, server.URL+"/error", &response, nil)
	assert.Error(t, err) // 这里会返回错误，因为GetJSON要求状态码2xx
}

func TestPostJSON(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	client := New(&Options{
		ConnectionTimeout: 1 * time.Second,
		RequestTimeout:    1 * time.Second,
		EnableLogging:     true,
	})

	ctx := context.Background()

	// 测试Echo请求
	requestBody := map[string]interface{}{
		"name": "测试用户",
		"age":  30,
		"tags": []string{"golang", "testing"},
	}
	var responseBody map[string]interface{}
	err := client.PostJSON(ctx, server.URL+"/echo", requestBody, &responseBody, nil)
	assert.NoError(t, err)
	assert.Equal(t, requestBody["name"], responseBody["name"])
	assert.Equal(t, float64(30), responseBody["age"])
	assert.ElementsMatch(t, []string{"golang", "testing"}, responseBody["tags"].([]interface{}))
}

func TestQueryParams(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	client := New(&Options{
		ConnectionTimeout: 1 * time.Second,
		RequestTimeout:    1 * time.Second,
		EnableLogging:     true,
	})

	ctx := context.Background()

	// 测试查询参数
	queryParams := map[string]string{
		"name":     "张三",
		"age":      "30",
		"language": "zh-CN",
	}
	var response map[string]string
	err := client.GetJSON(ctx, server.URL+"/query", &response, &RequestOptions{
		QueryParams: queryParams,
	})
	assert.NoError(t, err)
	assert.Equal(t, queryParams["name"], response["name"])
	assert.Equal(t, queryParams["age"], response["age"])
	assert.Equal(t, queryParams["language"], response["language"])
}

func TestCustomHeaders(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	client := New(&Options{
		ConnectionTimeout: 1 * time.Second,
		RequestTimeout:    1 * time.Second,
		EnableLogging:     true,
		Headers: map[string]string{
			"X-Client-ID": "test-client",
		},
	})

	ctx := context.Background()

	// 测试请求头
	customHeaders := map[string]string{
		"X-Custom-Header": "测试值",
		"Authorization":   "Bearer test-token",
	}
	var response map[string]string
	err := client.GetJSON(ctx, server.URL+"/headers", &response, &RequestOptions{
		Headers: customHeaders,
	})
	assert.NoError(t, err)
	assert.Equal(t, customHeaders["X-Custom-Header"], response["X-Custom-Header"])
	assert.Equal(t, customHeaders["Authorization"], response["Authorization"])
	assert.Equal(t, "test-client", response["X-Client-Id"]) // 注意: HTTP头不区分大小写
}

func TestTimeout(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	client := New(&Options{
		ConnectionTimeout: 1 * time.Second,
		RequestTimeout:    1 * time.Second, // 设置1秒超时，而服务器会等待2秒
		EnableLogging:     true,
	})

	ctx := context.Background()

	// 测试超时请求
	_, err := client.Get(ctx, server.URL+"/timeout", nil)
	assert.Error(t, err) // 预期会返回超时错误
}

func TestBaseURL(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	client := New(&Options{
		BaseURL:           server.URL,
		ConnectionTimeout: 1 * time.Second,
		RequestTimeout:    1 * time.Second,
		EnableLogging:     true,
	})

	ctx := context.Background()

	// 测试使用相对路径
	resp, err := client.Get(ctx, "/success", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// 测试使用绝对路径（不使用BaseURL）
	// 使用一个无效的URL方案来确保请求失败
	_, err = client.Get(ctx, "invalid://example.com/fake-path", nil)
	assert.Error(t, err) // 这个请求会失败，因为URL方案无效
}

func TestURLHandling(t *testing.T) {
	// 测试URL处理函数
	assert.True(t, isAbsoluteURL("http://example.com"))
	assert.True(t, isAbsoluteURL("https://example.com"))
	assert.False(t, isAbsoluteURL("/api/v1/users"))
	assert.False(t, isAbsoluteURL("api/v1/users"))

	assert.Equal(t, "http://example.com", trimTrailingSlash("http://example.com/"))
	assert.Equal(t, "http://example.com", trimTrailingSlash("http://example.com"))

	assert.Equal(t, "api/v1/users", trimLeadingSlash("/api/v1/users"))
	assert.Equal(t, "api/v1/users", trimLeadingSlash("api/v1/users"))
}
