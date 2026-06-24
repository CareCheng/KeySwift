package service_test

import (
	"encoding/json"
	"testing"

	"user-frontend/internal/model"
	pluginapi "user-frontend/internal/plugin"
	"user-frontend/internal/service"
	"user-frontend/internal/test"
)

func TestGovernanceServiceSyncHostPermissions(t *testing.T) {
	services, cleanup := test.SetupTestServices(t)
	defer cleanup()

	svc := service.NewGovernanceService(services.Repo)
	if err := svc.SyncHostPermissions(model.AllPermissions); err != nil {
		t.Fatalf("同步宿主权限失败: %v", err)
	}
	if err := svc.SyncHostPermissions(model.AllPermissions); err != nil {
		t.Fatalf("重复同步宿主权限失败: %v", err)
	}

	records, err := svc.ListPermissionDefinitions()
	if err != nil {
		t.Fatalf("读取权限定义失败: %v", err)
	}
	if len(records) != len(model.AllPermissions) {
		t.Fatalf("权限定义数量不匹配: got %d want %d", len(records), len(model.AllPermissions))
	}

	var pluginView *model.PermissionDefinition
	for i := range records {
		if records[i].PermissionCode == "plugin:view" {
			pluginView = &records[i]
			break
		}
	}
	if pluginView == nil {
		t.Fatal("未找到 plugin:view 权限定义")
	}
	if pluginView.OwnerType != service.PermissionOwnerHost {
		t.Fatalf("宿主权限 owner_type 不正确: %s", pluginView.OwnerType)
	}
}

func TestGovernanceServiceSyncRolePermissionGrants(t *testing.T) {
	services, cleanup := test.SetupTestServices(t)
	defer cleanup()

	svc := service.NewGovernanceService(services.Repo)
	if err := svc.SyncRolePermissionGrants(1, []string{"plugin:view", "plugin:manage", "plugin:view"}, "admin:root"); err != nil {
		t.Fatalf("同步角色权限失败: %v", err)
	}
	if err := svc.SyncRolePermissionGrants(1, []string{"plugin:view"}, "admin:root"); err != nil {
		t.Fatalf("裁剪角色权限失败: %v", err)
	}

	var grants []model.RolePermissionGrant
	if err := services.DB.Where("role_id = ?", 1).Order("permission_code ASC").Find(&grants).Error; err != nil {
		t.Fatalf("读取角色权限授权失败: %v", err)
	}
	if len(grants) != 1 {
		t.Fatalf("角色权限授权数量不正确: got %d want 1", len(grants))
	}
	if grants[0].PermissionCode != "plugin:view" {
		t.Fatalf("角色权限授权裁剪结果不正确: %s", grants[0].PermissionCode)
	}
}

