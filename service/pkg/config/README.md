# 配置管理库

本包提供了基于Viper的配置管理功能封装，支持从多种来源加载配置，并提供类型安全的访问方法。

## 特性

- 从多种来源加载配置：YAML/JSON/TOML文件、环境变量
- 支持配置热重载（可选）
- 提供类型安全的配置访问方法
- 支持嵌套配置结构
- 支持设置默认值
- 自动替换路径中的环境变量

## 使用示例

### 基本用法

```go
package main

import (
    "fmt"
    "github.com/XXXXD-cation/XXXXD-TIP/service/pkg/config"
)

func main() {
    // 使用默认选项初始化配置
    opts := config.DefaultOptions()
    
    // 也可以自定义选项
    opts = config.Options{
        ConfigName:  "config",
        ConfigType:  "yaml",
        ConfigPaths: []string{".", "./config", "$HOME/.config/app"},
        EnvPrefix:   "APP",
        WatchConfig: true,
    }
    
    // 创建配置提供器
    provider, err := config.New(opts)
    if err != nil {
        fmt.Printf("初始化配置警告: %v\n", err)
    }
    
    // 设置全局配置实例（可选）
    config.SetInstance(provider)
    
    // 获取配置值
    dbHost := config.GetString("database.host")
    dbPort := config.GetInt("database.port")
    
    fmt.Printf("数据库连接: %s:%d\n", dbHost, dbPort)
    
    // 使用默认值
    maxConns := config.GetOrDefault("database.max_connections", 100)
    fmt.Printf("最大连接数: %v\n", maxConns)
}
```

### 结构化配置

```go
package main

import (
    "fmt"
    "github.com/XXXXD-cation/XXXXD-TIP/service/pkg/config"
)

// 数据库配置结构体
type DatabaseConfig struct {
    Host     string `mapstructure:"host"`
    Port     int    `mapstructure:"port"`
    Username string `mapstructure:"username"`
    Password string `mapstructure:"password"`
    Database string `mapstructure:"database"`
}

// 应用配置结构体
type AppConfig struct {
    Server   struct {
        Port int `mapstructure:"port"`
    } `mapstructure:"server"`
    Database DatabaseConfig `mapstructure:"database"`
}

func main() {
    // 初始化配置（略）
    
    // 解析结构化配置
    var appConfig AppConfig
    if err := config.Unmarshal(&appConfig); err != nil {
        fmt.Printf("无法解析配置: %v\n", err)
        return
    }
    
    // 或者只解析配置的一部分
    var dbConfig DatabaseConfig
    if err := config.UnmarshalKey("database", &dbConfig); err != nil {
        fmt.Printf("无法解析数据库配置: %v\n", err)
        return
    }
    
    fmt.Printf("连接到数据库: %s@%s:%d/%s\n", 
        dbConfig.Username, dbConfig.Host, dbConfig.Port, dbConfig.Database)
}
```

### 环境变量支持

环境变量将自动转换为配置键。例如，对于前缀 `APP_`：

- 环境变量 `APP_SERVER_PORT=8080` 将映射到配置键 `server.port`
- 环境变量 `APP_DATABASE_HOST=localhost` 将映射到配置键 `database.host`

环境变量优先级高于配置文件。

## 配置文件示例

```yaml
# config.yaml
server:
  port: 8080
  host: localhost
  timeout: 30

database:
  driver: postgres
  host: localhost
  port: 5432
  username: app_user
  password: app_password
  database: app_db
```

## 详细文档

查看`example`目录中的完整示例来了解更多用法。 