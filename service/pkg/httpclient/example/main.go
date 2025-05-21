// Package main 提供HTTP客户端封装的使用示例
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/XXXXD-cation/XXXXD-TIP/service/pkg/httpclient"
	pkglog "github.com/XXXXD-cation/XXXXD-TIP/service/pkg/log"
)

// 一个简单的API响应结构
type JSONPlaceholderTodo struct {
	UserID    int    `json:"userId"`
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

// 一个简单的API请求结构
type CreateTodoRequest struct {
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

// 启动一个本地HTTP服务器用于演示
func startLocalServer() string {
	mux := http.NewServeMux()

	// GET /todos/1 处理
	mux.HandleFunc("/todos/1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(JSONPlaceholderTodo{
			UserID:    1,
			ID:        1,
			Title:     "测试待办事项",
			Completed: false,
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	// 同时处理GET和POST /todos请求
	mux.HandleFunc("/todos", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case http.MethodGet:
			// 处理GET请求
			// 检查查询参数
			userId := r.URL.Query().Get("userId")
			completed := r.URL.Query().Get("completed")

			fmt.Printf("接收到查询参数: userId=%s, completed=%s\n", userId, completed)

			// 返回待办事项列表
			if err := json.NewEncoder(w).Encode([]JSONPlaceholderTodo{
				{
					UserID:    1,
					ID:        1,
					Title:     "测试待办事项1",
					Completed: false,
				},
				{
					UserID:    1,
					ID:        2,
					Title:     "测试待办事项2",
					Completed: true,
				},
			}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		case http.MethodPost:
			// 处理POST请求
			var req CreateTodoRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			if err := json.NewEncoder(w).Encode(JSONPlaceholderTodo{
				UserID:    1,
				ID:        101,
				Title:     req.Title,
				Completed: req.Completed,
			}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		default:
			http.Error(w, "不支持的请求方法", http.StatusMethodNotAllowed)
		}
	})

	// 处理自定义请求头
	mux.HandleFunc("/headers", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		headers := make(map[string]string)
		for k, v := range r.Header {
			if len(v) > 0 {
				headers[k] = v[0]
			}
		}
		if err := json.NewEncoder(w).Encode(headers); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	// 启动本地服务器
	server := &http.Server{
		Addr:              ":8090",
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second, // 防止慢速攻击
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 等待服务器启动
	time.Sleep(100 * time.Millisecond)

	return "http://localhost:8090"
}

func main() {
	// 设置日志
	pkglog.Setup(pkglog.Config{
		Level:      "debug",
		Pretty:     true,
		WithCaller: true,
		Output:     "console",
	})

	// 启动本地服务器
	baseURL := startLocalServer()
	fmt.Printf("本地服务器启动在: %s\n", baseURL)

	// 创建带有自定义选项的客户端
	options := httpclient.DefaultOptions()
	options.BaseURL = baseURL
	options.RequestTimeout = 30 * time.Second
	options.MaxRetries = 2

	client := httpclient.New(options)
	defer client.Close()

	// 创建上下文，并添加请求和跟踪ID
	ctx := context.Background()
	ctx = pkglog.WithRequestID(ctx, "example-req-123")
	ctx = pkglog.WithTraceID(ctx, "example-trace-456")

	// 示例1: 发送GET请求获取JSON数据并自动解析
	fmt.Println("\n=== 示例1: GetJSON - 获取单个待办事项 ===")
	var todo JSONPlaceholderTodo
	err := client.GetJSON(ctx, "/todos/1", &todo, nil)
	if err != nil {
		log.Fatalf("GetJSON失败: %v", err)
	}
	fmt.Printf("获取到的待办事项: ID=%d, 标题=%s, 已完成=%v\n", todo.ID, todo.Title, todo.Completed)

	// 示例2: 发送带查询参数的GET请求
	fmt.Println("\n=== 示例2: Get - 带查询参数 ===")
	resp, err := client.Get(ctx, "/todos", &httpclient.RequestOptions{
		QueryParams: map[string]string{
			"userId":    "1",
			"completed": "false",
		},
	})
	if err != nil {
		log.Fatalf("Get失败: %v", err)
	}
	fmt.Printf("状态码: %d, 响应体长度: %d\n", resp.StatusCode, len(resp.Body))
	fmt.Printf("响应头: %v\n", resp.Headers.Get("Content-Type"))

	// 示例3: 发送POST请求创建资源
	fmt.Println("\n=== 示例3: PostJSON - 创建新待办事项 ===")
	createReq := CreateTodoRequest{
		Title:     "学习Go HTTP客户端",
		Completed: false,
	}
	var createdTodo JSONPlaceholderTodo
	err = client.PostJSON(ctx, "/todos", createReq, &createdTodo, nil)
	if err != nil {
		log.Fatalf("PostJSON失败: %v", err)
	}
	fmt.Printf("创建的待办事项: ID=%d, 标题=%s\n", createdTodo.ID, createdTodo.Title)

	// 示例4: 自定义请求头
	fmt.Println("\n=== 示例4: 带自定义请求头的请求 ===")
	respHeaders, headersErr := client.Get(ctx, "/headers", &httpclient.RequestOptions{
		Headers: map[string]string{
			"X-Custom-Header": "custom-value",
			"Accept":          "application/xml", // 覆盖默认的Accept头
		},
	})
	if headersErr != nil {
		log.Fatalf("带自定义请求头的请求失败: %v", headersErr)
	}
	fmt.Printf("状态码: %d\n", respHeaders.StatusCode)

	// 解析响应体，显示请求头
	var headersResp map[string]string
	if unmarshalErr := json.Unmarshal(respHeaders.Body, &headersResp); unmarshalErr == nil {
		fmt.Println("服务器收到的请求头:")
		for k, v := range headersResp {
			fmt.Printf("  %s: %s\n", k, v)
		}
	}

	// 示例5: 禁用重试的请求
	fmt.Println("\n=== 示例5: 禁用重试的请求 ===")
	respNoRetry, noRetryErr := client.Get(ctx, "/todos/1", &httpclient.RequestOptions{
		Retry: false,
	})
	if noRetryErr != nil {
		log.Fatalf("禁用重试的请求失败: %v", noRetryErr)
	}
	fmt.Printf("状态码: %d\n", respNoRetry.StatusCode)

	fmt.Println("\n所有示例完成!")
}
