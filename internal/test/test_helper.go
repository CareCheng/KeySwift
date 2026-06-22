// Package test 测试辅助包
// 本包提供单元测试和集成测试所需的辅助函数和 Mock 对象。
//
// 测试分类：
//   - 单元测试：测试单个函数或方法
//   - 集成测试：测试多个组件协作
//   - API 测试：测试 HTTP 接口
//
// 命名规范：
//   - 测试文件以 _test.go 结尾
//   - 测试函数以 Test 开头
//   - 表驱动测试使用 tests 切片
//   - 子测试使用 t.Run()
//
// 运行测试：
//
//	# 运行所有测试
//	go test ./...
//
//	# 运行特定包的测试
//	go test ./internal/service/...
//
//	# 显示详细输出
//	go test -v ./...
//
//	# 生成覆盖率报告
//	go test -cover -coverprofile=coverage.out ./...
//	go tool cover -html=coverage.out
//
// 测试环境：
//   - 使用 SQLite 内存数据库
//   - 每个测试用例独立的数据库实例
//   - 自动清理测试数据
package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"testing"

	"user-frontend/internal/config"
	"user-frontend/internal/dbschema"
	"user-frontend/internal/model"
	"user-frontend/internal/repository"
	"user-frontend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ==================== 测试数据库设置 ====================

// SetupTestDB 创建测试数据库
// 使用 SQLite 内存数据库，每次调用都会创建一个新的独立实例
// 返回：
//   - *gorm.DB: 数据库连接
//   - func(): 清理函数，测试结束后调用
func SetupTestDB(t *testing.T) (*gorm.DB, func()) {
	// 使用 SQLite 内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("无法创建测试数据库: %v", err)
	}

	applyTestSchema(t, db)

	// 清理函数
	cleanup := func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}

	return db, cleanup
}

func applyTestSchema(t *testing.T, db *gorm.DB) {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("无法定位测试辅助文件")
	}
	schemaPath := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "database", "main", "sqlite", "schema.sql"))
	seedPath := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "database", "main", "sqlite", "seed.sql"))

	schemaSQL, _, err := dbschema.ReadSQLFile(schemaPath)
	if err != nil {
		t.Fatalf("无法读取测试数据库 schema: %v", err)
	}
	if err := dbschema.ExecSQLScript(db, schemaSQL); err != nil {
		t.Fatalf("无法构建测试数据库 schema: %v", err)
	}
	seedSQL, _, err := dbschema.ReadSQLFile(seedPath)
	if err != nil {
		t.Fatalf("无法读取测试数据库 seed: %v", err)
	}
	if err := dbschema.ExecSQLScript(db, seedSQL); err != nil {
		t.Fatalf("无法写入测试数据库 seed: %v", err)
	}
}

// ==================== 测试服务设置 ====================

// TestServices 测试服务集合
type TestServices struct {
	DB         *gorm.DB
	Repo       *repository.Repository
	UserSvc    *service.UserService
	OrderSvc   *service.OrderService
	ProductSvc *service.ProductService
	SessionSvc *service.SessionService
	BalanceSvc *service.BalanceService
}

// SetupTestServices 创建测试服务实例
func SetupTestServices(t *testing.T) (*TestServices, func()) {
	db, cleanup := SetupTestDB(t)
	repo := repository.NewRepository(db)

	cfg := &config.Config{
		ServerConfig: config.ServerConfig{
			Port:        8080,
			SystemTitle: "Test System",
		},
	}

	services := &TestServices{
		DB:         db,
		Repo:       repo,
		UserSvc:    service.NewUserService(repo),
		OrderSvc:   service.NewOrderService(repo, cfg),
		ProductSvc: service.NewProductService(repo),
		SessionSvc: service.NewSessionService(repo),
		BalanceSvc: service.NewBalanceService(repo),
	}

	return services, cleanup
}

// ==================== 测试 HTTP 设置 ====================

// SetupTestRouter 创建测试路由
func SetupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	return r
}

// TestRequest HTTP 测试请求辅助结构
type TestRequest struct {
	Method  string
	Path    string
	Body    interface{}
	Headers map[string]string
	Cookies []*http.Cookie
}

