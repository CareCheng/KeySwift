// Package api 提供反向代理运行时配置、请求来源识别和外部访问地址解析。
package api

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"user-frontend/internal/service"

	"github.com/gin-gonic/gin"
)

// ClientIPInfo 描述一次请求的客户端 IP 解析结果。
type ClientIPInfo struct {
	ClientIP     string `json:"client_ip"`
	RemoteIP     string `json:"remote_ip"`
	ForwardedFor string `json:"forwarded_for"`
	RealIP       string `json:"real_ip"`
	TrustedProxy bool   `json:"trusted_proxy"`
	Source       string `json:"source"`
}

// ExternalRequestInfo 描述一次请求在公网侧看到的访问入口信息。
type ExternalRequestInfo struct {
	Scheme       string `json:"scheme"`
	Host         string `json:"host"`
	Port         string `json:"port"`
	BaseURL      string `json:"base_url"`
	PathPrefix   string `json:"path_prefix"`
	TrustedProxy bool   `json:"trusted_proxy"`
}

type trustedProxyMatcher struct {
	ips   []net.IP
	nets  []*net.IPNet
	valid bool
}

var (
	reverseProxyRuntimeMu sync.RWMutex
	reverseProxyRuntime   = service.DefaultReverseProxyConfig()
	trustedProxyRuntime   trustedProxyMatcher
)

// RefreshReverseProxyRuntime 从主业务库刷新反向代理运行时配置。
func RefreshReverseProxyRuntime() (*service.ReverseProxyConfig, error) {
	cfg := service.DefaultReverseProxyConfig()
	if ConfigSvc != nil {
		loaded, err := ConfigSvc.GetReverseProxyConfig()
		if err != nil {
			return nil, err
		}
		cfg = *loaded
	}

	matcher, err := buildTrustedProxyMatcher(cfg.TrustedProxies)
	if err != nil {
		return nil, err
	}

	reverseProxyRuntimeMu.Lock()
	reverseProxyRuntime = cfg
	trustedProxyRuntime = matcher
	reverseProxyRuntimeMu.Unlock()

	return cloneReverseProxyConfig(&cfg), nil
}

// CurrentReverseProxyConfig 返回当前内存中的反向代理配置副本。
func CurrentReverseProxyConfig() service.ReverseProxyConfig {
	reverseProxyRuntimeMu.RLock()
	defer reverseProxyRuntimeMu.RUnlock()
	return *cloneReverseProxyConfig(&reverseProxyRuntime)
}

// TrustedProxyListForGin 返回 Gin 启动期可信代理配置。
func TrustedProxyListForGin() []string {
	cfg := CurrentReverseProxyConfig()
	if !cfg.ReverseProxyEnabled || len(cfg.TrustedProxies) == 0 {
		return nil
	}
	return append([]string{}, cfg.TrustedProxies...)
}

// GetClientIP 返回统一解析后的客户端 IP。
func GetClientIP(c *gin.Context) string {
	return GetClientIPInfo(c).ClientIP
}

// GetClientIPInfo 返回统一解析后的客户端 IP 详情。
func GetClientIPInfo(c *gin.Context) ClientIPInfo {
	cfg := CurrentReverseProxyConfig()
	remoteIP := parseRemoteIP(c)
	info := ClientIPInfo{
		ClientIP:     ipToString(remoteIP),
		RemoteIP:     ipToString(remoteIP),
		ForwardedFor: c.GetHeader(cfg.ClientIPHeader),
		RealIP:       c.GetHeader(cfg.RealIPHeader),
		TrustedProxy: false,
		Source:       "remote",
	}

	if remoteIP == nil || !cfg.ReverseProxyEnabled || !isTrustedProxyIP(remoteIP) {
		return info
	}

	info.TrustedProxy = true
	if ip := firstValidForwardedIP(info.ForwardedFor); ip != nil {
		info.ClientIP = ip.String()
		info.Source = strings.ToLower(cfg.ClientIPHeader)
		return info
	}
	if ip := validHeaderIP(info.RealIP); ip != nil {
		info.ClientIP = ip.String()
		info.Source = strings.ToLower(cfg.RealIPHeader)
		return info
	}
	return info
}

// GetExternalRequestInfo 返回统一解析后的公网访问信息。
func GetExternalRequestInfo(c *gin.Context) ExternalRequestInfo {
	cfg := CurrentReverseProxyConfig()
	trusted := cfg.ReverseProxyEnabled && isTrustedProxyIP(parseRemoteIP(c))
	scheme := requestScheme(c)
	host := c.Request.Host
	port := portFromHostOrScheme(host, scheme)

	if cfg.PublicBaseURL != "" {
		if parsed, err := url.Parse(cfg.PublicBaseURL); err == nil {
			scheme = parsed.Scheme
			host = parsed.Host
			port = portFromHostOrScheme(host, scheme)
		}
	} else if trusted {
		if forwardedScheme := normalizeForwardedScheme(c.GetHeader(cfg.ProtoHeader)); forwardedScheme != "" {
			scheme = forwardedScheme
		}
		if forwardedHost := normalizeForwardedHost(c.GetHeader(cfg.HostHeader)); forwardedHost != "" {
			host = forwardedHost
		}
		if forwardedPort := normalizeForwardedPort(c.GetHeader(cfg.PortHeader)); forwardedPort != "" {
			port = forwardedPort
			host = ensureHostPort(host, port, scheme)
		} else {
			port = portFromHostOrScheme(host, scheme)
		}
	}

	return ExternalRequestInfo{
		Scheme:       scheme,
		Host:         host,
		Port:         port,
		BaseURL:      scheme + "://" + host,
		PathPrefix:   cfg.AppBasePath,
		TrustedProxy: trusted,
	}
}

