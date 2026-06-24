package service_test

import (
	"testing"

	pluginapi "user-frontend/internal/plugin"
	"user-frontend/internal/service"
	"user-frontend/internal/test"
)

func TestMockPaymentPluginPoC(t *testing.T) {
	services, cleanup := test.SetupTestServices(t)
	defer cleanup()

	order, err := services.OrderKernelSvc.CreateHostOrder(service.HostOrderMaterial{
		UserID:       1,
		Username:     "buyer",
		ProductID:    20,
		ProductName:  "支付 PoC 商品",
		Quantity:     1,
		UnitPrice:    5,
		Duration:     30,
		DurationUnit: "天",
	})
	if err != nil {
		t.Fatalf("创建订单失败: %v", err)
	}
	attempt, err := services.OrderKernelSvc.CreatePaymentAttempt(order.OrderNo, "poc.mock-payment", "mock", "poc-pay")
	if err != nil {
		t.Fatalf("创建支付尝试失败: %v", err)
	}

	rejected, err := services.OrderKernelSvc.SubmitPaymentFact(pluginapi.PaymentFact{
		AttemptNo:             attempt.AttemptNo,
		OrderNo:               order.OrderNo,
		PaymentPluginID:       "poc.mock-payment",
		PaymentChannel:        "mock",
		ProviderTransactionID: "tx-mismatch",
		ProviderStatus:        "paid",
		AmountCents:           order.AmountPayableCents + 100,
		Currency:              "CNY",
		Verified:              true,
		IdempotencyKey:        "poc-mismatch",
	})
	if err != nil {
		t.Fatalf("金额不一致不应造成系统错误: %v", err)
	}
	if rejected.Status != service.PaymentCallbackStatusRejected {
		t.Fatalf("金额不一致应被拒绝: %s", rejected.Status)
	}

	accepted, err := services.OrderKernelSvc.SubmitPaymentFact(pluginapi.PaymentFact{
		AttemptNo:             attempt.AttemptNo,
		OrderNo:               order.OrderNo,
		PaymentPluginID:       "poc.mock-payment",
		PaymentChannel:        "mock",
		ProviderTransactionID: "tx-success",
		ProviderStatus:        "paid",
		AmountCents:           order.AmountPayableCents,
		Currency:              "CNY",
		Verified:              true,
		IdempotencyKey:        "poc-success",
	})
	if err != nil {
		t.Fatalf("模拟支付成功事实提交失败: %v", err)
	}
	again, err := services.OrderKernelSvc.SubmitPaymentFact(pluginapi.PaymentFact{
		OrderNo:        order.OrderNo,
		AmountCents:    order.AmountPayableCents,
		Verified:       true,
		IdempotencyKey: "poc-success",
	})
	if err != nil {
		t.Fatalf("模拟支付重复事实应幂等: %v", err)
	}
	if again.ID != accepted.ID {
		t.Fatal("模拟支付重复事实未命中幂等记录")
	}
}

