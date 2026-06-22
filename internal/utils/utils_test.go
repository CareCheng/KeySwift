package utils

import (
	"strings"
	"testing"
)

// TestHashPassword 测试密码哈希
func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{"普通密码", "password123"},
		{"复杂密码", "P@ssw0rd!#$%"},
		{"长密码", "thisisaverylongpasswordthatneedstobehashedproperly"},
		{"短密码", "123456"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)
			if err != nil {
				t.Fatalf("密码哈希失败: %v", err)
			}

			// 验证哈希不为空
			if hash == "" {
				t.Error("哈希结果不应为空")
			}

			// 验证哈希与原密码不同
			if hash == tt.password {
				t.Error("哈希结果不应与原密码相同")
			}

			// 验证bcrypt前缀
			if !strings.HasPrefix(hash, "$2a$") && !strings.HasPrefix(hash, "$2b$") {
				t.Error("哈希结果应该是bcrypt格式")
			}
		})
	}
}

// TestCheckPassword 测试密码验证
func TestCheckPassword(t *testing.T) {
	password := "testpassword123"
	hash, _ := HashPassword(password)

	tests := []struct {
		name     string
		password string
		expected bool
	}{
		{"正确密码", password, true},
		{"错误密码", "wrongpassword", false},
		{"空密码", "", false},
		{"相似密码", "testpassword124", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckPassword(tt.password, hash)
			if result != tt.expected {
				t.Errorf("期望 %v, 实际 %v", tt.expected, result)
			}
		})
	}
}

// TestGenerateOrderNo 测试订单号生成
func TestGenerateOrderNo(t *testing.T) {
	t.Run("正常订单", func(t *testing.T) {
		orderNo := GenerateOrderNo()

		// 验证订单号不为空
		if orderNo == "" {
			t.Error("订单号不应为空")
		}

		// 验证唯一性（生成多个不应相同）
		orderNo2 := GenerateOrderNo()
		if orderNo == orderNo2 {
			t.Error("连续生成的订单号不应相同")
		}
	})
}

// TestGenerateLocalOrderNo 测试本地订单号生成
func TestGenerateLocalOrderNo(t *testing.T) {
	orderNo := GenerateLocalOrderNo()

	// 验证订单号不为空
	if orderNo == "" {
		t.Error("订单号不应为空")
	}

	// 验证包含订单前缀
	if !strings.HasPrefix(orderNo, "ORD_") {
		t.Error("本地订单应包含ORD_前缀")
	}

	// 验证唯一性（生成多个不应相同）
	orderNo2 := GenerateLocalOrderNo()
	if orderNo == orderNo2 {
		t.Error("连续生成的订单号不应相同")
	}
}

// TestToDays 测试时间单位转换
func TestToDays(t *testing.T) {
	tests := []struct {
		name     string
		duration int
		unit     string
		expected int
	}{
		{"30天", 30, "天", 30},
		{"1月", 1, "月", 30},
		{"1年", 1, "年", 365},
		{"2周", 2, "周", 14},
		{"默认天", 7, "", 7},
		{"未知单位", 10, "未知", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToDays(tt.duration, tt.unit)
			if result != tt.expected {
				t.Errorf("期望 %d, 实际 %d", tt.expected, result)
			}
		})
	}
}

// TestGenerateRandomString 测试随机字符串生成
func TestGenerateRandomString(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{"长度6", 6},
		{"长度12", 12},
		{"长度32", 32},
		{"长度64", 64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateRandomString(tt.length)

			// 验证长度
			if len(result) != tt.length {
				t.Errorf("期望长度 %d, 实际长度 %d", tt.length, len(result))
			}

			// 验证唯一性
			result2 := GenerateRandomString(tt.length)
			if result == result2 {
				t.Error("连续生成的随机字符串不应相同")
			}
		})
	}
}

// BenchmarkHashPassword 性能测试：密码哈希
func BenchmarkHashPassword(b *testing.B) {
	for i := 0; i < b.N; i++ {
		HashPassword("testpassword123")
	}
}

// BenchmarkCheckPassword 性能测试：密码验证
func BenchmarkCheckPassword(b *testing.B) {
	hash, _ := HashPassword("testpassword123")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CheckPassword("testpassword123", hash)
	}
}

// BenchmarkGenerateOrderNo 性能测试：订单号生成
func BenchmarkGenerateOrderNo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateOrderNo()
	}
}

// BenchmarkGenerateRandomString 性能测试：随机字符串生成
func BenchmarkGenerateRandomString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateRandomString(32)
	}
}
