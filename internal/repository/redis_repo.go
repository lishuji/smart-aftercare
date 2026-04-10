package repository

import (
	"smart-aftercare/config"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v9"
)

// QACacheData 问答缓存数据
type QACacheData struct {
	Answer  string   `json:"answer"`
	Sources []string `json:"sources"`
	Images  []string `json:"images"`
}

// RedisRepo Redis 缓存数据访问层
type RedisRepo struct {
	client *redis.Client
}

// NewRedisRepo 创建 Redis 数据访问层
func NewRedisRepo(cfg config.RedisConfig) *RedisRepo {
	client := redis.NewClient(&redis.Options{
		Addr:        cfg.Address(),
		Password:    cfg.Password,
		DB:          cfg.DB,
		PoolSize:    cfg.PoolSize,
		IdleTimeout: cfg.IdleTimeout,
	})

	return &RedisRepo{client: client}
}

// Ping 检查 Redis 连接
func (r *RedisRepo) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Close 关闭 Redis 连接
func (r *RedisRepo) Close() error {
	return r.client.Close()
}

// ==================== QA 缓存操作 ====================

// GetQACache 获取问答缓存
func (r *RedisRepo) GetQACache(cacheKey string) (string, []string, []string, error) {
	ctx := context.Background()

	data, err := r.client.Get(ctx, cacheKey).Result()
	if err != nil {
		return "", nil, nil, err
	}

	var cache QACacheData
	if err := json.Unmarshal([]byte(data), &cache); err != nil {
		return "", nil, nil, fmt.Errorf("解析缓存数据失败: %w", err)
	}

	return cache.Answer, cache.Sources, cache.Images, nil
}

// SetQACache 设置问答缓存
func (r *RedisRepo) SetQACache(cacheKey, answer string, sources, images []string, expiration time.Duration) error {
	ctx := context.Background()

	cache := QACacheData{
		Answer:  answer,
		Sources: sources,
		Images:  images,
	}

	data, err := json.Marshal(cache)
	if err != nil {
		return fmt.Errorf("序列化缓存数据失败: %w", err)
	}

	return r.client.Set(ctx, cacheKey, string(data), expiration).Err()
}

// DeleteQACache 删除问答缓存
func (r *RedisRepo) DeleteQACache(cacheKey string) error {
	ctx := context.Background()
	return r.client.Del(ctx, cacheKey).Err()
}

// DeleteQACacheByPattern 按模式删除缓存（如删除某型号的所有缓存）
func (r *RedisRepo) DeleteQACacheByPattern(pattern string) error {
	ctx := context.Background()

	var cursor uint64
	for {
		keys, nextCursor, err := r.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return fmt.Errorf("扫描缓存键失败: %w", err)
		}

		if len(keys) > 0 {
			if err := r.client.Del(ctx, keys...).Err(); err != nil {
				return fmt.Errorf("删除缓存键失败: %w", err)
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return nil
}

// ==================== 通用缓存操作 ====================

// Set 设置通用缓存
func (r *RedisRepo) Set(ctx context.Context, key, value string, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

// Get 获取通用缓存
func (r *RedisRepo) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// Delete 删除通用缓存
func (r *RedisRepo) Delete(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

// Exists 检查键是否存在
func (r *RedisRepo) Exists(ctx context.Context, key string) (bool, error) {
	result, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// Incr 自增（用于统计计数等场景）
func (r *RedisRepo) Incr(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

// GetClient 获取底层 Redis 客户端（谨慎使用）
func (r *RedisRepo) GetClient() *redis.Client {
	return r.client
}
