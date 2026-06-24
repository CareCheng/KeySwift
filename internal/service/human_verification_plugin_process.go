// Package service 提供人机验证独立插件进程调用器。
package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	pluginapi "user-frontend/internal/plugin"
)

const humanVerificationPluginTimeout = 8 * time.Second

type pluginProcessHumanVerificationProvider struct {
	service         *HumanVerificationService
	manifest        pluginapi.Manifest
	providerID      string
	providerType    string
	displayName     string
	renderMode      string
	supportedScopes []string
}

type humanVerificationPluginRequest struct {
	Action         string                          `json:"action"`
	Scope          string                          `json:"scope,omitempty"`
	ProviderID     string                          `json:"provider_id"`
	ProviderType   string                          `json:"provider_type"`
	Payload        *HumanVerificationPayload       `json:"payload,omitempty"`
	Config         map[string]any                  `json:"config,omitempty"`
	RequestContext HumanVerificationRequestContext `json:"request_context"`
}

type humanVerificationPluginResponse struct {
	Success      bool                                `json:"success"`
	Error        string                              `json:"error,omitempty"`
	ConfigStatus string                              `json:"config_status,omitempty"`
	HealthStatus string                              `json:"health_status,omitempty"`
	PublicConfig map[string]any                      `json:"public_config,omitempty"`
	Challenge    *HumanVerificationChallengeResponse `json:"challenge,omitempty"`
	Verified     bool                                `json:"verified,omitempty"`
	Reason       string                              `json:"reason,omitempty"`
}

func (p *pluginProcessHumanVerificationProvider) ProviderID() string {
	return p.providerID
}

func (p *pluginProcessHumanVerificationProvider) ProviderType() string {
	return p.providerType
}

func (p *pluginProcessHumanVerificationProvider) DisplayName() string {
	return p.displayName
}

func (p *pluginProcessHumanVerificationProvider) SupportedScopes() []string {
	return append([]string(nil), p.supportedScopes...)
}

func (p *pluginProcessHumanVerificationProvider) PublicConfig(scope string) map[string]any {
	ctx, cancel := context.WithTimeout(context.Background(), humanVerificationPluginTimeout)
	defer cancel()
	resp, err := p.invoke(ctx, humanVerificationPluginRequest{
		Action: "public_config",
		Scope:  scope,
	})
	if err != nil || resp == nil || !resp.Success {
		return map[string]any{}
	}
	if resp.PublicConfig == nil {
		return map[string]any{}
	}
	return resp.PublicConfig
}

func (p *pluginProcessHumanVerificationProvider) ConfigStatus() string {
	ctx, cancel := context.WithTimeout(context.Background(), humanVerificationPluginTimeout)
	defer cancel()
	resp, err := p.invoke(ctx, humanVerificationPluginRequest{Action: "config_status"})
	if err != nil || resp == nil || !resp.Success {
		return "unavailable"
	}
	if strings.TrimSpace(resp.ConfigStatus) == "" {
		return "ready"
	}
	return strings.TrimSpace(resp.ConfigStatus)
}

func (p *pluginProcessHumanVerificationProvider) HealthStatus(ctx context.Context) string {
	if ctx == nil {
		ctx = context.Background()
	}
	callCtx, cancel := context.WithTimeout(ctx, humanVerificationPluginTimeout)
	defer cancel()
	resp, err := p.invoke(callCtx, humanVerificationPluginRequest{Action: "health"})
	if err != nil || resp == nil || !resp.Success {
		return "unhealthy"
	}
	if strings.TrimSpace(resp.HealthStatus) == "" {
		return "ready"
	}
	return strings.TrimSpace(resp.HealthStatus)
}

func (p *pluginProcessHumanVerificationProvider) CreateChallenge(ctx context.Context, req HumanVerificationChallengeRequest) (*HumanVerificationChallengeResponse, error) {
	resp, err := p.invoke(ctx, humanVerificationPluginRequest{
		Action:         "create_challenge",
		Scope:          req.Scope,
		RequestContext: req.RequestContext,
	})
	if err != nil {
		return nil, err
	}
	if resp == nil || !resp.Success {
		return nil, errors.New(defaultText(responseError(resp), "人机验证插件创建挑战失败"))
	}
	if resp.Challenge == nil {
		return nil, errors.New("人机验证插件未返回挑战")
	}
	resp.Challenge.ProviderID = p.ProviderID()
	resp.Challenge.ProviderType = p.ProviderType()
	resp.Challenge.Scope = req.Scope
	return resp.Challenge, nil
}

