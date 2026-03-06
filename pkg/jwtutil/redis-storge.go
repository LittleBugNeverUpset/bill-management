// pkg/jwt/redis_storage.go
package jwtutil

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisBlacklistStorage Redis实现的黑名单存储
type RedisBlacklistStorage struct {
	client *redis.Client
	prefix string // Key前缀，默认"jwt:blacklist:"
	ctx    context.Context
}

// NewRedisBlacklistStorage 创建Redis存储实例
func NewRedisBlacklistStorage(client *redis.Client, prefix ...string) *RedisBlacklistStorage {
	p := "jwt:blacklist:"
	if len(prefix) > 0 && prefix[0] != "" {
		p = prefix[0]
	}
	return &RedisBlacklistStorage{
		client: client,
		prefix: p,
		ctx:    context.Background(),
	}
}

// Add 实现TokenBlacklistStorage接口：加入黑名单
func (r *RedisBlacklistStorage) Add(token string, ttl time.Duration) error {
	key := r.prefix + token
	return r.client.Set(r.ctx, key, "blacklisted", ttl).Err()
}

// Exists 实现TokenBlacklistStorage接口：检查是否在黑名单
func (r *RedisBlacklistStorage) Exists(token string) (bool, error) {
	key := r.prefix + token
	exists, err := r.client.Exists(r.ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}
