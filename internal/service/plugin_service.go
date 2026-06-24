package service

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"user-frontend/internal/model"
	pluginapi "user-frontend/internal/plugin"
	"user-frontend/internal/repository"

	"gorm.io/gorm"
)

// PluginService 管理宿主插件发现、注册表和运行时能力快照。
type PluginService struct {
	repo       *repository.Repository
	pluginRoot string
	registry   *pluginapi.MemoryRegistry
	governance *GovernanceService
}

// PluginDatabaseSnapshot 是后台读取插件数据库治理状态的完整快照。
type PluginDatabaseSnapshot struct {
	Declaration *model.PluginDatabaseDeclaration `json:"declaration"`
	Tables      []model.PluginDatabaseTable      `json:"tables"`
	Columns     []model.PluginDatabaseColumn     `json:"columns"`
	Indexes     []model.PluginDatabaseIndex      `json:"indexes"`
	Relations   []model.PluginDatabaseRelation   `json:"relations"`
	Operations  []model.PluginDatabaseOperation  `json:"operations"`
}

// PluginInstallResult 描述一次后台安装插件包后的刷新结果。
type PluginInstallResult struct {
	InstalledRoots []string                    `json:"installed_roots"`
	Results        []pluginapi.DiscoveryResult `json:"results"`
}

// PluginUninstallOptions 描述插件卸载时的用户选择。
type PluginUninstallOptions struct {
	DeleteConfig  bool `json:"delete_config"`
	DeleteDatabase bool `json:"delete_database"`
}

// PluginUninstallResult 描述插件卸载执行结果。
type PluginUninstallResult struct {
	PluginID         string `json:"plugin_id"`
	RemovedDirectory  bool   `json:"removed_directory"`
	RemovedConfig     bool   `json:"removed_config"`
	RemovedDatabase   bool   `json:"removed_database"`
	RemovedState      bool   `json:"removed_state"`
}

// NewPluginService 创建插件服务。
func NewPluginService(repo *repository.Repository, pluginRoot string) *PluginService {
	return &PluginService{
		repo:       repo,
		pluginRoot: pluginRoot,
		registry:   pluginapi.NewMemoryRegistry(),
	}
}

// Registry 返回宿主内存注册中心。
func (s *PluginService) Registry() *pluginapi.MemoryRegistry {
	return s.registry
}

// PluginRoot 返回标准插件目录。
func (s *PluginService) PluginRoot() string {
	return s.pluginRoot
}

// InstallRoot 返回插件实际安装目录；优先使用最近一次扫描登记的目录。
func (s *PluginService) InstallRoot(pluginID string) string {
	pluginID = strings.TrimSpace(pluginID)
	if s == nil || pluginID == "" {
		return ""
	}
	if s.repo != nil {
		if record, err := s.repo.GetPluginRegistry(pluginID); err == nil {
			if root := strings.TrimSpace(record.InstallRoot); root != "" {
				if info, statErr := os.Stat(root); statErr == nil && info.IsDir() {
					return filepath.Clean(root)
				}
			}
		}
	}
	return filepath.Clean(filepath.Join(s.pluginRoot, pluginID))
}

// ReleaseRoot 返回插件指定版本的实际 release 目录。
func (s *PluginService) ReleaseRoot(pluginID string, version string) string {
	return filepath.Join(s.InstallRoot(pluginID), "releases", strings.TrimSpace(version))
}

// DataRoot 返回插件运行数据目录。
func (s *PluginService) DataRoot(pluginID string) string {
	return filepath.Join(s.InstallRoot(pluginID), "data")
}

// SetGovernanceService 绑定插件治理落库服务。
func (s *PluginService) SetGovernanceService(governance *GovernanceService) {
	s.governance = governance
}

// Refresh 扫描磁盘插件目录，注册 manifest 并同步治理表。
func (s *PluginService) Refresh(ctx context.Context) ([]pluginapi.DiscoveryResult, error) {
	results, err := pluginapi.DiscoverManifests(ctx, pluginapi.DiscoverOptions{PluginRoot: s.pluginRoot})
	if err != nil {
		return nil, err
	}

	for i := range results {
		result := &results[i]
		if len(result.Errors) > 0 || result.Manifest.ID == "" {
			s.saveDiscoveryError(*result)
			continue
		}

		manifest := result.Manifest
		configSchema, hasConfigSchema, err := s.loadConfigSchema(*result)
		if err != nil {
			result.Errors = append(result.Errors, err.Error())
			s.saveDiscoveryError(*result)
			continue
		}
		if err := s.registry.RegisterManifest(ctx, manifest); err != nil {
			result.Errors = append(result.Errors, err.Error())
			s.saveDiscoveryError(*result)
			continue
		}
		if hasConfigSchema {
			if err := s.registry.RegisterConfigSchema(manifest.ID, configSchema); err != nil {
				result.Errors = append(result.Errors, err.Error())
				s.saveDiscoveryError(*result)
				continue
			}
		}
		s.registry.SetRuntime(manifest.ID, pluginapi.RuntimePlugin{
			PluginID:                manifest.ID,
			InstallID:               buildInstallID(manifest),
			State:                   pluginapi.PluginRuntimeStopped,
			TrafficEnabled:          false,
			SelectedProtocolVersion: pluginapi.BackendProtocol,
			ControlTransport:        manifest.Backend.ControlProtocol,
			DataTransport:           manifest.Backend.DataProtocol,
			TrustLevel:              "local-approved",
			IntegrityState:          "unchecked",
			ConfigVersion:           1,
			Extensions:              pluginapi.ExtensionMap{},
		})
		s.persistManifest(*result)
		s.persistBindings(manifest)
		s.persistMigrations(manifest)
		s.persistDatabaseDeclarations(manifest)
		s.persistGovernanceDeclarations(*result)
	}

	return results, nil
}

