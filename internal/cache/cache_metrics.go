// Package cache 提供统一的缓存抽象层
// cache_metrics.go - 缓存监控指标
package cache

import (
	"sync/atomic"
	"time"
)

// CacheMetrics 缓存监控指标
type CacheMetrics struct {
	// 命中统计
	hits   atomic.Int64
	misses atomic.Int64

	// 启动时间
	startTime time.Time

	// 最后一次错误
	lastError     atomic.Value // string
	lastErrorTime atomic.Value // time.Time
}

// NewCacheMetrics 创建缓存监控指标
func NewCacheMetrics() *CacheMetrics {
	m := &CacheMetrics{
		startTime: time.Now(),
	}
	m.lastError.Store("")
	m.lastErrorTime.Store(time.Time{})
	return m
}

// RecordHit 记录缓存命中
func (m *CacheMetrics) RecordHit() {
	m.hits.Add(1)
}

// RecordMiss 记录缓存未命中
func (m *CacheMetrics) RecordMiss() {
	m.misses.Add(1)
}

// RecordError 记录错误
func (m *CacheMetrics) RecordError(err string) {
	m.lastError.Store(err)
	m.lastErrorTime.Store(time.Now())
}

// GetHits 获取命中次数
func (m *CacheMetrics) GetHits() int64 {
	return m.hits.Load()
}

// GetMisses 获取未命中次数
func (m *CacheMetrics) GetMisses() int64 {
	return m.misses.Load()
}

// GetHitRate 获取命中率
func (m *CacheMetrics) GetHitRate() float64 {
	hits := m.hits.Load()
	misses := m.misses.Load()
	total := hits + misses
	if total == 0 {
		return 0
	}
	return float64(hits) / float64(total) * 100
}

// GetUptime 获取运行时间
func (m *CacheMetrics) GetUptime() time.Duration {
	return time.Since(m.startTime)
}

// GetLastError 获取最后一次错误
func (m *CacheMetrics) GetLastError() string {
	if v := m.lastError.Load(); v != nil {
		return v.(string)
	}
	return ""
}

// GetLastErrorTime 获取最后一次错误时间
func (m *CacheMetrics) GetLastErrorTime() time.Time {
	if v := m.lastErrorTime.Load(); v != nil {
		return v.(time.Time)
	}
	return time.Time{}
}

// CacheStats 缓存统计信息（用于 API 返回）
type CacheStats struct {
	LocalCacheSize int     `json:"local_cache_size"`
	Hits           int64   `json:"hits"`
	Misses         int64   `json:"misses"`
	HitRate        string  `json:"hit_rate"`
	Uptime         string  `json:"uptime"`
	LastError      string  `json:"last_error,omitempty"`
}

// CacheDashboard 缓存仪表盘数据（详细统计）
type CacheDashboard struct {
	// 基本信息
	Mode           string `json:"mode"`             // 缓存模式：local
	Status         string `json:"status"`           // 状态：connected
	Uptime         string `json:"uptime"`           // 运行时间
	UptimeSeconds  int64  `json:"uptime_seconds"`   // 运行时间（秒）
	
	// 性能指标
	HitRate        float64 `json:"hit_rate"`        // 命中率（百分比）
	HitRateStr     string  `json:"hit_rate_str"`    // 命中率（字符串）
	Hits           int64   `json:"hits"`            // 命中次数
	Misses         int64   `json:"misses"`          // 未命中次数
	TotalRequests  int64   `json:"total_requests"`  // 总请求次数
	OpsPerSecond   float64 `json:"ops_per_second"`  // 每秒操作数
	
	// 内存信息
	MemoryUsed      string  `json:"memory_used"`       // 已用内存
	MemoryUsedBytes int64   `json:"memory_used_bytes"` // 已用内存（字节）
	// 键空间信息
	KeysCount      int64  `json:"keys_count"`       // 总键数

	// 错误信息
	LastError        string `json:"last_error"`          // 最后一次错误
	LastErrorTime    string `json:"last_error_time"`     // 最后错误时间
	
	// 本地缓存信息（降级模式时使用）
	LocalCacheSize   int    `json:"local_cache_size"`    // 本地缓存条目数
	LocalCacheMemory string `json:"local_cache_memory"`  // 本地缓存估算内存
}

