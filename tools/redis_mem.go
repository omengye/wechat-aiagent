package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/omengye/wechat_aiagent/types"
	"github.com/silenceper/wechat/v2/cache"
)

// RedisMem Redis内存客户端封装
type RedisMem struct {
	Client *cache.Redis
}

// NewRedisMem 创建Redis客户端实例
func NewRedisMem(ctx context.Context, config *types.Config) *RedisMem {
	newRedis := cache.NewRedis(ctx, &cache.RedisOpts{
		Host:     config.Redis.Addr,
		Password: config.Redis.Password,
	})
	return &RedisMem{
		Client: newRedis,
	}
}

// Set 存储值到Redis,并设置过期时间(单位:秒)
func (r *RedisMem) Set(key string, val interface{}, timeoutSeconds int64) error {
	return r.Client.Set(key, val, time.Duration(timeoutSeconds)*time.Second)
}

// Get 从Redis获取字符串值
func (r *RedisMem) Get(key string) (string, error) {
	val := r.Client.Get(key)
	if val == nil {
		return "", nil
	}

	// 尝试将值转换为字符串
	switch v := val.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	default:
		// 如果不是字符串或字节数组，返回格式化错误
		return "", fmt.Errorf("unexpected value type: %T, expected string or []byte", val)
	}
}

// Del 删除Redis中的键
func (r *RedisMem) Del(key string) error {
	return r.Client.Delete(key)
}
