package service_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"user-frontend/internal/model"
	pluginapi "user-frontend/internal/plugin"
	"user-frontend/internal/service"
	"user-frontend/internal/test"
)

func TestPluginRuntimeServiceResolveBinary(t *testing.T) {
	services, cleanup := test.SetupTestServices(t)
	defer cleanup()

	releaseDir := t.TempDir()
	binaryPath := filepath.Join(releaseDir, "diagnostic.bin")
	if err := os.WriteFile(binaryPath, []byte("diagnostic"), 0755); err != nil {
		t.Fatalf("写入测试二进制失败: %v", err)
	}
	hash, err := fileHashForTest(binaryPath)
	if err != nil {
		t.Fatalf("计算测试二进制 hash 失败: %v", err)
	}

	runtimeSvc := service.NewPluginRuntimeService(services.Repo, nil, nil)
	manifest := pluginapi.Manifest{
		ID:      "diagnostic",
		Version: "1.0.0",
		Package: pluginapi.PackageInfo{Binaries: []pluginapi.BinaryInfo{
			{
				Platform: runtime.GOOS,
				Arch:     runtime.GOARCH,
				Path:     "diagnostic.bin",
				Extensions: pluginapi.ExtensionMap{
					"sha256": hash,
				},
			},
		}},
	}
	resolved, err := runtimeSvc.ResolveBinary(manifest, releaseDir)
	if err != nil {
		t.Fatalf("解析插件二进制失败: %v", err)
	}
	if resolved.Path != binaryPath || resolved.SHA256 != hash {
		t.Fatalf("解析结果不符合预期: %#v", resolved)
	}

	manifest.Package.Binaries[0].Extensions["sha256"] = "bad"
	if _, err := runtimeSvc.ResolveBinary(manifest, releaseDir); err == nil {
		t.Fatal("hash 不匹配时应拒绝插件二进制")
	}
}

func TestPluginRuntimeServiceInvocationGate(t *testing.T) {
	services, cleanup := test.SetupTestServices(t)
	defer cleanup()

	govSvc := service.NewGovernanceService(services.Repo)
	runtimeSvc := service.NewPluginRuntimeService(services.Repo, nil, govSvc)
	session := model.PluginRuntimeSession{
		PluginID:   "diagnostic",
		Version:    "1.0.0",
		InstanceID: "pin-test",
		PID:        1,
		State:      service.PluginRuntimeStateStarting,
		StartedAt:  testNow(),
	}
	if err := services.DB.Create(&session).Error; err != nil {
		t.Fatalf("创建运行会话失败: %v", err)
	}

	result, err := runtimeSvc.Invoke(context.Background(), service.PluginInvocationRequest{PluginID: "diagnostic", RouteID: "health"})
	if err != nil {
		t.Fatalf("未 Ready 调用不应返回系统错误: %v", err)
	}
	if result.Accepted {
		t.Fatal("未 Ready 插件不应通过调用门禁")
	}

	if err := runtimeSvc.MarkReady("diagnostic", "pin-test"); err != nil {
		t.Fatalf("标记 Ready 失败: %v", err)
	}
	result, err = runtimeSvc.Invoke(context.Background(), service.PluginInvocationRequest{PluginID: "diagnostic", RouteID: "health"})
	if err != nil {
		t.Fatalf("Ready 后调用门禁失败: %v", err)
	}
	if !result.Accepted {
		t.Fatalf("Ready 后应通过调用门禁: %#v", result)
	}
}

func fileHashForTest(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return serviceTestSHA256(data), nil
}

func serviceTestSHA256(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func testNow() time.Time {
	return time.Date(2026, 6, 22, 1, 0, 0, 0, time.UTC)
}
