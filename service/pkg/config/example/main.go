package main

import (
	"fmt"
	"os"

	"github.com/XXXXD-cation/XXXXD-TIP/service/pkg/config"
	"github.com/XXXXD-cation/XXXXD-TIP/service/pkg/log"
)

// AppConfig 应用配置结构体
type AppConfig struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Logging  LogConfig      `mapstructure:"logging"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port    int    `mapstructure:"port"`
	Host    string `mapstructure:"host"`
	Timeout int    `mapstructure:"timeout"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver   string `mapstructure:"driver"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `mapstructure:"level"`
	Output     string `mapstructure:"output"`
	Pretty     bool   `mapstructure:"pretty"`
	WithCaller bool   `mapstructure:"with_caller"`
}

func main() {
	// 初始化配置
	configureApplication()

	// 使用结构化配置
	var appConfig AppConfig
	if err := config.Unmarshal(&appConfig); err != nil {
		fmt.Printf("无法解析配置: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	log.Setup(log.Config{
		Level:       appConfig.Logging.Level,
		Pretty:      appConfig.Logging.Pretty,
		WithCaller:  appConfig.Logging.WithCaller,
		Output:      appConfig.Logging.Output,
		ServiceName: "config-example",
	})

	// 输出配置信息
	log.Info().
		Int("server_port", appConfig.Server.Port).
		Str("server_host", appConfig.Server.Host).
		Int("server_timeout", appConfig.Server.Timeout).
		Msg("服务器配置")

	log.Info().
		Str("db_type", appConfig.Database.Driver).
		Str("db_host", appConfig.Database.Host).
		Int("db_port", appConfig.Database.Port).
		Str("db_name", appConfig.Database.Database).
		Msg("数据库配置")

	// 使用单个配置值
	serverPort := config.GetInt("server.port")
	log.Info().Int("port", serverPort).Msg("从配置中获取单个值")

	// 使用默认值
	maxConnections := config.GetOrDefault("server.max_connections", 100)
	log.Info().Interface("max_connections", maxConnections).Msg("使用默认值")

	log.Info().Msg("配置示例程序运行完成")
}

func configureApplication() {
	// 首先检查工作目录中是否存在配置文件
	if _, err := os.Stat("config.yaml"); os.IsNotExist(err) {
		// 没有找到配置文件，创建示例配置
		createExampleConfig()
	}

	// 配置选项
	opts := config.Options{
		ConfigName:  "config",
		ConfigType:  "yaml",
		ConfigPaths: []string{".", "./config", "$HOME/.config/app"},
		EnvPrefix:   "APP",
		WatchConfig: true,
		WatchConfigCallback: func() {
			fmt.Println("检测到配置文件变更，已重新加载")
		},
	}

	// 初始化配置
	provider, err := config.New(opts)
	if err != nil {
		fmt.Printf("配置初始化警告: %v\n", err)
	}

	// 设置全局配置实例
	config.SetInstance(provider)
}

func createExampleConfig() {
	configYaml := `# 应用配置示例
server:
  port: 8080
  host: localhost
  timeout: 30

database:
  driver: postgres
  host: localhost
  port: 5432
  username: tip_user
  password: tip_password
  database: tip_db

logging:
  level: info
  output: console
  pretty: true
  with_caller: true
`

	err := os.WriteFile("config.yaml", []byte(configYaml), 0644)
	if err != nil {
		fmt.Printf("无法创建示例配置文件: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("已创建示例配置文件: config.yaml")
}
