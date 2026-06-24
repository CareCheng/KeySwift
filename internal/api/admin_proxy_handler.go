package api

import (
	"user-frontend/internal/service"

	"github.com/gin-gonic/gin"
)

// AdminGetReverseProxyConfig 获取反向代理访问配置。
func AdminGetReverseProxyConfig(c *gin.Context) {
	if ConfigSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}
	cfg, err := ConfigSvc.GetReverseProxyConfig()
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取反向代理配置失败: " + err.Error()})
		return
	}
	c.JSON(200, gin.H{
		"success": true,
		"config":  cfg,
	})
}

// AdminSaveReverseProxyConfig 保存反向代理访问配置。
func AdminSaveReverseProxyConfig(c *gin.Context) {
	if ConfigSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	var req service.ReverseProxyConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	cfg, err := ConfigSvc.SaveReverseProxyConfig(&req)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}
	if _, err := RefreshReverseProxyRuntime(); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "配置已保存，但刷新运行时配置失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success":      true,
		"message":      "反向代理配置已保存",
		"need_restart": true,
		"config":       cfg,
	})
}

// AdminReverseProxyDiagnostics 返回当前请求的反向代理识别诊断信息。
func AdminReverseProxyDiagnostics(c *gin.Context) {
	c.JSON(200, gin.H{
		"success": true,
		"diagnostics": gin.H{
			"client_ip":       GetClientIPInfo(c),
			"external_access": GetExternalRequestInfo(c),
			"headers": gin.H{
				"x_forwarded_for":   c.GetHeader("X-Forwarded-For"),
				"x_real_ip":         c.GetHeader("X-Real-IP"),
				"x_forwarded_proto": c.GetHeader("X-Forwarded-Proto"),
				"x_forwarded_host":  c.GetHeader("X-Forwarded-Host"),
				"x_forwarded_port":  c.GetHeader("X-Forwarded-Port"),
				"host":              c.Request.Host,
			},
		},
	})
}
