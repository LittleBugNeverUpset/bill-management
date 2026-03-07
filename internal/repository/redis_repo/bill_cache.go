package redis_repo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"bill-management/pkg/logger"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// BillCache 账单缓存接口
type BillCache interface {
	// SetBillStat 缓存用户月度统计数据
	SetBillStat(ctx context.Context, userID uint64, month string, stat interface{}, expire time.Duration) error
	// GetBillStat 获取用户月度统计数据
	GetBillStat(ctx context.Context, userID uint64, month string, stat interface{}) (bool, error)
	// DelBillStat 删除用户月度统计缓存
	DelBillStat(ctx context.Context, userID uint64, month string) error
	// DelUserAllBillStat 删除用户所有账单统计缓存
	DelUserAllBillStat(ctx context.Context, userID uint64) error
}

// billCache 实现BillCache接口
type billCache struct {
	client *redis.Client
}

// NewBillCache 创建账单缓存实例
func NewBillCache(client *redis.Client) BillCache {
	return &billCache{
		client: client,
	}
}

// 缓存key前缀
const (
	BillStatKeyPrefix = "bill:stat:" // 账单统计缓存key：bill:stat:用户ID:月份
)

// SetBillStat 缓存用户月度统计数据
func (c *billCache) SetBillStat(ctx context.Context, userID uint64, month string, stat interface{}, expire time.Duration) error {
	key := fmt.Sprintf("%s%d:%s", BillStatKeyPrefix, userID, month)
	data, err := json.Marshal(stat)
	if err != nil {
		logger.Error("序列化账单统计数据失败", zap.Uint64("userID", userID), zap.String("month", month), zap.Error(err))
		return err
	}
	return c.client.Set(ctx, key, data, expire).Err()
}

// GetBillStat 获取用户月度统计数据
func (c *billCache) GetBillStat(ctx context.Context, userID uint64, month string, stat interface{}) (bool, error) {
	key := fmt.Sprintf("%s%d:%s", BillStatKeyPrefix, userID, month)
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil // 缓存不存在
		}
		logger.Error("获取账单统计缓存失败", zap.Uint64("userID", userID), zap.String("month", month), zap.Error(err))
		return false, err
	}
	err = json.Unmarshal(data, stat)
	if err != nil {
		logger.Error("反序列化账单统计数据失败", zap.Uint64("userID", userID), zap.String("month", month), zap.Error(err))
		return false, err
	}
	return true, nil
}

// DelBillStat 删除用户月度统计缓存
func (c *billCache) DelBillStat(ctx context.Context, userID uint64, month string) error {
	key := fmt.Sprintf("%s%d:%s", BillStatKeyPrefix, userID, month)
	return c.client.Del(ctx, key).Err()
}

// DelUserAllBillStat 删除用户所有账单统计缓存
func (c *billCache) DelUserAllBillStat(ctx context.Context, userID uint64) error {
	pattern := fmt.Sprintf("%s%d:*", BillStatKeyPrefix, userID)
	keys, err := c.client.Keys(ctx, pattern).Result()
	if err != nil {
		logger.Error("获取用户账单统计缓存key失败", zap.Uint64("userID", userID), zap.Error(err))
		return err
	}
	if len(keys) == 0 {
		return nil
	}
	return c.client.Del(ctx, keys...).Err()
}
