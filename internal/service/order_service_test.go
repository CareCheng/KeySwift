package service_test

import (
	"testing"

	"user-frontend/internal/model"
	"user-frontend/internal/test"
)

// TestOrderService_CreateOrder 测试创建订单
func TestOrderService_CreateOrder(t *testing.T) {
	services, cleanup := test.SetupTestServices(t)
	defer cleanup()

	// 创建测试用户和商品
	testUser := test.CreateTestUser(t, services, "orderuser", "order@example.com", "password123")
	testProduct := test.CreateTestProduct(t, services, "测试商品", 99.99)

	tests := []struct {
		name        string
		userID      uint
		productID   uint
		expectError bool
		errorMsg    string
	}{
		{
			name:        "正常创建订单",
			userID:      testUser.ID,
			productID:   testProduct.ID,
			expectError: false,
		},
		{
			name:        "商品不存在",
			userID:      testUser.ID,
			productID:   99999,
			expectError: true,
			errorMsg:    "商品不存在",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order, err := services.OrderSvc.CreateOrder(tt.userID, "testuser", tt.productID, "127.0.0.1")

			if tt.expectError {
				test.AssertError(t, err, tt.name)
				if err != nil && tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("错误信息不匹配: 期望 %s, 实际 %s", tt.errorMsg, err.Error())
				}
			} else {
				test.AssertNoError(t, err, tt.name)
				test.AssertNotNil(t, order, "订单对象")
				test.AssertEqual(t, tt.userID, order.UserID, "用户ID")
				test.AssertEqual(t, tt.productID, order.ProductID, "商品ID")
				test.AssertEqual(t, 0, order.Status, "订单状态应为待支付")
			}
		})
	}
}

// TestOrderService_GetOrderByOrderNo 测试根据订单号获取订单
func TestOrderService_GetOrderByOrderNo(t *testing.T) {
	services, cleanup := test.SetupTestServices(t)
	defer cleanup()

	// 创建测试数据
	testUser := test.CreateTestUser(t, services, "orderuser", "order@example.com", "password123")
	testProduct := test.CreateTestProduct(t, services, "测试商品", 99.99)
	testOrder := test.CreateTestOrder(t, services, testUser.ID, testProduct.ID)

	// 测试获取存在的订单
	order, err := services.OrderSvc.GetOrderByOrderNo(testOrder.OrderNo)
	test.AssertNoError(t, err, "获取订单")
	test.AssertNotNil(t, order, "订单对象")
	test.AssertEqual(t, testOrder.OrderNo, order.OrderNo, "订单号")

	// 测试获取不存在的订单
	_, err = services.OrderSvc.GetOrderByOrderNo("NONEXISTENT")
	test.AssertError(t, err, "获取不存在的订单")
}

// TestOrderService_CancelOrder 测试取消订单
func TestOrderService_CancelOrder(t *testing.T) {
	services, cleanup := test.SetupTestServices(t)
	defer cleanup()

	// 创建测试数据
	testUser := test.CreateTestUser(t, services, "canceluser", "cancel@example.com", "password123")
	testProduct := test.CreateTestProduct(t, services, "测试商品", 99.99)
	testOrder := test.CreateTestOrder(t, services, testUser.ID, testProduct.ID)

	tests := []struct {
		name        string
		orderNo     string
		userID      uint
		expectError bool
		errorMsg    string
	}{
		{
			name:        "正常取消订单",
			orderNo:     testOrder.OrderNo,
			userID:      testUser.ID,
			expectError: false,
		},
		{
			name:        "订单不存在",
			orderNo:     "NONEXISTENT",
			userID:      testUser.ID,
			expectError: true,
			errorMsg:    "订单不存在",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := services.OrderSvc.CancelOrder(tt.orderNo, tt.userID)

			if tt.expectError {
				test.AssertError(t, err, tt.name)
				if err != nil && tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("错误信息不匹配: 期望 %s, 实际 %s", tt.errorMsg, err.Error())
				}
			} else {
				test.AssertNoError(t, err, tt.name)
				// 验证订单状态
				order, _ := services.OrderSvc.GetOrderByOrderNo(tt.orderNo)
				test.AssertEqual(t, 3, order.Status, "订单状态应为已取消")
			}
		})
	}
}

