# HTTP响应格式化工具

本包提供了威胁情报平台(TIP)的HTTP响应格式化功能，包括统一的响应结构、成功/错误处理和中间件等。

## 功能特性

- 统一的API响应结构
- 支持标准HTTP库和Gin框架
- 自动错误处理与日志记录
- 响应中间件
- 支持功能选项模式自定义响应

## 使用方法

### 标准HTTP库中使用

```go
import (
    "context"
    "net/http"
    
    "github.com/XXXXD-cation/XXXXD-TIP/service/pkg/errors"
    httputil "github.com/XXXXD-cation/XXXXD-TIP/service/pkg/http"
)

func userHandler(w http.ResponseWriter, r *http.Request) {
    ctx := context.Background()
    
    // 成功响应
    user := map[string]interface{}{
        "id": "123",
        "name": "张三",
    }
    httputil.Success(w, user)
    
    // 错误响应
    err := errors.NotFound("用户不存在")
    httputil.Fail(ctx, w, err)
}
```

### Gin框架中使用

```go
import (
    "github.com/gin-gonic/gin"
    
    "github.com/XXXXD-cation/XXXXD-TIP/service/pkg/errors"
    httputil "github.com/XXXXD-cation/XXXXD-TIP/service/pkg/http"
)

func setupRouter() *gin.Engine {
    r := gin.Default()
    
    // 添加全局错误处理中间件
    r.Use(httputil.GinErrorHandler())
    
    r.GET("/user/:id", func(c *gin.Context) {
        id := c.Param("id")
        
        if id == "" {
            err := errors.ValidationError("用户ID不能为空")
            // 方式1: 直接处理错误
            httputil.GinFail(c, err)
            return
        }
        
        if id != "123" {
            // 方式2: 通过Gin的错误系统，会被GinErrorHandler捕获
            c.Error(errors.NotFound("用户不存在"))
            return
        }
        
        // 成功响应
        user := map[string]interface{}{
            "id": id,
            "name": "张三",
        }
        httputil.GinSuccess(c, user)
    })
    
    return r
}
```

### 自定义响应

```go
// 创建自定义响应
resp := httputil.NewResponse(
    0,                 // 状态码 
    "操作成功",          // 消息
    httputil.WithData(userData),                 // 添加数据
    httputil.WithError("仅用于调试的详细信息"),      // 添加错误信息
)

// 发送自定义响应
httputil.JSONResponse(w, http.StatusOK, resp)
```

### 错误处理中间件

```go
// 错误处理中间件
func errorHandlerMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
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
        
        next(w, r)
    }
}

// 使用中间件
http.HandleFunc("/api/users", errorHandlerMiddleware(getUserHandler))
```

## 响应格式

### 成功响应

```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "id": "123",
    "name": "张三"
  }
}
```

### 错误响应

```json
{
  "code": 4000,
  "message": "用户不存在",
  "error": "找不到ID为xyz的用户记录"
}
```

## 示例代码

请参考 `example/main.go` 获取完整示例。 