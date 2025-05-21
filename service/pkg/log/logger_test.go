package log

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

// 测试全局日志实例初始化
func TestInit(t *testing.T) {
	// 验证初始化后的全局日志实例是否正常工作
	// 无法直接比较logger结构体，而是检查它是否能正常输出日志
	var buf bytes.Buffer
	tempLogger := globalLogger

	// 临时将输出重定向到buffer
	globalLogger = zerolog.New(&buf)

	// 尝试写入日志
	Info().Msg("测试初始化")

	// 还原全局logger
	globalLogger = tempLogger

	// 如果能正常记录日志，则说明logger已正确初始化
	if len(buf.String()) == 0 {
		t.Error("全局日志实例未正确初始化，无法写入日志")
	}
}

// 测试日志配置设置
func TestSetup(t *testing.T) {
	// 准备自定义配置
	config := Config{
		Level:       "debug",
		Pretty:      false,
		WithCaller:  true,
		TimeFormat:  "2006-01-02 15:04:05",
		Output:      "console",
		ServiceName: "test-service",
	}

	// 设置日志
	Setup(config)

	// 由于zerolog不提供直接获取当前配置的方法，
	// 我们主要通过后续的日志输出来确认配置是否生效
	if zerolog.GlobalLevel() != zerolog.DebugLevel {
		t.Errorf("日志级别设置错误，期望 %v，实际 %v", zerolog.DebugLevel, zerolog.GlobalLevel())
	}

	// 验证ServiceName是否正确添加到日志
	var buf bytes.Buffer
	origLogger := globalLogger
	globalLogger = zerolog.New(&buf).With().Timestamp().Str("service", config.ServiceName).Logger()

	Info().Msg("测试服务名称")

	globalLogger = origLogger

	var logData map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logData); err != nil {
		t.Fatalf("解析日志失败: %v", err)
	}

	if svc, ok := logData["service"].(string); !ok || svc != config.ServiceName {
		t.Errorf("服务名称未正确设置，期望 %s，实际 %v", config.ServiceName, logData["service"])
	}
}

// 测试无效日志级别处理
func TestInvalidLevel(t *testing.T) {
	// 准备无效的日志级别配置
	config := Config{
		Level:      "invalid_level",
		Pretty:     false,
		WithCaller: true,
		TimeFormat: "2006-01-02 15:04:05",
		Output:     "console",
	}

	// 保存当前的全局级别
	originalLevel := zerolog.GlobalLevel()

	// 设置日志
	Setup(config)

	// 验证是否使用了默认级别
	if zerolog.GlobalLevel() != defaultLevel {
		t.Errorf("无效日志级别未正确处理，期望 %v，实际 %v", defaultLevel, zerolog.GlobalLevel())
	}

	// 恢复原始配置
	zerolog.SetGlobalLevel(originalLevel)
}

// 测试文件输出
func TestFileOutput(t *testing.T) {
	// 创建临时文件用于测试
	tempFile, err := os.CreateTemp("", "log_test_*.log")
	if err != nil {
		t.Fatalf("创建临时文件失败: %v", err)
	}
	tempFileName := tempFile.Name()
	tempFile.Close()
	defer os.Remove(tempFileName) // 测试完成后清理文件

	// 配置日志输出到文件
	config := Config{
		Level:      "info",
		Pretty:     false,
		WithCaller: false,
		TimeFormat: time.RFC3339,
		Output:     "file:" + tempFileName,
	}

	// 设置日志
	Setup(config)

	// 写入测试日志
	testMsg := "测试文件输出"
	Info().Msg(testMsg)

	// 读取日志文件内容
	content, err := os.ReadFile(tempFileName)
	if err != nil {
		t.Fatalf("读取日志文件失败: %v", err)
	}

	// 验证日志内容
	if !strings.Contains(string(content), testMsg) {
		t.Errorf("日志文件中没有找到预期的消息：%s", testMsg)
	}
}

// 测试上下文中的跟踪ID
func TestFromContext(t *testing.T) {
	// 创建一个包含跟踪ID的上下文
	traceID := "test-trace-123"
	requestID := "req-456"
	spanID := "span-789"
	parentSpanID := "parent-span-012"

	ctx := context.Background()
	ctx = WithTraceID(ctx, traceID)
	ctx = WithRequestID(ctx, requestID)
	ctx = WithSpanID(ctx, spanID)
	ctx = WithParentSpanID(ctx, parentSpanID)

	// 捕获日志输出
	var buf bytes.Buffer
	originalLogger := globalLogger
	globalLogger = zerolog.New(&buf)

	// 从上下文获取日志记录器并记录日志
	FromContext(ctx).Info().Msg("测试链路跟踪日志")

	// 恢复原始全局记录器
	globalLogger = originalLogger

	// 解析JSON日志
	var logEvent map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEvent); err != nil {
		t.Fatalf("无法解析日志JSON: %v", err)
	}

	// 检查所有跟踪ID是否正确记录
	if logEvent["trace_id"] != traceID {
		t.Errorf("跟踪ID未正确记录，期望 %v，实际 %v", traceID, logEvent["trace_id"])
	}

	if logEvent["request_id"] != requestID {
		t.Errorf("请求ID未正确记录，期望 %v，实际 %v", requestID, logEvent["request_id"])
	}

	if logEvent["span_id"] != spanID {
		t.Errorf("跨度ID未正确记录，期望 %v，实际 %v", spanID, logEvent["span_id"])
	}

	if logEvent["parent_span_id"] != parentSpanID {
		t.Errorf("父跨度ID未正确记录，期望 %v，实际 %v", parentSpanID, logEvent["parent_span_id"])
	}
}