// TestOrderService_GetUserOrders 测试获取用户订单列表
func TestOrderService_GetUserOrders(t *testing.T) {
	services, cleanup := test.SetupTestServices(t)
	defer cleanup()

	// 创建测试数据
	testUser := test.CreateTestUser(t, services, "listuser", "list@example.com", "password123")
	testProduct := test.CreateTestProduct(t, services, "测试商品", 99.99)

	// 创建多个订单
	for i := 0; i < 5; i++ {
		test.CreateTestOrder(t, services, testUser.ID, testProduct.ID)
	}

	// 测试获取订单列表
	orders, total, err := services.OrderSvc.GetUserOrders(testUser.ID, 1, 10)
	test.AssertNoError(t, err, "获取订单列表")
	test.AssertEqual(t, int64(5), total, "订单总数")
	test.AssertEqual(t, 5, len(orders), "返回订单数")

	// 测试分页
	orders, _, err = services.OrderSvc.GetUserOrders(testUser.ID, 1, 2)
	test.AssertNoError(t, err, "获取分页订单")
	test.AssertEqual(t, 2, len(orders), "分页订单数")
}

// TestOrderService_CreateOrderWithRemark 测试创建带备注的订单
func TestOrderService_CreateOrderWithRemark(t *testing.T) {
	services, cleanup := test.SetupTestServices(t)
	defer cleanup()

	testUser := test.CreateTestUser(t, services, "remarkuser", "remark@example.com", "password123")
	testProduct := test.CreateTestProduct(t, services, "测试商品", 99.99)

	remark := "这是测试备注"
	order, err := services.OrderSvc.CreateOrderWithRemark(
		testUser.ID, "remarkuser", testProduct.ID, "127.0.0.1", remark,
	)
	test.AssertNoError(t, err, "创建带备注订单")
	test.AssertNotNil(t, order, "订单对象")
	test.AssertEqual(t, remark, order.Remark, "订单备注")
}

// TestOrderService_OrderStatus 测试订单状态流转
func TestOrderService_OrderStatus(t *testing.T) {
	services, cleanup := test.SetupTestServices(t)
	defer cleanup()

	testUser := test.CreateTestUser(t, services, "statususer", "status@example.com", "password123")

	// 创建一个手动卡密类型的商品用于测试
	product := &model.Product{
		Name:         "手动卡密商品",
		Price:        99.99,
		Duration:     30,
		DurationUnit: "天",
		Status:       1,
		Stock:        100,
		ProductType:  model.ProductTypeManual,
	}
	services.Repo.CreateProduct(product)

	// 创建订单
	order, err := services.OrderSvc.CreateOrder(testUser.ID, "statususer", product.ID, "127.0.0.1")
	test.AssertNoError(t, err, "创建订单")
	test.AssertEqual(t, 0, order.Status, "初始状态为待支付")

	// 取消订单
	err = services.OrderSvc.CancelOrder(order.OrderNo, testUser.ID)
	test.AssertNoError(t, err, "取消订单")

	updatedOrder, _ := services.OrderSvc.GetOrderByOrderNo(order.OrderNo)
	test.AssertEqual(t, 3, updatedOrder.Status, "取消后状态为已取消")

	// 尝试再次取消已取消的订单
	err = services.OrderSvc.CancelOrder(order.OrderNo, testUser.ID)
	test.AssertError(t, err, "不能取消已取消的订单")
}

// TestOrderService_GetOrderStats 测试订单统计
func TestOrderService_GetOrderStats(t *testing.T) {
	services, cleanup := test.SetupTestServices(t)
	defer cleanup()

	testUser := test.CreateTestUser(t, services, "statsuser", "stats@example.com", "password123")
	testProduct := test.CreateTestProduct(t, services, "统计商品", 100.00)

	// 创建多个订单
	for i := 0; i < 3; i++ {
		test.CreateTestOrder(t, services, testUser.ID, testProduct.ID)
	}

	// 获取统计
	stats, err := services.OrderSvc.GetOrderStats()
	test.AssertNoError(t, err, "获取订单统计")
	test.AssertNotNil(t, stats, "统计数据")
}
