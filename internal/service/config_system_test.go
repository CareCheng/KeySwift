package service_test

import (
	"testing"

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
