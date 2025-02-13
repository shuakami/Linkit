package cache

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
)

// NewRedisClient 创建Redis客户端连接
func NewRedisClient() (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", viper.GetString("redis.host"), viper.GetInt("redis.port")),
		Password: viper.GetString("redis.password"),
		DB:       viper.GetInt("redis.db"),
	})

	// 测试连接
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return client, nil
}