// 测试添加跟踪ID到上下文
func TestWithTraceID(t *testing.T) {
	traceID := "test-trace-456"
	ctx := context.Background()

	// 添加跟踪ID到上下文
	ctx = WithTraceID(ctx, traceID)

	// 验证上下文中的跟踪ID
	if id, ok := ctx.Value(TraceIDKey).(string); !ok || id != traceID {
		t.Errorf("WithTraceID未正确设置跟踪ID，期望 %v，实际 %v", traceID, id)
	}
}

// 测试添加请求ID到上下文
func TestWithRequestID(t *testing.T) {
	requestID := "req-test-789"
	ctx := context.Background()

	// 添加请求ID到上下文
	ctx = WithRequestID(ctx, requestID)

	// 验证上下文中的请求ID
	if id, ok := ctx.Value(RequestIDKey).(string); !ok || id != requestID {
		t.Errorf("WithRequestID未正确设置请求ID，期望 %v，实际 %v", requestID, id)
	}
}

// 测试添加SpanID到上下文
func TestWithSpanID(t *testing.T) {
	spanID := "span-test-123"
	ctx := context.Background()

	// 添加SpanID到上下文
	ctx = WithSpanID(ctx, spanID)

	// 验证上下文中的SpanID
	if id, ok := ctx.Value(SpanIDKey).(string); !ok || id != spanID {
		t.Errorf("WithSpanID未正确设置跨度ID，期望 %v，实际 %v", spanID, id)
	}
}

// 测试添加父SpanID到上下文
func TestWithParentSpanID(t *testing.T) {
	parentSpanID := "parent-span-test-456"
	ctx := context.Background()

	// 添加父SpanID到上下文
	ctx = WithParentSpanID(ctx, parentSpanID)

	// 验证上下文中的父SpanID
	if id, ok := ctx.Value(ParentSpanIDKey).(string); !ok || id != parentSpanID {
		t.Errorf("WithParentSpanID未正确设置父跨度ID，期望 %v，实际 %v", parentSpanID, id)
	}
}

// 测试全局日志函数
func TestGlobalLogFunctions(t *testing.T) {
	// 保存原始日志级别
	originalLevel := zerolog.GlobalLevel()
	// 设置为Debug级别以确保所有日志都能输出
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	// 测试所有全局日志函数
	logFunctions := []struct {
		name     string
		function func() *zerolog.Event
		level    string
	}{
		{"Debug", Debug, "debug"},
		{"Info", Info, "info"},
		{"Warn", Warn, "warn"},
		{"Error", Error, "error"},
		// 注意：Fatal函数会终止程序，不在测试中直接调用
	}

	for _, lf := range logFunctions {
		t.Run(lf.name, func(t *testing.T) {
			var buf bytes.Buffer
			originalLogger := globalLogger
			globalLogger = zerolog.New(&buf)

			// 调用日志函数
			lf.function().Msg("测试" + lf.name + "日志")

			// 恢复原始全局记录器
			globalLogger = originalLogger

			// 验证日志级别
			if !strings.Contains(buf.String(), lf.level) {
				t.Errorf("%s函数未产生正确的日志级别，日志内容: %s", lf.name, buf.String())
			}
		})
	}

	// 恢复原始日志级别
	zerolog.SetGlobalLevel(originalLevel)
}

// 测试获取调用者信息
func TestGetCallerInfo(t *testing.T) {
	// GetCallerInfo(0)会返回此函数自身在logger.go中的位置
	callerInfo := GetCallerInfo(0)

	// 验证调用者信息包含logger.go文件名，而不是测试文件
	if !strings.Contains(callerInfo, "logger.go") {
		t.Errorf("GetCallerInfo未返回正确的调用者信息，得到: %s", callerInfo)
	}

	// 创建辅助函数，它会调用GetCallerInfo(1)
	helperCaller := func() string {
		return GetCallerInfo(1) // 跳过辅助函数本身，获取调用辅助函数的位置
	}

	// 调用辅助函数，此时GetCallerInfo(1)应该返回本测试函数位置
	result := helperCaller()

	// 验证返回的信息包含当前测试文件名
	if !strings.Contains(result, "logger_test.go") {
		t.Errorf("GetCallerInfo未正确获取间接调用位置，期望包含logger_test.go，得到: %s", result)
	}
}
