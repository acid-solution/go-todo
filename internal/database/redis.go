package database

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// 初始化 Redis 客户端
func InitRedis(
	addr string,
	password string,
	db int,
) *redis.Client {

	// 创建 Redis 客户端
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	// 本函数最多等待 5 秒钟来尝试连接 Redis，如果超过这个时间还没有连接成功，就会返回错误。
	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	// Ping Redis 服务器，检查连接是否成功
	//接收参数ctx，用来控制操作何时停止
	if err := client.Ping(ctx).Err(); err != nil {
		// 关闭 Redis 客户端连接
		_ = client.Close()
		// 记录错误日志并终止程序
		log.Fatal("连接 Redis 失败:", err)
	}
	log.Println("Redis 连接成功")

	return client
}