func (s *PluginService) loadConfigSchema(result pluginapi.DiscoveryResult) (pluginapi.ConfigSchema, bool, error) {
	settingsRef := strings.TrimSpace(result.Manifest.Backend.SettingsRef)
	if settingsRef == "" {
		return pluginapi.ConfigSchema{}, false, nil
	}
	cleanRef := filepath.Clean(settingsRef)
	if filepath.IsAbs(cleanRef) || strings.HasPrefix(cleanRef, "..") {
		return pluginapi.ConfigSchema{}, false, errors.New("插件配置 schema 路径不合法")
	}
	releaseRoot := filepath.Dir(result.ManifestPath)
	content, err := os.ReadFile(filepath.Join(releaseRoot, cleanRef))
	if err != nil {
		return pluginapi.ConfigSchema{}, false, errors.New("无法读取插件配置 schema")
	}
	var schema pluginapi.ConfigSchema
	if err := json.Unmarshal(content, &schema); err != nil {
		return pluginapi.ConfigSchema{}, false, errors.New("插件配置 schema 解析失败")
	}
	if strings.TrimSpace(schema.PluginID) == "" {
		schema.PluginID = result.Manifest.ID
	}
	if schema.PluginID != result.Manifest.ID {
		return pluginapi.ConfigSchema{}, false, errors.New("插件配置 schema 的 pluginId 与 manifest 不一致")
	}
	if strings.TrimSpace(schema.SchemaVersion) == "" {
		schema.SchemaVersion = pluginapi.ManifestVersion
	}
	if strings.TrimSpace(schema.ConfigVersion) == "" {
		schema.ConfigVersion = result.Manifest.Version
	}
	return schema, true, nil
}

// InstallPackage 解压 .ksplugin.zip 安装包到标准插件目录并刷新插件注册表。
func (s *PluginService) InstallPackage(ctx context.Context, packagePath string) (*PluginInstallResult, error) {
	if s == nil {
		return nil, errors.New("插件服务未初始化")
	}
	if strings.TrimSpace(packagePath) == "" {
		return nil, errors.New("插件安装包路径不能为空")
	}
	installedRoots, err := s.extractPluginPackage(packagePath)
	if err != nil {
		return nil, err
	}
	results, err := s.Refresh(ctx)
	if err != nil {
		return nil, err
	}
	return &PluginInstallResult{InstalledRoots: installedRoots, Results: results}, nil
}

// UninstallPlugin 卸载指定插件，并按选项清理配置和独立数据表。
func (s *PluginService) UninstallPlugin(ctx context.Context, pluginID string, options PluginUninstallOptions) (*PluginUninstallResult, error) {
	pluginID = strings.TrimSpace(pluginID)
	if pluginID == "" {
		return nil, errors.New("插件ID不能为空")
	}
	if s == nil || s.repo == nil {
		return nil, errors.New("插件仓储未初始化")
	}
	if ctx != nil && ctx.Err() != nil {
		return nil, ctx.Err()
	}

	record, err := s.repo.GetPluginRegistry(pluginID)
	if err != nil {
		return nil, err
	}

	installRoot := strings.TrimSpace(record.InstallRoot)
	if installRoot == "" {
		installRoot = s.InstallRoot(pluginID)
	}

	if ctx != nil && ctx.Err() != nil {
		return nil, ctx.Err()
	}

	if err := s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
		if options.DeleteDatabase {
			if err := dropPluginDatabaseTables(tx, pluginID); err != nil {
				return err
			}
		}
		if err := tx.Where("plugin_id = ?", pluginID).Delete(&model.PluginArtifact{}).Error; err != nil {
			return err
		}
		if err := tx.Where("plugin_id = ?", pluginID).Delete(&model.PluginBinding{}).Error; err != nil {
			return err
		}
		if err := tx.Where("plugin_id = ?", pluginID).Delete(&model.PluginMigration{}).Error; err != nil {
			return err
		}
		if err := tx.Where("plugin_id = ?", pluginID).Delete(&model.PluginRegistry{}).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	if s.governance != nil {
		if err := s.governance.DeletePluginGovernanceData(pluginID); err != nil {
			return nil, err
		}
		if options.DeleteConfig {
			if err := s.governance.DeletePluginConfigData(pluginID); err != nil {
				return nil, err
			}
		}
	}

	removedDirectory := false
	if installRoot != "" {
		rootClean := filepath.Clean(s.pluginRoot)
		installClean := filepath.Clean(installRoot)
		if pathWithin(rootClean, installClean) {
			if err := os.RemoveAll(installClean); err != nil {
				return nil, err
			}
			removedDirectory = true
		}
	}

	if s.registry != nil {
		s.registry.UnregisterPlugin(pluginID)
	}

	return &PluginUninstallResult{
		PluginID:        pluginID,
		RemovedDirectory: removedDirectory,
		RemovedConfig:   options.DeleteConfig,
		RemovedDatabase:  options.DeleteDatabase,
		RemovedState:     true,
	}, nil
}

