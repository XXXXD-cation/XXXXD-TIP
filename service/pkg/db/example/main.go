package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/XXXXD-cation/XXXXD-TIP/service/pkg/db"
	"github.com/XXXXD-cation/XXXXD-TIP/service/pkg/db/elasticsearch"
	"github.com/XXXXD-cation/XXXXD-TIP/service/pkg/db/postgres"
	"github.com/XXXXD-cation/XXXXD-TIP/service/pkg/db/redis"
	"github.com/XXXXD-cation/XXXXD-TIP/service/pkg/log"
)

// User 示例用户模型
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"size:50;not null" json:"username"`
	Email     string    `gorm:"size:100;not null" json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func main() {
	// 初始化日志
	log.Setup(log.Config{
		Level:       "debug",
		Pretty:      true,
		WithCaller:  true,
		Output:      "console",
		ServiceName: "db-example",
	})

	// 创建上下文，支持信号取消
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 监听终止信号
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChan
		log.Info().Msg("接收到终止信号，正在关闭连接...")
		cancel()
	}()

	// 示例：PostgreSQL
	if err := postgresExample(ctx); err != nil {
		log.Error().Err(err).Msg("PostgreSQL示例失败")
	}

	// 示例：Redis
	if err := redisExample(ctx); err != nil {
		log.Error().Err(err).Msg("Redis示例失败")
	}

	// 示例：Elasticsearch
	if err := elasticsearchExample(ctx); err != nil {
		log.Error().Err(err).Msg("Elasticsearch示例失败")
	}

	log.Info().Msg("所有示例完成")
}

// postgresExample 展示PostgreSQL客户端用法
func postgresExample(ctx context.Context) error {
	log.Info().Msg("===== PostgreSQL示例 =====")

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
	opts.Timeout = 5 * time.Second
	pgClient := postgres.New(pgConfig, opts)

	// 连接到数据库
	log.Info().Msg("连接到PostgreSQL...")
	if err := pgClient.Connect(ctx); err != nil {
		log.Warn().Err(err).Msg("PostgreSQL连接失败，跳过示例")
		return nil // 如果无法连接，跳过但不报错
	}
	defer pgClient.Close()

	// 自动迁移模型
	log.Info().Msg("迁移User模型...")

	// 先尝试删除表以避免迁移冲突
	if pgClient.DB().Migrator().HasTable(&User{}) {
		log.Info().Msg("表已存在，先删除...")
		if err := pgClient.DB().Migrator().DropTable(&User{}); err != nil {
			return fmt.Errorf("删除表失败: %w", err)
		}
	}

	if err := pgClient.AutoMigrate(&User{}); err != nil {
		return fmt.Errorf("迁移模型失败: %w", err)
	}

	// 创建用户
	log.Info().Msg("创建用户...")
	newUser := User{
		Username: "testuser",
		Email:    "test@example.com",
	}
	if err := pgClient.DB().Create(&newUser).Error; err != nil {
		return fmt.Errorf("创建用户失败: %w", err)
	}
	log.Info().Uint("id", newUser.ID).Str("username", newUser.Username).Msg("用户已创建")

	// 查询用户
	log.Info().Msg("查询用户...")
	var users []User
	if err := pgClient.DB().Find(&users).Error; err != nil {
		return fmt.Errorf("查询用户失败: %w", err)
	}
	log.Info().Int("count", len(users)).Msg("查询到的用户数")

	// 打印连接池统计信息
	stats := pgClient.Stats()
	log.Info().Interface("stats", stats).Msg("PostgreSQL连接池统计")

	log.Info().Msg("PostgreSQL示例完成")
	return nil
}

// redisExample 展示Redis客户端用法
func redisExample(ctx context.Context) error {
	log.Info().Msg("===== Redis示例 =====")

	// 配置Redis连接
	redisConfig := redis.Config{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}

	// 初始化客户端
	opts := db.DefaultOptions()
	opts.Timeout = 5 * time.Second
	redisClient := redis.New(redisConfig, opts)

	// 连接到Redis
	log.Info().Msg("连接到Redis...")
	if err := redisClient.Connect(ctx); err != nil {
		log.Warn().Err(err).Msg("Redis连接失败，跳过示例")
		return nil // 如果无法连接，跳过但不报错
	}
	defer redisClient.Close()

	// 设置键值
	key := "test:key"
	value := "Hello Redis"
	expiration := time.Minute * 5

	log.Info().Str("key", key).Str("value", value).Msg("设置键值...")
	if err := redisClient.Set(ctx, key, value, expiration); err != nil {
		return fmt.Errorf("设置键值失败: %w", err)
	}

	// 获取键值
	log.Info().Str("key", key).Msg("获取键值...")
	result, err := redisClient.Get(ctx, key)
	if err != nil {
		return fmt.Errorf("获取键值失败: %w", err)
	}
	log.Info().Str("key", key).Str("value", result).Msg("获取的键值")

	// 检查键是否存在
	log.Info().Str("key", key).Msg("检查键是否存在...")
	exists, err := redisClient.Exists(ctx, key)
	if err != nil {
		return fmt.Errorf("检查键是否存在失败: %w", err)
	}
	log.Info().Str("key", key).Bool("exists", exists > 0).Msg("键存在状态")

	// 删除键
	log.Info().Str("key", key).Msg("删除键...")
	if err := redisClient.Delete(ctx, key); err != nil {
		return fmt.Errorf("删除键失败: %w", err)
	}

	// 打印连接池统计信息
	stats := redisClient.Stats()
	log.Info().Interface("stats", stats).Msg("Redis连接池统计")

	log.Info().Msg("Redis示例完成")
	return nil
}

