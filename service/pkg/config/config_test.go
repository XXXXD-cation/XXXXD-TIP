package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestConfigBasic(t *testing.T) {
	// 创建临时配置文件
	dir := t.TempDir()
	configFile := filepath.Join(dir, "config.yaml")
	configContent := `
db:
  host: localhost
  port: 5432
  username: test_user
  password: test_password
  database: test_db
app:
  name: test-app
  port: 8080
  debug: true
  features:
    - feature1
    - feature2
  timeout: 30s
`
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("无法创建测试配置文件: %v", err)
	}

	// 初始化配置
	opts := Options{
		ConfigName:  "config",
		ConfigType:  "yaml",
		ConfigPaths: []string{dir},
		EnvPrefix:   "TEST",
	}

	provider, err := New(opts)
	if err != nil {
		t.Fatalf("初始化配置失败: %v", err)
	}

	// 测试获取字符串值
	dbHost := provider.GetString("db.host")
	if dbHost != "localhost" {
		t.Errorf("GetString() = %v, 期望 %v", dbHost, "localhost")
	}

	// 测试获取整数值
	dbPort := provider.GetInt("db.port")
	if dbPort != 5432 {
		t.Errorf("GetInt() = %v, 期望 %v", dbPort, 5432)
	}

	// 测试获取布尔值
	appDebug := provider.GetBool("app.debug")
	if !appDebug {
		t.Errorf("GetBool() = %v, 期望 %v", appDebug, true)
	}

	// 测试获取字符串切片
	features := provider.GetStringSlice("app.features")
	if len(features) != 2 || features[0] != "feature1" || features[1] != "feature2" {
		t.Errorf("GetStringSlice() = %v, 期望 %v", features, []string{"feature1", "feature2"})
	}

	// 测试获取持续时间
	timeout := provider.GetDuration("app.timeout")
	if timeout != 30*time.Second {
		t.Errorf("GetDuration() = %v, 期望 %v", timeout, 30*time.Second)
	}

	// 测试GetOrDefault
	nonExistent := provider.GetOrDefault("non.existent", "default")
	if nonExistent != "default" {
		t.Errorf("GetOrDefault() = %v, 期望 %v", nonExistent, "default")
	}

	// 测试IsSet
	if !provider.IsSet("app.name") {
		t.Errorf("IsSet(app.name) = false, 期望 true")
	}
	if provider.IsSet("non.existent") {
		t.Errorf("IsSet(non.existent) = true, 期望 false")
	}

	// 测试Unmarshal
	type DBConfig struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
		Database string `mapstructure:"database"`
	}

	var dbConfig DBConfig
	err = provider.UnmarshalKey("db", &dbConfig)
	if err != nil {
		t.Errorf("UnmarshalKey失败: %v", err)
	}

	if dbConfig.Host != "localhost" || dbConfig.Port != 5432 ||
		dbConfig.Username != "test_user" || dbConfig.Password != "test_password" ||
		dbConfig.Database != "test_db" {
		t.Errorf("UnmarshalKey结果不正确: %+v", dbConfig)
	}
}

func TestEnvironmentVariables(t *testing.T) {
	// 设置环境变量
	os.Setenv("TEST_APP_NAME", "env-app")
	os.Setenv("TEST_DB_PORT", "6543")
	defer func() {
		os.Unsetenv("TEST_APP_NAME")
		os.Unsetenv("TEST_DB_PORT")
	}()

	// 初始化配置
	opts := Options{
		ConfigName:   "non-existent", // 使用不存在的配置文件名，强制使用环境变量
		EnvPrefix:    "TEST",
		AutomaticEnv: true,
	}

	provider, err := New(opts)
	if err != nil {
		t.Logf("初始化配置警告: %v", err)
	}

	// 测试从环境变量读取
	appName := provider.GetString("app.name")
	if appName != "env-app" {
		t.Errorf("从环境变量GetString() = %v, 期望 %v", appName, "env-app")
	}

	dbPort := provider.GetInt("db.port")
	if dbPort != 6543 {
		t.Errorf("从环境变量GetInt() = %v, 期望 %v", dbPort, 6543)
	}
}

func TestGlobalFunctions(t *testing.T) {
	// 先重置全局实例
	Reset()

	// 创建临时配置文件
	dir := t.TempDir()
	configFile := filepath.Join(dir, "config.yaml")
	configContent := `
key: value
number: 123
`
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("无法创建测试配置文件: %v", err)
	}

	// 初始化配置
	opts := Options{
		ConfigName:  "config",
		ConfigType:  "yaml",
		ConfigPaths: []string{dir},
	}

	provider, err := New(opts)
	if err != nil {
		t.Fatalf("初始化配置失败: %v", err)
	}

	// 设置全局实例
	SetInstance(provider)

	// 测试全局函数
	val := GetString("key")
	if val != "value" {
		t.Errorf("GetString() = %v, 期望 %v", val, "value")
	}

	num := GetInt("number")
	if num != 123 {
		t.Errorf("GetInt() = %v, 期望 %v", num, 123)
	}

	def := GetOrDefault("non.existent", "default")
	if def != "default" {
		t.Errorf("GetOrDefault() = %v, 期望 %v", def, "default")
	}
}