func (s *PluginService) extractPluginPackage(packagePath string) ([]string, error) {
	if err := os.MkdirAll(s.pluginRoot, 0755); err != nil {
		return nil, err
	}

	reader, err := zip.OpenReader(packagePath)
	if err != nil {
		return nil, errors.New("插件安装包无法读取或格式错误")
	}
	defer reader.Close()

	topDirs := map[string]bool{}
	for _, file := range reader.File {
		name := cleanZipEntryName(file.Name)
		if name == "" {
			continue
		}
		top := strings.Split(name, "/")[0]
		if !validPluginTopDirectory(top) {
			return nil, fmt.Errorf("插件安装包顶层目录不合法: %s", top)
		}
		topDirs[top] = true
	}
	if len(topDirs) == 0 {
		return nil, errors.New("插件安装包为空")
	}
	if len(topDirs) > 1 {
		return nil, errors.New("插件安装包只能包含一个插件顶层目录")
	}

	installedRoots := make([]string, 0, len(topDirs))
	for top := range topDirs {
		installedRoots = append(installedRoots, filepath.Join(s.pluginRoot, top))
	}

	rootClean := filepath.Clean(s.pluginRoot)
	for _, file := range reader.File {
		name := cleanZipEntryName(file.Name)
		if name == "" {
			continue
		}
		target := filepath.Join(rootClean, filepath.FromSlash(name))
		targetClean := filepath.Clean(target)
		if !pathWithin(rootClean, targetClean) {
			return nil, errors.New("插件安装包包含非法路径")
		}
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(targetClean, 0755); err != nil {
				return nil, err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(targetClean), 0755); err != nil {
			return nil, err
		}
		src, err := file.Open()
		if err != nil {
			return nil, err
		}
		if err := writeZipFile(targetClean, src, file.FileInfo().Mode()); err != nil {
			_ = src.Close()
			return nil, err
		}
		if err := src.Close(); err != nil {
			return nil, err
		}
	}

	return installedRoots, nil
}

// Summary 返回插件平台概览。
func (s *PluginService) Summary() map[string]any {
	plugins := s.registry.ListManifests()
	frontend := s.registry.FrontendContribution()
	return map[string]any{
		"plugin_root":       s.pluginRoot,
		"protocol_version":  pluginapi.ManifestVersion,
		"backend_protocol":  pluginapi.BackendProtocol,
		"frontend_protocol": pluginapi.FrontendProtocol,
		"plugins":           len(plugins),
		"frontend_pages":    len(frontend.Pages),
		"frontend_menus":    len(frontend.Menus),
		"themes":            len(frontend.Themes),
	}
}

// ListPlugins 返回后台插件列表。
func (s *PluginService) ListPlugins() []map[string]any {
	manifests := s.registry.ListManifests()
	registryRecords := s.registryRecordMap()
	items := make([]map[string]any, 0, len(manifests))
	for _, manifest := range manifests {
		runtimeState := pluginapi.PluginRuntimeStopped
		trafficEnabled := false
		if runtimeInfo, ok := s.registry.GetRuntime(manifest.ID); ok {
			runtimeState = runtimeInfo.State
			trafficEnabled = runtimeInfo.TrafficEnabled
		}

		displayName := manifest.Identity.DisplayName
		if displayName == "" {
			displayName = manifest.Identity.Name
		}
		if displayName == "" {
			displayName = manifest.ID
		}

		enabled := false
		desiredState := pluginapi.PluginStateApprovedDisabled
		lifecycleState := pluginapi.PluginStateDiscovered
		verifyStatus := ""
		healthStatus := runtimeState
		if record, ok := registryRecords[manifest.ID]; ok {
			enabled = record.Enabled
			desiredState = record.DesiredState
			lifecycleState = record.LifecycleState
			verifyStatus = record.LastVerifyStatus
			healthStatus = record.HealthStatus
		}

		items = append(items, map[string]any{
			"id":               manifest.ID,
			"version":          manifest.Version,
			"plugin_kind":      manifest.PluginKind,
			"categories":       normalizeStringList(manifest.Identity.Categories),
			"tags":             normalizeStringList(manifest.Identity.Tags),
			"keywords":         normalizeStringList(manifest.Identity.Keywords),
			"display_name":     displayName,
			"description":      manifest.Identity.Description,
			"author":           manifest.Identity.Author,
			"runtime_state":    runtimeState,
			"traffic_enabled":  trafficEnabled,
			"enabled":          enabled,
			"desired_state":    desiredState,
			"lifecycle_state":  lifecycleState,
			"verify_status":    verifyStatus,
			"health_status":    healthStatus,
			"frontend_enabled": manifest.Frontend.Enabled,
			"theme_enabled":    manifest.UI.Enabled || manifest.PluginKind == pluginapi.PluginKindUITheme,
			"permissions":      len(manifest.Permissions),
			"pages":            len(manifest.Frontend.Pages),
			"menus":            len(manifest.Frontend.Menus),
			"routes":           len(manifest.Backend.Routes),
			"events":           len(manifest.Backend.Events),
			"jobs":             len(manifest.Backend.Jobs),
			"migrations":       len(manifest.Backend.Migrations),
			"database_tables":  len(manifest.Database.Tables),
		})
	}
	return items
}