// IsRequestSecure 判断当前请求在外部访问入口是否为 HTTPS。
func IsRequestSecure(c *gin.Context) bool {
	return GetExternalRequestInfo(c).Scheme == "https"
}

// IsCookieSecure 根据配置和当前请求判断 Cookie 是否需要 Secure。
func IsCookieSecure(c *gin.Context) bool {
	cfg := CurrentReverseProxyConfig()
	switch cfg.CookieSecureMode {
	case service.CookieSecureAlways:
		return true
	case service.CookieSecureNever:
		return false
	default:
		return IsRequestSecure(c)
	}
}

// ReverseProxyCORSMiddleware 根据后台配置处理可选 CORS。
func ReverseProxyCORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := CurrentReverseProxyConfig()
		if !cfg.CORSEnabled {
			c.Next()
			return
		}

		origin := strings.TrimRight(strings.TrimSpace(c.GetHeader("Origin")), "/")
		if origin == "" {
			c.Next()
			return
		}

		if isAllowedCORSOrigin(origin, cfg.CORSAllowOrigins) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			if cfg.CORSAllowCredentials {
				c.Header("Access-Control-Allow-Credentials", "true")
			}
			c.Header("Access-Control-Allow-Headers", "Content-Type, X-CSRF-Token, X-Requested-With")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Max-Age", "600")
		} else if c.Request.Method == http.MethodOptions {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "CORS 来源未被允许"})
			c.Abort()
			return
		}

		if c.Request.Method == http.MethodOptions {
			c.Status(http.StatusNoContent)
			c.Abort()
			return
		}
		c.Next()
	}
}

func cloneReverseProxyConfig(cfg *service.ReverseProxyConfig) *service.ReverseProxyConfig {
	if cfg == nil {
		defaultCfg := service.DefaultReverseProxyConfig()
		cfg = &defaultCfg
	}
	clone := *cfg
	clone.TrustedProxies = append([]string{}, cfg.TrustedProxies...)
	clone.CORSAllowOrigins = append([]string{}, cfg.CORSAllowOrigins...)
	return &clone
}

func buildTrustedProxyMatcher(values []string) (trustedProxyMatcher, error) {
	matcher := trustedProxyMatcher{valid: true}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if ip := net.ParseIP(value); ip != nil {
			matcher.ips = append(matcher.ips, ip)
			continue
		}
		_, network, err := net.ParseCIDR(value)
		if err != nil {
			return matcher, fmt.Errorf("可信代理地址无效: %s", value)
		}
		matcher.nets = append(matcher.nets, network)
	}
	return matcher, nil
}

func isTrustedProxyIP(ip net.IP) bool {
	if ip == nil {
		return false
	}
	reverseProxyRuntimeMu.RLock()
	matcher := trustedProxyRuntime
	reverseProxyRuntimeMu.RUnlock()
	if !matcher.valid {
		return false
	}
	for _, item := range matcher.ips {
		if item.Equal(ip) {
			return true
		}
	}
	for _, network := range matcher.nets {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

func parseRemoteIP(c *gin.Context) net.IP {
	remote := ""
	if c != nil && c.Request != nil {
		remote = c.Request.RemoteAddr
	}
	if remote == "" {
		return nil
	}
	if host, _, err := net.SplitHostPort(remote); err == nil {
		return net.ParseIP(host)
	}
	return net.ParseIP(remote)
}

func ipToString(ip net.IP) string {
	if ip == nil {
		return ""
	}
	return ip.String()
}

func firstValidForwardedIP(value string) net.IP {
	for _, item := range strings.Split(value, ",") {
		if ip := validHeaderIP(item); ip != nil {
			return ip
		}
	}
	return nil
}

func validHeaderIP(value string) net.IP {
	ip := net.ParseIP(strings.TrimSpace(value))
	if ip == nil || ip.IsUnspecified() || ip.IsMulticast() {
		return nil
	}
	return ip
}

func requestScheme(c *gin.Context) string {
	if c != nil && c.Request != nil && c.Request.TLS != nil {
		return "https"
	}
	return "http"
}

func normalizeForwardedScheme(value string) string {
	value = strings.ToLower(strings.TrimSpace(firstHeaderValue(value)))
	if value == "http" || value == "https" {
		return value
	}
	return ""
}

func normalizeForwardedHost(value string) string {
	value = strings.TrimSpace(firstHeaderValue(value))
	if value == "" || strings.ContainsAny(value, "\r\n/") {
		return ""
	}
	return value
}

func normalizeForwardedPort(value string) string {
	value = strings.TrimSpace(firstHeaderValue(value))
	if value == "" {
		return ""
	}
	port, err := strconv.Atoi(value)
	if err != nil || port < 1 || port > 65535 {
		return ""
	}
	return strconv.Itoa(port)
}

func firstHeaderValue(value string) string {
	if idx := strings.Index(value, ","); idx >= 0 {
		return value[:idx]
	}
	return value
}

func portFromHostOrScheme(host, scheme string) string {
	if _, port, err := net.SplitHostPort(host); err == nil && port != "" {
		return port
	}
	if scheme == "https" {
		return "443"
	}
	return "80"
}

func ensureHostPort(host, port, scheme string) string {
	if host == "" || port == "" {
		return host
	}
	if _, _, err := net.SplitHostPort(host); err == nil {
		return host
	}
	if (scheme == "https" && port == "443") || (scheme == "http" && port == "80") {
		return host
	}
	if strings.Contains(host, ":") && !strings.HasPrefix(host, "[") {
		return "[" + host + "]:" + port
	}
	return host + ":" + port
}

func isAllowedCORSOrigin(origin string, allowed []string) bool {
	for _, item := range allowed {
		if item == "*" || strings.EqualFold(origin, item) {
			return true
		}
	}
	return false
}