// ExecuteRequest 执行测试请求
// 参数：
//   - router: Gin 路由实例
//   - req: 测试请求配置
//
// 返回：
//   - *httptest.ResponseRecorder: 响应记录器
func ExecuteRequest(router *gin.Engine, req TestRequest) *httptest.ResponseRecorder {
	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, _ = json.Marshal(req.Body)
	}

	httpReq, _ := http.NewRequest(req.Method, req.Path, bytes.NewBuffer(bodyBytes))
	httpReq.Header.Set("Content-Type", "application/json")

	// 设置自定义请求头
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// 设置 Cookie
	for _, cookie := range req.Cookies {
		httpReq.AddCookie(cookie)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	return w
}

// ParseResponse 解析 JSON 响应
func ParseResponse(t *testing.T, w *httptest.ResponseRecorder, v interface{}) {
	if err := json.Unmarshal(w.Body.Bytes(), v); err != nil {
		t.Fatalf("无法解析响应: %v, body: %s", err, w.Body.String())
	}
}

// ==================== 测试数据生成 ====================

// TestUser 测试用户数据
type TestUser struct {
	ID       uint
	Username string
	Email    string
	Password string
}

// CreateTestUser 创建测试用户
func CreateTestUser(t *testing.T, services *TestServices, username, email, password string) *TestUser {
	user, err := services.UserSvc.Register(username, email, password, "")
	if err != nil {
		t.Fatalf("创建测试用户失败: %v", err)
	}
	return &TestUser{
		ID:       user.ID,
		Username: username,
		Email:    email,
		Password: password,
	}
}

// CreateTestProduct 创建测试商品
func CreateTestProduct(t *testing.T, services *TestServices, name string, price float64) *model.Product {
	product := &model.Product{
		Name:         name,
		Price:        price,
		Duration:     30,
		DurationUnit: "天",
		Status:       1,
		Stock:        100,
	}
	if err := services.Repo.CreateProduct(product); err != nil {
		t.Fatalf("创建测试商品失败: %v", err)
	}
	return product
}

// CreateTestOrder 创建测试订单
func CreateTestOrder(t *testing.T, services *TestServices, userID uint, productID uint) *model.Order {
	order, err := services.OrderSvc.CreateOrder(userID, "testuser", productID, "127.0.0.1")
	if err != nil {
		t.Fatalf("创建测试订单失败: %v", err)
	}
	return order
}

// ==================== 断言辅助函数 ====================

// AssertEqual 断言两个值相等
func AssertEqual(t *testing.T, expected, actual interface{}, msg string) {
	t.Helper()
	if expected != actual {
		t.Errorf("%s: 期望 %v, 实际 %v", msg, expected, actual)
	}
}

// AssertNotNil 断言值不为 nil
func AssertNotNil(t *testing.T, v interface{}, msg string) {
	t.Helper()
	if v == nil {
		t.Errorf("%s: 不应该为 nil", msg)
	}
}

// AssertNil 断言值为 nil
func AssertNil(t *testing.T, v interface{}, msg string) {
	t.Helper()
	if v != nil {
		t.Errorf("%s: 应该为 nil, 实际为 %v", msg, v)
	}
}

// AssertNoError 断言没有错误
func AssertNoError(t *testing.T, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Errorf("%s: 不应该有错误, 实际错误: %v", msg, err)
	}
}

// AssertError 断言有错误
func AssertError(t *testing.T, err error, msg string) {
	t.Helper()
	if err == nil {
		t.Errorf("%s: 应该有错误", msg)
	}
}

// AssertHTTPStatus 断言 HTTP 状态码
func AssertHTTPStatus(t *testing.T, expected int, w *httptest.ResponseRecorder, msg string) {
	t.Helper()
	if w.Code != expected {
		t.Errorf("%s: 期望状态码 %d, 实际 %d, 响应: %s", msg, expected, w.Code, w.Body.String())
	}
}

// AssertJSONSuccess 断言 JSON 响应成功
func AssertJSONSuccess(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("无法解析 JSON 响应: %v", err)
	}
	if success, ok := resp["success"].(bool); !ok || !success {
		t.Errorf("响应应该成功, 实际: %s", w.Body.String())
	}
}

// AssertJSONError 断言 JSON 响应失败
func AssertJSONError(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("无法解析 JSON 响应: %v", err)
	}
	if success, ok := resp["success"].(bool); ok && success {
		t.Errorf("响应应该失败, 实际: %s", w.Body.String())
	}
}