// GetPluginDetail 返回单个插件完整声明与运行态。
func (s *PluginService) GetPluginDetail(pluginID string) (map[string]any, bool) {
	manifest, ok := s.registry.GetManifest(pluginID)
	if !ok {
		return nil, false
	}
	runtimeInfo, _ := s.registry.GetRuntime(pluginID)
	var registry *model.PluginRegistry
	if s.repo != nil {
		if record, err := s.repo.GetPluginRegistry(pluginID); err == nil {
			registry = record
		}
	}
	return map[string]any{
		"manifest": manifest,
		"runtime":  runtimeInfo,
		"registry": registry,
	}, true
}

// FrontendContribution 返回前端挂载声明。
func (s *PluginService) FrontendContribution() pluginapi.FrontendContribution {
	return s.registry.FrontendContribution()
}

// Permissions 返回插件权限声明。
func (s *PluginService) Permissions() []pluginapi.PermissionDeclaration {
	return s.registry.ListPermissions()
}

// PermissionDefinitionInputs 返回带插件归属的权限定义输入。
func (s *PluginService) PermissionDefinitionInputs() []PermissionDefinitionInput {
	manifests := s.registry.ListManifests()
	inputs := make([]PermissionDefinitionInput, 0)
	for _, manifest := range manifests {
		for _, item := range manifest.Permissions {
			name := item.Title
			if name == "" {
				name = item.Key
			}
			group := item.Namespace
			if group == "" {
				group = "插件权限"
			}
			inputs = append(inputs, PermissionDefinitionInput{
				PermissionCode:     item.Key,
				OwnerType:          PermissionOwnerPlugin,
				OwnerPluginID:      manifest.ID,
				RiskLevel:          defaultString(item.RiskLevel, DefaultPermissionRiskLevel),
				GroupKey:           group,
				Name:               name,
				Description:        item.Description,
				DefaultGrantPolicy: defaultString(item.DefaultVisibility, DefaultGrantPolicyManual),
				Extensions:         item.Extensions,
			})
		}
	}
	return inputs
}

// ConfigSchemas 返回配置 schema 声明。
func (s *PluginService) ConfigSchemas() []pluginapi.ConfigSchema {
	return s.registry.ListConfigSchemas()
}

func (s *PluginService) registryRecordMap() map[string]model.PluginRegistry {
	records := map[string]model.PluginRegistry{}
	if s.repo == nil {
		return records
	}
	items, err := s.repo.ListPluginRegistries()
	if err != nil {
		return records
	}
	for _, item := range items {
		records[item.PluginID] = item
	}
	return records
}

// GetPluginBindings 返回插件绑定声明。
func (s *PluginService) GetPluginBindings(pluginID string) ([]model.PluginBinding, bool) {
	if s.repo == nil {
		return nil, false
	}
	items, err := s.repo.ListPluginBindings(pluginID)
	if err != nil {
		return nil, false
	}
	return items, true
}

// GetPluginMigrations 返回插件迁移声明。
func (s *PluginService) GetPluginMigrations(pluginID string) ([]model.PluginMigration, bool) {
	if s.repo == nil {
		return nil, false
	}
	items, err := s.repo.ListPluginMigrations(pluginID)
	if err != nil {
		return nil, false
	}
	return items, true
}

// GetPluginDatabaseTables 返回插件数据库表声明。
func (s *PluginService) GetPluginDatabaseTables(pluginID string) ([]model.PluginDatabaseTable, bool) {
	if s.repo == nil {
		return nil, false
	}
	items, err := s.repo.ListPluginDatabaseTables(pluginID)
	if err != nil {
		return nil, false
	}
	return items, true
}

// GetPluginDatabaseSnapshot 返回插件数据库声明和治理明细。
func (s *PluginService) GetPluginDatabaseSnapshot(pluginID string) (PluginDatabaseSnapshot, bool) {
	var snapshot PluginDatabaseSnapshot
	if s.repo == nil {
		return snapshot, false
	}

	if declaration, err := s.repo.GetPluginDatabaseDeclaration(pluginID); err == nil {
		snapshot.Declaration = declaration
	}

	tables, err := s.repo.ListPluginDatabaseTables(pluginID)
	if err != nil {
		return snapshot, false
	}
	columns, err := s.repo.ListPluginDatabaseColumns(pluginID)
	if err != nil {
		return snapshot, false
	}
	indexes, err := s.repo.ListPluginDatabaseIndexes(pluginID)
	if err != nil {
		return snapshot, false
	}
	relations, err := s.repo.ListPluginDatabaseRelations(pluginID)
	if err != nil {
		return snapshot, false
	}
	operations, err := s.repo.ListPluginDatabaseOperations(pluginID)
	if err != nil {
		return snapshot, false
	}

	snapshot.Tables = tables
	snapshot.Columns = columns
	snapshot.Indexes = indexes
	snapshot.Relations = relations
	snapshot.Operations = operations
	return snapshot, true
}

