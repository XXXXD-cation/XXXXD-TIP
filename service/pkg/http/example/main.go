// Package main 提供HTTP响应格式化工具的使用示例
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/XXXXD-cation/XXXXD-TIP/service/pkg/errors"
	httputil "github.com/XXXXD-cation/XXXXD-TIP/service/pkg/http"
)

// 模拟的用户数据
var users = map[string]map[string]interface{}{
	"1": {
		"id":   "1",
		"name": "张三",
		"age":  30,
	},
	"2": {
		"id":   "2",
		"name": "李四",
		"age":  25,
	},
}

// getUserHandler 处理获取用户的请求
func getUserHandler(w http.ResponseWriter, r *http.Request) {
	// 创建上下文(在实际应用中，这通常由中间件提供)
	ctx := context.Background()

	// 获取用户ID
	userID := r.URL.Query().Get("id")
	if userID == "" {
		err := errors.ValidationError("用户ID不能为空")
		httputil.Fail(ctx, w, err)
		return
	}

	// 查找用户
	user, exists := users[userID]
	if !exists {
		err := errors.NotFound("用户不存在")
		httputil.Fail(ctx, w, err)
		return
	}

	// 返回成功响应
	httputil.Success(w, user)
}

// createUserHandler 处理创建用户的请求
func createUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// 解析请求体
	var userData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&userData); err != nil {
		httputil.Fail(ctx, w, errors.BadRequestWithError(err, "无效的请求数据"))
		return
	}

	// 验证必填字段
	if userData["name"] == nil {
		httputil.Fail(ctx, w, errors.ValidationError("用户名不能为空"))
		return
	}

	// 生成新ID (简单示例)
	newID := strconv.Itoa(len(users) + 1)

	// 添加ID
	userData["id"] = newID

	// 保存用户
	users[newID] = userData

	// 返回成功响应
	httputil.Success(w, userData)
}

// authMiddleware 模拟身份验证中间件
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()

		// 检查授权头
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			httputil.Fail(ctx, w, errors.Unauthorized("未提供认证信息"))
			return
		}

		// 模拟令牌验证 (简化示例)
		if authHeader != "Bearer valid-token" {
			httputil.Fail(ctx, w, errors.Unauthorized("无效的认证令牌"))
			return
		}

		// 认证成功，继续处理请求
		next(w, r)
	}
}

// 实现通用的错误处理中间件
func errorHandlerMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 设置panic恢复
		defer func() {
			if err := recover(); err != nil {
				// 将panic转换为错误响应
				ctx := context.Background()
				var appErr error
				switch e := err.(type) {
				case error:
					appErr = errors.InternalServerErrorWithError(e, "服务器内部错误")
				default:
					appErr = errors.InternalServerError(fmt.Sprintf("未处理的异常: %v", err))
				}
				httputil.Fail(ctx, w, appErr)
			}
		}()

		// 继续处理请求
		next(w, r)
	}
}

func main() {
	// 注册路由
	http.HandleFunc("/api/users", errorHandlerMiddleware(authMiddleware(getUserHandler)))
	http.HandleFunc("/api/users/create", errorHandlerMiddleware(authMiddleware(createUserHandler)))

	// 启动服务器
	fmt.Println("启动HTTP服务器在 :8088 端口...")
	fmt.Println("示例API:")
	fmt.Println("1. GET /api/users?id=1 (需要 Authorization: Bearer valid-token 头)")
	fmt.Println("2. POST /api/users/create (需要 Authorization: Bearer valid-token 头和JSON请求体)")
	log.Fatal(http.ListenAndServe(":8088", nil))
}
