// Package service 提供人机验证插件化接入服务。
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"user-frontend/internal/config"
	"user-frontend/internal/model"
	pluginapi "user-frontend/internal/plugin"
)

const (
	HumanVerificationCapability = "human_verification.provider"

	HumanScopeAdminLogin   = "admin_login"
	HumanScopeUserLogin    = "user_login"
	HumanScopeUserRegister = "user_register"

	HumanProviderTypeImageCaptcha        = "image_captcha"
	HumanProviderTypeCloudflareTurnstile = "cloudflare_turnstile"
)

var (
	errHumanVerificationProviderMissing  = errors.New("人机验证插件不存在或未安装")
	errHumanVerificationProviderDisabled = errors.New("人机验证插件未启用")
)

// HumanVerificationPayload 是登录、注册等接口提交的人机验证结果。
type HumanVerificationPayload struct {
	ProviderID   string         `json:"provider_id"`
	ProviderType string         `json:"provider_type"`
	Scope        string         `json:"scope"`
	Token        string         `json:"token"`
	ChallengeID  string         `json:"challenge_id"`
	Answer       string         `json:"answer"`
	ClientNonce  string         `json:"client_nonce"`
	Metadata     map[string]any `json:"metadata"`
}

// HumanVerificationRequestContext 是 provider 校验时需要的最小请求上下文。
type HumanVerificationRequestContext struct {
	ClientIP  string `json:"client_ip"`
	UserAgent string `json:"user_agent"`
	RequestID string `json:"request_id"`
}

// HumanVerificationChallengeRequest 是创建挑战的服务层请求。
type HumanVerificationChallengeRequest struct {
	Scope          string
	ProviderID     string
	RequestContext HumanVerificationRequestContext
}

// HumanVerificationChallengeResponse 是公开 challenge 接口返回结构。
type HumanVerificationChallengeResponse struct {
	ProviderID   string         `json:"provider_id"`
	ProviderType string         `json:"provider_type"`
	Scope        string         `json:"scope"`
	ChallengeID  string         `json:"challenge_id,omitempty"`
	Image        string         `json:"image,omitempty"`
	ExpiresAt    *time.Time     `json:"expires_at,omitempty"`
	PublicConfig map[string]any `json:"public_config,omitempty"`
}

// HumanVerificationVerifyRequest 是 provider 校验请求。
type HumanVerificationVerifyRequest struct {
	Scope          string
	Payload        *HumanVerificationPayload
	RequestContext HumanVerificationRequestContext
}

// HumanVerificationVerifyResult 是 provider 校验结果。
type HumanVerificationVerifyResult struct {
	Success bool
	Reason  string
}

// PublicHumanVerificationConfig 是前端认证页可读取的非敏感配置。
type PublicHumanVerificationConfig struct {
	Enabled           bool           `json:"enabled"`
	Scope             string         `json:"scope"`
	ProviderID        string         `json:"provider_id,omitempty"`
	ProviderType      string         `json:"provider_type,omitempty"`
	DisplayName       string         `json:"display_name,omitempty"`
	RenderMode        string         `json:"render_mode,omitempty"`
	FrontendURL       string         `json:"frontend_url,omitempty"`
	FrontendHeight    int            `json:"frontend_height,omitempty"`
	ChallengeEndpoint string         `json:"challenge_endpoint,omitempty"`
	PublicConfig      map[string]any `json:"public_config,omitempty"`
	Error             string         `json:"error,omitempty"`
}

// HumanVerificationProviderSummary 是后台 provider 选择弹窗使用的候选项。
type HumanVerificationProviderSummary struct {
	ProviderID     string         `json:"provider_id"`
	ProviderType   string         `json:"provider_type"`
	DisplayName    string         `json:"display_name"`
	PluginID       string         `json:"plugin_id"`
	Enabled        bool           `json:"enabled"`
	HealthStatus   string         `json:"health_status"`
	ConfigStatus   string         `json:"config_status"`
	SupportedScope []string       `json:"supported_scopes"`
	PublicConfig   map[string]any `json:"public_config,omitempty"`
}

// HumanVerificationPolicy 表示某个业务场景的人机验证策略。
type HumanVerificationPolicy struct {
	Enabled    bool
	ProviderID string
}