// EnablePlugin 更新插件治理状态为已启用。
func (s *PluginService) EnablePlugin(pluginID string) error {
	if strings.TrimSpace(pluginID) == "" {
		return errors.New("插件ID不能为空")
	}
	if s.repo == nil {
		return errors.New("插件仓储未初始化")
	}
	record, err := s.repo.GetPluginRegistry(pluginID)
	if err != nil {
		return err
	}
	if s.governance != nil {
		if err := s.governance.AssertPluginTrusted(record.PluginID, record.CurrentVersion); err != nil {
			return err
		}
	}
	now := time.Now()
	record.Enabled = true
	record.Autostart = true
	record.DesiredState = pluginapi.PluginStateEnabled
	record.LifecycleState = pluginapi.PluginStateEnabled
	record.ActualState = pluginapi.PluginRuntimeStopped
	record.HealthStatus = pluginapi.PluginRuntimeStopped
	record.LastVerifiedAt = &now
	record.LastVerifyStatus = "passed"
	return s.repo.UpsertPluginRegistry(record)
}

// DisablePlugin 更新插件治理状态为已停用。
func (s *PluginService) DisablePlugin(pluginID string) error {
	if strings.TrimSpace(pluginID) == "" {
		return errors.New("插件ID不能为空")
	}
	if s.repo == nil {
		return errors.New("插件仓储未初始化")
	}
	record, err := s.repo.GetPluginRegistry(pluginID)
	if err != nil {
		return err
	}
	now := time.Now()
	record.Enabled = false
	record.Autostart = false
	record.DesiredState = pluginapi.PluginStateApprovedDisabled
	record.LifecycleState = pluginapi.PluginStateApprovedDisabled
	record.ActualState = pluginapi.PluginRuntimeStopped
	record.HealthStatus = pluginapi.PluginRuntimeStopped
	record.LastStopAt = &now
	return s.repo.UpsertPluginRegistry(record)
}

// ListEnabledFrontendContribution 返回仅启用插件的前端贡献。
func (s *PluginService) ListEnabledFrontendContribution() pluginapi.FrontendContribution {
	if s.repo == nil {
		return s.registry.FrontendContribution()
	}

	enabled := map[string]bool{}
	records, err := s.repo.ListPluginRegistries()
	if err == nil {
		for _, record := range records {
			if record.Enabled {
				enabled[record.PluginID] = true
			}
		}
	}

	frontend := pluginapi.FrontendContribution{
		ProtocolVersion: pluginapi.FrontendProtocol,
		Pages:           make([]pluginapi.PageDeclaration, 0),
		Menus:           make([]pluginapi.MenuDeclaration, 0),
		Forms:           make([]pluginapi.FormDeclaration, 0),
		Views:           make([]pluginapi.ViewDeclaration, 0),
		Themes:          make([]pluginapi.UIContribution, 0),
		Extensions:      pluginapi.ExtensionMap{},
	}

	for _, manifest := range s.registry.ListManifests() {
		if !enabled[manifest.ID] {
			continue
		}
		if manifest.Frontend.Enabled {
			frontend.Pages = append(frontend.Pages, manifest.Frontend.Pages...)
			frontend.Menus = append(frontend.Menus, manifest.Frontend.Menus...)
			frontend.Forms = append(frontend.Forms, manifest.Frontend.Forms...)
			frontend.Views = append(frontend.Views, manifest.Frontend.Views...)
		}
		if manifest.UI.Enabled || manifest.PluginKind == pluginapi.PluginKindUITheme {
			frontend.Themes = append(frontend.Themes, manifest.UI)
		}
	}

	sort.Slice(frontend.Pages, func(i, j int) bool { return frontend.Pages[i].ID < frontend.Pages[j].ID })
	sort.Slice(frontend.Menus, func(i, j int) bool { return frontend.Menus[i].ID < frontend.Menus[j].ID })
	sort.Slice(frontend.Forms, func(i, j int) bool { return frontend.Forms[i].ID < frontend.Forms[j].ID })
	sort.Slice(frontend.Views, func(i, j int) bool { return frontend.Views[i].ID < frontend.Views[j].ID })
	sort.Slice(frontend.Themes, func(i, j int) bool { return frontend.Themes[i].ThemeScope < frontend.Themes[j].ThemeScope })

	return frontend
}

