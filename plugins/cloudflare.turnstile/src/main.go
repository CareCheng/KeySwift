// Package main 是 Cloudflare Turnstile 人机验证独立插件入口。
package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	providerID   = "cloudflare.turnstile"
	providerType = "cloudflare_turnstile"
	verifyURL    = "https://challenges.cloudflare.com/turnstile/v0/siteverify"
)

type request struct {
	Action         string         `json:"action"`
	Scope          string         `json:"scope"`
	ProviderID     string         `json:"provider_id"`
	ProviderType   string         `json:"provider_type"`
	Payload        *payload       `json:"payload"`
	Config         map[string]any `json:"config"`
	RequestContext requestContext `json:"request_context"`
}

type payload struct {
	Token       string `json:"token"`
	ClientNonce string `json:"client_nonce"`
}

type requestContext struct {
	ClientIP  string `json:"client_ip"`
	UserAgent string `json:"user_agent"`
	RequestID string `json:"request_id"`
}

type response struct {
	Success      bool               `json:"success"`
	Error        string             `json:"error,omitempty"`
	ConfigStatus string             `json:"config_status,omitempty"`
	HealthStatus string             `json:"health_status,omitempty"`
	PublicConfig map[string]any     `json:"public_config,omitempty"`
	Challenge    *challengeResponse `json:"challenge,omitempty"`
	Verified     bool               `json:"verified,omitempty"`
	Reason       string             `json:"reason,omitempty"`
}

type challengeResponse struct {
	ProviderID   string         `json:"provider_id"`
	ProviderType string         `json:"provider_type"`
	Scope        string         `json:"scope"`
	PublicConfig map[string]any `json:"public_config,omitempty"`
}

type siteVerifyResponse struct {
	Success    bool     `json:"success"`
	ErrorCodes []string `json:"error-codes"`
}

func main() {
	var req request
	if err := json.NewDecoder(os.Stdin).Decode(&req); err != nil {
		write(response{Success: false, Error: "请求解析失败"})
		return
	}
	write(handle(req))
}

func handle(req request) response {
	switch strings.TrimSpace(req.Action) {
	case "public_config":
		return response{Success: true, PublicConfig: publicConfig(req.Config)}
	case "config_status":
		if configReady(req.Config) {
			return response{Success: true, ConfigStatus: "ready"}
		}
		return response{Success: true, ConfigStatus: "missing_config"}
	case "health":
		if configReady(req.Config) {
			return response{Success: true, HealthStatus: "ready"}
		}
		return response{Success: true, HealthStatus: "degraded"}
	case "create_challenge":
		return response{
			Success: true,
			Challenge: &challengeResponse{
				ProviderID:   providerID,
				ProviderType: providerType,
				Scope:        strings.TrimSpace(req.Scope),
				PublicConfig: publicConfig(req.Config),
			},
		}
	case "verify":
		return verify(req)
	default:
		return response{Success: false, Error: "不支持的人机验证插件动作"}
	}
}

func verify(req request) response {
	if req.Payload == nil || strings.TrimSpace(req.Payload.Token) == "" {
		return response{Success: true, Verified: false, Reason: "请完成人机验证"}
	}
	secretKey := strings.TrimSpace(stringFromAny(req.Config["secret_key"]))
	if secretKey == "" {
		return response{Success: false, Error: "Cloudflare Turnstile 服务端密钥未配置"}
	}
	result, err := siteVerify(secretKey, strings.TrimSpace(req.Payload.Token), req.Payload.ClientNonce, req.RequestContext.ClientIP)
	if err != nil {
		return response{Success: false, Error: err.Error()}
	}
	if !result.Success {
		reason := "人机验证失败"
		if len(result.ErrorCodes) > 0 {
			reason += "：" + strings.Join(result.ErrorCodes, ",")
		}
		return response{Success: true, Verified: false, Reason: reason}
	}
	return response{Success: true, Verified: true}
}

func siteVerify(secretKey string, token string, nonce string, clientIP string) (siteVerifyResponse, error) {
	form := url.Values{}
	form.Set("secret", secretKey)
	form.Set("response", token)
	if strings.TrimSpace(clientIP) != "" {
		form.Set("remoteip", strings.TrimSpace(clientIP))
	}
	if strings.TrimSpace(nonce) != "" {
		form.Set("idempotency_key", strings.TrimSpace(nonce))
	}
	req, err := http.NewRequest(http.MethodPost, verifyURL, strings.NewReader(form.Encode()))
	if err != nil {
		return siteVerifyResponse{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return siteVerifyResponse{}, errors.New("Cloudflare Turnstile 校验请求失败")
	}
	defer resp.Body.Close()
	var result siteVerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return siteVerifyResponse{}, errors.New("Cloudflare Turnstile 校验响应解析失败")
	}
	return result, nil
}

func publicConfig(values map[string]any) map[string]any {
	result := map[string]any{
		"site_key": stringFromAny(values["site_key"]),
		"theme":    defaultString(stringFromAny(values["theme"]), "auto"),
		"size":     defaultString(stringFromAny(values["size"]), "normal"),
	}
	if appearance := stringFromAny(values["appearance"]); appearance != "" {
		result["appearance"] = appearance
	}
	return result
}

func configReady(values map[string]any) bool {
	return strings.TrimSpace(stringFromAny(values["site_key"])) != "" &&
		strings.TrimSpace(stringFromAny(values["secret_key"])) != ""
}

func stringFromAny(value any) string {
	if text, ok := value.(string); ok {
		return strings.TrimSpace(text)
	}
	return ""
}

func defaultString(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func write(resp response) {
	_ = json.NewEncoder(os.Stdout).Encode(resp)
}
