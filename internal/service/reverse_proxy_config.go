// Package service 提供反向代理访问配置的读取、保存和校验能力。
package service

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
)

const (
	// ReverseProxyConfigSettingKey 是主业务库 system_settings 中保存反向代理配置的键。
	ReverseProxyConfigSettingKey = "reverse_proxy_config"

	CookieSecureAuto   = "auto"
	CookieSecureAlways = "always"
	CookieSecureNever  = "never"
)

// ReverseProxyConfig 描述访问入口与反向代理相关配置。
//
// 配置保存在主业务数据库 system_settings 中，不依赖环境变量，也不写入配置数据库。
type ReverseProxyConfig struct {
	PublicBaseURL        string   `json:"public_base_url"`
	ReverseProxyEnabled  bool     `json:"reverse_proxy_enabled"`
	TrustedProxies       []string `json:"trusted_proxies"`
	ClientIPHeader       string   `json:"client_ip_header"`
	RealIPHeader         string   `json:"real_ip_header"`
	ProtoHeader          string   `json:"proto_header"`
	HostHeader           string   `json:"host_header"`
	PortHeader           string   `json:"port_header"`
	CookieSecureMode     string   `json:"cookie_secure_mode"`
	CookieDomain         string   `json:"cookie_domain"`
	AppBasePath          string   `json:"app_base_path"`
	CORSEnabled          bool     `json:"cors_enabled"`
	CORSAllowOrigins     []string `json:"cors_allow_origins"`
	CORSAllowCredentials bool     `json:"cors_allow_credentials"`
	HSTSEnabled          bool     `json:"hsts_enabled"`
}

// DefaultReverseProxyConfig 返回首次启动使用的安全默认配置。
func DefaultReverseProxyConfig() ReverseProxyConfig {
	return ReverseProxyConfig{
		PublicBaseURL:        "",
		ReverseProxyEnabled:  false,
		TrustedProxies:       []string{},
		ClientIPHeader:       "X-Forwarded-For",
		RealIPHeader:         "X-Real-IP",
		ProtoHeader:          "X-Forwarded-Proto",
		HostHeader:           "X-Forwarded-Host",
		PortHeader:           "X-Forwarded-Port",
		CookieSecureMode:     CookieSecureAuto,
		CookieDomain:         "",
		AppBasePath:          "/",
		CORSEnabled:          false,
		CORSAllowOrigins:     []string{},
		CORSAllowCredentials: true,
		HSTSEnabled:          false,
	}
}

// GetReverseProxyConfig 从主业务数据库读取反向代理配置。
func (s *ConfigService) GetReverseProxyConfig() (*ReverseProxyConfig, error) {
	defaultCfg := DefaultReverseProxyConfig()
	if s == nil || s.repo == nil {
		return &defaultCfg, nil
	}

	raw, err := s.repo.GetSetting(ReverseProxyConfigSettingKey)
	if err != nil || strings.TrimSpace(raw) == "" {
		return &defaultCfg, nil
	}

	var cfg ReverseProxyConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return nil, fmt.Errorf("解析反向代理配置失败: %w", err)
	}
	return NormalizeReverseProxyConfig(&cfg)
}

// SaveReverseProxyConfig 校验并保存反向代理配置到主业务数据库。
func (s *ConfigService) SaveReverseProxyConfig(cfg *ReverseProxyConfig) (*ReverseProxyConfig, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("数据库未连接")
	}

	normalized, err := NormalizeReverseProxyConfig(cfg)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(normalized)
	if err != nil {
		return nil, fmt.Errorf("序列化反向代理配置失败: %w", err)
	}
	if err := s.repo.SetSetting(ReverseProxyConfigSettingKey, string(data), "反向代理访问配置"); err != nil {
		return nil, err
	}
	return normalized, nil
}