type humanVerificationProvider interface {
	ProviderID() string
	ProviderType() string
	DisplayName() string
	SupportedScopes() []string
	PublicConfig(scope string) map[string]any
	ConfigStatus() string
	HealthStatus(context.Context) string
	CreateChallenge(context.Context, HumanVerificationChallengeRequest) (*HumanVerificationChallengeResponse, error)
	Verify(context.Context, HumanVerificationVerifyRequest) (*HumanVerificationVerifyResult, error)
}

// HumanVerificationService 统一管理人机验证 provider 发现、公开配置、挑战生成和结果校验。
type HumanVerificationService struct {
	pluginSvc  *PluginService
	governance *GovernanceService
}

// NewHumanVerificationService 创建人机验证服务。
func NewHumanVerificationService(pluginSvc *PluginService, governance *GovernanceService) *HumanVerificationService {
	return &HumanVerificationService{
		pluginSvc:  pluginSvc,
		governance: governance,
	}
}

// PolicyForScope 返回指定场景的当前启用策略。
// 注意：此处返回的是"配置层声明的策略"，不包含 provider 可用性降级。
// 需要"实际生效策略"（含 provider 不可用时的降级）时使用 EffectivePolicyForScope。
func (s *HumanVerificationService) PolicyForScope(scope string) HumanVerificationPolicy {
	cfg := config.GlobalConfig.ServerConfig
	switch scope {
	case HumanScopeAdminLogin:
		return HumanVerificationPolicy{
			Enabled:    cfg.EnableLogin && cfg.AdminHumanVerificationEnabled,
			ProviderID: defaultProviderID(cfg.AdminHumanVerificationProviderID),
		}
	case HumanScopeUserLogin:
		return HumanVerificationPolicy{
			Enabled:    cfg.UserLoginHumanVerificationEnabled,
			ProviderID: defaultProviderID(cfg.UserLoginHumanVerificationProviderID),
		}
	case HumanScopeUserRegister:
		providerID := cfg.UserRegisterHumanVerificationProviderID
		if cfg.UserRegisterHumanVerificationFollowLogin {
			providerID = cfg.UserLoginHumanVerificationProviderID
		}
		return HumanVerificationPolicy{
			Enabled:    cfg.UserRegisterHumanVerificationEnabled,
			ProviderID: defaultProviderID(providerID),
		}
	default:
		return HumanVerificationPolicy{}
	}
}

// EffectivePolicyForScope 返回指定场景的实际生效策略。
// 在配置层策略基础上，叠加 provider 可用性降级：
//   - 默认 provider 缺失/未启用/配置不全时，静默降级为关闭（兼容旧默认开启配置，避免阻断登录）；
//   - 非默认 provider 不可用时保留启用并交由调用方处理错误（管理员显式选择，不应静默放行）。
//
// 这是展示层与运行时共用的单一权威判定，避免两者状态来源不一致。
func (s *HumanVerificationService) EffectivePolicyForScope(scope string) HumanVerificationPolicy {
	policy := s.PolicyForScope(scope)
	if !policy.Enabled {
		return policy
	}
	_, err := s.availableProvider(policy.ProviderID, scope)
	if err == nil {
		return policy
	}
	if shouldSkipUnavailableDefaultProvider(policy, err) {
		return HumanVerificationPolicy{Enabled: false, ProviderID: policy.ProviderID}
	}
	// 非默认 provider 不可用：保留启用状态，由 PublicConfigForScope/Verify 向调用方暴露错误
	return policy
}

// PublicConfigForScope 返回前端渲染人机验证所需的非敏感配置。
func (s *HumanVerificationService) PublicConfigForScope(scope string) PublicHumanVerificationConfig {
	policy := s.EffectivePolicyForScope(scope)
	if !policy.Enabled {
		return PublicHumanVerificationConfig{Enabled: false, Scope: scope}
	}
	provider, err := s.availableProvider(policy.ProviderID, scope)
	if err != nil {
		// 走到此处说明是非默认 provider 不可用（默认 provider 不可用已在 EffectivePolicyForScope 降级为关闭）
		return PublicHumanVerificationConfig{
			Enabled:    true,
			Scope:      scope,
			ProviderID: policy.ProviderID,
			Error:      err.Error(),
		}
	}
	return PublicHumanVerificationConfig{
		Enabled:           true,
		Scope:             scope,
		ProviderID:        provider.ProviderID(),
		ProviderType:      provider.ProviderType(),
		DisplayName:       provider.DisplayName(),
		RenderMode:        s.providerRenderMode(provider.ProviderID()),
		FrontendURL:       s.providerFrontendURL(provider.ProviderID()),
		FrontendHeight:    s.providerFrontendHeight(provider.ProviderID()),
		ChallengeEndpoint: "/api/human-verification/challenge",
		PublicConfig:      provider.PublicConfig(scope),
	}
}

