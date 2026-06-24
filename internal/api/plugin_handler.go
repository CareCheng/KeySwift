package api

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	plugin "user-frontend/internal/plugin"
	"user-frontend/internal/service"

	"github.com/gin-gonic/gin"
)

// AdminGetPluginSummary 获取插件平台概览。
func AdminGetPluginSummary(c *gin.Context) {
	if PluginSvc == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"summary": gin.H{
				"plugins":        0,
				"frontend_pages": 0,
				"frontend_menus": 0,
				"themes":         0,
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"summary": PluginSvc.Summary(),
	})
}

// AdminListPlugins 获取插件列表。
func AdminListPlugins(c *gin.Context) {
	if PluginSvc == nil {
		c.JSON(http.StatusOK, gin.H{"success": true, "plugins": []any{}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"plugins": PluginSvc.ListPlugins(),
	})
}

// AdminGetPluginDetail 获取插件详情。
func AdminGetPluginDetail(c *gin.Context) {
	if PluginSvc == nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "插件服务未初始化"})
		return
	}

	pluginID := c.Param("id")
	detail, ok := PluginSvc.GetPluginDetail(pluginID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "插件不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "plugin": detail})
}

// AdminGetPluginBindings 获取插件绑定声明。
func AdminGetPluginBindings(c *gin.Context) {
	if PluginSvc == nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "插件服务未初始化"})
		return
	}
	pluginID := c.Param("id")
	bindings, ok := PluginSvc.GetPluginBindings(pluginID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "插件绑定不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "bindings": bindings})
}

// AdminGetPluginMigrations 获取插件迁移声明。
func AdminGetPluginMigrations(c *gin.Context) {
	if PluginSvc == nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "插件服务未初始化"})
		return
	}
	pluginID := c.Param("id")
	migrations, ok := PluginSvc.GetPluginMigrations(pluginID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "插件迁移不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "migrations": migrations})
}

// AdminGetPluginDatabaseTables 获取插件数据库治理声明。
func AdminGetPluginDatabaseTables(c *gin.Context) {
	if PluginSvc == nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "插件服务未初始化"})
		return
	}
	pluginID := c.Param("id")
	database, ok := PluginSvc.GetPluginDatabaseSnapshot(pluginID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "插件数据库声明不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"database": database,
		"tables":   database.Tables,
	})
}

// AdminEnablePlugin 启用插件治理状态。
func AdminEnablePlugin(c *gin.Context) {
	if PluginSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "插件服务未初始化"})
		return
	}
	pluginID := c.Param("id")
	if err := PluginSvc.EnablePlugin(pluginID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	recordPluginGovernanceAction(c, pluginID, "plugin.enable", "enabled")
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "插件已启用"})
}

// AdminDisablePlugin 停用插件治理状态。
func AdminDisablePlugin(c *gin.Context) {
	if PluginSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "插件服务未初始化"})
		return
	}
	pluginID := c.Param("id")
	if err := PluginSvc.DisablePlugin(pluginID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	recordPluginGovernanceAction(c, pluginID, "plugin.disable", "approved-disabled")
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "插件已停用"})
}