// elasticsearchExample 展示Elasticsearch客户端用法
func elasticsearchExample(ctx context.Context) error {
	log.Info().Msg("===== Elasticsearch示例 =====")

	// 配置Elasticsearch连接
	esConfig := elasticsearch.Config{
		Addresses: []string{"http://localhost:9200"},
	}

	// 初始化客户端
	opts := db.DefaultOptions()
	opts.Timeout = 5 * time.Second
	esClient := elasticsearch.New(esConfig, opts)

	// 连接到Elasticsearch
	log.Info().Msg("连接到Elasticsearch...")
	if err := esClient.Connect(ctx); err != nil {
		log.Warn().Err(err).Msg("Elasticsearch连接失败，跳过示例")
		return nil // 如果无法连接，跳过但不报错
	}
	defer esClient.Close()

	// 索引名称
	indexName := "test-index"

	// 检查索引是否存在
	log.Info().Str("index", indexName).Msg("检查索引是否存在...")
	exists, err := esClient.IndexExists(ctx, indexName)
	if err != nil {
		return fmt.Errorf("检查索引是否存在失败: %w", err)
	}

	// 如果索引存在，删除它
	if exists {
		log.Info().Str("index", indexName).Msg("索引已存在，正在删除...")
		if err := esClient.DeleteIndex(ctx, indexName); err != nil {
			return fmt.Errorf("删除索引失败: %w", err)
		}
	}

	// 创建索引
	log.Info().Str("index", indexName).Msg("创建索引...")
	indexMapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"id":         map[string]string{"type": "keyword"},
				"username":   map[string]string{"type": "keyword"},
				"email":      map[string]string{"type": "keyword"},
				"created_at": map[string]string{"type": "date"},
				"updated_at": map[string]string{"type": "date"},
			},
		},
	}

	if err := esClient.CreateIndex(ctx, indexName, indexMapping); err != nil {
		return fmt.Errorf("创建索引失败: %w", err)
	}

	// 索引文档
	docID := "1"
	user := User{
		ID:        1,
		Username:  "esuser",
		Email:     "es@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	log.Info().Str("index", indexName).Str("id", docID).Msg("索引文档...")
	if err := esClient.Index(ctx, indexName, docID, user); err != nil {
		return fmt.Errorf("索引文档失败: %w", err)
	}

	// 搜索文档
	log.Info().Str("index", indexName).Msg("搜索文档...")
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"username": "esuser",
			},
		},
	}

	results, err := esClient.Search(ctx, indexName, query)
	if err != nil {
		return fmt.Errorf("搜索文档失败: %w", err)
	}

	hits, ok := results["hits"].(map[string]interface{})
	if ok {
		total, _ := hits["total"].(map[string]interface{})
		value, _ := total["value"].(float64)
		log.Info().Int("count", int(value)).Msg("搜索结果数")
	}

	// 获取文档
	log.Info().Str("index", indexName).Str("id", docID).Msg("获取文档...")
	doc, err := esClient.Get(ctx, indexName, docID)
	if err != nil {
		return fmt.Errorf("获取文档失败: %w", err)
	}
	log.Info().Interface("doc", doc).Msg("获取的文档")

	// 删除文档
	log.Info().Str("index", indexName).Str("id", docID).Msg("删除文档...")
	if err := esClient.Delete(ctx, indexName, docID); err != nil {
		return fmt.Errorf("删除文档失败: %w", err)
	}

	// 删除索引
	log.Info().Str("index", indexName).Msg("删除索引...")
	if err := esClient.DeleteIndex(ctx, indexName); err != nil {
		return fmt.Errorf("删除索引失败: %w", err)
	}

	log.Info().Msg("Elasticsearch示例完成")
	return nil
}
