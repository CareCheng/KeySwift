// Package api 提供人机验证公开接口与后台配置候选项接口。
package api

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"user-frontend/internal/service"

	"github.com/gin-gonic/gin"
)

// PluginFrontendAsset 公开提供插件声明的前端资源，仅允许访问当前版本 frontend 目录。
func PluginFrontendAsset(c *gin.Context) {
	if PluginSvc == nil {
		c.Status(http.StatusNotFound)
		return
	}
	pluginID := strings.TrimSpace(c.Param("plugin_id"))
	version := strings.TrimSpace(c.Param("version"))
	manifest, ok := PluginSvc.Registry().GetManifest(pluginID)
	if !ok || manifest.Version != version {
		c.Status(http.StatusNotFound)
		return
	}

	rawPath := strings.TrimPrefix(c.Param("filepath"), "/")
	if rawPath == "" {
		c.Status(http.StatusNotFound)
		return
	}
	cleanPath := filepath.Clean(rawPath)
	if filepath.IsAbs(cleanPath) || cleanPath == "." || strings.HasPrefix(cleanPath, "..") {
		c.Status(http.StatusBadRequest)
		return
	}

	root := filepath.Join(PluginSvc.ReleaseRoot(pluginID, version), "frontend")
	target := filepath.Join(root, cleanPath)
	rootClean := filepath.Clean(root)
	targetClean := filepath.Clean(target)
	relativeTarget, err := filepath.Rel(rootClean, targetClean)
	if err != nil || relativeTarget == "." || filepath.IsAbs(relativeTarget) || strings.HasPrefix(relativeTarget, ".."+string(os.PathSeparator)) || relativeTarget == ".." {
		c.Status(http.StatusBadRequest)
		return
	}
	info, err := os.Stat(targetClean)
	if err != nil || info.IsDir() {
		c.Status(http.StatusNotFound)
		return
	}
	c.File(targetClean)
}

// HumanVerificationChallenge 创建当前启用 provider 的人机验证挑战。
func HumanVerificationChallenge(c *gin.Context) {
	humanSvc := currentHumanVerificationService()
	var req struct {
		Scope      string `json:"scope" binding:"required"`
		ProviderID string `json:"provider_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	challenge, err := humanSvc.CreateChallenge(ctx, service.HumanVerificationChallengeRequest{
		Scope:      strings.TrimSpace(req.Scope),
		ProviderID: strings.TrimSpace(req.ProviderID),
		RequestContext: service.HumanVerificationRequestContext{
			ClientIP:  GetClientIP(c),
			UserAgent: c.GetHeader("User-Agent"),
			RequestID: c.GetHeader("X-Request-ID"),
		},
	})
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"success": true, "challenge": challenge})
}

// AdminListHumanVerificationProviders 返回已安装且已启用的人机验证插件 provider。
func AdminListHumanVerificationProviders(c *gin.Context) {
	if HumanVerificationSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "人机验证服务未初始化"})
		return
	}
	scopes := parseHumanVerificationScopes(c.Query("scope"))
	c.JSON(200, gin.H{
		"success":   true,
		"providers": HumanVerificationSvc.ListProviders(scopes),
	})
}

func parseHumanVerificationScopes(raw string) []string {
	items := strings.Split(raw, ",")
	scopes := make([]string, 0, len(items))
	for _, item := range items {
		scope := strings.TrimSpace(item)
		if scope != "" {
			scopes = append(scopes, scope)
		}
	}
	return scopes
}

func verifyHumanVerificationForRequest(c *gin.Context, scope string, payload *service.HumanVerificationPayload) bool {
	humanSvc := currentHumanVerificationService()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	err := humanSvc.Verify(ctx, service.HumanVerificationVerifyRequest{
		Scope:   scope,
		Payload: payload,
		RequestContext: service.HumanVerificationRequestContext{
			ClientIP:  GetClientIP(c),
			UserAgent: c.GetHeader("User-Agent"),
			RequestID: c.GetHeader("X-Request-ID"),
		},
	})
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return false
	}
	return true
}

func currentHumanVerificationService() *service.HumanVerificationService {
	if HumanVerificationSvc != nil {
		return HumanVerificationSvc
	}
	return service.NewHumanVerificationService(nil, nil)
}
