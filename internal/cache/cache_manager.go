// Package cache 提供统一的缓存抽象层
// cache_manager.go - 缓存管理器（核心入口）
package cache

import (
	"sync"
	"time"
)

// CacheManager 缓存管理器
//
// 统一管理本地内存缓存，通过 GetCacheManager() 获取全局单例实例。
type CacheManager struct {
	local   *LocalCache
	metrics *CacheMetrics
}

// 全局缓存管理器实例
var (
	globalManager *CacheManager
	managerOnce   sync.Once
)

// InitCacheManager 初始化缓存管理器
func InitCacheManager() error {
	managerOnce.Do(func() {
		globalManager = &CacheManager{
			local:   NewLocalCache(),
			metrics: NewCacheMetrics(),
		}
	})
	return nil
}

// GetCacheManager 获取全局缓存管理器实例
//
// 返回 nil 表示缓存管理器未初始化。
// 服务应该能够处理返回 nil 的情况。
func GetCacheManager() *CacheManager {
	return globalManager
}

// GetManager 获取全局缓存管理器实例（别名）
//
// 这是 GetCacheManager 的简短别名，方便在服务层使用。
func GetManager() *CacheManager {
	return globalManager
}

// getActiveCache 获取当前活跃的缓存实现
func (cm *CacheManager) getActiveCache() Cache {
	return cm.local
}

// Get 获取缓存值
func (cm *CacheManager) Get(key string) (interface{}, bool) {
	val, ok := cm.getActiveCache().Get(key)
	if ok {
		cm.metrics.RecordHit()
	} else {
		cm.metrics.RecordMiss()
	}
	return val, ok
}

// GetString 获取字符串类型缓存值
func (cm *CacheManager) GetString(key string) (string, bool) {
	val, ok := cm.getActiveCache().GetString(key)
	if ok {
		cm.metrics.RecordHit()
	} else {
		cm.metrics.RecordMiss()
	}
	return val, ok
}

// Set 设置缓存值
func (cm *CacheManager) Set(key string, value interface{}, ttl time.Duration) error {
	return cm.getActiveCache().Set(key, value, ttl)
}

// SetString 设置字符串类型缓存值
func (cm *CacheManager) SetString(key string, value string, ttl time.Duration) error {
	return cm.getActiveCache().SetString(key, value, ttl)
}

// Delete 删除缓存
func (cm *CacheManager) Delete(key string) error {
	return cm.local.Delete(key)
}

// DeleteWithDelay 延迟双删
//
// 用于保证数据一致性：先删除缓存，延迟后再删除一次。
func (cm *CacheManager) DeleteWithDelay(key string, delayMs int) {
	// 第一次删除
	cm.Delete(key)

	// 延迟第二次删除（异步）
	go func() {
		time.Sleep(time.Duration(delayMs) * time.Millisecond)
		cm.Delete(key)
	}()
}

// Exists 检查键是否存在
func (cm *CacheManager) Exists(key string) bool {
	return cm.getActiveCache().Exists(key)
}

// Expire 设置过期时间
func (cm *CacheManager) Expire(key string, ttl time.Duration) error {
	return cm.getActiveCache().Expire(key, ttl)
}

// TTL 获取剩余过期时间
func (cm *CacheManager) TTL(key string) (time.Duration, error) {
	return cm.getActiveCache().TTL(key)
}

// Incr 原子自增
func (cm *CacheManager) Incr(key string) (int64, error) {
	return cm.getActiveCache().Incr(key)
}

// IncrBy 原子自增指定值
func (cm *CacheManager) IncrBy(key string, delta int64) (int64, error) {
	return cm.getActiveCache().IncrBy(key, delta)
}

// Decr 原子自减
func (cm *CacheManager) Decr(key string) (int64, error) {
	return cm.getActiveCache().Decr(key)
}

// Keys 获取匹配模式的所有键
func (cm *CacheManager) Keys(pattern string) ([]string, error) {
	return cm.getActiveCache().Keys(pattern)
}

// GetOrLoad 获取缓存值，如果不存在则通过 loader 加载
//
// 这是一个便捷方法，实现了 Cache-Aside 模式。
// 如果缓存未命中，会调用 loader 加载数据并写入缓存。
func (cm *CacheManager) GetOrLoad(key string, loader func() (interface{}, error), ttl time.Duration) (interface{}, error) {
	// 1. 查询缓存
	if val, ok := cm.Get(key); ok {
		return val, nil
	}

	// 2. 缓存未命中，调用 loader 加载
	data, err := loader()
	if err != nil {
		return nil, err
	}

	// 3. 写入缓存
	if data != nil {
		cm.Set(key, data, ttl)
	}

	return data, nil
}

// Ping 健康检查
func (cm *CacheManager) Ping() error {
	return cm.getActiveCache().Ping()
}

// GetMetrics 获取缓存统计指标
func (cm *CacheManager) GetMetrics() *CacheMetrics {
	return cm.metrics
}

// GetLocalCacheSize 获取本地缓存大小
func (cm *CacheManager) GetLocalCacheSize() int {
	return cm.local.Size()
}

// FlushAll 清空所有缓存（危险操作）
func (cm *CacheManager) FlushAll() error {
	cm.local.Close()
	cm.local = NewLocalCache()
	return nil
}

// Close 关闭缓存管理器
func (cm *CacheManager) Close() error {
	cm.local.Close()
	return nil
}
