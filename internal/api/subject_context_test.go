package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"user-frontend/internal/service"

	"github.com/gin-gonic/gin"
)

func TestCurrentSubjectContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)
	subject := &service.SubjectContext{
		SubjectID:   "admin:1",
		SubjectType: service.SubjectTypeAdmin,
		Roles:       []string{"admin"},
		Permissions: []string{"plugin:view"},
	}
	c.Set(subjectContextKey, subject)

	got, ok := CurrentSubjectContext(c)
	if !ok {
		t.Fatal("应能读取当前主体上下文")
	}
	if got.SubjectID != subject.SubjectID {
		t.Fatalf("主体 ID 不符合预期: %s", got.SubjectID)
	}
}

func TestPermissionRequiredUsesSubjectContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/protected", func(c *gin.Context) {
		c.Set(subjectContextKey, &service.SubjectContext{
			SubjectID:   "admin:1",
			SubjectType: service.SubjectTypeAdmin,
			Roles:       []string{"admin"},
			Permissions: []string{"plugin:view"},
		})
	}, PermissionRequired("plugin:view"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("主体权限应允许访问，状态码: %d，响应: %s", w.Code, w.Body.String())
	}
}

func TestPermissionRequiredRejectsSubjectWithoutPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/protected", func(c *gin.Context) {
		c.Set(subjectContextKey, &service.SubjectContext{
			SubjectID:   "admin:1",
			SubjectType: service.SubjectTypeAdmin,
			Roles:       []string{"admin"},
			Permissions: []string{"dashboard:view"},
		})
	}, PermissionRequired("plugin:view"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("缺少权限应拒绝访问，状态码: %d，响应: %s", w.Code, w.Body.String())
	}
}