func TestFulfillmentPluginPoC(t *testing.T) {
	services, cleanup := test.SetupTestServices(t)
	defer cleanup()

	order, err := services.OrderKernelSvc.CreateHostOrder(service.HostOrderMaterial{
		UserID:       1,
		Username:     "buyer",
		ProductID:    21,
		ProductName:  "交付 PoC 商品",
		Quantity:     1,
		UnitPrice:    8,
		Duration:     30,
		DurationUnit: "天",
	})
	if err != nil {
		t.Fatalf("创建订单失败: %v", err)
	}
	if _, err := services.OrderKernelSvc.CreateDeliveryTask(order.OrderNo, "poc.fulfillment", "kami", "before-pay"); err == nil {
		t.Fatal("订单未支付时不应允许创建交付任务")
	}
	attempt, err := services.OrderKernelSvc.CreatePaymentAttempt(order.OrderNo, "poc.mock-payment", "mock", "pay-before-delivery")
	if err != nil {
		t.Fatalf("创建支付尝试失败: %v", err)
	}
	if _, err := services.OrderKernelSvc.SubmitPaymentFact(pluginapi.PaymentFact{
		AttemptNo:             attempt.AttemptNo,
		OrderNo:               order.OrderNo,
		PaymentPluginID:       "poc.mock-payment",
		PaymentChannel:        "mock",
		ProviderTransactionID: "tx-delivery",
		ProviderStatus:        "paid",
		AmountCents:           order.AmountPayableCents,
		Currency:              "CNY",
		Verified:              true,
		IdempotencyKey:        "delivery-payment",
	}); err != nil {
		t.Fatalf("支付事实提交失败: %v", err)
	}
	delivery, err := services.OrderKernelSvc.CreateDeliveryTask(order.OrderNo, "poc.fulfillment", "kami", "after-pay")
	if err != nil {
		t.Fatalf("支付后创建交付任务失败: %v", err)
	}
	failed, err := services.OrderKernelSvc.SubmitDeliveryFact(pluginapi.FulfillmentFact{
		DeliveryNo:          delivery.DeliveryNo,
		OrderNo:             order.OrderNo,
		FulfillmentPluginID: "poc.fulfillment",
		FactType:            "kami.issue",
		Status:              service.DeliveryStatusFailed,
		ErrorCode:           "stock_empty",
		ErrorMessage:        "卡密库存不足",
		Retryable:           true,
		IdempotencyKey:      "delivery-stock-empty",
	})
	if err != nil {
		t.Fatalf("库存不足交付事实提交失败: %v", err)
	}
	if !failed.Retryable || failed.ErrorCode != "stock_empty" {
		t.Fatalf("库存不足事实不符合预期: %#v", failed)
	}

	delivery2, err := services.OrderKernelSvc.CreateDeliveryTask(order.OrderNo, "poc.fulfillment", "kami", "retry-after-stock")
	if err != nil {
		t.Fatalf("重试交付任务创建失败: %v", err)
	}
	if _, err := services.OrderKernelSvc.SubmitDeliveryFact(pluginapi.FulfillmentFact{
		DeliveryNo:          delivery2.DeliveryNo,
		OrderNo:             order.OrderNo,
		FulfillmentPluginID: "poc.fulfillment",
		FactType:            "kami.issue",
		Status:              service.DeliveryStatusSuccess,
		ResultPayload:       map[string]any{"kami_count": 1},
		IdempotencyKey:      "delivery-success",
	}); err != nil {
		t.Fatalf("交付成功事实提交失败: %v", err)
	}
}

func TestProductPluginMaterialPoC(t *testing.T) {
	services, cleanup := test.SetupTestServices(t)
	defer cleanup()

	order, err := services.OrderKernelSvc.CreatePluginMaterialOrder(1, "buyer", "127.0.0.1", pluginapi.ProductOrderMaterial{
		ProductRef:          "plugin-product-basic",
		SKURef:              "sku-basic",
		ItemType:            "plugin_product",
		Quantity:            2,
		UnitPriceCents:      399,
		Currency:            "CNY",
		OwnerPluginID:       "poc.product",
		IdempotencyKey:      "plugin-material-order",
		ItemSnapshot:        map[string]any{"title": "插件商品"},
		DeliveryRequirement: map[string]any{"type": "external"},
	})
	if err != nil {
		t.Fatalf("插件订单素材创建正式订单失败: %v", err)
	}
	if order.OrderNo == "" || order.AmountPayableCents != 798 {
		t.Fatalf("插件素材订单金额或订单号不正确: %#v", order)
	}
	snapshot, err := services.OrderKernelSvc.SnapshotByOrderNo(order.OrderNo)
	if err != nil {
		t.Fatalf("读取插件素材订单快照失败: %v", err)
	}
	if len(snapshot.Items) != 1 || snapshot.Items[0].OwnerPluginID != "poc.product" {
		t.Fatalf("插件素材订单明细不正确: %#v", snapshot.Items)
	}
}
