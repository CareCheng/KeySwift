package api

import (
	"user-frontend/internal/service"

	"github.com/gin-gonic/gin"
)

const subjectContextKey = "subject_context"

// CurrentSubjectContext 读取当前请求的统一主体上下文。
func CurrentSubjectContext(c *gin.Context) (*service.SubjectContext, bool) {
	value, exists := c.Get(subjectContextKey)
	if !exists {
		return nil, false
	}
	subject, ok := value.(*service.SubjectContext)
	return subject, ok && subject != nil
}

// MustCurrentSubjectContext 读取当前请求主体，缺失时直接返回 401。
func MustCurrentSubjectContext(c *gin.Context) (*service.SubjectContext, bool) {
	subject, ok := CurrentSubjectContext(c)
	if !ok {
		c.JSON(401, gin.H{"success": false, "error": "未登录"})
		c.Abort()
		return nil, false
	}
	return subject, true
}

func subjectHasRole(subject *service.SubjectContext, role string) bool {
	if subject == nil {
		return false
	}
	for _, item := range subject.Roles {
		if item == role {
			return true
		}
	}
	return false
}

func subjectHasPermission(subject *service.SubjectContext, permission string) bool {
	if subject == nil {
		return false
	}
	for _, item := range subject.Permissions {
		if item == permission {
			return true
		}
	}
	return false
}
