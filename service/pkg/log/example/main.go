package main

import (
	"context"
	"fmt"
	"time"

	"github.com/XXXXD-cation/XXXXD-TIP/service/pkg/log"
	"github.com/google/uuid"
)

// 模拟一个HTTP请求处理函数
func handleRequest(ctx context.Context, endpoint string) {
	// 从上下文中获取日志记录器
	logger := log.FromContext(ctx)

	// 记录请求开始
	logger.Info().
		Str("endpoint", endpoint).
		Str("method", "GET").
		Msg("开始处理请求")

	// 模拟处理时间
	time.Sleep(100 * time.Millisecond)

	// 模拟调用其他服务
	callExternalService(ctx, "user-service")

	// 记录请求完成
	logger.Info().
		Str("endpoint", endpoint).
		Int("status", 200).
		Dur("duration", 150*time.Millisecond).
		Msg("请求处理完成")
}

// 模拟调用外部服务
func callExternalService(ctx context.Context, serviceName string) {
	// 从父上下文创建子Span上下文
	childSpanID := uuid.New().String()
	parentSpanID, _ := ctx.Value(log.SpanIDKey).(string)
	
	childCtx := ctx
	if parentSpanID != "" {
		childCtx = log.WithParentSpanID(ctx, parentSpanID)
	}
	childCtx = log.WithSpanID(childCtx, childSpanID)
	
	// 从子上下文获取日志记录器
	logger := log.FromContext(childCtx)
	
	// 记录服务调用
	logger.Info().
		Str("external_service", serviceName).
		Msg("调用外部服务")
	
	// 模拟处理时间
	time.Sleep(50 * time.Millisecond)
	
	// 记录服务调用完成
	logger.Info().
		Str("external_service", serviceName).
		Dur("duration", 50*time.Millisecond).
		Msg("外部服务调用完成")
}

func main() {
	// 配置日志
	config := log.Config{
		Level:       "debug",
		Pretty:      true, // 在开发环境使用美化输出
		WithCaller:  true,
		TimeFormat:  "2006-01-02 15:04:05",
		Output:      "console",
		ServiceName: "example-service",
	}
	log.Setup(config)

	// 输出一些日志示例
	log.Info().
		Str("app", "example").
		Int("version", 1).
		Msg("应用启动")

	log.Debug().
		Str("config", fmt.Sprintf("%+v", config)).
		Msg("加载配置")

	// 模拟处理HTTP请求
	// 创建带有跟踪信息的上下文
	traceID := uuid.New().String()
	requestID := uuid.New().String()
	spanID := uuid.New().String()
	
	ctx := context.Background()
	ctx = log.WithTraceID(ctx, traceID)
	ctx = log.WithRequestID(ctx, requestID)
	ctx = log.WithSpanID(ctx, spanID)
	
	// 处理请求
	handleRequest(ctx, "/api/v1/users")
	
	// 模拟错误日志
	log.Error().
		Str("module", "database").
		Err(fmt.Errorf("connection timeout")).
		Msg("数据库连接失败")
	
	log.Info().
		Int("processed_requests", 1).
		Msg("应用正常退出")
} 