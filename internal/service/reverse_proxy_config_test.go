package service

import "testing"

func TestNormalizeReverseProxyConfigDefaults(t *testing.T) {
	cfg, err := NormalizeReverseProxyConfig(nil)
	if err != nil {
		t.Fatalf("默认配置不应失败: %v", err)
	}
	if cfg.ReverseProxyEnabled {
		t.Fatal("首次启动默认不应启用反向代理")
	}
	if cfg.CookieSecureMode != CookieSecureAuto {
		t.Fatalf("默认 Cookie Secure 模式应为 auto，得到 %s", cfg.CookieSecureMode)
	}
	if cfg.AppBasePath != "/" {
		t.Fatalf("默认挂载路径应为 /，得到 %s", cfg.AppBasePath)
	}
}

func TestNormalizeReverseProxyConfigRejectsInvalidCIDR(t *testing.T) {
	cfg := DefaultReverseProxyConfig()
	cfg.TrustedProxies = []string{"not-a-cidr"}
	if _, err := NormalizeReverseProxyConfig(&cfg); err == nil {
		t.Fatal("非法可信代理地址应被拒绝")
	}
}

func TestNormalizeReverseProxyConfigRejectsPathBaseURL(t *testing.T) {
	cfg := DefaultReverseProxyConfig()
	cfg.PublicBaseURL = "https://example.com/keyswift"
	if _, err := NormalizeReverseProxyConfig(&cfg); err == nil {
		t.Fatal("当前版本不支持带路径的公网地址")
	}
}

func TestNormalizeReverseProxyConfigRejectsWildcardCORSWithCredentials(t *testing.T) {
	cfg := DefaultReverseProxyConfig()
	cfg.CORSEnabled = true
	cfg.CORSAllowCredentials = true
	cfg.CORSAllowOrigins = []string{"*"}
	if _, err := NormalizeReverseProxyConfig(&cfg); err == nil {
		t.Fatal("允许 Cookie 时不应接受 CORS 通配来源")
	}
}

func TestNormalizeReverseProxyConfigAcceptsCORSOrigins(t *testing.T) {
	cfg := DefaultReverseProxyConfig()
	cfg.CORSEnabled = true
	cfg.CORSAllowOrigins = []string{"https://web.example.com/", "https://web.example.com"}
	normalized, err := NormalizeReverseProxyConfig(&cfg)
	if err != nil {
		t.Fatalf("合法 CORS 来源不应失败: %v", err)
	}
	if len(normalized.CORSAllowOrigins) != 1 || normalized.CORSAllowOrigins[0] != "https://web.example.com" {
		t.Fatalf("CORS 来源应归一化去重，得到 %#v", normalized.CORSAllowOrigins)
	}
}
