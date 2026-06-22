package service

import (
	"encoding/json"
	"time"

	"user-frontend/internal/cache"
	"user-frontend/internal/model"
	"user-frontend/internal/repository"

	"github.com/google/uuid"
)

// SessionService 会话服务（数据库持久化 + 缓存加速）
type SessionService struct {
	repo *repository.Repository
}

// 会话过期时间
const (
	UserSessionDuration      = 2 * time.Hour      // 用户会话2小时
	AdminSessionDuration     = 1 * time.Hour      // 管理员会话1小时
	RememberMeDuration       = 7 * 24 * time.Hour // 记住我7天
	LongLivedSessionDuration = 10 * 365 * 24 * time.Hour
)

func NewSessionService(repo *repository.Repository) *SessionService {
	return &SessionService{repo: repo}
}

// CreateUserSession 创建用户会话
func (s *SessionService) CreateUserSession(userID uint, username, ip, userAgent string, remember bool) (string, error) {
	duration := UserSessionDuration
	if remember {
		duration = RememberMeDuration
	}
	return s.CreateUserSessionWithDuration(userID, username, ip, userAgent, duration)
}

// CreateUserSessionWithDuration 按指定有效期创建用户会话。
func (s *SessionService) CreateUserSessionWithDuration(userID uint, username, ip, userAgent string, duration time.Duration) (string, error) {
	sessionID := uuid.New().String()
	if duration <= 0 {
		duration = UserSessionDuration
	}
	session := &model.UserSession{
		SessionID: sessionID,
		UserID:    userID,
		Username:  username,
		IP:        ip,
		UserAgent: userAgent,
		ExpiresAt: time.Now().Add(duration),
	}

	// 先写数据库
	if err := s.repo.CreateUserSession(session); err != nil {
		return "", err
	}

	// 写缓存（异步方式，不阻塞主流程）
	go s.cacheUserSession(session, duration)

	return sessionID, nil
}

// GetUserSession 获取用户会话
func (s *SessionService) GetUserSession(sessionID string) (*model.UserSession, error) {
	// 优先从缓存读取
	if session := s.getUserSessionFromCache(sessionID); session != nil {
		return session, nil
	}

	// 缓存未命中，从数据库读取
	session, err := s.repo.GetUserSession(sessionID)
	if err != nil {
		return nil, err
	}

	// 回填缓存
	if session != nil {
		remaining := time.Until(session.ExpiresAt)
		if remaining > 0 {
			go s.cacheUserSession(session, remaining)
		}
	}

	return session, nil
}

// RefreshUserSession 刷新用户会话
func (s *SessionService) RefreshUserSession(sessionID string) error {
	session, err := s.GetUserSession(sessionID)
	if err != nil {
		return err
	}

	// 如果会话还有超过一半的时间，不刷新
	remaining := time.Until(session.ExpiresAt)
	if remaining > UserSessionDuration/2 {
		return nil
	}

	session.ExpiresAt = time.Now().Add(UserSessionDuration)

	// 更新数据库
	if err := s.repo.UpdateUserSession(session); err != nil {
		return err
	}

	// 更新缓存
	go s.cacheUserSession(session, UserSessionDuration)

	return nil
}

// DeleteUserSession 删除用户会话
func (s *SessionService) DeleteUserSession(sessionID string) error {
	// 删除缓存
	s.deleteUserSessionFromCache(sessionID)

	// 删除数据库
	return s.repo.DeleteUserSession(sessionID)
}

// DeleteUserSessionsByUserID 删除用户的所有会话
func (s *SessionService) DeleteUserSessionsByUserID(userID uint) error {
	// 先获取用户的所有会话ID，用于删除缓存
	sessions, _ := s.repo.GetUserSessionsByUserID(userID)
	for _, session := range sessions {
		s.deleteUserSessionFromCache(session.SessionID)
	}

	return s.repo.DeleteUserSessionsByUserID(userID)
}

// CreateAdminSession 创建管理员会话
func (s *SessionService) CreateAdminSession(username, role, ip, userAgent string, remember bool) (string, error) {
	duration := AdminSessionDuration
	if remember {
		duration = 24 * time.Hour // 管理员记住我24小时
	}
	return s.CreateAdminSessionWithDuration(username, role, ip, userAgent, duration)
}

// CreateAdminSessionWithDuration 按指定有效期创建管理员会话。
func (s *SessionService) CreateAdminSessionWithDuration(username, role, ip, userAgent string, duration time.Duration) (string, error) {
	sessionID := uuid.New().String()
	if duration <= 0 {
		duration = AdminSessionDuration
	}
	session := &model.AdminSession{
		SessionID: sessionID,
		Username:  username,
		Role:      role,
		IP:        ip,
		UserAgent: userAgent,
		Verified:  false,
		ExpiresAt: time.Now().Add(duration),
	}

	// 先写数据库
	if err := s.repo.CreateAdminSession(session); err != nil {
		return "", err
	}

	// 写缓存
	go s.cacheAdminSession(session, duration)

	return sessionID, nil
}

