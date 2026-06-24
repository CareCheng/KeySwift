package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"user-frontend/internal/service"

	"github.com/gin-gonic/gin"
)

func TestGetClientIPIgnoresForwardedHeaderByDefault(t *testing.T) {
	setReverseProxyRuntimeForTest(service.DefaultReverseProxyConfig())

	c := newReverseProxyTestContext("10.0.0.8:12345")
	c.Request.Header.Set("X-Forwarded-For", "1.2.3.4")

	info := GetClientIPInfo(c)
	if info.ClientIP != "10.0.0.8" {
		t.Fatalf("默认关闭代理时应使用直连 IP，得到 %s", info.ClientIP)
	}
	if info.TrustedProxy {
		t.Fatal("默认关闭代理时不应标记可信代理")
	}
}

func TestGetClientIPUsesForwardedHeaderFromTrustedProxy(t *testing.T) {
	cfg := service.DefaultReverseProxyConfig()
	cfg.ReverseProxyEnabled = true
	cfg.TrustedProxies = []string{"10.0.0.0/24"}
	setReverseProxyRuntimeForTest(cfg)

	c := newReverseProxyTestContext("10.0.0.8:12345")
	c.Request.Header.Set("X-Forwarded-For", "1.2.3.4, 10.0.0.8")

	info := GetClientIPInfo(c)
	if info.ClientIP != "1.2.3.4" {
		t.Fatalf("可信代理应使用 X-Forwarded-For 首个有效 IP，得到 %s", info.ClientIP)
	}
	if !info.TrustedProxy {
		t.Fatal("可信代理来源应标记 trusted_proxy=true")
	}
}

func TestGetClientIPRejectsUntrustedForwardedHeader(t *testing.T) {
	cfg := service.DefaultReverseProxyConfig()
	cfg.ReverseProxyEnabled = true
	cfg.TrustedProxies = []string{"10.0.0.0/24"}
	setReverseProxyRuntimeForTest(cfg)

	c := newReverseProxyTestContext("192.168.1.10:12345")
	c.Request.Header.Set("X-Forwarded-For", "1.2.3.4")

	info := GetClientIPInfo(c)
	if info.ClientIP != "192.168.1.10" {
		t.Fatalf("非可信来源伪造代理头不应生效，得到 %s", info.ClientIP)
	}
	if info.TrustedProxy {
		t.Fatal("非可信来源不应标记 trusted_proxy=true")
	}
}

func TestExternalRequestInfoUsesTrustedForwardedProto(t *testing.T) {
	cfg := service.DefaultReverseProxyConfig()
	cfg.ReverseProxyEnabled = true
	cfg.TrustedProxies = []string{"127.0.0.1"}
	setReverseProxyRuntimeForTest(cfg)

	c := newReverseProxyTestContext("127.0.0.1:12345")
	c.Request.Host = "internal.local:8080"
	c.Request.Header.Set("X-Forwarded-Proto", "https")
	c.Request.Header.Set("X-Forwarded-Host", "example.com")
	c.Request.Header.Set("X-Forwarded-Port", "443")

	info := GetExternalRequestInfo(c)
	if info.Scheme != "https" || info.BaseURL != "https://example.com" {
		t.Fatalf("外部访问信息解析错误: %#v", info)
	}
	if !IsCookieSecure(c) {
		t.Fatal("外部 HTTPS 且 Cookie auto 时应启用 Secure")
	}
}

func TestExternalRequestInfoPrefersPublicBaseURL(t *testing.T) {
	cfg := service.DefaultReverseProxyConfig()
	cfg.ReverseProxyEnabled = true
	cfg.PublicBaseURL = "https://shop.example.com"
	cfg.TrustedProxies = []string{"127.0.0.1"}
	setReverseProxyRuntimeForTest(cfg)

	c := newReverseProxyTestContext("127.0.0.1:12345")
	c.Request.Host = "internal.local:8080"
	c.Request.Header.Set("X-Forwarded-Proto", "http")
	c.Request.Header.Set("X-Forwarded-Host", "wrong.example.com")

	info := GetExternalRequestInfo(c)
	if info.BaseURL != "https://shop.example.com" {
		t.Fatalf("public_base_url 应优先，得到 %s", info.BaseURL)
	}
}

func setReverseProxyRuntimeForTest(cfg service.ReverseProxyConfig) {
	normalized, err := service.NormalizeReverseProxyConfig(&cfg)
	if err != nil {
		panic(err)
	}
	matcher, err := buildTrustedProxyMatcher(normalized.TrustedProxies)
	if err != nil {
		panic(err)
	}
	reverseProxyRuntimeMu.Lock()
	reverseProxyRuntime = *normalized
	trustedProxyRuntime = matcher
	reverseProxyRuntimeMu.Unlock()
}

func newReverseProxyTestContext(remoteAddr string) *gin.Context {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "http://internal.local/api/health", nil)
	c.Request.RemoteAddr = remoteAddr
	return c
}