func TestGovernanceServiceSyncPluginConfigSchemasEnsuresDefaultValues(t *testing.T) {
	services, cleanup := test.SetupTestServices(t)
	defer cleanup()

	svc := service.NewGovernanceService(services.Repo)
	schema := pluginapi.ConfigSchema{
		SchemaVersion:  "1.0.0",
		PluginID:       "cloudflare.turnstile",
		ConfigVersion:  "1.0.0",
		SecretPolicies: []string{"secret_key"},
		Sections: []pluginapi.ConfigSection{
			{
				ID: "turnstile",
				Fields: []pluginapi.FieldSchema{
					{Key: "site_key", Type: "text", Required: true},
					{Key: "secret_key", Type: "secret", Required: true, Secret: true},
					{Key: "theme", Type: "select", Default: "auto"},
					{Key: "size", Type: "select", Default: "normal"},
					{Key: "fail_open", Type: "boolean", Default: false},
				},
			},
		},
	}

	if err := svc.SyncPluginConfigSchemas([]pluginapi.ConfigSchema{schema}); err != nil {
		t.Fatalf("同步插件配置 schema 失败: %v", err)
	}

	var value model.PluginConfigValue
	if err := services.DB.Where("plugin_id = ? AND config_key = ?", "cloudflare.turnstile", "default").First(&value).Error; err != nil {
		t.Fatalf("默认配置值未落库: %v", err)
	}
	values := map[string]any{}
	if err := json.Unmarshal([]byte(value.ValueJSON), &values); err != nil {
		t.Fatalf("默认配置 JSON 解析失败: %v", err)
	}
	if values["site_key"] != "" {
		t.Fatalf("未配置文本字段应初始化为空字符串: %#v", values["site_key"])
	}
	if values["theme"] != "auto" || values["size"] != "normal" {
		t.Fatalf("默认配置值不正确: %#v", values)
	}
	if values["fail_open"] != false {
		t.Fatalf("布尔默认值不正确: %#v", values["fail_open"])
	}
	if _, ok := values["secret_key"]; ok {
		t.Fatal("敏感字段不应写入普通配置 JSON")
	}
	if value.SecretJSON != "" {
		t.Fatalf("未填写密钥时不应写入 secret_json: %s", value.SecretJSON)
	}

	var revisionCount int64
	if err := services.DB.Model(&model.PluginConfigRevision{}).
		Where("plugin_id = ?", "cloudflare.turnstile").
		Count(&revisionCount).Error; err != nil {
		t.Fatalf("统计配置修订失败: %v", err)
	}
	if revisionCount != 1 {
		t.Fatalf("首次初始化应只生成一条修订记录: %d", revisionCount)
	}

	if err := svc.SyncPluginConfigSchemas([]pluginapi.ConfigSchema{schema}); err != nil {
		t.Fatalf("重复同步插件配置 schema 失败: %v", err)
	}
	if err := services.DB.Model(&model.PluginConfigRevision{}).
		Where("plugin_id = ?", "cloudflare.turnstile").
		Count(&revisionCount).Error; err != nil {
		t.Fatalf("统计重复同步后的配置修订失败: %v", err)
	}
	if revisionCount != 1 {
		t.Fatalf("无变更重复同步不应新增修订记录: %d", revisionCount)
	}
}

func TestGovernanceServiceSyncPluginConfigSchemasBackfillsWithoutOverwrite(t *testing.T) {
	services, cleanup := test.SetupTestServices(t)
	defer cleanup()

	svc := service.NewGovernanceService(services.Repo)
	initial := pluginapi.ConfigSchema{
		SchemaVersion: "1.0.0",
		PluginID:      "keyswift.image_captcha",
		ConfigVersion: "1.0.0",
		Sections: []pluginapi.ConfigSection{
			{
				ID: "image",
				Fields: []pluginapi.FieldSchema{
					{Key: "width", Type: "number", Default: 240},
					{Key: "height", Type: "number", Default: 80},
				},
			},
		},
	}
	if err := svc.SyncPluginConfigSchemas([]pluginapi.ConfigSchema{initial}); err != nil {
		t.Fatalf("首次同步图片验证码配置失败: %v", err)
	}
	if _, err := svc.SavePluginConfigValue(service.PluginConfigValueInput{
		PluginID:  "keyswift.image_captcha",
		ConfigKey: "default",
		Value: map[string]any{
			"width":  320,
			"height": 80,
		},
		UpdatedBy: "admin:test",
	}); err != nil {
		t.Fatalf("保存管理员配置失败: %v", err)
	}

	updated := initial
	updated.Sections[0].Fields = append(updated.Sections[0].Fields, pluginapi.FieldSchema{
		Key:     "ttl_seconds",
		Type:    "number",
		Default: 300,
	})
	if err := svc.SyncPluginConfigSchemas([]pluginapi.ConfigSchema{updated}); err != nil {
		t.Fatalf("补齐新增配置字段失败: %v", err)
	}

	var value model.PluginConfigValue
	if err := services.DB.Where("plugin_id = ? AND config_key = ?", "keyswift.image_captcha", "default").First(&value).Error; err != nil {
		t.Fatalf("读取补齐后的配置失败: %v", err)
	}
	values := map[string]any{}
	if err := json.Unmarshal([]byte(value.ValueJSON), &values); err != nil {
		t.Fatalf("补齐后配置 JSON 解析失败: %v", err)
	}
	if values["width"] != float64(320) {
		t.Fatalf("已有管理员配置不应被默认值覆盖: %#v", values)
	}
	if values["ttl_seconds"] != float64(300) {
		t.Fatalf("新增字段应按默认值补齐: %#v", values)
	}
	if value.Revision != 3 {
		t.Fatalf("修订号应包含初始化、管理员保存和补齐: %d", value.Revision)
	}
}