func (s *PluginService) saveDiscoveryError(result pluginapi.DiscoveryResult) {
	if s.repo == nil || result.PluginID == "" {
		return
	}
	now := time.Now()
	record := model.PluginRegistry{
		PluginID:         result.PluginID,
		InstallID:        result.PluginID + ":unknown",
		CurrentVersion:   result.Version,
		InstallRoot:      result.InstallRoot,
		DesiredState:     pluginapi.PluginStateQuarantined,
		ActualState:      pluginapi.PluginRuntimeStopped,
		LifecycleState:   pluginapi.PluginStateQuarantined,
		HealthStatus:     pluginapi.PluginRuntimeStopped,
		LastVerifiedAt:   &now,
		LastVerifyStatus: "failed",
		QuarantineReason: joinMessages(result.Errors),
	}
	_ = s.repo.UpsertPluginRegistry(&record)
}

func (s *PluginService) persistManifest(result pluginapi.DiscoveryResult) {
	if s.repo == nil {
		return
	}
	manifest := result.Manifest
	now := time.Now()
	manifestJSON, _ := json.Marshal(manifest)
	manifestHash := fileSHA256(result.ManifestPath)

	enabled := false
	autostart := false
	desiredState := pluginapi.PluginStateApprovedDisabled
	lifecycleState := pluginapi.PluginStateDiscovered
	lastStartAt := (*time.Time)(nil)
	lastReadyAt := (*time.Time)(nil)
	lastStopAt := (*time.Time)(nil)
	lastFaultAt := (*time.Time)(nil)
	if existing, err := s.repo.GetPluginRegistry(manifest.ID); err == nil {
		enabled = existing.Enabled
		autostart = existing.Autostart
		desiredState = existing.DesiredState
		lifecycleState = existing.LifecycleState
		lastStartAt = existing.LastStartAt
		lastReadyAt = existing.LastReadyAt
		lastStopAt = existing.LastStopAt
		lastFaultAt = existing.LastFaultAt
	}

	record := model.PluginRegistry{
		PluginID:            manifest.ID,
		InstallID:           buildInstallID(manifest),
		CurrentVersion:      manifest.Version,
		InstallRoot:         result.InstallRoot,
		SourceType:          "local-directory",
		Enabled:             enabled,
		Autostart:           autostart,
		DesiredState:        desiredState,
		ActualState:         pluginapi.PluginRuntimeStopped,
		LifecycleState:      lifecycleState,
		TrustLevel:          "local-approved",
		SignatureStatus:     "unknown",
		CurrentManifestHash: manifestHash,
		LastVerifiedAt:      &now,
		LastVerifyStatus:    "passed",
		ConfigVersion:       1,
		SelectedOS:          runtime.GOOS,
		SelectedArch:        runtime.GOARCH,
		HealthStatus:        pluginapi.PluginRuntimeStopped,
		LastStartAt:         lastStartAt,
		LastReadyAt:         lastReadyAt,
		LastStopAt:          lastStopAt,
		LastFaultAt:         lastFaultAt,
		ManifestJSON:        string(manifestJSON),
	}
	_ = s.repo.UpsertPluginRegistry(&record)
}

func (s *PluginService) persistBindings(manifest pluginapi.Manifest) {
	if s.repo == nil {
		return
	}
	_ = s.repo.DeletePluginBindings(manifest.ID)
	for _, page := range manifest.Frontend.Pages {
		_ = s.repo.UpsertPluginBinding(&model.PluginBinding{
			PluginID:        manifest.ID,
			BindingType:     "frontend.page",
			BindingKey:      page.ID,
			TargetScope:     page.Area,
			MountArea:       page.Area,
			RouteOrViewID:   page.ViewID,
			Enabled:         page.Visible,
			PermissionGuard: firstString(page.PermissionKeys),
		})
	}
	for _, menu := range manifest.Frontend.Menus {
		_ = s.repo.UpsertPluginBinding(&model.PluginBinding{
			PluginID:        manifest.ID,
			BindingType:     "frontend.menu",
			BindingKey:      menu.ID,
			TargetScope:     menu.DefaultGroup,
			MountArea:       menu.DefaultGroup,
			RouteOrViewID:   menu.TargetPageID,
			Enabled:         menu.Visible,
			OrderHint:       menu.Order,
			PermissionGuard: firstString(menu.PermissionKeys),
		})
	}
	for _, permission := range manifest.Permissions {
		_ = s.repo.UpsertPluginBinding(&model.PluginBinding{
			PluginID:        manifest.ID,
			BindingType:     "permission",
			BindingKey:      permission.Key,
			TargetScope:     permission.Scope,
			MountArea:       permission.Namespace,
			Enabled:         true,
			PermissionGuard: permission.Key,
		})
	}
}

func (s *PluginService) persistMigrations(manifest pluginapi.Manifest) {
	if s.repo == nil {
		return
	}
	_ = s.repo.DeletePluginMigrations(manifest.ID)
	for _, migration := range manifest.Backend.Migrations {
		_ = s.repo.UpsertPluginMigration(&model.PluginMigration{
			PluginID:    manifest.ID,
			MigrationID: migration.ID,
			Version:     migration.Version,
			Direction:   migration.Direction,
			Path:        migration.Path,
			Checksum:    migration.Checksum,
			Status:      "declared",
		})
	}
}

