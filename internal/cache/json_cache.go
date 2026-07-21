package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// JSONCache 封装使用 JSON 格式读写 Redis 缓存的能力。
type JSONCache struct {
	client *redis.Client
}

// NewJSONCache 创建 JSONCache，并复用应用启动时创建的 Redis 客户端。
func NewJSONCache(client *redis.Client) *JSONCache {
	return &JSONCache{
		client: client,
	}
}

// marshalJSON 将 Go 值序列化为 JSON 字节。
func marshalJSON(value any) ([]byte, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("序列化缓存数据失败: %w", err)
	}

	return data, nil
}

// unmarshalJSON 将 JSON 字节反序列化到 destination 指向的目标中。
func unmarshalJSON(data []byte, destination any) error {
	if err := json.Unmarshal(data, destination); err != nil {
		return fmt.Errorf("反序列化缓存数据失败: %w", err)
	}

	return nil
}

// Set 将 Go 值序列化为 JSON，并写入 Redis。
func (c *JSONCache) Set(
	ctx context.Context, //
	key string, //键
	value any, //值
	ttl time.Duration, //过期时间
) error {
	data, err := marshalJSON(value)
	if err != nil {
		return err
	}

	if err := c.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("写入缓存失败: %w", err)
	}

	return nil
}

// Get 从 Redis 读取 JSON，并反序列化到 destination。
func (c *JSONCache) Get(
	ctx context.Context,
	key string,
	destination any,
) (bool, error) { //键是否存在，是否有错误
	data, err := c.client.Get(ctx, key).Bytes() //顺手转成字符数组
	//错误是键不存在
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	//其他错误
	if err != nil {
		return false, fmt.Errorf("读取缓存失败: %w", err)
	}
	//反序列化并检测错误
	if err := unmarshalJSON(data, destination); err != nil {
		return false, err
	}

	return true, nil
}

// Delete 删除一个或多个缓存 key，并返回实际删除的数量。
func (c *JSONCache) Delete(
	ctx context.Context,
	keys ...string,
) (int64, error) {
	if len(keys) == 0 {
		return 0, nil
	}

	deleted, err := c.client.Del(ctx, keys...).Result()
	if err != nil {
		return 0, fmt.Errorf("删除缓存失败: %w", err)
	}

	return deleted, nil
}

// DeleteByPattern 扫描并删除所有匹配 pattern 的缓存 key。
func (c *JSONCache) DeleteByPattern(
	ctx context.Context,
	pattern string,
) (int64, error) {
	var cursor uint64
	var totalDeleted int64

	for {
		//找到的keys，下一批的位置，错误
		keys, nextCursor, err := c.client.Scan(
			ctx,     // 控制本次 Redis 操作何时取消
			cursor,  // 当前扫描游标
			pattern, // key 匹配规则
			100,     // 建议每次扫描的数量
		).Result()
		if err != nil {
			return totalDeleted, fmt.Errorf(
				"扫描缓存 key 失败: %w",
				err,
			)
		}
		//展开切片并删除缓存
		if len(keys) > 0 {
			deleted, err := c.client.Del(ctx, keys...).Result()
			if err != nil {
				return totalDeleted, fmt.Errorf(
					"批量删除缓存失败: %w",
					err,
				)
			}
			//计算删除数量
			totalDeleted += deleted
		}
		//扫下一批，扫完退出
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	//返回删除数量
	return totalDeleted, nil
}

// TTL 查询缓存 key 的剩余生存时间。
func (c *JSONCache) TTL(
	ctx context.Context,
	key string,
) (time.Duration, error) {
	ttl, err := c.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("查询缓存 TTL 失败: %w", err)
	}

	return ttl, nil
}