// PublicConfigMap 返回全部认证场景的人机验证公开配置。
func (s *HumanVerificationService) PublicConfigMap() map[string]PublicHumanVerificationConfig {
	return map[string]PublicHumanVerificationConfig{
		HumanScopeAdminLogin:   s.PublicConfigForScope(HumanScopeAdminLogin),
		HumanScopeUserLogin:    s.PublicConfigForScope(HumanScopeUserLogin),
		HumanScopeUserRegister: s.PublicConfigForScope(HumanScopeUserRegister),
	}
}

// ListProviders 返回已安装且启用的人机验证 provider。
func (s *HumanVerificationService) ListProviders(scopes []string) []HumanVerificationProviderSummary {
	providers := s.humanVerificationProviders()
	items := make([]HumanVerificationProviderSummary, 0, len(providers))
	for _, provider := range providers {
		if !s.providerPluginEnabled(provider.ProviderID()) || !s.manifestDeclaresHumanVerification(provider.ProviderID()) {
			continue
		}
		if !supportsAllScopes(provider.SupportedScopes(), scopes) {
			continue
		}
		items = append(items, HumanVerificationProviderSummary{
			ProviderID:     provider.ProviderID(),
			ProviderType:   provider.ProviderType(),
			DisplayName:    provider.DisplayName(),
			PluginID:       provider.ProviderID(),
			Enabled:        true,
			HealthStatus:   provider.HealthStatus(context.Background()),
			ConfigStatus:   provider.ConfigStatus(),
			SupportedScope: provider.SupportedScopes(),
			PublicConfig:   provider.PublicConfig(firstScope(scopes)),
		})
	}
	return items
}

// ValidatePolicy 校验某个场景是否具备可启用的人机验证 provider。
func (s *HumanVerificationService) ValidatePolicy(scope string, enabled bool, providerID string) error {
	if !enabled {
		return nil
	}
	providerID = defaultProviderID(providerID)
	providers := s.ListProviders([]string{scope})
	if len(providers) == 0 {
		return errors.New("请先安装并启用至少一个人机验证插件")
	}
	provider, err := s.availableProvider(providerID, scope)
	if err != nil {
		return err
	}
	if provider.ConfigStatus() != "ready" {
		return fmt.Errorf("人机验证插件 %s 配置不完整", provider.DisplayName())
	}
	return nil
}

// CreateChallenge 创建需要服务端 challenge 的人机验证挑战。
func (s *HumanVerificationService) CreateChallenge(ctx context.Context, req HumanVerificationChallengeRequest) (*HumanVerificationChallengeResponse, error) {
	if strings.TrimSpace(req.Scope) == "" {
		return nil, errors.New("人机验证场景不能为空")
	}
	policy := s.EffectivePolicyForScope(req.Scope)
	if !policy.Enabled {
		return &HumanVerificationChallengeResponse{Scope: req.Scope}, nil
	}
	providerID := strings.TrimSpace(req.ProviderID)
	if providerID == "" {
		providerID = policy.ProviderID
	}
	if providerID != policy.ProviderID {
		return nil, errors.New("人机验证类型与当前配置不一致")
	}
	provider, err := s.availableProvider(providerID, req.Scope)
	if err != nil {
		// 默认 provider 不可用已在 EffectivePolicyForScope 降级为关闭；走到此处为非默认 provider 不可用
		return nil, err
	}
	return provider.CreateChallenge(ctx, req)
}