// AdminUninstallPlugin 卸载插件并按需清理配置与独立数据表。
func AdminUninstallPlugin(c *gin.Context) {
	if PluginSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "插件服务未初始化"})
		return
	}

	var req struct {
		DeleteConfig  bool `json:"delete_config"`
		DeleteDatabase bool `json:"delete_database"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "参数错误"})
		return
	}

	pluginID := c.Param("id")
	if req.DeleteDatabase {
		if snapshot, ok := PluginSvc.GetPluginDatabaseSnapshot(pluginID); !ok || len(snapshot.Tables) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "当前插件没有可删除的独立数据库"})
			return
		}
	}

	if PluginRuntimeSvc != nil {
		if err := PluginRuntimeSvc.Stop(pluginID, "plugin uninstall"); err != nil && !strings.Contains(err.Error(), "插件运行会话不存在") {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
			return
		}
	}

	result, err := PluginSvc.UninstallPlugin(c.Request.Context(), pluginID, service.PluginUninstallOptions{
		DeleteConfig:  req.DeleteConfig,
		DeleteDatabase: req.DeleteDatabase,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if GovernanceSvc != nil {
		_ = GovernanceSvc.RecordAudit(service.AuditInput{
			ActorSubjectID: currentSubjectID(c),
			Action:         "plugin.uninstall",
			ResourceType:   "plugin",
			ResourceID:     pluginID,
			RiskLevel:      "high",
			IP:             GetClientIP(c),
			UserAgent:      c.GetHeader("User-Agent"),
			Payload: gin.H{
				"delete_config":  req.DeleteConfig,
				"delete_database": req.DeleteDatabase,
			},
		})
	}

	message := "插件已卸载"
	switch {
	case req.DeleteConfig && req.DeleteDatabase:
		message = "插件已卸载，配置和独立数据表已删除"
	case req.DeleteConfig:
		message = "插件已卸载，配置已删除"
	case req.DeleteDatabase:
		message = "插件已卸载，独立数据表已删除"
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
		"result":  result,
	})
}

// AdminInstallPluginPackage 上传并安装 .ksplugin.zip 插件包。
func AdminInstallPluginPackage(c *gin.Context) {
	if PluginSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "插件服务未初始化"})
		return
	}

	file, err := c.FormFile("package")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "请选择插件安装包"})
		return
	}
	filename := filepath.Base(file.Filename)
	if !strings.HasSuffix(strings.ToLower(filename), ".ksplugin.zip") {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "仅支持 .ksplugin.zip 插件安装包"})
		return
	}

	tempDir, err := os.MkdirTemp("", "keyswift-plugin-upload-*")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "创建临时目录失败"})
		return
	}
	defer os.RemoveAll(tempDir)

	tempPath := filepath.Join(tempDir, filename)
	if err := c.SaveUploadedFile(file, tempPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "保存安装包失败"})
		return
	}

	result, err := PluginSvc.InstallPackage(c.Request.Context(), tempPath)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	if GovernanceSvc != nil {
		_ = GovernanceSvc.RegisterPermissions(PluginSvc.PermissionDefinitionInputs())
		_ = GovernanceSvc.SyncPluginConfigSchemas(PluginSvc.ConfigSchemas())
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "插件安装完成",
		"install": result,
		"results": result.Results,
	})
}

// AdminRefreshPlugins 刷新插件目录。
func AdminRefreshPlugins(c *gin.Context) {
	if PluginSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "插件服务未初始化"})
		return
	}

	results, err := PluginSvc.Refresh(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	if GovernanceSvc != nil {
		_ = GovernanceSvc.RegisterPermissions(PluginSvc.PermissionDefinitionInputs())
		_ = GovernanceSvc.SyncPluginConfigSchemas(PluginSvc.ConfigSchemas())
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"results": results,
	})
}

// AdminGetPluginFrontendContribution 获取插件前端挂载声明。
func AdminGetPluginFrontendContribution(c *gin.Context) {
	if PluginSvc == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"frontend": gin.H{
				"protocolVersion": "1.0.0",
				"pages":           []any{},
				"menus":           []any{},
				"forms":           []any{},
				"views":           []any{},
				"themes":          []any{},
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"frontend": PluginSvc.ListEnabledFrontendContribution(),
	})
}

// AdminGetPluginPermissions 获取插件权限声明。
func AdminGetPluginPermissions(c *gin.Context) {
	if GovernanceSvc == nil {
		c.JSON(http.StatusOK, gin.H{"success": true, "permissions": []any{}})
		return
	}

	permissions, err := GovernanceSvc.ListPermissionDefinitions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "读取权限定义失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"permissions": permissions,
	})
}

// AdminGetPluginConfigSchemas 获取插件配置 schema 声明。
func AdminGetPluginConfigSchemas(c *gin.Context) {
	if GovernanceSvc == nil {
		c.JSON(http.StatusOK, gin.H{"success": true, "schemas": []any{}})
		return
	}

	schemas, err := GovernanceSvc.ListPluginConfigSchemas("")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "读取插件配置 schema 失败"})
		return
	}
	parsedSchemas := make([]plugin.ConfigSchema, 0, len(schemas))
	for _, record := range schemas {
		var schema plugin.ConfigSchema
		if err := json.Unmarshal([]byte(record.SchemaJSON), &schema); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "插件配置 schema 解析失败"})
			return
		}
		if strings.TrimSpace(schema.PluginID) == "" {
			schema.PluginID = record.PluginID
		}
		if strings.TrimSpace(schema.SchemaVersion) == "" {
			schema.SchemaVersion = record.SchemaVersion
		}
		if strings.TrimSpace(schema.ConfigVersion) == "" {
			schema.ConfigVersion = record.SchemaVersion
		}
		parsedSchemas = append(parsedSchemas, schema)
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"schemas": parsedSchemas,
	})
}

// AdminGetPluginRuntimeRecords 获取插件状态事件和故障日志。
func AdminGetPluginRuntimeRecords(c *gin.Context) {
	if GovernanceSvc == nil {
		c.JSON(http.StatusOK, gin.H{"success": true, "sessions": []any{}, "events": []any{}, "faults": []any{}})
		return
	}
	pluginID := c.Param("id")
	sessions := []any{}
	if PluginRuntimeSvc != nil {
		if records, err := PluginRuntimeSvc.ListSessions(pluginID); err == nil {
			casted := make([]any, 0, len(records))
			for _, record := range records {
				casted = append(casted, record)
			}
			sessions = casted
		}
	}
	events, err := GovernanceSvc.ListPluginStateEvents(pluginID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "读取插件状态事件失败"})
		return
	}
	faults, err := GovernanceSvc.ListPluginFaultLogs(pluginID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "读取插件故障日志失败"})
		return
	}
	trustRecords, _ := GovernanceSvc.ListPluginTrustRecords(pluginID)
	c.JSON(http.StatusOK, gin.H{"success": true, "sessions": sessions, "events": events, "faults": faults, "trust_records": trustRecords})
}

// AdminGetPluginPermissionDefinitions 获取单插件权限定义。
func AdminGetPluginPermissionDefinitions(c *gin.Context) {
	if GovernanceSvc == nil {
		c.JSON(http.StatusOK, gin.H{"success": true, "permissions": []any{}})
		return
	}
	permissions, err := GovernanceSvc.ListPluginPermissionDefinitions(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "读取插件权限定义失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "permissions": permissions})
}

// AdminGetPluginConfigValues 获取单插件配置值和修订记录。
func AdminGetPluginConfigValues(c *gin.Context) {
	if GovernanceSvc == nil {
		c.JSON(http.StatusOK, gin.H{"success": true, "values": []any{}, "revisions": []any{}})
		return
	}
	pluginID := c.Param("id")
	values, err := GovernanceSvc.ListPluginConfigValues(pluginID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "读取插件配置值失败"})
		return
	}
	revisions, err := GovernanceSvc.ListPluginConfigRevisions(pluginID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "读取插件配置修订失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "values": values, "revisions": revisions})
}

// AdminSavePluginConfigValue 保存插件配置值。
func AdminSavePluginConfigValue(c *gin.Context) {
	if GovernanceSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "治理服务未初始化"})
		return
	}
	var req struct {
		ConfigKey     string `json:"config_key"`
		Value         any    `json:"value"`
		SecretJSON    string `json:"secret_json"`
		ChangeSummary string `json:"change_summary"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "参数错误"})
		return
	}
	subjectID := currentSubjectID(c)
	value, err := GovernanceSvc.SavePluginConfigValue(service.PluginConfigValueInput{
		PluginID:      c.Param("id"),
		ConfigKey:     req.ConfigKey,
		Value:         req.Value,
		SecretJSON:    req.SecretJSON,
		UpdatedBy:     subjectID,
		ChangeSummary: req.ChangeSummary,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	_ = GovernanceSvc.RecordAudit(service.AuditInput{
		ActorSubjectID: subjectID,
		Action:         "plugin.config.save",
		ResourceType:   "plugin",
		ResourceID:     c.Param("id"),
		RiskLevel:      "high",
		IP:             GetClientIP(c),
		UserAgent:      c.GetHeader("User-Agent"),
		Payload:        gin.H{"config_key": req.ConfigKey, "revision": value.Revision},
	})
	c.JSON(http.StatusOK, gin.H{"success": true, "value": value})
}

// AdminStartPluginRuntime 启动插件运行进程。
func AdminStartPluginRuntime(c *gin.Context) {
	if PluginSvc == nil || PluginRuntimeSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "插件运行时未初始化"})
		return
	}
	pluginID := c.Param("id")
	detail, ok := PluginSvc.GetPluginDetail(pluginID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "插件不存在"})
		return
	}
	manifest, ok := detail["manifest"].(plugin.Manifest)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "插件声明异常"})
		return
	}
	releaseDir := PluginSvc.ReleaseRoot(pluginID, manifest.Version)
	session, err := PluginRuntimeSvc.Start(c.Request.Context(), manifest, releaseDir)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "session": session})
}