func (s *PluginService) persistDatabaseDeclarations(manifest pluginapi.Manifest) {
	if s.repo == nil {
		return
	}
	_ = s.repo.DeletePluginDatabaseDeclarations(manifest.ID)
	if strings.TrimSpace(manifest.Database.Namespace) == "" && len(manifest.Database.Tables) == 0 {
		return
	}
	_ = s.repo.UpsertPluginDatabaseDeclaration(&model.PluginDatabaseDeclaration{
		PluginID:       manifest.ID,
		PluginVersion:  manifest.Version,
		Namespace:      manifest.Database.Namespace,
		StorageMode:    defaultString(manifest.Database.StorageMode, "host-main-db"),
		TableCount:     len(manifest.Database.Tables),
		Status:         "declared",
		ExtensionsJSON: jsonString(manifest.Database.Extensions),
	})
	for _, tableDeclaration := range manifest.Database.Tables {
		table := &model.PluginDatabaseTable{
			PluginID:          manifest.ID,
			PluginVersion:     manifest.Version,
			Namespace:         manifest.Database.Namespace,
			TableKey:          tableDeclaration.TableKey,
			PhysicalTableName: tableDeclaration.PhysicalName,
			TableKind:         tableDeclaration.TableKind,
			SchemaVersion:     tableDeclaration.SchemaVersion,
			SchemaChecksum:    tableDeclaration.SchemaChecksum,
			Status:            "declared",
			Sensitivity:       defaultString(tableDeclaration.Sensitivity, "internal"),
			CreatePolicy:      defaultString(tableDeclaration.CreatePolicy, "on_enable"),
			DropPolicy:        defaultString(tableDeclaration.DropPolicy, "manual_only"),
			BackupPolicy:      defaultString(tableDeclaration.BackupPolicy, "include"),
			RetentionPolicy:   tableDeclaration.RetentionPolicy,
			Description:       tableDeclaration.Description,
			ExtensionsJSON:    jsonString(tableDeclaration.Extensions),
		}
		if err := s.repo.CreatePluginDatabaseTable(table); err != nil {
			continue
		}
		for _, columnDeclaration := range tableDeclaration.Columns {
			_ = s.repo.CreatePluginDatabaseColumn(&model.PluginDatabaseColumn{
				PluginID:         manifest.ID,
				TableID:          table.ID,
				ColumnKey:        columnDeclaration.ColumnKey,
				ColumnName:       columnDeclaration.ColumnName,
				DBType:           columnDeclaration.DBType,
				LogicalType:      columnDeclaration.LogicalType,
				Nullable:         columnDeclaration.Nullable,
				DefaultValueJSON: jsonString(columnDeclaration.DefaultValue),
				PrimaryKey:       columnDeclaration.PrimaryKey,
				AutoIncrement:    columnDeclaration.AutoIncrement,
				UniqueKey:        columnDeclaration.Unique,
				Indexed:          columnDeclaration.Indexed,
				Encrypted:        columnDeclaration.Encrypted,
				Secret:           columnDeclaration.Secret,
				ReferenceType:    columnDeclaration.ReferenceType,
				ReferenceTarget:  columnDeclaration.ReferenceTarget,
				Description:      columnDeclaration.Description,
				ExtensionsJSON:   jsonString(columnDeclaration.Extensions),
			})
		}
		for _, indexDeclaration := range tableDeclaration.Indexes {
			_ = s.repo.CreatePluginDatabaseIndex(&model.PluginDatabaseIndex{
				PluginID:       manifest.ID,
				TableID:        table.ID,
				IndexKey:       indexDeclaration.IndexKey,
				IndexName:      indexDeclaration.IndexName,
				ColumnsJSON:    jsonString(indexDeclaration.Columns),
				UniqueIndex:    indexDeclaration.Unique,
				Status:         "declared",
				ExtensionsJSON: jsonString(indexDeclaration.Extensions),
			})
		}
		for _, relationDeclaration := range tableDeclaration.Relations {
			_ = s.repo.CreatePluginDatabaseRelation(&model.PluginDatabaseRelation{
				PluginID:           manifest.ID,
				TableID:            table.ID,
				RelationKey:        relationDeclaration.RelationKey,
				LocalColumn:        relationDeclaration.LocalColumn,
				TargetResourceType: relationDeclaration.TargetResourceType,
				TargetKey:          relationDeclaration.TargetKey,
				RelationType:       defaultString(relationDeclaration.RelationType, "many_to_one"),
				Required:           relationDeclaration.Required,
				OnDeletePolicy:     defaultString(relationDeclaration.OnDeletePolicy, "restrict"),
				ExtensionsJSON:     jsonString(relationDeclaration.Extensions),
			})
		}
		for _, operationDeclaration := range tableDeclaration.Operations {
			operationID := operationDeclaration.OperationID
			if operationID == "" {
				operationID = manifest.ID + ":" + tableDeclaration.TableKey + ":" + operationDeclaration.OperationType
			}
			_ = s.repo.UpsertPluginDatabaseOperation(&model.PluginDatabaseOperation{
				OperationID:    operationID,
				PluginID:       manifest.ID,
				PluginVersion:  manifest.Version,
				TableKey:       tableDeclaration.TableKey,
				OperationType:  operationDeclaration.OperationType,
				Path:           operationDeclaration.Path,
				RequiresReview: operationDeclaration.RequiresReview,
				Status:         "declared",
				SchemaChecksum: operationDeclaration.Checksum,
				ExtensionsJSON: jsonString(operationDeclaration.Extensions),
			})
		}
	}
}