// NormalizeReverseProxyConfig 归一化并校验反向代理配置。
func NormalizeReverseProxyConfig(input *ReverseProxyConfig) (*ReverseProxyConfig, error) {
	defaultCfg := DefaultReverseProxyConfig()
	cfg := defaultCfg
	if input != nil {
		cfg = *input
	}

	var err error
	cfg.PublicBaseURL, err = normalizePublicBaseURL(cfg.PublicBaseURL)
	if err != nil {
		return nil, err
	}

	cfg.TrustedProxies, err = normalizeTrustedProxies(cfg.TrustedProxies)
	if err != nil {
		return nil, err
	}

	if cfg.ClientIPHeader, err = normalizeHeaderName(cfg.ClientIPHeader, defaultCfg.ClientIPHeader); err != nil {
		return nil, err
	}
	if cfg.RealIPHeader, err = normalizeHeaderName(cfg.RealIPHeader, defaultCfg.RealIPHeader); err != nil {
		return nil, err
	}
	if cfg.ProtoHeader, err = normalizeHeaderName(cfg.ProtoHeader, defaultCfg.ProtoHeader); err != nil {
		return nil, err
	}
	if cfg.HostHeader, err = normalizeHeaderName(cfg.HostHeader, defaultCfg.HostHeader); err != nil {
		return nil, err
	}
	if cfg.PortHeader, err = normalizeHeaderName(cfg.PortHeader, defaultCfg.PortHeader); err != nil {
		return nil, err
	}

	cfg.CookieSecureMode = strings.ToLower(strings.TrimSpace(cfg.CookieSecureMode))
	if cfg.CookieSecureMode == "" {
		cfg.CookieSecureMode = defaultCfg.CookieSecureMode
	}
	if cfg.CookieSecureMode != CookieSecureAuto &&
		cfg.CookieSecureMode != CookieSecureAlways &&
		cfg.CookieSecureMode != CookieSecureNever {
		return nil, fmt.Errorf("Cookie Secure 模式只能是 auto、always 或 never")
	}

	cfg.CookieDomain = strings.TrimSpace(cfg.CookieDomain)
	if strings.ContainsAny(cfg.CookieDomain, "\r\n/") {
		return nil, fmt.Errorf("Cookie 域名不能包含换行或路径字符")
	}

	cfg.AppBasePath = strings.TrimSpace(cfg.AppBasePath)
	if cfg.AppBasePath == "" {
		cfg.AppBasePath = "/"
	}
	if cfg.AppBasePath != "/" {
		return nil, fmt.Errorf("当前版本仅支持根路径部署，应用挂载路径必须为 /")
	}

	cfg.CORSAllowOrigins, err = normalizeCORSOrigins(cfg.CORSAllowOrigins, cfg.CORSAllowCredentials)
	if err != nil {
		return nil, err
	}
	if cfg.CORSEnabled && len(cfg.CORSAllowOrigins) == 0 {
		return nil, fmt.Errorf("启用 CORS 时必须配置允许的来源")
	}

	return &cfg, nil
}

func normalizePublicBaseURL(value string) (string, error) {
	value = strings.TrimRight(strings.TrimSpace(value), "/")
	if value == "" {
		return "", nil
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return "", fmt.Errorf("公网访问地址格式错误: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("公网访问地址必须以 http:// 或 https:// 开头")
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("公网访问地址必须包含域名或 IP")
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", fmt.Errorf("公网访问地址不能包含查询参数或片段")
	}
	if parsed.Path != "" && parsed.Path != "/" {
		return "", fmt.Errorf("当前版本仅支持根路径部署，公网访问地址不能包含路径")
	}
	return parsed.Scheme + "://" + parsed.Host, nil
}

func normalizeTrustedProxies(values []string) ([]string, error) {
	seen := make(map[string]struct{})
	result := make([]string, 0, len(values))
	for _, item := range values {
		value := strings.TrimSpace(item)
		if value == "" {
			continue
		}
		if ip := net.ParseIP(value); ip == nil {
			if _, _, err := net.ParseCIDR(value); err != nil {
				return nil, fmt.Errorf("可信代理地址无效: %s", value)
			}
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result, nil
}

func normalizeHeaderName(value, fallback string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		value = fallback
	}
	if strings.ContainsAny(value, "\r\n:") {
		return "", fmt.Errorf("请求头名称无效: %s", value)
	}
	return http.CanonicalHeaderKey(value), nil
}

func normalizeCORSOrigins(values []string, allowCredentials bool) ([]string, error) {
	seen := make(map[string]struct{})
	result := make([]string, 0, len(values))
	for _, item := range values {
		value := strings.TrimRight(strings.TrimSpace(item), "/")
		if value == "" {
			continue
		}
		if value == "*" {
			if allowCredentials {
				return nil, fmt.Errorf("允许携带凭证时 CORS 来源不能使用 *")
			}
			result = append(result, value)
			continue
		}
		parsed, err := url.Parse(value)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			return nil, fmt.Errorf("CORS 来源格式错误: %s", value)
		}
		if parsed.Path != "" && parsed.Path != "/" || parsed.RawQuery != "" || parsed.Fragment != "" {
			return nil, fmt.Errorf("CORS 来源只能包含协议、域名和端口: %s", value)
		}
		origin := parsed.Scheme + "://" + parsed.Host
		if _, exists := seen[origin]; exists {
			continue
		}
		seen[origin] = struct{}{}
		result = append(result, origin)
	}
	return result, nil
}
