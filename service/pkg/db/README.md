# 数据库客户端封装库

本包提供了对PostgreSQL、Redis和Elasticsearch数据库客户端的封装，简化了连接管理、错误处理和日志记录。

## 特性

- 统一的数据库客户端接口
- 支持多种数据库类型：PostgreSQL、Redis、Elasticsearch
- 自动重试连接和错误恢复
- 连接池管理和配置
- 结构化日志集成
- 健康检查和监控

## 支持的数据库

### PostgreSQL (基于GORM)

- 自动迁移模型
- 事务支持
- 连接池管理
- 安全的连接字符串处理

### Redis

- 支持单节点、集群和哨兵模式
- 通用的键值操作API
- 连接池统计

### Elasticsearch

- 索引管理
- 文档CRUD操作
- 搜索功能
- 支持Cloud ID和API Key认证

## 快速开始

### PostgreSQL 示例

```go
// 配置PostgreSQL连接
pgConfig := postgres.Config{
    Host:     "localhost",
    Port:     5432,
    Username: "tip_user",
    Password: "tip_password",
    Database: "tip_db",
    SSLMode:  "disable",
}

// 初始化客户端
opts := db.DefaultOptions()
pgClient := postgres.New(pgConfig, opts)

// 连接到数据库
if err := pgClient.Connect(ctx); err != nil {
    log.Error().Err(err).Msg("PostgreSQL连接失败")
    return
}
defer pgClient.Close()

// 使用GORM
if err := pgClient.DB().Create(&User{...}).Error; err != nil {
    log.Error().Err(err).Msg("创建用户失败")
}
```

### Redis 示例

```go
// 配置Redis连接
redisConfig := redis.Config{
    Addr:     "localhost:6379",
    Password: "",
    DB:       0,
}

// 初始化客户端
opts := db.DefaultOptions()
redisClient := redis.New(redisConfig, opts)

// 连接到Redis
if err := redisClient.Connect(ctx); err != nil {
    log.Error().Err(err).Msg("Redis连接失败")
    return
}
defer redisClient.Close()

// 使用Redis
if err := redisClient.Set(ctx, "key", "value", time.Minute); err != nil {
    log.Error().Err(err).Msg("设置键值失败")
}
```

### Elasticsearch 示例

```go
// 配置Elasticsearch连接
esConfig := elasticsearch.Config{
    Addresses: []string{"http://localhost:9200"},
}

// 初始化客户端
opts := db.DefaultOptions()
esClient := elasticsearch.New(esConfig, opts)

// 连接到Elasticsearch
if err := esClient.Connect(ctx); err != nil {
    log.Error().Err(err).Msg("Elasticsearch连接失败")
    return
}
defer esClient.Close()

// 使用Elasticsearch
doc := map[string]interface{}{...}
if err := esClient.Index(ctx, "index", "id", doc); err != nil {
    log.Error().Err(err).Msg("索引文档失败")
}
```

## 配置选项

通过`db.Options`结构体可以配置数据库连接选项：

```go
opts := db.Options{
    MaxOpenConns:        100,      // 最大打开连接数
    MaxIdleConns:        10,       // 最大空闲连接数
    ConnMaxLifetime:     time.Hour, // 连接最大生存时间
    ConnMaxIdleTime:     time.Minute * 30, // 连接最大空闲时间
    Timeout:             time.Second * 10, // 连接超时时间
    HealthCheckInterval: time.Minute, // 健康检查间隔
    RetryAttempts:       3,        // 重试次数
    RetryDelay:          time.Second * 2, // 重试延迟
}
```

## 详细文档

查看`example`目录中的完整示例来了解更多用法。 