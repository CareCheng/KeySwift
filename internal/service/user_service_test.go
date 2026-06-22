package service_test

import (
	"testing"

	"user-frontend/internal/test"
)

// TestUserService_Register 测试用户注册
func TestUserService_Register(t *testing.T) {
	services, cleanup := test.SetupTestServices(t)
	defer cleanup()

	// 测试用例表
	tests := []struct {
		name        string
		username    string
		email       string
		password    string
		phone       string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "正常注册",
			username:    "testuser",
			email:       "test@example.com",
			password:    "password123",
			phone:       "13800138000",
			expectError: false,
		},
		{
			name:        "用户名重复",
			username:    "testuser", // 与上一个测试用例相同
			email:       "test2@example.com",
			password:    "password123",
			phone:       "13800138001",
			expectError: true,
			errorMsg:    "用户名已存在",
		},
		{
			name:        "邮箱重复",
			username:    "testuser2",
			email:       "test@example.com", // 与第一个测试用例相同
			password:    "password123",
			phone:       "13800138002",
			expectError: true,
			errorMsg:    "邮箱已被注册",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := services.UserSvc.Register(tt.username, tt.email, tt.password, tt.phone)

			if tt.expectError {
				test.AssertError(t, err, tt.name)
				if err != nil && tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("错误信息不匹配: 期望 %s, 实际 %s", tt.errorMsg, err.Error())
				}
			} else {
				test.AssertNoError(t, err, tt.name)
				test.AssertNotNil(t, user, "用户对象")
				test.AssertEqual(t, tt.username, user.Username, "用户名")
				test.AssertEqual(t, tt.email, user.Email, "邮箱")
			}
		})
	}
}

// TestUserService_Login 测试用户登录
func TestUserService_Login(t *testing.T) {
	services, cleanup := test.SetupTestServices(t)
	defer cleanup()

	// 先创建测试用户
	test.CreateTestUser(t, services, "loginuser", "login@example.com", "password123")

	tests := []struct {
		name        string
		username    string
		password    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "正常登录",
			username:    "loginuser",
			password:    "password123",
			expectError: false,
		},
		{
			name:        "用户名错误",
			username:    "wronguser",
			password:    "password123",
			expectError: true,
			errorMsg:    "用户名或密码错误",
		},
		{
			name:        "密码错误",
			username:    "loginuser",
			password:    "wrongpassword",
			expectError: true,
			errorMsg:    "用户名或密码错误",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := services.UserSvc.Login(tt.username, tt.password, "127.0.0.1")

			if tt.expectError {
				test.AssertError(t, err, tt.name)
				if err != nil && tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("错误信息不匹配: 期望 %s, 实际 %s", tt.errorMsg, err.Error())
				}
			} else {
				test.AssertNoError(t, err, tt.name)
				test.AssertNotNil(t, user, "用户对象")
				test.AssertEqual(t, tt.username, user.Username, "用户名")
			}
		})
	}
}

// TestUserService_UpdatePassword 测试修改密码
func TestUserService_UpdatePassword(t *testing.T) {
	services, cleanup := test.SetupTestServices(t)
	defer cleanup()

	// 创建测试用户
	testUser := test.CreateTestUser(t, services, "pwduser", "pwd@example.com", "oldpassword")

	tests := []struct {
		name        string
		oldPassword string
		newPassword string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "正常修改密码",
			oldPassword: "oldpassword",
			newPassword: "newpassword",
			expectError: false,
		},
		{
			name:        "旧密码错误",
			oldPassword: "wrongpassword",
			newPassword: "newpassword2",
			expectError: true,
			errorMsg:    "原密码错误",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := services.UserSvc.UpdatePassword(testUser.ID, tt.oldPassword, tt.newPassword)

			if tt.expectError {
				test.AssertError(t, err, tt.name)
				if err != nil && tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("错误信息不匹配: 期望 %s, 实际 %s", tt.errorMsg, err.Error())
				}
			} else {
				test.AssertNoError(t, err, tt.name)
				// 验证新密码可以登录
				_, err := services.UserSvc.Login(testUser.Username, tt.newPassword, "127.0.0.1")
				test.AssertNoError(t, err, "使用新密码登录")
			}
		})
	}
}

// TestUserService_Enable2FA 测试启用两步验证
func TestUserService_Enable2FA(t *testing.T) {
	services, cleanup := test.SetupTestServices(t)
	defer cleanup()

	testUser := test.CreateTestUser(t, services, "2fauser", "2fa@example.com", "password123")

	// 测试启用2FA
	secret := "JBSWY3DPEHPK3PXP"
	err := services.UserSvc.Enable2FA(testUser.ID, secret)
	test.AssertNoError(t, err, "启用2FA")

	// 验证2FA状态
	enabled, savedSecret, err := services.UserSvc.GetUser2FAStatus(testUser.ID)
	test.AssertNoError(t, err, "获取2FA状态")
	test.AssertEqual(t, true, enabled, "2FA启用状态")
	test.AssertEqual(t, secret, savedSecret, "2FA密钥")

	// 测试禁用2FA
	err = services.UserSvc.Disable2FA(testUser.ID)
	test.AssertNoError(t, err, "禁用2FA")

	// 再次验证2FA状态
	enabled, _, err = services.UserSvc.GetUser2FAStatus(testUser.ID)
	test.AssertNoError(t, err, "获取2FA状态")
	test.AssertEqual(t, false, enabled, "2FA禁用状态")
}

// TestUserService_GetUserByID 测试根据ID获取用户
func TestUserService_GetUserByID(t *testing.T) {
	services, cleanup := test.SetupTestServices(t)
	defer cleanup()

	testUser := test.CreateTestUser(t, services, "getuser", "get@example.com", "password123")

	// 测试获取存在的用户
	user, err := services.UserSvc.GetUserByID(testUser.ID)
	test.AssertNoError(t, err, "获取用户")
	test.AssertNotNil(t, user, "用户对象")
	test.AssertEqual(t, testUser.Username, user.Username, "用户名")

	// 测试获取不存在的用户
	_, err = services.UserSvc.GetUserByID(99999)
	test.AssertError(t, err, "获取不存在的用户")
}

// TestUserService_BindEmail 测试绑定邮箱
func TestUserService_BindEmail(t *testing.T) {
	services, cleanup := test.SetupTestServices(t)
	defer cleanup()

	// 创建无邮箱用户
	user, err := services.UserSvc.Register("noemail", "", "password123", "")
	test.AssertNoError(t, err, "创建用户")

	// 绑定邮箱
	err = services.UserSvc.BindEmail(user.ID, "newemail@example.com")
	test.AssertNoError(t, err, "绑定邮箱")

	// 验证邮箱已绑定
	updatedUser, _ := services.UserSvc.GetUserByID(user.ID)
	test.AssertEqual(t, "newemail@example.com", updatedUser.Email, "绑定的邮箱")
	test.AssertEqual(t, true, updatedUser.EmailVerified, "邮箱验证状态")

	// 测试绑定已存在的邮箱
	user2, _ := services.UserSvc.Register("user2", "", "password123", "")
	err = services.UserSvc.BindEmail(user2.ID, "newemail@example.com")
	test.AssertError(t, err, "绑定已存在的邮箱")
}
