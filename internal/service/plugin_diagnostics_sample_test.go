package service_test

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"

	"user-frontend/internal/model"
	"user-frontend/internal/service"
	"user-frontend/internal/test"
)

func TestDiagnosticsSamplePluginDiscoveryAndGovernance(t *testing.T) {
	services, cleanup := test.SetupTestServices(t)
	defer cleanup()

	pluginRoot := filepath.Clean(filepath.Join("..", "..", "plugins"))
	pluginSvc := service.NewPluginService(services.Repo, pluginRoot)
	pluginSvc.SetGovernanceService(services.GovernanceSvc)

	results, err := pluginSvc.Refresh(context.Background())
	if err != nil {
		t.Fatalf("刷新示例插件失败: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("未发现示例诊断插件")
	}

	detail, ok := pluginSvc.GetPluginDetail("example-diagnostics")
	if !ok || detail == nil {
		t.Fatal("示例诊断插件未进入注册中心")
	}

	var permissions []model.PermissionDefinition
	if err := services.DB.Where("owner_plugin_id = ?", "example-diagnostics").Find(&permissions).Error; err != nil {
		t.Fatalf("读取插件权限定义失败: %v", err)
	}
	if len(permissions) != 2 {
		t.Fatalf("插件权限定义数量不正确: %d", len(permissions))
	}

	var declaration model.PluginDatabaseDeclaration
	if err := services.DB.Where("plugin_id = ?", "example-diagnostics").First(&declaration).Error; err != nil {
		t.Fatalf("读取插件数据库声明失败: %v", err)
	}
	if declaration.Namespace != "example_diagnostics" || declaration.TableCount != 1 {
		t.Fatalf("插件数据库声明不正确: %#v", declaration)
	}

	var trust model.PluginTrustRecord
	if err := services.DB.Where("plugin_id = ? AND version = ?", "example-diagnostics", "1.0.0").First(&trust).Error; err != nil {
		t.Fatalf("读取插件信任记录失败: %v", err)
	}
	if trust.TrustLevel != "local-approved" {
		t.Fatalf("示例插件信任状态不正确: %s", trust.TrustLevel)
	}

	manifest, ok := pluginSvc.Registry().GetManifest("example-diagnostics")
	if !ok {
		t.Fatal("示例插件 manifest 不存在")
	}
	releaseDir := filepath.Join(pluginRoot, "example-diagnostics", "releases", "1.0.0")
	runtimeSvc := service.NewPluginRuntimeService(services.Repo, pluginSvc, services.GovernanceSvc)
	resolved, err := runtimeSvc.ResolveBinary(manifest, releaseDir)
	if err != nil {
		t.Fatalf("解析示例插件二进制失败: %v", err)
	}
	if resolved.Platform != runtime.GOOS && resolved.Platform != "all" {
		t.Fatalf("二进制平台选择不正确: %#v", resolved)
	}
}
