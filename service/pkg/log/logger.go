// Package log 提供了应用程序的日志功能封装，基于zerolog实现
// 支持结构化日志输出、日志级别控制和链路追踪
package log

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"time"

	"github.com/rs/zerolog"
)

// contextKey 是用于context的键类型，避免与其他包的键冲突
type contextKey string

// 上下文中存储的键
const (
	// TraceIDKey 是上下文中存储跟踪ID的键
	TraceIDKey = contextKey("trace_id")
	// RequestIDKey 是上下文中存储请求ID的键
	RequestIDKey = contextKey("request_id")
	// SpanIDKey 是上下文中存储跨度ID的键
	SpanIDKey = contextKey("span_id")
	// ParentSpanIDKey 是上下文中存储父跨度ID的键
	ParentSpanIDKey = contextKey("parent_span_id")
)

// Config 日志配置
type Config struct {
	// Level 日志级别: debug, info, warn, error, fatal
	Level string `yaml:"level" json:"level"`
	// Pretty 是否启用美化输出（适合开发环境）
	Pretty bool `yaml:"pretty" json:"pretty"`
	// WithCaller 是否记录调用者信息
	WithCaller bool `yaml:"with_caller" json:"with_caller"`
	// TimeFormat 时间格式
	TimeFormat string `yaml:"time_format" json:"time_format"`
	// Output 输出位置 (可以是"console", "file:/path/to/file", 或其他io.Writer)
	Output string `yaml:"output" json:"output"`
	// ServiceName 服务名称，便于在日志中标识来源服务
	ServiceName string `yaml:"service_name" json:"service_name"`
}

// 全局日志实例
var (
	globalLogger zerolog.Logger
	defaultLevel = zerolog.InfoLevel
)

// 初始化默认日志配置
func init() {
	// 默认配置
	config := Config{
		Level:       "info",
		Pretty:      false,
		WithCaller:  true,
		TimeFormat:  time.RFC3339,
		Output:      "console",
		ServiceName: "unknown-service",
	}

	// 初始化全局日志实例
	Setup(config)
}

// Setup 根据配置初始化日志
func Setup(config Config) {
	// 设置日志级别
	level, err := zerolog.ParseLevel(config.Level)
	if err != nil {
		level = defaultLevel
	}
	zerolog.SetGlobalLevel(level)

	// 配置输出
	var output io.Writer = os.Stdout
	if config.Output != "console" && config.Output != "" {
		// 如果是文件路径，打开文件
		if len(config.Output) > 5 && config.Output[:5] == "file:" {
			filePath := config.Output[5:]
			file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err == nil {
				output = file
			} else {
				// 如果打开文件失败，记录错误并使用标准输出
				fmt.Fprintf(os.Stderr, "无法打开日志文件 %s: %v, 使用标准输出\n", filePath, err)
			}
		}
	}

	// 配置美化输出
	if config.Pretty {
		output = zerolog.ConsoleWriter{
			Out:        output,
			TimeFormat: config.TimeFormat,
		}
	}

	// 创建日志实例
	logger := zerolog.New(output).With().Timestamp()

	// 添加服务名称
	if config.ServiceName != "" {
		logger = logger.Str("service", config.ServiceName)
	}

	// 添加调用者信息
	if config.WithCaller {
		logger = logger.Caller()
	}

	// 设置全局实例
	globalLogger = logger.Logger()
}

// FromContext 从上下文中获取或创建一个带链路跟踪信息的日志记录器
func FromContext(ctx context.Context) *zerolog.Logger {
	// 基础日志实例
	logger := globalLogger

	// 尝试从上下文获取跟踪信息
	if ctx != nil {
		// 添加跟踪ID
		if traceID, ok := ctx.Value(TraceIDKey).(string); ok && traceID != "" {
			logger = logger.With().Str("trace_id", traceID).Logger()
		}

		// 添加请求ID
		if requestID, ok := ctx.Value(RequestIDKey).(string); ok && requestID != "" {
			logger = logger.With().Str("request_id", requestID).Logger()
		}

		// 添加跨度ID
		if spanID, ok := ctx.Value(SpanIDKey).(string); ok && spanID != "" {
			logger = logger.With().Str("span_id", spanID).Logger()
		}

		// 添加父跨度ID
		if parentSpanID, ok := ctx.Value(ParentSpanIDKey).(string); ok && parentSpanID != "" {
			logger = logger.With().Str("parent_span_id", parentSpanID).Logger()
		}
	}

	return &logger
}

// WithTraceID 向上下文添加跟踪ID并返回新的上下文
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

// WithRequestID 向上下文添加请求ID并返回新的上下文
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// WithSpanID 向上下文添加跨度ID并返回新的上下文
func WithSpanID(ctx context.Context, spanID string) context.Context {
	return context.WithValue(ctx, SpanIDKey, spanID)
}

// WithParentSpanID 向上下文添加父跨度ID并返回新的上下文
func WithParentSpanID(ctx context.Context, parentSpanID string) context.Context {
	return context.WithValue(ctx, ParentSpanIDKey, parentSpanID)
}

// 下面是一系列全局日志函数，用于方便直接调用

// Debug 记录调试级别日志
func Debug() *zerolog.Event {
	return globalLogger.Debug()
}

// Info 记录信息级别日志
func Info() *zerolog.Event {
	return globalLogger.Info()
}

// Warn 记录警告级别日志
func Warn() *zerolog.Event {
	return globalLogger.Warn()
}

// Error 记录错误级别日志
func Error() *zerolog.Event {
	return globalLogger.Error()
}

// Fatal 记录致命错误并退出程序
func Fatal() *zerolog.Event {
	return globalLogger.Fatal()
}

// GetCallerInfo 获取调用者信息
func GetCallerInfo(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown"
	}
	return fmt.Sprintf("%s:%d", file, line)
}
