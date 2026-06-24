package service_test

import (
	"testing"

	"user-frontend/internal/model"
	pluginapi "user-frontend/internal/plugin"
	"user-frontend/internal/service"
	"user-frontend/internal/test"
)

func TestOrderKernelServiceCreatesOrderFacts(t *testing.T) {
	services, cleanup := test.SetupTestServices(t)
	defer cleanup()

	order, err := services.OrderKernelSvc.CreateHostOrder(service.HostOrderMaterial{
		UserID:       1,
		Username:     "buyer",
		ProductID:    10,
		ProductName:  "测试卡密",
		Quantity:     2,
		UnitPrice:    9.9,
		Duration:     30,
		DurationUnit: "天",
		ItemSnapshot: map[string]any{"name": "测试卡密"},
	})
	if err != nil {
		t.Fatalf("创建订单失败: %v", err)
	}
	if order.OrderStatus != model.OrderKernelStatusPendingPayment {
		t.Fatalf("订单主状态不正确: %s", order.OrderStatus)
	}

	snapshot, err := services.OrderKernelSvc.SnapshotByOrderNo(order.OrderNo)
	if err != nil {
		t.Fatalf("读取订单快照失败: %v", err)
	}
	if !snapshot.HasItems() || !snapshot.HasStateEvents() {
		t.Fatalf("订单明细或状态事件未写入: %#v", snapshot)
	}
	if snapshot.Items[0].Quantity != 2 {
		t.Fatalf("订单明细数量不正确: %d", snapshot.Items[0].Quantity)
	}
}

func TestOrderKernelServicePaymentAndDeliveryFacts(t *testing.T) {
	services, cleanup := test.SetupTestServices(t)
	defer cleanup()

	order, err := services.OrderKernelSvc.CreateHostOrder(service.HostOrderMaterial{
		UserID:       1,
		Username:     "buyer",
		ProductID:    10,
		ProductName:  "测试卡密",
		Quantity:     1,
		UnitPrice:    12.34,
		Duration:     30,
		DurationUnit: "天",
	})
	if err != nil {
		t.Fatalf("创建订单失败: %v", err)
	}

	attempt, err := services.OrderKernelSvc.CreatePaymentAttempt(order.OrderNo, "mock-pay", "mock", "pay-key")
	if err != nil {
		t.Fatalf("创建支付尝试失败: %v", err)
	}
	rejected, err := services.OrderKernelSvc.SubmitPaymentFact(pluginapi.PaymentFact{
		AttemptNo:             attempt.AttemptNo,
		OrderNo:               order.OrderNo,
		PaymentPluginID:       "mock-pay",
		PaymentChannel:        "mock",
		ProviderTransactionID: "tx-bad",
		ProviderStatus:        "paid",
		AmountCents:           order.AmountPayableCents + 1,
		Currency:              "CNY",
		Verified:              true,
		IdempotencyKey:        "bad-callback",
	})
	if err != nil {
		t.Fatalf("金额不一致事实应被记录为拒绝而不是系统错误: %v", err)
	}
	if rejected.Status != service.PaymentCallbackStatusRejected {
		t.Fatalf("金额不一致回调状态不正确: %s", rejected.Status)
	}

	accepted, err := services.OrderKernelSvc.SubmitPaymentFact(pluginapi.PaymentFact{
		AttemptNo:             attempt.AttemptNo,
		OrderNo:               order.OrderNo,
		PaymentPluginID:       "mock-pay",
		PaymentChannel:        "mock",
		ProviderTransactionID: "tx-ok",
		ProviderStatus:        "paid",
		AmountCents:           order.AmountPayableCents,
		Currency:              "CNY",
		Verified:              true,
		IdempotencyKey:        "ok-callback",
	})
	if err != nil {
		t.Fatalf("提交支付事实失败: %v", err)
	}
	if accepted.Status != service.PaymentCallbackStatusAccepted {
		t.Fatalf("支付回调状态不正确: %s", accepted.Status)
	}
	acceptedAgain, err := services.OrderKernelSvc.SubmitPaymentFact(pluginapi.PaymentFact{
		OrderNo:        order.OrderNo,
		AmountCents:    order.AmountPayableCents,
		Verified:       true,
		IdempotencyKey: "ok-callback",
	})
	if err != nil {
		t.Fatalf("重复支付事实应幂等返回: %v", err)
	}
	if acceptedAgain.ID != accepted.ID {
		t.Fatalf("重复支付事实没有命中幂等记录")
	}

	delivery, err := services.OrderKernelSvc.CreateDeliveryTask(order.OrderNo, "mock-fulfill", "kami", "delivery-key")
	if err != nil {
		t.Fatalf("创建交付任务失败: %v", err)
	}
	if _, err := services.OrderKernelSvc.SubmitDeliveryFact(pluginapi.FulfillmentFact{
		DeliveryNo:          delivery.DeliveryNo,
		OrderNo:             order.OrderNo,
		FulfillmentPluginID: "mock-fulfill",
		FactType:            "kami.issued",
		Status:              service.DeliveryStatusSuccess,
		ResultPayload:       map[string]any{"count": 1},
		IdempotencyKey:      "delivery-fact",
	}); err != nil {
		t.Fatalf("提交交付事实失败: %v", err)
	}

	snapshot, err := services.OrderKernelSvc.SnapshotByOrderNo(order.OrderNo)
	if err != nil {
		t.Fatalf("读取订单追踪失败: %v", err)
	}
	if snapshot.Order.OrderStatus != model.OrderKernelStatusCompleted {
		t.Fatalf("交付成功后订单未完成: %s", snapshot.Order.OrderStatus)
	}
	if !snapshot.HasPaymentCallbacks() || !snapshot.HasDeliveryFacts() || !snapshot.HasStateEvents() {
		t.Fatalf("订单追踪事实不完整: %#v", snapshot)
	}
}
