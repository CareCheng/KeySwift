package service_test

import (
	"testing"

	"user-frontend/internal/model"
	"user-frontend/internal/repository"
	"user-frontend/internal/service"
	"user-frontend/internal/test"
)

// TestConfigService_SetInitialPassword_AllowsDefaultPassword 测试首次设置允许用户主动选择默认种子密码。
func TestConfigService_SetInitialPassword_AllowsDefaultPassword(t *testing.T) {
	db, cleanup := test.SetupTestDB(t)
	defer cleanup()

	configSvc := service.NewConfigService(repository.NewRepository(db))

	if !configSvc.NeedsInitialSetup() {
		t.Fatal("种子配置应标记为需要首次设置")
	}

	if err := configSvc.SetInitialPassword("admin123"); err != nil {
		t.Fatalf("首次设置为 admin123 不应失败: %v", err)
	}

	if configSvc.NeedsInitialSetup() {
		t.Fatal("首次设置成功后不应继续提示初始化")
	}

	cfg, err := configSvc.GetSystemConfig()
	if err != nil {
		t.Fatalf("读取系统配置失败: %v", err)
	}
	if cfg.AdminPassword != "admin123" {
		t.Fatalf("管理员密码保存不符合预期: %s", cfg.AdminPassword)
	}
	if !cfg.AdminPasswordInitialized {
		t.Fatal("首次设置成功后应持久化初始化完成状态")
	}

	if err := configSvc.SetInitialPassword("another123"); err == nil {
		t.Fatal("首次设置完成后不应允许重复设置")
	}
}

// TestConfigService_HumanVerificationDefaultsOff 测试初始化种子配置默认关闭人机验证。
func TestConfigService_HumanVerificationDefaultsOff(t *testing.T) {
	db, cleanup := test.SetupTestDB(t)
	defer cleanup()

	configSvc := service.NewConfigService(repository.NewRepository(db))
	cfg, err := configSvc.GetSystemConfig()
	if err != nil {
		t.Fatalf("读取系统配置失败: %v", err)
	}
	if cfg.AdminHumanVerificationEnabled ||
		cfg.UserLoginHumanVerificationEnabled ||
		cfg.UserRegisterHumanVerificationEnabled {
		t.Fatal("种子配置应默认关闭管理员、用户登录和用户注册人机验证")
	}
}

// TestConfigService_HumanVerificationLegacyFlagsIgnored 测试未显式配置时旧库残留开关不会让人机验证默认开启。
func TestConfigService_HumanVerificationLegacyFlagsIgnored(t *testing.T) {
	db, cleanup := test.SetupTestDB(t)
	defer cleanup()

	if err := db.Model(&model.SystemConfigDB{}).Where("id = ?", 1).Updates(map[string]any{
		"admin_human_verification_enabled":         true,
		"user_login_human_verification_enabled":    true,
		"user_register_human_verification_enabled": true,
	}).Error; err != nil {
		t.Fatalf("准备旧库残留配置失败: %v", err)
	}

	configSvc := service.NewConfigService(repository.NewRepository(db))
	cfg, err := configSvc.GetSystemConfig()
	if err != nil {
		t.Fatalf("读取系统配置失败: %v", err)
	}
	if cfg.AdminHumanVerificationEnabled ||
		cfg.UserLoginHumanVerificationEnabled ||
		cfg.UserRegisterHumanVerificationEnabled {
		t.Fatal("未显式配置人机验证时应默认关闭管理员、用户登录和用户注册人机验证")
	}
}