// GetStats 获取缓存统计信息
func (cm *CacheManager) GetStats() *CacheStats {
	hitRate := cm.metrics.GetHitRate()
	return &CacheStats{
		LocalCacheSize: cm.local.Size(),
		Hits:           cm.metrics.GetHits(),
		Misses:         cm.metrics.GetMisses(),
		HitRate:        formatPercent(hitRate),
		Uptime:         formatDuration(cm.metrics.GetUptime()),
		LastError:      cm.metrics.GetLastError(),
	}
}

// GetDashboard 获取缓存仪表盘数据
func (cm *CacheManager) GetDashboard() *CacheDashboard {
	dashboard := &CacheDashboard{
		Mode:            "local",
		Status:          "connected",
		Uptime:          formatDuration(cm.metrics.GetUptime()),
		UptimeSeconds:   int64(cm.metrics.GetUptime().Seconds()),
		HitRate:         cm.metrics.GetHitRate(),
		HitRateStr:      formatPercent(cm.metrics.GetHitRate()),
		Hits:            cm.metrics.GetHits(),
		Misses:          cm.metrics.GetMisses(),
		TotalRequests:   cm.metrics.GetHits() + cm.metrics.GetMisses(),
		LastError:       cm.metrics.GetLastError(),
		KeysCount:       int64(cm.local.Size()),
		LocalCacheSize:  cm.local.Size(),
		LocalCacheMemory: formatBytes(int64(cm.local.Size() * 256)), // 估算每个条目256字节
	}
	
	// 计算每秒操作数
	uptimeSeconds := cm.metrics.GetUptime().Seconds()
	if uptimeSeconds > 0 {
		dashboard.OpsPerSecond = float64(dashboard.TotalRequests) / uptimeSeconds
	}
	
	// 最后错误时间
	if lastErrTime := cm.metrics.GetLastErrorTime(); !lastErrTime.IsZero() {
		dashboard.LastErrorTime = lastErrTime.Format("2006-01-02 15:04:05")
	}
	
	return dashboard
}

// formatBytes 格式化字节数
func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)
	
	if bytes < KB {
		return itoa(bytes) + " B"
	}
	if bytes < MB {
		return formatFloat(float64(bytes)/float64(KB), 2) + " KB"
	}
	if bytes < GB {
		return formatFloat(float64(bytes)/float64(MB), 2) + " MB"
	}
	return formatFloat(float64(bytes)/float64(GB), 2) + " GB"
}

// formatPercent 格式化百分比
func formatPercent(value float64) string {
	return formatFloat(value, 2) + "%"
}

// formatFloat 格式化浮点数
func formatFloat(value float64, precision int) string {
	format := "%." + string(rune('0'+precision)) + "f"
	return sprintf(format, value)
}

// sprintf 简单的格式化函数
func sprintf(format string, value float64) string {
	// 使用简单的整数和小数分离方式
	intPart := int64(value)
	decPart := int64((value - float64(intPart)) * 100)
	if decPart < 0 {
		decPart = -decPart
	}
	if decPart < 10 {
		return itoa(intPart) + ".0" + itoa(decPart)
	}
	return itoa(intPart) + "." + itoa(decPart)
}

// itoa 整数转字符串
func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	negative := n < 0
	if negative {
		n = -n
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if negative {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}

// formatDuration 格式化时间间隔
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return itoa(int64(d.Seconds())) + "s"
	}
	if d < time.Hour {
		return itoa(int64(d.Minutes())) + "m " + itoa(int64(d.Seconds())%60) + "s"
	}
	if d < 24*time.Hour {
		return itoa(int64(d.Hours())) + "h " + itoa(int64(d.Minutes())%60) + "m"
	}
	days := int64(d.Hours()) / 24
	hours := int64(d.Hours()) % 24
	return itoa(days) + "d " + itoa(hours) + "h"
}