// AdminStopPluginRuntime 停止插件运行进程。
func AdminStopPluginRuntime(c *gin.Context) {
	if PluginRuntimeSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "插件运行时未初始化"})
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&req)
	if err := PluginRuntimeSvc.Stop(c.Param("id"), req.Reason); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// AdminKillPluginRuntime 强制结束插件运行进程。
func AdminKillPluginRuntime(c *gin.Context) {
	if PluginRuntimeSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "插件运行时未初始化"})
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&req)
	if err := PluginRuntimeSvc.Kill(c.Param("id"), req.Reason); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// AdminRestartPluginRuntime 重启插件运行进程。
func AdminRestartPluginRuntime(c *gin.Context) {
	if PluginSvc == nil || PluginRuntimeSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "插件运行时未初始化"})
		return
	}
	pluginID := c.Param("id")
	manifest, ok := PluginSvc.Registry().GetManifest(pluginID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "插件不存在"})
		return
	}
	releaseDir := PluginSvc.ReleaseRoot(pluginID, manifest.Version)
	session, err := PluginRuntimeSvc.Restart(c.Request.Context(), manifest, releaseDir)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "session": session})
}

// AdminMarkPluginReady 将插件运行会话标记为可接流量。
func AdminMarkPluginReady(c *gin.Context) {
	if PluginRuntimeSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "插件运行时未初始化"})
		return
	}
	var req struct {
		InstanceID string `json:"instance_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "参数错误"})
		return
	}
	if err := PluginRuntimeSvc.MarkReady(c.Param("id"), req.InstanceID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// AdminInvokePluginGateway 执行插件调用门禁检查。
func AdminInvokePluginGateway(c *gin.Context) {
	if PluginRuntimeSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "插件运行时未初始化"})
		return
	}
	var req struct {
		RouteID string `json:"route_id"`
		Payload any    `json:"payload"`
	}
	_ = c.ShouldBindJSON(&req)
	result, err := PluginRuntimeSvc.Invoke(c.Request.Context(), service.PluginInvocationRequest{
		PluginID:  c.Param("id"),
		RouteID:   req.RouteID,
		Payload:   req.Payload,
		SubjectID: currentSubjectID(c),
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error(), "result": result})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "result": result})
}

// AdminApprovePluginTrust 批准插件版本信任状态。
func AdminApprovePluginTrust(c *gin.Context) {
	if GovernanceSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "治理服务未初始化"})
		return
	}
	var req struct {
		Version string `json:"version" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "参数错误"})
		return
	}
	subjectID := currentSubjectID(c)
	if err := GovernanceSvc.ApprovePluginTrust(c.Param("id"), req.Version, subjectID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	_ = GovernanceSvc.RecordAudit(service.AuditInput{
		ActorSubjectID: subjectID,
		Action:         "plugin.trust.approve",
		ResourceType:   "plugin",
		ResourceID:     c.Param("id"),
		RiskLevel:      "high",
		IP:             GetClientIP(c),
		UserAgent:      c.GetHeader("User-Agent"),
		Payload:        req,
	})
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func recordPluginGovernanceAction(c *gin.Context, pluginID, action, toState string) {
	if GovernanceSvc == nil {
		return
	}
	subjectID := currentSubjectID(c)
	_ = GovernanceSvc.RecordPluginStateEvent(pluginID, "", toState, action, "", subjectID)
	_ = GovernanceSvc.RecordAudit(service.AuditInput{
		ActorSubjectID: subjectID,
		Action:         action,
		ResourceType:   "plugin",
		ResourceID:     pluginID,
		RiskLevel:      "high",
		IP:             GetClientIP(c),
		UserAgent:      c.GetHeader("User-Agent"),
		Payload:        gin.H{"plugin_id": pluginID, "to_state": toState},
	})
}

func currentSubjectID(c *gin.Context) string {
	if subject, ok := CurrentSubjectContext(c); ok {
		return subject.SubjectID
	}
	return ""
}
