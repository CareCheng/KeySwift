// Package cache 提供统一的缓存抽象层
// cache_interface.go - 缓存接口定义
package cache

import "time"

// Cache 缓存接口
//
// 定义统一的本地缓存操作接口。
type Cache interface {
	// Get 获取缓存值
	Get(key string) (interface{}, bool)

	// GetString 获取字符串类型缓存值
	GetString(key string) (string, bool)

	// Set 设置缓存值
	Set(key string, value interface{}, ttl time.Duration) error

	// SetString 设置字符串类型缓存值
	SetString(key string, value string, ttl time.Duration) error

	// Delete 删除缓存
	Delete(key string) error

	// Exists 检查键是否存在
	Exists(key string) bool

	// Expire 设置过期时间
	Expire(key string, ttl time.Duration) error

	// TTL 获取剩余过期时间
	TTL(key string) (time.Duration, error)

	// Incr 原子自增
	Incr(key string) (int64, error)

	// IncrBy 原子自增指定值
	IncrBy(key string, delta int64) (int64, error)

	// Decr 原子自减
	Decr(key string) (int64, error)

	// Keys 获取匹配模式的所有键（本地缓存支持前缀匹配）
	Keys(pattern string) ([]string, error)

	// Ping 健康检查
	Ping() error

	// Close 关闭连接
	Close() error
}
