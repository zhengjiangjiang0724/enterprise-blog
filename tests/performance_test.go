package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"enterprise-blog/internal/config"
	"enterprise-blog/internal/database"
	"enterprise-blog/pkg/logger"
)

// 性能测试辅助函数
var (
	perfDBReady    bool
	perfRedisReady bool
)

func init() {
	// 加载配置和日志
	_ = config.Load()
	_ = logger.Init("error", "")

	// 初始化数据库和 Redis，不强制失败，后面基准用例按可用性决定是否跳过
	if err := database.Init(); err == nil {
		perfDBReady = true
	}
	if err := database.InitRedis(); err == nil {
		perfRedisReady = true
	}
}

func BenchmarkDatabaseQuery(b *testing.B) {
	if !perfDBReady {
		b.Skip("Database not available")
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var count int
		if err := database.DB.WithContext(ctx).
			Raw("SELECT COUNT(*) FROM articles").
			Scan(&count).Error; err != nil {
			b.Fatalf("db query failed: %v", err)
		}
	}
}

func BenchmarkRedisSet(b *testing.B) {
	if !perfRedisReady {
		b.Skip("Redis not available")
	}

	ctx := context.Background()
	key := "benchmark:test"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		database.RedisClient.Set(ctx, fmt.Sprintf("%s:%d", key, i), "value", time.Hour)
	}
}

func BenchmarkRedisGet(b *testing.B) {
	if !perfRedisReady {
		b.Skip("Redis not available")
	}

	ctx := context.Background()
	key := "benchmark:test"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		database.RedisClient.Get(ctx, fmt.Sprintf("%s:%d", key, i%1000))
	}
}