func TestGovernanceServiceDataScopeAuditEventAndJob(t *testing.T) {
	services, cleanup := test.SetupTestServices(t)
	defer cleanup()

	svc := service.NewGovernanceService(services.Repo)
	if err := svc.RegisterResourceScopes([]service.ResourceScopeDefinitionInput{
		{
			ResourceType: "order",
			ScopeType:    "own",
			Name:         "本人订单",
			Description:  "只允许读取当前主体自己的订单",
		},
	}); err != nil {
		t.Fatalf("注册资源范围失败: %v", err)
	}

	scope := service.DataScope{ResourceType: "order", ScopeType: "own", ScopeValue: "self"}
	if err := svc.GrantSubjectDataScope("admin:1", "admin", scope, "admin:root"); err != nil {
		t.Fatalf("授予数据范围失败: %v", err)
	}
	scopes, err := svc.ListSubjectDataScopes("admin:1")
	if err != nil {
		t.Fatalf("读取数据范围失败: %v", err)
	}
	if len(scopes) != 1 || scopes[0].ResourceType != "order" || scopes[0].ScopeType != "own" {
		t.Fatalf("数据范围读取结果不正确: %#v", scopes)
	}

	if err := svc.RecordAudit(service.AuditInput{
		ActorSubjectID: "admin:1",
		Action:         "plugin.enable",
		ResourceType:   "plugin",
		ResourceID:     "diagnostic",
		RiskLevel:      "high",
		Payload:        map[string]any{"enabled": true},
	}); err != nil {
		t.Fatalf("写入审计日志失败: %v", err)
	}

	eventID, err := svc.RecordEvent(service.EventInput{
		EventType:     "plugin.state.changed",
		SourceType:    "plugin",
		SourceID:      "diagnostic",
		OwnerPluginID: "diagnostic",
		Payload:       map[string]any{"to": "ready"},
	})
	if err != nil {
		t.Fatalf("写入事件日志失败: %v", err)
	}
	if eventID == "" {
		t.Fatal("事件 ID 不能为空")
	}

	jobID, err := svc.CreateSystemJob(service.SystemJobInput{
		JobType:       "plugin.healthcheck",
		OwnerPluginID: "diagnostic",
		Payload:       map[string]any{"probe": "runtime"},
	})
	if err != nil {
		t.Fatalf("创建系统任务失败: %v", err)
	}
	if jobID == "" {
		t.Fatal("任务 ID 不能为空")
	}

	var auditCount int64
	if err := services.DB.Model(&model.AuditLog{}).Count(&auditCount).Error; err != nil {
		t.Fatalf("统计审计日志失败: %v", err)
	}
	if auditCount != 1 {
		t.Fatalf("审计日志数量不正确: %d", auditCount)
	}

	var job model.SystemJob
	if err := services.DB.Where("job_id = ?", jobID).First(&job).Error; err != nil {
		t.Fatalf("读取系统任务失败: %v", err)
	}
	if job.OwnerPluginID != "diagnostic" {
		t.Fatalf("系统任务 owner_plugin_id 不正确: %s", job.OwnerPluginID)
	}
}