// Verify 校验登录、注册等接口提交的人机验证结果。
func (s *HumanVerificationService) Verify(ctx context.Context, req HumanVerificationVerifyRequest) error {
	policy := s.EffectivePolicyForScope(req.Scope)
	if !policy.Enabled {
		return nil
	}
	provider, err := s.availableProvider(policy.ProviderID, req.Scope)
	if err != nil {
		// 默认 provider 不可用已在 EffectivePolicyForScope 降级为关闭；走到此处为非默认 provider 不可用，应阻断
		return err
	}
	if req.Payload == nil {
		return errors.New("请完成人机验证")
	}
	req.Payload.ProviderID = strings.TrimSpace(req.Payload.ProviderID)
	req.Payload.Scope = strings.TrimSpace(req.Payload.Scope)
	if req.Payload.Scope != "" && req.Payload.Scope != req.Scope {
		return errors.New("人机验证场景不匹配")
	}
	if req.Payload.ProviderID == "" {
		return errors.New("人机验证类型不能为空")
	}
	if req.Payload.ProviderID != policy.ProviderID {
		return errors.New("人机验证类型与当前配置不一致")
	}
	result, err := provider.Verify(ctx, req)
	if err != nil {
		return err
	}
	if result == nil || !result.Success {
		if result != nil && result.Reason != "" {
			return errors.New(result.Reason)
		}
		return errors.New("人机验证失败")
	}
	return nil
}

func (s *HumanVerificationService) availableProvider(providerID string, scope string) (humanVerificationProvider, error) {
	providerID = defaultProviderID(providerID)
	provider, ok := s.humanVerificationProvider(providerID)
	if !ok {
		return nil, errHumanVerificationProviderMissing
	}
	if !s.providerPluginEnabled(providerID) || !s.manifestDeclaresHumanVerification(providerID) {
		return nil, errHumanVerificationProviderDisabled
	}
	if !supportsScope(provider.SupportedScopes(), scope) {
		return nil, errors.New("人机验证插件不支持当前场景")
	}
	if provider.ConfigStatus() != "ready" {
		return nil, fmt.Errorf("人机验证插件 %s 配置不完整", provider.DisplayName())
	}
	return provider, nil
}

func shouldSkipUnavailableDefaultProvider(policy HumanVerificationPolicy, err error) bool {
	if !policy.Enabled || defaultProviderID(policy.ProviderID) != DefaultHumanVerificationProviderID {
		return false
	}
	return errors.Is(err, errHumanVerificationProviderMissing) || errors.Is(err, errHumanVerificationProviderDisabled)
}

func (s *HumanVerificationService) providerPluginEnabled(providerID string) bool {
	if s.pluginSvc == nil || s.pluginSvc.repo == nil {
		return false
	}
	record, err := s.pluginSvc.repo.GetPluginRegistry(providerID)
	if err != nil {
		return false
	}
	return record.Enabled && record.LifecycleState == pluginapi.PluginStateEnabled
}

func (s *HumanVerificationService) manifestDeclaresHumanVerification(providerID string) bool {
	if s.pluginSvc == nil || s.pluginSvc.registry == nil {
		return false
	}
	manifest, ok := s.pluginSvc.registry.GetManifest(providerID)
	if !ok {
		return false
	}
	return containsString(manifest.Capabilities.Security, HumanVerificationCapability)
}

func (s *HumanVerificationService) humanVerificationProvider(providerID string) (humanVerificationProvider, bool) {
	if s == nil || s.pluginSvc == nil || s.pluginSvc.registry == nil {
		return nil, false
	}
	manifest, ok := s.pluginSvc.registry.GetManifest(providerID)
	if !ok {
		return nil, false
	}
	return s.providerFromManifest(manifest)
}

func (s *HumanVerificationService) humanVerificationProviders() []humanVerificationProvider {
	if s == nil || s.pluginSvc == nil || s.pluginSvc.registry == nil {
		return nil
	}
	providers := make([]humanVerificationProvider, 0)
	for _, manifest := range s.pluginSvc.registry.ListManifests() {
		provider, ok := s.providerFromManifest(manifest)
		if ok {
			providers = append(providers, provider)
		}
	}
	sort.Slice(providers, func(i, j int) bool {
		return providers[i].DisplayName() < providers[j].DisplayName()
	})
	return providers
}

func (s *HumanVerificationService) providerFromManifest(manifest pluginapi.Manifest) (humanVerificationProvider, bool) {
	if strings.TrimSpace(manifest.ID) == "" || !containsString(manifest.Capabilities.Security, HumanVerificationCapability) {
		return nil, false
	}
	hv, ok := manifest.Extensions["humanVerification"].(map[string]any)
	if !ok {
		return nil, false
	}
	providerID := extensionString(hv, "providerId")
	if providerID == "" {
		providerID = manifest.ID
	}
	providerType := extensionString(hv, "providerType")
	displayName := extensionString(hv, "displayName")
	if displayName == "" {
		displayName = manifest.Identity.DisplayName
	}
	if displayName == "" {
		displayName = providerID
	}
	supportedScopes := extensionStringSlice(hv["supportedScopes"])
	if len(supportedScopes) == 0 {
		return nil, false
	}
	return &pluginProcessHumanVerificationProvider{
		service:         s,
		manifest:        manifest,
		providerID:      providerID,
		providerType:    providerType,
		displayName:     displayName,
		renderMode:      extensionString(hv, "renderMode"),
		supportedScopes: supportedScopes,
	}, true
}