func (s *PluginService) persistGovernanceDeclarations(result pluginapi.DiscoveryResult) {
	if s.governance == nil {
		return
	}
	manifest := result.Manifest
	_ = s.governance.RegisterPermissions(s.PermissionDefinitionInputs())
	_ = s.governance.SyncPluginConfigSchemas(s.ConfigSchemas())
	_ = s.governance.UpsertPluginVersion(PluginVersionInput{
		PluginID:     manifest.ID,
		Version:      manifest.Version,
		ManifestHash: fileSHA256(result.ManifestPath),
		PackageHash:  manifest.Integrity.PackageDigest,
		InstallPath:  result.InstallRoot,
		Status:       "installed",
	})
	_ = s.governance.EnsurePluginTrustRecord(
		manifest.ID,
		manifest.Version,
		strings.Join(manifest.Integrity.RequiredScopes, ","),
		manifest.Integrity.RequireApprovedFingerprint,
	)
}

func buildInstallID(manifest pluginapi.Manifest) string {
	return manifest.ID + ":" + manifest.Version
}

func fileSHA256(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

func joinMessages(items []string) string {
	data, _ := json.Marshal(items)
	return string(data)
}

func firstString(items []string) string {
	if len(items) == 0 {
		return ""
	}
	return items[0]
}

func normalizeStringList(items []string) []string {
	values := make([]string, 0, len(items))
	seen := map[string]bool{}
	for _, item := range items {
		value := strings.TrimSpace(item)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		values = append(values, value)
	}
	return values
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func jsonString(value any) string {
	if value == nil {
		return ""
	}
	data, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(data)
}

func cleanZipEntryName(name string) string {
	name = strings.ReplaceAll(strings.TrimSpace(name), "\\", "/")
	name = strings.TrimPrefix(name, "/")
	cleaned := filepath.ToSlash(filepath.Clean(name))
	if cleaned == "." || strings.HasPrefix(cleaned, "../") || cleaned == ".." {
		return ""
	}
	return cleaned
}

func validPluginTopDirectory(name string) bool {
	if strings.TrimSpace(name) == "" {
		return false
	}
	if strings.ContainsAny(name, `/\:`) || strings.HasPrefix(name, ".") {
		return false
	}
	if strings.HasSuffix(strings.ToLower(name), ".zip") {
		return false
	}
	return true
}

func pathWithin(root, target string) bool {
	relative, err := filepath.Rel(root, target)
	if err != nil {
		return false
	}
	return relative == "." || (!filepath.IsAbs(relative) && relative != ".." && !strings.HasPrefix(relative, ".."+string(os.PathSeparator)))
}

func dropPluginDatabaseTables(tx *gorm.DB, pluginID string) error {
	if tx == nil {
		return errors.New("数据库事务未初始化")
	}

	var tables []model.PluginDatabaseTable
	if err := tx.Where("plugin_id = ?", strings.TrimSpace(pluginID)).Find(&tables).Error; err != nil {
		return err
	}

	tableIDs := make([]uint, 0, len(tables))
	for _, table := range tables {
		tableIDs = append(tableIDs, table.ID)
		name := strings.TrimSpace(table.PhysicalTableName)
		if name == "" {
			continue
		}
		if err := tx.Migrator().DropTable(name); err != nil {
			return err
		}
	}

	if len(tableIDs) > 0 {
		if err := tx.Where("table_id IN ?", tableIDs).Delete(&model.PluginDatabaseRelation{}).Error; err != nil {
			return err
		}
		if err := tx.Where("table_id IN ?", tableIDs).Delete(&model.PluginDatabaseIndex{}).Error; err != nil {
			return err
		}
		if err := tx.Where("table_id IN ?", tableIDs).Delete(&model.PluginDatabaseColumn{}).Error; err != nil {
			return err
		}
	}

	if err := tx.Where("plugin_id = ?", pluginID).Delete(&model.PluginDatabaseOperation{}).Error; err != nil {
		return err
	}
	if err := tx.Where("plugin_id = ?", pluginID).Delete(&model.PluginDatabaseTable{}).Error; err != nil {
		return err
	}
	if err := tx.Where("plugin_id = ?", pluginID).Delete(&model.PluginDatabaseDeclaration{}).Error; err != nil {
		return err
	}
	return nil
}

func writeZipFile(target string, src io.Reader, mode os.FileMode) error {
	perm := mode.Perm()
	if perm == 0 {
		perm = 0644
	}
	dst, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, perm)
	if err != nil {
		return err
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)
	return err
}

// DefaultPluginRoot 返回程序标准插件目录。
func DefaultPluginRoot(configDir string) string {
	if configDir == "" {
		return "plugins"
	}
	return filepath.Clean(filepath.Join(filepath.Dir(configDir), "plugins"))
}
