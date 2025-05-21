# 日志库

本包提供了基于 zerolog 的结构化日志封装，支持链路追踪、日志级别控制和多种输出格式。

## 特性

- 结构化日志：所有日志以JSON格式输出，便于机器解析和分析
- 链路追踪：支持Trace ID、Request ID、Span ID等，方便跟踪请求流程
- 日志级别：支持Debug, Info, Warn, Error, Fatal多个日志级别
- 灵活配置：支持控制台输出、文件输出，可配置美化输出等
- 上下文感知：支持从请求上下文中提取跟踪信息，自动附加到日志中
- 简便API：提供简洁的全局API和上下文感知API

## 使用示例

### 基本用法

```go
package main

import "github.com/XXXXD-cation/XXXXD-TIP/service/pkg/log"

func main() {
    // 使用全局日志函数记录日志
    log.Info().Str("app", "example").Int("version", 1).Msg("应用启动")
    log.Debug().Str("config", "debug mode").Msg("调试信息")
    log.Error().Err(err).Msg("发生错误")
}
```

### 配置

```go
// 自定义日志配置
config := log.Config{
    Level:       "debug",     // 日志级别：debug, info, warn, error, fatal
    Pretty:      true,        // 美化输出（适合开发环境）
    WithCaller:  true,        // 是否记录调用者信息
    TimeFormat:  time.RFC3339, // 时间格式
    Output:      "console",   // 输出位置: console, file:/path/to/file
    ServiceName: "my-service", // 服务名称
}
log.Setup(config)
```

### 链路追踪

```go
// 创建带有跟踪信息的上下文
ctx := context.Background()
ctx = log.WithTraceID(ctx, "trace-123")
ctx = log.WithRequestID(ctx, "req-456")
ctx = log.WithSpanID(ctx, "span-789")

// 从上下文中获取日志记录器
logger := log.FromContext(ctx)
logger.Info().Str("action", "process").Msg("处理请求")

// 子调用中创建子Span
childCtx := log.WithParentSpanID(ctx, "span-789")
childCtx = log.WithSpanID(childCtx, "child-span-001")
childLogger := log.FromContext(childCtx)
childLogger.Info().Msg("子操作")
```

## 与Web框架集成示例

### 集成到Gin中间件

```go
func LoggerMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 为每个请求生成唯一ID
        requestID := uuid.New().String()
        traceID := c.GetHeader("X-Trace-ID")
        if traceID == "" {
            traceID = requestID
        }
        spanID := uuid.New().String()

        // 将跟踪信息添加到上下文
        ctx := log.WithRequestID(c.Request.Context(), requestID)
        ctx = log.WithTraceID(ctx, traceID)
        ctx = log.WithSpanID(ctx, spanID)

        // 替换原始上下文
        c.Request = c.Request.WithContext(ctx)

        // 添加跟踪ID到响应头
        c.Header("X-Request-ID", requestID)
        c.Header("X-Trace-ID", traceID)

        // 记录请求开始
        logger := log.FromContext(ctx)
        logger.Info().
            Str("method", c.Request.Method).
            Str("path", c.Request.URL.Path).
            Str("client_ip", c.ClientIP()).
            Str("user_agent", c.Request.UserAgent()).
            Msg("开始处理请求")

        // 处理请求
        start := time.Now()
        c.Next()

        // 记录请求完成
        latency := time.Since(start)
        logger.Info().
            Int("status", c.Writer.Status()).
            Dur("latency", latency).
            Int("body_size", c.Writer.Size()).
            Msg("请求处理完成")
    }
}
```

## 详细文档

查看`example`目录中的完整示例来了解更多用法。 