func (s *HumanVerificationService) providerRenderMode(providerID string) string {
	if s.pluginSvc == nil || s.pluginSvc.registry == nil {
		return "plugin_iframe"
	}
	manifest, ok := s.pluginSvc.registry.GetManifest(providerID)
	if !ok {
		return "plugin_iframe"
	}
	hv, ok := manifest.Extensions["humanVerification"].(map[string]any)
	if !ok {
		return "plugin_iframe"
	}
	if renderMode, ok := hv["renderMode"].(string); ok && renderMode != "" {
		return renderMode
	}
	return "plugin_iframe"
}

func (s *HumanVerificationService) providerFrontendURL(providerID string) string {
	hv, ok := s.humanVerificationExtension(providerID)
	if !ok {
		return ""
	}
	if frontendURL := extensionString(hv, "frontendUrl"); frontendURL != "" {
		return frontendURL
	}
	return ""
}

func (s *HumanVerificationService) providerFrontendHeight(providerID string) int {
	hv, ok := s.humanVerificationExtension(providerID)
	if !ok {
		return 96
	}
	height := intFromAny(hv["frontendHeight"], 96)
	if height < 48 {
		return 48
	}
	if height > 420 {
		return 420
	}
	return height
}

func (s *HumanVerificationService) humanVerificationExtension(providerID string) (map[string]any, bool) {
	if s.pluginSvc == nil || s.pluginSvc.registry == nil {
		return nil, false
	}
	manifest, ok := s.pluginSvc.registry.GetManifest(providerID)
	if !ok {
		return nil, false
	}
	hv, ok := manifest.Extensions["humanVerification"].(map[string]any)
	return hv, ok
}

func (s *HumanVerificationService) pluginConfig(pluginID string) map[string]any {
	if s.governance == nil || s.governance.repo == nil {
		return map[string]any{}
	}
	var record model.PluginConfigValue
	err := s.governance.repo.GetDB().
		Where("plugin_id = ? AND config_key = ?", strings.TrimSpace(pluginID), "default").
		First(&record).Error
	if err != nil {
		return map[string]any{}
	}
	values := map[string]any{}
	if record.ValueJSON != "" {
		_ = json.Unmarshal([]byte(record.ValueJSON), &values)
	}
	if record.SecretJSON != "" {
		mergeSecretConfig(values, record.SecretJSON)
	}
	return values
}

func mergeSecretConfig(values map[string]any, secretJSON string) {
	secretJSON = strings.TrimSpace(secretJSON)
	if secretJSON == "" {
		return
	}
	secretValues := map[string]any{}
	if err := json.Unmarshal([]byte(secretJSON), &secretValues); err == nil {
		for key, value := range secretValues {
			if strings.TrimSpace(key) == "" {
				continue
			}
			if _, exists := values[key]; !exists {
				values[key] = value
			}
		}
		return
	}
	if _, exists := values["secret_key"]; !exists {
		values["secret_key"] = secretJSON
	}
}

func supportsScope(scopes []string, scope string) bool {
	for _, item := range scopes {
		if item == scope {
			return true
		}
	}
	return false
}

func supportsAllScopes(providerScopes []string, requiredScopes []string) bool {
	for _, scope := range requiredScopes {
		if !supportsScope(providerScopes, scope) {
			return false
		}
	}
	return true
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func firstScope(scopes []string) string {
	if len(scopes) == 0 {
		return HumanScopeAdminLogin
	}
	return scopes[0]
}

func stringFromAny(value any) string {
	switch item := value.(type) {
	case string:
		return strings.TrimSpace(item)
	case fmt.Stringer:
		return strings.TrimSpace(item.String())
	default:
		return ""
	}
}

func intFromAny(value any, fallback int) int {
	switch item := value.(type) {
	case int:
		return item
	case int64:
		return int(item)
	case float64:
		return int(item)
	case string:
		var result int
		if _, err := fmt.Sscanf(item, "%d", &result); err == nil {
			return result
		}
	}
	return fallback
}
