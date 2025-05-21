package db

import (
	"testing"
	"time"
)

// 测试默认选项
func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	if opts.MaxOpenConns != 100 {
		t.Errorf("默认MaxOpenConns = %v, 期望 %v", opts.MaxOpenConns, 100)
	}

	if opts.MaxIdleConns != 10 {
		t.Errorf("默认MaxIdleConns = %v, 期望 %v", opts.MaxIdleConns, 10)
	}

	if opts.ConnMaxLifetime != time.Hour {
		t.Errorf("默认ConnMaxLifetime = %v, 期望 %v", opts.ConnMaxLifetime, time.Hour)
	}

	if opts.ConnMaxIdleTime != time.Minute*30 {
		t.Errorf("默认ConnMaxIdleTime = %v, 期望 %v", opts.ConnMaxIdleTime, time.Minute*30)
	}

	if opts.Timeout != time.Second*10 {
		t.Errorf("默认Timeout = %v, 期望 %v", opts.Timeout, time.Second*10)
	}

	if opts.HealthCheckInterval != time.Minute {
		t.Errorf("默认HealthCheckInterval = %v, 期望 %v", opts.HealthCheckInterval, time.Minute)
	}

	if opts.RetryAttempts != 3 {
		t.Errorf("默认RetryAttempts = %v, 期望 %v", opts.RetryAttempts, 3)
	}

	if opts.RetryDelay != time.Second*2 {
		t.Errorf("默认RetryDelay = %v, 期望 %v", opts.RetryDelay, time.Second*2)
	}
}

// 空接口实现测试，确保ConnectionInfo和Client接口能够被不同的实现使用
type testConnectionInfo struct{}

func (t testConnectionInfo) DSN() string {
	return "test-dsn"
}

func (t testConnectionInfo) String() string {
	return "test-connection-string"
}

// 测试ConnectionInfo接口
func TestConnectionInfoInterface(t *testing.T) {
	var info ConnectionInfo = testConnectionInfo{}

	if info.DSN() != "test-dsn" {
		t.Errorf("DSN() = %v, 期望 %v", info.DSN(), "test-dsn")
	}

	if info.String() != "test-connection-string" {
		t.Errorf("String() = %v, 期望 %v", info.String(), "test-connection-string")
	}
} 