func (p *pluginProcessHumanVerificationProvider) Verify(ctx context.Context, req HumanVerificationVerifyRequest) (*HumanVerificationVerifyResult, error) {
	resp, err := p.invoke(ctx, humanVerificationPluginRequest{
		Action:         "verify",
		Scope:          req.Scope,
		Payload:        req.Payload,
		RequestContext: req.RequestContext,
	})
	if err != nil {
		return nil, err
	}
	if resp == nil || !resp.Success {
		return nil, errors.New(defaultText(responseError(resp), "人机验证插件校验失败"))
	}
	return &HumanVerificationVerifyResult{
		Success: resp.Verified,
		Reason:  resp.Reason,
	}, nil
}

func (p *pluginProcessHumanVerificationProvider) invoke(ctx context.Context, req humanVerificationPluginRequest) (*humanVerificationPluginResponse, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	callCtx, cancel := context.WithTimeout(ctx, humanVerificationPluginTimeout)
	defer cancel()

	binaryPath, releaseRoot, err := p.binaryPath()
	if err != nil {
		return nil, err
	}
	dataDir, err := p.dataDir()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("创建插件数据目录失败: %w", err)
	}

	req.ProviderID = p.ProviderID()
	req.ProviderType = p.ProviderType()
	req.Config = p.service.pluginConfig(p.ProviderID())
	input, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(callCtx, binaryPath)
	cmd.Dir = releaseRoot
	cmd.Stdin = bytes.NewReader(input)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Env = append(os.Environ(),
		"KEYSWIFT_PLUGIN_ID="+p.manifest.ID,
		"KEYSWIFT_PLUGIN_VERSION="+p.manifest.Version,
		"KEYSWIFT_PLUGIN_RELEASE_DIR="+releaseRoot,
		"KEYSWIFT_PLUGIN_DATA_DIR="+dataDir,
	)
	if err := cmd.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = err.Error()
		}
		return nil, fmt.Errorf("人机验证插件调用失败: %s", message)
	}
	var resp humanVerificationPluginResponse
	if err := json.Unmarshal(bytes.TrimSpace(stdout.Bytes()), &resp); err != nil {
		return nil, fmt.Errorf("人机验证插件响应解析失败: %w", err)
	}
	return &resp, nil
}

func (p *pluginProcessHumanVerificationProvider) binaryPath() (string, string, error) {
	if p == nil || p.service == nil || p.service.pluginSvc == nil {
		return "", "", errors.New("插件服务未初始化")
	}
	releaseRoot := p.service.pluginSvc.ReleaseRoot(p.manifest.ID, p.manifest.Version)
	binaryRef := p.selectedBinaryRef()
	if strings.TrimSpace(binaryRef) == "" {
		return "", releaseRoot, errors.New("人机验证插件未声明当前平台二进制")
	}
	cleanRef := filepath.Clean(binaryRef)
	if filepath.IsAbs(cleanRef) || strings.HasPrefix(cleanRef, "..") {
		return "", releaseRoot, errors.New("人机验证插件二进制路径不合法")
	}
	binaryPath := filepath.Join(releaseRoot, cleanRef)
	if _, err := os.Stat(binaryPath); err != nil {
		return "", releaseRoot, errors.New("人机验证插件二进制不存在，请先构建或安装插件")
	}
	return binaryPath, releaseRoot, nil
}

func (p *pluginProcessHumanVerificationProvider) selectedBinaryRef() string {
	for _, item := range p.manifest.Package.Binaries {
		if platformMatches(item.Platform, runtime.GOOS) && platformMatches(item.Arch, runtime.GOARCH) {
			if strings.TrimSpace(item.Path) != "" {
				return item.Path
			}
			if strings.TrimSpace(item.Entrypoint) != "" {
				return item.Entrypoint
			}
		}
	}
	if strings.TrimSpace(p.manifest.Backend.EntryExecutable) != "" {
		return p.manifest.Backend.EntryExecutable
	}
	return p.manifest.Package.DefaultBinary
}

func (p *pluginProcessHumanVerificationProvider) dataDir() (string, error) {
	if p == nil || p.service == nil || p.service.pluginSvc == nil {
		return "", errors.New("插件服务未初始化")
	}
	return p.service.pluginSvc.DataRoot(p.manifest.ID), nil
}

func platformMatches(value string, expected string) bool {
	value = strings.TrimSpace(strings.ToLower(value))
	return value == "all" || value == strings.ToLower(expected)
}

func responseError(resp *humanVerificationPluginResponse) string {
	if resp == nil {
		return ""
	}
	return strings.TrimSpace(resp.Error)
}

func extensionString(values map[string]any, key string) string {
	if values == nil {
		return ""
	}
	return stringFromAny(values[key])
}

func extensionStringSlice(value any) []string {
	switch item := value.(type) {
	case []string:
		return append([]string(nil), item...)
	case []any:
		result := make([]string, 0, len(item))
		for _, entry := range item {
			if text := stringFromAny(entry); text != "" {
				result = append(result, text)
			}
		}
		return result
	default:
		return nil
	}
}