// GetAdminSession 获取管理员会话
func (s *SessionService) GetAdminSession(sessionID string) (*model.AdminSession, error) {
	// 优先从缓存读取
	if session := s.getAdminSessionFromCache(sessionID); session != nil {
		return session, nil
	}

	// 缓存未命中，从数据库读取
	session, err := s.repo.GetAdminSession(sessionID)
	if err != nil {
		return nil, err
	}

	// 回填缓存
	if session != nil {
		remaining := time.Until(session.ExpiresAt)
		if remaining > 0 {
			go s.cacheAdminSession(session, remaining)
		}
	}

	return session, nil
}

// SetAdminSessionVerified 设置管理员会话已验证
func (s *SessionService) SetAdminSessionVerified(sessionID string) error {
	session, err := s.GetAdminSession(sessionID)
	if err != nil {
		return err
	}
	session.Verified = true

	// 更新数据库
	if err := s.repo.UpdateAdminSession(session); err != nil {
		return err
	}

	// 更新缓存
	remaining := time.Until(session.ExpiresAt)
	if remaining > 0 {
		go s.cacheAdminSession(session, remaining)
	}

	return nil
}

// DeleteAdminSession 删除管理员会话
func (s *SessionService) DeleteAdminSession(sessionID string) error {
	// 删除缓存
	s.deleteAdminSessionFromCache(sessionID)

	// 删除数据库
	return s.repo.DeleteAdminSession(sessionID)
}

// CleanupExpiredSessions 清理过期会话
func (s *SessionService) CleanupExpiredSessions() {
	s.repo.DeleteExpiredUserSessions()
	s.repo.DeleteExpiredAdminSessions()
	// 缓存中的过期会话会自动过期，无需手动清理
}

// ==================== 缓存辅助方法 ====================

// cacheUserSession 缓存用户会话
func (s *SessionService) cacheUserSession(session *model.UserSession, ttl time.Duration) {
	cm := cache.GetCacheManager()
	if cm == nil {
		return
	}

	key := cache.UserSessionKey(session.SessionID)
	data, err := json.Marshal(session)
	if err != nil {
		return
	}

	cm.Set(key, string(data), ttl)
}

// getUserSessionFromCache 从缓存获取用户会话
func (s *SessionService) getUserSessionFromCache(sessionID string) *model.UserSession {
	cm := cache.GetCacheManager()
	if cm == nil {
		return nil
	}

	key := cache.UserSessionKey(sessionID)
	data, ok := cm.GetString(key)
	if !ok {
		return nil
	}

	var session model.UserSession
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return nil
	}

	// 检查是否过期
	if time.Now().After(session.ExpiresAt) {
		s.deleteUserSessionFromCache(sessionID)
		return nil
	}

	return &session
}

// deleteUserSessionFromCache 从缓存删除用户会话
func (s *SessionService) deleteUserSessionFromCache(sessionID string) {
	cm := cache.GetCacheManager()
	if cm == nil {
		return
	}

	key := cache.UserSessionKey(sessionID)
	cm.Delete(key)
}

// cacheAdminSession 缓存管理员会话
func (s *SessionService) cacheAdminSession(session *model.AdminSession, ttl time.Duration) {
	cm := cache.GetCacheManager()
	if cm == nil {
		return
	}

	key := cache.AdminSessionKey(session.SessionID)
	data, err := json.Marshal(session)
	if err != nil {
		return
	}

	cm.Set(key, string(data), ttl)
}

// getAdminSessionFromCache 从缓存获取管理员会话
func (s *SessionService) getAdminSessionFromCache(sessionID string) *model.AdminSession {
	cm := cache.GetCacheManager()
	if cm == nil {
		return nil
	}

	key := cache.AdminSessionKey(sessionID)
	data, ok := cm.GetString(key)
	if !ok {
		return nil
	}

	var session model.AdminSession
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return nil
	}

	// 检查是否过期
	if time.Now().After(session.ExpiresAt) {
		s.deleteAdminSessionFromCache(sessionID)
		return nil
	}

	return &session
}

// deleteAdminSessionFromCache 从缓存删除管理员会话
func (s *SessionService) deleteAdminSessionFromCache(sessionID string) {
	cm := cache.GetCacheManager()
	if cm == nil {
		return
	}

	key := cache.AdminSessionKey(sessionID)
	cm.Delete(key)
}
