// Package plugin 提供宿主插件的内存注册中心实现。
package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"sync"
)

// MemoryRegistry 是宿主进程内存中的插件注册中心。
type MemoryRegistry struct {
	mu          sync.RWMutex
	manifests   map[string]Manifest
	runtimes    map[string]RuntimePlugin
	permissions map[string][]PermissionDeclaration
	configs     map[string]ConfigSchema
	themes      map[string]UIContribution
	frontend    FrontendContribution
}

// NewMemoryRegistry 创建内存注册中心。
func NewMemoryRegistry() *MemoryRegistry {
	return &MemoryRegistry{
		manifests:   make(map[string]Manifest),
		runtimes:    make(map[string]RuntimePlugin),
		permissions: make(map[string][]PermissionDeclaration),
		configs:     make(map[string]ConfigSchema),
		themes:      make(map[string]UIContribution),
		frontend: FrontendContribution{
			ProtocolVersion: FrontendProtocol,
			Pages:           make([]PageDeclaration, 0),
			Menus:           make([]MenuDeclaration, 0),
			Forms:           make([]FormDeclaration, 0),
			Views:           make([]ViewDeclaration, 0),
			Themes:          make([]UIContribution, 0),
			Extensions:      ExtensionMap{},
		},
	}
}

// RegisterManifest 注册或更新插件 manifest。
func (r *MemoryRegistry) RegisterManifest(_ context.Context, manifest Manifest) error {
	if manifest.ID == "" {
		return errors.New("插件ID不能为空")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.manifests[manifest.ID] = manifest
	r.permissions[manifest.ID] = append([]PermissionDeclaration(nil), manifest.Permissions...)
	if manifest.Backend.SettingsRef != "" || manifest.Frontend.Enabled {
		r.configs[manifest.ID] = ConfigSchema{
			SchemaVersion: ManifestVersion,
			PluginID:      manifest.ID,
			ConfigVersion: ManifestVersion,
			Sections:      nil,
			Extensions:    ExtensionMap{},
		}
	}
	if manifest.UI.Enabled || manifest.PluginKind == PluginKindUITheme {
		r.themes[manifest.ID] = manifest.UI
	}
	r.rebuildFrontendLocked()
	return nil
}

// SetRuntime 更新插件运行态快照。
func (r *MemoryRegistry) SetRuntime(pluginID string, runtime RuntimePlugin) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.runtimes[pluginID] = runtime
}

// GetManifest 获取单个插件 manifest。
func (r *MemoryRegistry) GetManifest(pluginID string) (Manifest, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	manifest, ok := r.manifests[pluginID]
	return manifest, ok
}

// ListManifests 获取全部 manifest。
func (r *MemoryRegistry) ListManifests() []Manifest {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]Manifest, 0, len(r.manifests))
	for _, item := range r.manifests {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})
	return items
}

// GetRuntime 获取单个插件运行态。
func (r *MemoryRegistry) GetRuntime(pluginID string) (RuntimePlugin, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	runtime, ok := r.runtimes[pluginID]
	return runtime, ok
}

// ListRuntimes 获取全部运行态。
func (r *MemoryRegistry) ListRuntimes() []RuntimePlugin {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]RuntimePlugin, 0, len(r.runtimes))
	for _, item := range r.runtimes {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].PluginID < items[j].PluginID
	})
	return items
}

// FrontendContribution 返回宿主可直接消费的前端聚合结果。
func (r *MemoryRegistry) FrontendContribution() FrontendContribution {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.frontend
}

// ThemeContribution 返回全部主题贡献。
func (r *MemoryRegistry) ThemeContribution() []UIContribution {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]UIContribution, 0, len(r.themes))
	for _, item := range r.themes {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].ThemeScope < items[j].ThemeScope
	})
	return items
}

// ListPermissions 返回全部插件权限。
func (r *MemoryRegistry) ListPermissions() []PermissionDeclaration {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var items []PermissionDeclaration
	for _, manifest := range r.manifests {
		items = append(items, manifest.Permissions...)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Key < items[j].Key
	})
	return items
}

// ListConfigSchemas 返回全部配置 schema。
func (r *MemoryRegistry) ListConfigSchemas() []ConfigSchema {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]ConfigSchema, 0, len(r.configs))
	for _, item := range r.configs {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].PluginID < items[j].PluginID
	})
	return items
}

func (r *MemoryRegistry) rebuildFrontendLocked() {
	pages := make([]PageDeclaration, 0)
	menus := make([]MenuDeclaration, 0)
	forms := make([]FormDeclaration, 0)
	views := make([]ViewDeclaration, 0)
	themes := make([]UIContribution, 0)

	for _, manifest := range r.manifests {
		if manifest.Frontend.Enabled {
			pages = append(pages, manifest.Frontend.Pages...)
			menus = append(menus, manifest.Frontend.Menus...)
			forms = append(forms, manifest.Frontend.Forms...)
			views = append(views, manifest.Frontend.Views...)
		}
		if manifest.UI.Enabled || manifest.PluginKind == PluginKindUITheme {
			themes = append(themes, manifest.UI)
		}
	}

	sort.Slice(pages, func(i, j int) bool { return pages[i].ID < pages[j].ID })
	sort.Slice(menus, func(i, j int) bool { return menus[i].ID < menus[j].ID })
	sort.Slice(forms, func(i, j int) bool { return forms[i].ID < forms[j].ID })
	sort.Slice(views, func(i, j int) bool { return views[i].ID < views[j].ID })
	sort.Slice(themes, func(i, j int) bool { return themes[i].ThemeScope < themes[j].ThemeScope })

	r.frontend.Pages = pages
	r.frontend.Menus = menus
	r.frontend.Forms = forms
	r.frontend.Views = views
	r.frontend.Themes = themes
}

// MarshalJSONSummary 生成插件注册中心的简要 JSON。
func (r *MemoryRegistry) MarshalJSONSummary() ([]byte, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	summary := struct {
		Plugins     int `json:"plugins"`
		Runtimes    int `json:"runtimes"`
		Permissions int `json:"permissions"`
		Themes      int `json:"themes"`
	}{
		Plugins:     len(r.manifests),
		Runtimes:    len(r.runtimes),
		Permissions: countPermissionsLocked(r.manifests),
		Themes:      len(r.themes),
	}

	return json.Marshal(summary)
}

func countPermissionsLocked(manifests map[string]Manifest) int {
	total := 0
	for _, manifest := range manifests {
		total += len(manifest.Permissions)
	}
	return total
}
