package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// DiscoveryResult 是一次插件发现的结果。
type DiscoveryResult struct {
	PluginID     string   `json:"pluginId"`
	Version      string   `json:"version"`
	ManifestPath string   `json:"manifestPath"`
	InstallRoot  string   `json:"installRoot"`
	State        string   `json:"state"`
	Errors       []string `json:"errors"`
	Warnings     []string `json:"warnings"`
	Manifest     Manifest `json:"manifest"`
}

// DiscoverOptions 控制插件发现行为。
type DiscoverOptions struct {
	PluginRoot string
}

// DiscoverManifests 扫描标准插件目录并读取 manifest.json。
func DiscoverManifests(ctx context.Context, options DiscoverOptions) ([]DiscoveryResult, error) {
	if strings.TrimSpace(options.PluginRoot) == "" {
		return nil, errors.New("插件根目录不能为空")
	}

	if err := os.MkdirAll(options.PluginRoot, 0755); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(options.PluginRoot)
	if err != nil {
		return nil, err
	}

	results := make([]DiscoveryResult, 0)
	for _, entry := range entries {
		if ctx.Err() != nil {
			return results, ctx.Err()
		}
		if !entry.IsDir() {
			continue
		}

		pluginDir := filepath.Join(options.PluginRoot, entry.Name())
		results = append(results, discoverPluginDir(pluginDir)...)
	}

	return results, nil
}

func discoverPluginDir(pluginDir string) []DiscoveryResult {
	releasesDir := filepath.Join(pluginDir, "releases")
	releaseEntries, err := os.ReadDir(releasesDir)
	if err != nil {
		return []DiscoveryResult{{
			PluginID:    filepath.Base(pluginDir),
			InstallRoot: pluginDir,
			State:       PluginStateDiscovered,
			Errors:      []string{"缺少 releases 目录或无法读取版本目录"},
		}}
	}

	results := make([]DiscoveryResult, 0, len(releaseEntries))
	for _, release := range releaseEntries {
		if !release.IsDir() {
			continue
		}
		releaseRoot := filepath.Join(releasesDir, release.Name())
		result := readManifest(releaseRoot)
		result.InstallRoot = pluginDir
		results = append(results, result)
	}
	return results
}

func readManifest(releaseRoot string) DiscoveryResult {
	manifestPath := filepath.Join(releaseRoot, "manifest.json")
	result := DiscoveryResult{
		ManifestPath: manifestPath,
		State:        PluginStateDiscovered,
		Warnings:     make([]string, 0),
		Errors:       make([]string, 0),
	}

	content, err := os.ReadFile(manifestPath)
	if err != nil {
		result.Errors = append(result.Errors, "无法读取 manifest.json")
		return result
	}

	var manifest Manifest
	if err := json.Unmarshal(content, &manifest); err != nil {
		result.Errors = append(result.Errors, "manifest.json 解析失败")
		return result
	}

	result.Manifest = manifest
	result.PluginID = manifest.ID
	result.Version = manifest.Version
	result.Errors = append(result.Errors, ValidateManifest(manifest, releaseRoot)...)
	result.Warnings = append(result.Warnings, inspectOptionalFiles(releaseRoot, manifest)...)
	return result
}

// ValidateManifest 执行安装阶段的最小必校验。
func ValidateManifest(manifest Manifest, releaseRoot string) []string {
	var errs []string
	if strings.TrimSpace(manifest.ManifestVersion) == "" {
		errs = append(errs, "缺少 manifestVersion")
	}
	if strings.TrimSpace(manifest.ID) == "" {
		errs = append(errs, "缺少插件ID")
	}
	if strings.TrimSpace(manifest.Version) == "" {
		errs = append(errs, "缺少插件版本")
	}
	if strings.TrimSpace(manifest.PluginKind) == "" {
		errs = append(errs, "缺少插件类型")
	}
	if strings.TrimSpace(manifest.Identity.Name) == "" && strings.TrimSpace(manifest.Identity.DisplayName) == "" {
		errs = append(errs, "缺少插件显示名称")
	}
	if strings.TrimSpace(manifest.Compatibility.BackendProtocol) == "" {
		errs = append(errs, "缺少后端协议版本")
	}
	if len(manifest.Compatibility.SupportedPlatforms) == 0 {
		errs = append(errs, "缺少支持平台声明")
	} else if !containsString(manifest.Compatibility.SupportedPlatforms, runtime.GOOS) && !containsString(manifest.Compatibility.SupportedPlatforms, "all") {
		errs = append(errs, "当前平台不在插件支持列表中")
	}
	if strings.TrimSpace(manifest.Package.PackageFormat) == "" {
		errs = append(errs, "缺少插件包格式声明")
	}
	if strings.TrimSpace(manifest.Integrity.HashAlgorithm) == "" {
		errs = append(errs, "缺少哈希算法声明")
	}
	if strings.TrimSpace(manifest.Integrity.ChecksumFile) == "" {
		errs = append(errs, "缺少 checksum 文件声明")
	}

	if requiresBackendBinary(manifest) && !hasPlatformBinary(manifest) {
		errs = append(errs, "当前平台缺少可用插件二进制声明")
	}

	checksumFile := manifest.Integrity.ChecksumFile
	if checksumFile == "" {
		checksumFile = "checksums.json"
	}
	if _, err := os.Stat(filepath.Join(releaseRoot, checksumFile)); err != nil {
		errs = append(errs, "缺少 checksums.json")
	}

	return errs
}

func inspectOptionalFiles(releaseRoot string, manifest Manifest) []string {
	var warnings []string
	if manifest.Integrity.SignatureTrustPolicy != "" && manifest.Integrity.SignatureFile != "" {
		if _, err := os.Stat(filepath.Join(releaseRoot, manifest.Integrity.SignatureFile)); err != nil {
			warnings = append(warnings, "签名策略已声明但未找到签名文件")
		}
	}
	if manifest.Backend.SettingsRef != "" {
		if _, err := os.Stat(filepath.Join(releaseRoot, manifest.Backend.SettingsRef)); err != nil {
			warnings = append(warnings, "已声明配置 schema 但未找到对应文件")
		}
	}
	return warnings
}

func requiresBackendBinary(manifest Manifest) bool {
	if manifest.Backend.EntryExecutable != "" {
		return true
	}
	return len(manifest.Backend.Routes) > 0 ||
		len(manifest.Backend.Webhooks) > 0 ||
		len(manifest.Backend.Events) > 0 ||
		len(manifest.Backend.Jobs) > 0
}

func hasPlatformBinary(manifest Manifest) bool {
	if len(manifest.Package.Binaries) == 0 {
		return false
	}
	for _, binary := range manifest.Package.Binaries {
		platformOK := binary.Platform == runtime.GOOS || binary.Platform == "all"
		archOK := binary.Arch == runtime.GOARCH || binary.Arch == "all"
		if platformOK && archOK && binary.Path != "" {
			return true
		}
	}
	return false
}

func containsString(items []string, expected string) bool {
	for _, item := range items {
		if strings.EqualFold(item, expected) {
			return true
		}
	}
	return false
}

// FindStandaloneManifests 支持开发期直接扫描散装插件根目录。
func FindStandaloneManifests(root string) ([]string, error) {
	paths := make([]string, 0)
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Base(path) == "manifest.json" {
			paths = append(paths, path)
		}
		return nil
	})
	return paths, err
}
