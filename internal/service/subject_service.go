package service

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

const (
	SubjectTypeUser  = "user"
	SubjectTypeAdmin = "admin"
)

// DataScope 描述主体可访问的数据范围。
//
// 当前阶段先提供统一结构，具体授权落库在后续 PRE-S3-006 中实现。
type DataScope struct {
	ResourceType  string `json:"resource_type"`
	ScopeType     string `json:"scope_type"`
	ScopeValue    string `json:"scope_value"`
	OwnerPluginID string `json:"owner_plugin_id"`
}

// SessionContext 是用户、管理员和后续插件调用共享的会话视图。
type SessionContext struct {
	SessionID   string    `json:"session_id"`
	SubjectID   string    `json:"subject_id"`
	SubjectType string    `json:"subject_type"`
	ExpiresAt   time.Time `json:"expires_at"`
	IP          string    `json:"ip"`
	UserAgent   string    `json:"user_agent"`
}

// SubjectContext 是宿主和插件统一使用的调用主体上下文。
type SubjectContext struct {
	SubjectID   string         `json:"subject_id"`
	SubjectType string         `json:"subject_type"`
	UserID      uint           `json:"user_id,omitempty"`
	AdminID     uint           `json:"admin_id,omitempty"`
	Username    string         `json:"username"`
	DisplayName string         `json:"display_name"`
	Roles       []string       `json:"roles"`
	Permissions []string       `json:"permissions"`
	DataScopes  []DataScope    `json:"data_scopes"`
	Session     SessionContext `json:"session"`
}

// SubjectService 将现有用户会话和管理员会话映射为统一主体上下文。
type SubjectService struct {
	sessionSvc *SessionService
	userSvc    *UserService
	roleSvc    *RoleService
}

func NewSubjectService(sessionSvc *SessionService, userSvc *UserService, roleSvc *RoleService) *SubjectService {
	return &SubjectService{
		sessionSvc: sessionSvc,
		userSvc:    userSvc,
		roleSvc:    roleSvc,
	}
}

// BuildUserSubjectBySession 根据用户会话构建统一主体上下文。
func (s *SubjectService) BuildUserSubjectBySession(sessionID string) (*SubjectContext, error) {
	if s == nil || s.sessionSvc == nil {
		return nil, errors.New("主体服务未初始化")
	}
	session, err := s.sessionSvc.GetUserSession(sessionID)
	if err != nil {
		return nil, err
	}
	displayName := session.Username
	if s.userSvc != nil {
		if user, err := s.userSvc.GetUserByID(session.UserID); err == nil && user != nil && user.Username != "" {
			displayName = user.Username
		}
	}
	subjectID := fmt.Sprintf("%s:%d", SubjectTypeUser, session.UserID)
	return &SubjectContext{
		SubjectID:   subjectID,
		SubjectType: SubjectTypeUser,
		UserID:      session.UserID,
		Username:    session.Username,
		DisplayName: displayName,
		Roles:       []string{SubjectTypeUser},
		Permissions: nil,
		DataScopes:  nil,
		Session: SessionContext{
			SessionID:   session.SessionID,
			SubjectID:   subjectID,
			SubjectType: SubjectTypeUser,
			ExpiresAt:   session.ExpiresAt,
			IP:          session.IP,
			UserAgent:   session.UserAgent,
		},
	}, nil
}

// BuildAdminSubjectBySession 根据管理员会话构建统一主体上下文。
func (s *SubjectService) BuildAdminSubjectBySession(sessionID string) (*SubjectContext, error) {
	if s == nil || s.sessionSvc == nil {
		return nil, errors.New("主体服务未初始化")
	}
	session, err := s.sessionSvc.GetAdminSession(sessionID)
	if err != nil {
		return nil, err
	}
	if !session.Verified {
		return nil, errors.New("管理员会话未完成验证")
	}
	return s.BuildAdminSubject(session.SessionID, session.Username, session.Role, session.ExpiresAt, session.IP, session.UserAgent)
}

// BuildAdminSubject 构建管理员统一主体上下文。
func (s *SubjectService) BuildAdminSubject(sessionID, username, roleName string, expiresAt time.Time, ip, userAgent string) (*SubjectContext, error) {
	if username == "" {
		return nil, errors.New("管理员用户名不能为空")
	}
	if roleName == "" {
		roleName = "admin"
	}
	adminID := uint(0)
	displayName := username
	permissions := make([]string, 0)
	if roleName == "super_admin" && s != nil && s.roleSvc != nil {
		for _, permission := range s.roleSvc.GetAllPermissions() {
			permissions = append(permissions, permission.Code)
		}
	} else if s != nil && s.roleSvc != nil {
		if admin, err := s.roleSvc.GetAdminByUsername(username); err == nil && admin != nil {
			adminID = admin.ID
			if admin.Nickname != "" {
				displayName = admin.Nickname
			}
			if admin.Role != nil && admin.Role.Name != "" {
				roleName = admin.Role.Name
			}
			if perms, err := s.roleSvc.GetAdminPermissions(admin.ID); err == nil {
				permissions = perms
			}
		}
	}
	subjectID := fmt.Sprintf("%s:%s", SubjectTypeAdmin, username)
	if adminID > 0 {
		subjectID = SubjectTypeAdmin + ":" + strconv.FormatUint(uint64(adminID), 10)
	}
	return &SubjectContext{
		SubjectID:   subjectID,
		SubjectType: SubjectTypeAdmin,
		AdminID:     adminID,
		Username:    username,
		DisplayName: displayName,
		Roles:       []string{roleName},
		Permissions: permissions,
		DataScopes:  nil,
		Session: SessionContext{
			SessionID:   sessionID,
			SubjectID:   subjectID,
			SubjectType: SubjectTypeAdmin,
			ExpiresAt:   expiresAt,
			IP:          ip,
			UserAgent:   userAgent,
		},
	}, nil
}
