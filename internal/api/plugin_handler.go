package api

import (
	"context"
	"net/http"

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

// AdminGetPluginConfigs 获取插件配置记录。
func AdminGetPluginConfigs(c *gin.Context) {
	if PluginSvc == nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "插件服务未初始化"})
		return
	}
	pluginID := c.Param("id")
	configs, ok := PluginSvc.GetPluginConfigs(pluginID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "插件配置不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "configs": configs})
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
	if err := PluginSvc.EnablePlugin(c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "插件已启用"})
}

// AdminDisablePlugin 停用插件治理状态。
func AdminDisablePlugin(c *gin.Context) {
	if PluginSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "插件服务未初始化"})
		return
	}
	if err := PluginSvc.DisablePlugin(c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "插件已停用"})
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
	if PluginSvc == nil {
		c.JSON(http.StatusOK, gin.H{"success": true, "permissions": []any{}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"permissions": PluginSvc.Permissions(),
	})
}

// AdminGetPluginConfigSchemas 获取插件配置 schema 声明。
func AdminGetPluginConfigSchemas(c *gin.Context) {
	if PluginSvc == nil {
		c.JSON(http.StatusOK, gin.H{"success": true, "schemas": []any{}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"schemas": PluginSvc.ConfigSchemas(),
	})
}
