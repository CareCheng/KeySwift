// Package cache 提供统一的缓存抽象层
// local_cache.go - 本地内存缓存实现
package cache

import (
	"strings"
	"sync"
	"time"
)

// LocalCache 本地内存缓存实现
//
// 基于 sync.Map 实现的本地内存缓存，支持 TTL 过期。
type LocalCache struct {
	data    sync.Map
	mu      sync.RWMutex
	closeCh chan struct{}
}

// cacheItem 缓存项
type cacheItem struct {
	value     interface{}
	expiresAt time.Time
}

// NewLocalCache 创建本地缓存实例
func NewLocalCache() *LocalCache {
	lc := &LocalCache{
		closeCh: make(chan struct{}),
	}
	// 启动过期清理协程
	go lc.cleanupLoop()
	return lc
}

// cleanupLoop 定期清理过期数据
func (c *LocalCache) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanup()
		case <-c.closeCh:
			return
		}
	}
}

// cleanup 清理过期数据
func (c *LocalCache) cleanup() {
	now := time.Now()
	c.data.Range(func(key, value interface{}) bool {
		item, ok := value.(*cacheItem)
		if ok && !item.expiresAt.IsZero() && now.After(item.expiresAt) {
			c.data.Delete(key)
		}
		return true
	})
}

// Get 获取缓存值
func (c *LocalCache) Get(key string) (interface{}, bool) {
	value, ok := c.data.Load(key)
	if !ok {
		return nil, false
	}

	item, ok := value.(*cacheItem)
	if !ok {
		return nil, false
	}

	// 检查是否过期
	if !item.expiresAt.IsZero() && time.Now().After(item.expiresAt) {
		c.data.Delete(key)
		return nil, false
	}

	return item.value, true
}

// GetString 获取字符串类型缓存值
func (c *LocalCache) GetString(key string) (string, bool) {
	value, ok := c.Get(key)
	if !ok {
		return "", false
	}
	str, ok := value.(string)
	return str, ok
}

// Set 设置缓存值
func (c *LocalCache) Set(key string, value interface{}, ttl time.Duration) error {
	item := &cacheItem{
		value: value,
	}
	if ttl > 0 {
		item.expiresAt = time.Now().Add(ttl)
	}
	c.data.Store(key, item)
	return nil
}

// SetString 设置字符串类型缓存值
func (c *LocalCache) SetString(key string, value string, ttl time.Duration) error {
	return c.Set(key, value, ttl)
}

// Delete 删除缓存
func (c *LocalCache) Delete(key string) error {
	c.data.Delete(key)
	return nil
}

// Exists 检查键是否存在
func (c *LocalCache) Exists(key string) bool {
	_, ok := c.Get(key)
	return ok
}

// Expire 设置过期时间
func (c *LocalCache) Expire(key string, ttl time.Duration) error {
	value, ok := c.data.Load(key)
	if !ok {
		return nil
	}

	item, ok := value.(*cacheItem)
	if !ok {
		return nil
	}

	item.expiresAt = time.Now().Add(ttl)
	c.data.Store(key, item)
	return nil
}

// TTL 获取剩余过期时间
func (c *LocalCache) TTL(key string) (time.Duration, error) {
	value, ok := c.data.Load(key)
	if !ok {
		return -2, nil // 键不存在
	}

	item, ok := value.(*cacheItem)
	if !ok {
		return -2, nil
	}

	if item.expiresAt.IsZero() {
		return -1, nil // 无过期时间
	}

	remaining := time.Until(item.expiresAt)
	if remaining <= 0 {
		c.data.Delete(key)
		return -2, nil
	}

	return remaining, nil
}

// Incr 原子自增
func (c *LocalCache) Incr(key string) (int64, error) {
	return c.IncrBy(key, 1)
}

// IncrBy 原子自增指定值
func (c *LocalCache) IncrBy(key string, delta int64) (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	value, ok := c.data.Load(key)
	var current int64
	var expiresAt time.Time

	if ok {
		item, ok := value.(*cacheItem)
		if ok {
			if v, ok := item.value.(int64); ok {
				current = v
			}
			expiresAt = item.expiresAt
		}
	}

	current += delta
	item := &cacheItem{
		value:     current,
		expiresAt: expiresAt,
	}
	c.data.Store(key, item)
	return current, nil
}

// Decr 原子自减
func (c *LocalCache) Decr(key string) (int64, error) {
	return c.IncrBy(key, -1)
}

// Keys 获取匹配模式的所有键（前缀匹配）
func (c *LocalCache) Keys(pattern string) ([]string, error) {
	var keys []string
	prefix := strings.TrimSuffix(pattern, "*")

	c.data.Range(func(key, value interface{}) bool {
		keyStr, ok := key.(string)
		if !ok {
			return true
		}
		if strings.HasPrefix(keyStr, prefix) {
			// 检查是否过期
			item, ok := value.(*cacheItem)
			if ok && (item.expiresAt.IsZero() || time.Now().Before(item.expiresAt)) {
				keys = append(keys, keyStr)
			}
		}
		return true
	})

	return keys, nil
}

// Ping 健康检查
func (c *LocalCache) Ping() error {
	return nil
}

// Close 关闭缓存
func (c *LocalCache) Close() error {
	close(c.closeCh)
	return nil
}

// Size 获取缓存大小（用于监控）
func (c *LocalCache) Size() int {
	count := 0
	c.data.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}
