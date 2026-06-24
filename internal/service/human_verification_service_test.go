package service_test

import (
	"context"
	"testing"

	"user-frontend/internal/config"
	"user-frontend/internal/service"
)

// TestHumanVerification_DefaultProviderUnavailableFallsBackDisabled 验证旧默认开启配置不会在插件缺失时阻断登录。
func TestHumanVerification_DefaultProviderUnavailableFallsBackDisabled(t *testing.T) {
	restoreGlobalConfig(t, config.ServerConfig{
		EnableLogin:                          true,
		AdminHumanVerificationEnabled:        true,
		AdminHumanVerificationProviderID:     service.DefaultHumanVerificationProviderID,
		UserLoginHumanVerificationEnabled:    true,
		UserLoginHumanVerificationProviderID: service.DefaultHumanVerificationProviderID,
	})

	humanSvc := service.NewHumanVerificationService(nil, nil)

	publicConfig := humanSvc.PublicConfigForScope(service.HumanScopeAdminLogin)
	if publicConfig.Enabled {
		t.Fatalf("默认人机验证插件缺失时，公开配置应降级为关闭")
	}
	if publicConfig.Error != "" {
		t.Fatalf("默认人机验证插件缺失时，不应向认证页暴露错误: %s", publicConfig.Error)
	}

	challenge, err := humanSvc.CreateChallenge(context.Background(), service.HumanVerificationChallengeRequest{
		Scope: service.HumanScopeAdminLogin,
	})
	if err != nil {
		t.Fatalf("默认人机验证插件缺失时，创建挑战不应失败: %v", err)
	}
	if challenge == nil || challenge.Scope != service.HumanScopeAdminLogin {
		t.Fatalf("默认人机验证插件缺失时，应返回当前场景的空挑战")
	}

	if err := humanSvc.Verify(context.Background(), service.HumanVerificationVerifyRequest{
		Scope: service.HumanScopeAdminLogin,
	}); err != nil {
		t.Fatalf("默认人机验证插件缺失时，登录校验不应阻断: %v", err)
	}
}

// TestHumanVerification_CustomProviderUnavailableStillErrors 验证显式配置的非默认 provider 不会被静默关闭。
func TestHumanVerification_CustomProviderUnavailableStillErrors(t *testing.T) {
	restoreGlobalConfig(t, config.ServerConfig{
		EnableLogin:                          true,
		AdminHumanVerificationEnabled:        true,
		AdminHumanVerificationProviderID:     "custom.turnstile",
		UserLoginHumanVerificationEnabled:    true,
		UserLoginHumanVerificationProviderID: "custom.turnstile",
	})

	humanSvc := service.NewHumanVerificationService(nil, nil)

	publicConfig := humanSvc.PublicConfigForScope(service.HumanScopeAdminLogin)
	if !publicConfig.Enabled {
		t.Fatalf("非默认 provider 缺失时不应静默关闭")
	}
	if publicConfig.Error == "" {
		t.Fatalf("非默认 provider 缺失时应保留错误提示")
	}

	if err := humanSvc.Verify(context.Background(), service.HumanVerificationVerifyRequest{
		Scope: service.HumanScopeAdminLogin,
	}); err == nil {
		t.Fatalf("非默认 provider 缺失时，登录校验应返回错误")
	}
}

func restoreGlobalConfig(t *testing.T, serverConfig config.ServerConfig) {
	t.Helper()
	previous := config.GlobalConfig
	config.GlobalConfig = &config.Config{ServerConfig: serverConfig}
	t.Cleanup(func() {
		config.GlobalConfig = previous
	})
}
