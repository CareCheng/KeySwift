package service_test

import (
	"testing"

	"user-frontend/internal/model"
	"user-frontend/internal/service"
	"user-frontend/internal/test"
)

func TestSubjectService_BuildUserSubjectBySession(t *testing.T) {
	services, cleanup := test.SetupTestServices(t)
	defer cleanup()

	user := test.CreateTestUser(t, services, "subject_user", "subject_user@example.com", "password123")
	sessionID, err := services.SessionSvc.CreateUserSession(user.ID, user.Username, "127.0.0.1", "test-agent", false)
	if err != nil {
		t.Fatalf("创建用户会话失败: %v", err)
	}

	subjectSvc := service.NewSubjectService(services.SessionSvc, services.UserSvc, nil)
	subject, err := subjectSvc.BuildUserSubjectBySession(sessionID)
	if err != nil {
		t.Fatalf("构建用户主体上下文失败: %v", err)
	}

	if subject.SubjectType != service.SubjectTypeUser {
		t.Fatalf("主体类型不符合预期: %s", subject.SubjectType)
	}
	if subject.SubjectID != "user:1" {
		t.Fatalf("主体 ID 不符合预期: %s", subject.SubjectID)
	}
	if subject.UserID != user.ID {
		t.Fatalf("用户 ID 不符合预期: %d", subject.UserID)
	}
	if subject.Session.SessionID != sessionID {
		t.Fatalf("会话 ID 不符合预期: %s", subject.Session.SessionID)
	}
}

func TestSubjectService_BuildAdminSubjectBySession(t *testing.T) {
	services, cleanup := test.SetupTestServices(t)
	defer cleanup()

	roleSvc := service.NewRoleService(services.Repo)
	if err := roleSvc.InitDefaultRoles(); err != nil {
		t.Fatalf("初始化默认角色失败: %v", err)
	}
	admin, err := roleSvc.CreateAdmin("subject_admin", "password123", "admin@example.com", "主体管理员", 2)
	if err != nil {
		t.Fatalf("创建管理员失败: %v", err)
	}
	sessionID, err := services.SessionSvc.CreateAdminSession(admin.Username, "admin", "127.0.0.1", "test-agent", false)
	if err != nil {
		t.Fatalf("创建管理员会话失败: %v", err)
	}
	if err := services.SessionSvc.SetAdminSessionVerified(sessionID); err != nil {
		t.Fatalf("设置管理员会话验证状态失败: %v", err)
	}

	subjectSvc := service.NewSubjectService(services.SessionSvc, services.UserSvc, roleSvc)
	subject, err := subjectSvc.BuildAdminSubjectBySession(sessionID)
	if err != nil {
		t.Fatalf("构建管理员主体上下文失败: %v", err)
	}

	if subject.SubjectType != service.SubjectTypeAdmin {
		t.Fatalf("主体类型不符合预期: %s", subject.SubjectType)
	}
	if subject.SubjectID != "admin:1" {
		t.Fatalf("主体 ID 不符合预期: %s", subject.SubjectID)
	}
	if subject.AdminID != admin.ID {
		t.Fatalf("管理员 ID 不符合预期: %d", subject.AdminID)
	}
	if len(subject.Roles) != 1 || subject.Roles[0] != "admin" {
		t.Fatalf("管理员角色不符合预期: %#v", subject.Roles)
	}
	if len(subject.Permissions) == 0 {
		t.Fatal("管理员主体上下文应包含角色权限")
	}
}

func TestSubjectService_RejectsUnverifiedAdminSession(t *testing.T) {
	services, cleanup := test.SetupTestServices(t)
	defer cleanup()

	if err := services.Repo.GetDB().Create(&model.AdminRole{
		Name:        "admin",
		Description: "测试管理员",
		Permissions: `["dashboard:view"]`,
		Status:      1,
	}).Error; err != nil {
		t.Fatalf("创建测试角色失败: %v", err)
	}
	sessionID, err := services.SessionSvc.CreateAdminSession("unverified_admin", "admin", "127.0.0.1", "test-agent", false)
	if err != nil {
		t.Fatalf("创建管理员会话失败: %v", err)
	}

	subjectSvc := service.NewSubjectService(services.SessionSvc, services.UserSvc, service.NewRoleService(services.Repo))
	if _, err := subjectSvc.BuildAdminSubjectBySession(sessionID); err == nil {
		t.Fatal("未完成验证的管理员会话不应生成主体上下文")
	}
}
