package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"user-frontend/internal/model"
	pluginapi "user-frontend/internal/plugin"
	"user-frontend/internal/repository"
	"user-frontend/internal/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	PaymentAttemptStatusCreated = "created"
	PaymentAttemptStatusPaid    = "paid"
	PaymentAttemptStatusFailed  = "failed"

	PaymentCallbackStatusAccepted = "accepted"
	PaymentCallbackStatusRejected = "rejected"

	DeliveryStatusCreated = "created"
	DeliveryStatusSuccess = "success"
	DeliveryStatusFailed  = "failed"
)

// HostOrderMaterial 是宿主核心商品创建订单时使用的订单素材。
type HostOrderMaterial struct {
	UserID              uint
	Username            string
	ProductID           uint
	ProductName         string
	Quantity            int
	UnitPrice           float64
	Duration            int
	DurationUnit        string
	ClientIP            string
	Remark              string
	ItemSnapshot        any
	DeliveryRequirement any
	OwnerPluginID       string
	IdempotencyKey      string
}

// OrderSnapshot 是后台订单追踪视图。
type OrderSnapshot struct {
	Order            model.Order                  `json:"order"`
	Items            []model.OrderItem            `json:"items"`
	PaymentAttempts  []model.OrderPaymentAttempt  `json:"payment_attempts"`
	PaymentCallbacks []model.OrderPaymentCallback `json:"payment_callbacks"`
	Deliveries       []model.OrderDelivery        `json:"deliveries"`
	DeliveryFacts    []model.OrderDeliveryFact    `json:"delivery_facts"`
	StateEvents      []model.OrderStateEvent      `json:"state_events"`
}

// OrderKernelService 是宿主订单状态与事实裁决内核。
type OrderKernelService struct {
	repo   *repository.Repository
	govSvc *GovernanceService
}

func NewOrderKernelService(repo *repository.Repository, govSvc *GovernanceService) *OrderKernelService {
	return &OrderKernelService{repo: repo, govSvc: govSvc}
}

func (s *OrderKernelService) CreateHostOrder(material HostOrderMaterial) (*model.Order, error) {
	if material.UserID == 0 || material.ProductID == 0 {
		return nil, errors.New("用户和商品不能为空")
	}
	quantity := material.Quantity
	if quantity < 1 {
		quantity = 1
	}
	total := material.UnitPrice * float64(quantity)
	amountCents := moneyToCents(total)
	order := &model.Order{
		OrderNo:            utils.GenerateLocalOrderNo(),
		BuyerSubjectID:     fmt.Sprintf("user:%d", material.UserID),
		UserID:             material.UserID,
		Username:           material.Username,
		ProductID:          material.ProductID,
		ProductName:        material.ProductName,
		Quantity:           quantity,
		OriginalPrice:      total,
		Price:              total,
		AmountTotalCents:   amountCents,
		AmountPayableCents: amountCents,
		AmountPaidCents:    0,
		Currency:           "CNY",
		Duration:           material.Duration,
		DurationUnit:       material.DurationUnit,
		Status:             model.OrderStatusPending,
		OrderStatus:        model.OrderKernelStatusPendingPayment,
		PaymentStatus:      model.PaymentStatusUnpaid,
		DeliveryStatus:     model.DeliveryStatusUnfulfilled,
		IdempotencyKey:     material.IdempotencyKey,
		Version:            1,
		Remark:             material.Remark,
		ClientIP:           material.ClientIP,
	}

	err := s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(order).Error; err != nil {
			return err
		}
		item := model.OrderItem{
			OrderID:                 order.ID,
			ItemType:                defaultText("host_product", "host_product"),
			ProductRef:              fmt.Sprintf("%d", material.ProductID),
			Quantity:                quantity,
			UnitPriceCents:          moneyToCents(material.UnitPrice),
			ItemSnapshotJSON:        jsonForDB(material.ItemSnapshot),
			DeliveryRequirementJSON: jsonForDB(material.DeliveryRequirement),
			OwnerPluginID:           material.OwnerPluginID,
		}
		if err := tx.Create(&item).Error; err != nil {
			return err
		}
		return s.createOrderEvent(tx, order, "order.created", "", order.OrderStatus, "", jsonForDB(material.ItemSnapshot))
	})
	if err != nil {
		return nil, err
	}
	_ = s.recordEvent("order.created", "order", order.OrderNo, material.OwnerPluginID, order)
	return order, nil
}

// CreatePluginMaterialOrder 接收商品插件订单素材，由宿主生成正式订单。
func (s *OrderKernelService) CreatePluginMaterialOrder(userID uint, username, clientIP string, material pluginapi.ProductOrderMaterial) (*model.Order, error) {
	if strings.TrimSpace(material.ProductRef) == "" {
		return nil, errors.New("商品引用不能为空")
	}
	quantity := material.Quantity
	if quantity < 1 {
		quantity = 1
	}
	unitPrice := float64(material.UnitPriceCents) / 100
	order := &model.Order{
		OrderNo:            utils.GenerateLocalOrderNo(),
		BuyerSubjectID:     fmt.Sprintf("user:%d", userID),
		UserID:             userID,
		Username:           username,
		ProductName:        material.ProductRef,
		Quantity:           quantity,
		OriginalPrice:      unitPrice * float64(quantity),
		Price:              unitPrice * float64(quantity),
		AmountTotalCents:   material.UnitPriceCents * int64(quantity),
		AmountPayableCents: material.UnitPriceCents * int64(quantity),
		Currency:           defaultText(material.Currency, "CNY"),
		Status:             model.OrderStatusPending,
		OrderStatus:        model.OrderKernelStatusPendingPayment,
		PaymentStatus:      model.PaymentStatusUnpaid,
		DeliveryStatus:     model.DeliveryStatusUnfulfilled,
		IdempotencyKey:     material.IdempotencyKey,
		Version:            1,
		ClientIP:           clientIP,
	}
	err := s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(order).Error; err != nil {
			return err
		}
		item := model.OrderItem{
			OrderID:                 order.ID,
			ItemType:                defaultText(material.ItemType, "plugin_product"),
			ProductRef:              material.ProductRef,
			SKURef:                  material.SKURef,
			Quantity:                quantity,
			UnitPriceCents:          material.UnitPriceCents,
			ItemSnapshotJSON:        jsonForDB(material.ItemSnapshot),
			DeliveryRequirementJSON: jsonForDB(material.DeliveryRequirement),
			OwnerPluginID:           material.OwnerPluginID,
		}
		if err := tx.Create(&item).Error; err != nil {
			return err
		}
		return s.createOrderEvent(tx, order, "order.created.from_plugin_material", "", order.OrderStatus, material.OwnerPluginID, jsonForDB(material))
	})
	if err != nil {
		return nil, err
	}
	_ = s.recordEvent("order.created.from_plugin_material", "order", order.OrderNo, material.OwnerPluginID, material)
	return order, nil
}

func (s *OrderKernelService) CreatePaymentAttempt(orderNo, paymentPluginID, channel, idempotencyKey string) (*model.OrderPaymentAttempt, error) {
	order, err := s.repo.GetOrderByOrderNo(orderNo)
	if err != nil {
		return nil, errors.New("订单不存在")
	}
	if strings.TrimSpace(idempotencyKey) == "" {
		idempotencyKey = "pay_" + uuid.NewString()
	}
	attempt := &model.OrderPaymentAttempt{
		AttemptNo:       "pa_" + uuid.NewString(),
		OrderID:         order.ID,
		OrderNo:         order.OrderNo,
		PaymentPluginID: paymentPluginID,
		PaymentChannel:  channel,
		AmountCents:     order.AmountPayableCents,
		Currency:        defaultText(order.Currency, "CNY"),
		Status:          PaymentAttemptStatusCreated,
		IdempotencyKey:  idempotencyKey,
		CallbackToken:   "cb_" + uuid.NewString(),
	}
	return attempt, s.repo.GetDB().Create(attempt).Error
}

func (s *OrderKernelService) SubmitPaymentFact(fact pluginapi.PaymentFact) (*model.OrderPaymentCallback, error) {
	order, err := s.repo.GetOrderByOrderNo(fact.OrderNo)
	if err != nil {
		return nil, errors.New("订单不存在")
	}
	callback := &model.OrderPaymentCallback{
		CallbackID:            "pcb_" + uuid.NewString(),
		AttemptNo:             fact.AttemptNo,
		OrderNo:               fact.OrderNo,
		ProviderTransactionID: fact.ProviderTransactionID,
		ProviderStatus:        fact.ProviderStatus,
		AmountCents:           fact.AmountCents,
		Currency:              defaultText(fact.Currency, "CNY"),
		Verified:              fact.Verified,
		IdempotencyKey:        fact.IdempotencyKey,
		CallbackPayloadJSON:   jsonForDB(fact.CallbackPayload),
		ReceivedAt:            time.Now(),
	}
	err = s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
		var existing model.OrderPaymentCallback
		if fact.IdempotencyKey != "" {
			err := tx.Where("idempotency_key = ? AND order_no = ?", fact.IdempotencyKey, fact.OrderNo).First(&existing).Error
			if err == nil {
				*callback = existing
				return nil
			}
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
		}

		if !fact.Verified || fact.AmountCents != order.AmountPayableCents {
			callback.Status = PaymentCallbackStatusRejected
			callback.ErrorMessage = "支付事实未验证或金额不一致"
			return tx.Create(callback).Error
		}

		callback.Status = PaymentCallbackStatusAccepted
		if err := tx.Create(callback).Error; err != nil {
			return err
		}
		now := time.Now()
		fromStatus := order.OrderStatus
		order.PaymentStatus = model.PaymentStatusPaid
		order.AmountPaidCents = fact.AmountCents
		order.PaymentNo = fact.ProviderTransactionID
		order.PaymentMethod = fact.PaymentChannel
		order.PaymentTime = &now
		order.Status = model.OrderStatusPaid
		order.Version++
		if err := tx.Save(order).Error; err != nil {
			return err
		}
		if fact.AttemptNo != "" {
			if err := tx.Model(&model.OrderPaymentAttempt{}).
				Where("attempt_no = ?", fact.AttemptNo).
				Updates(map[string]any{
					"status":                  PaymentAttemptStatusPaid,
					"provider_transaction_id": fact.ProviderTransactionID,
				}).Error; err != nil {
				return err
			}
		}
		return s.createOrderEvent(tx, order, "payment.accepted", fromStatus, order.OrderStatus, fact.PaymentPluginID, callback.CallbackPayloadJSON)
	})
	if err != nil {
		return nil, err
	}
	_ = s.recordEvent("payment.accepted", "order", order.OrderNo, fact.PaymentPluginID, fact)
	return callback, nil
}

func (s *OrderKernelService) CreateDeliveryTask(orderNo, fulfillmentPluginID, deliveryType, idempotencyKey string) (*model.OrderDelivery, error) {
	order, err := s.repo.GetOrderByOrderNo(orderNo)
	if err != nil {
		return nil, errors.New("订单不存在")
	}
	if order.PaymentStatus != model.PaymentStatusPaid {
		return nil, errors.New("订单未完成支付，不能创建交付任务")
	}
	if strings.TrimSpace(idempotencyKey) == "" {
		idempotencyKey = "del_" + uuid.NewString()
	}
	delivery := &model.OrderDelivery{
		DeliveryNo:          "od_" + uuid.NewString(),
		OrderID:             order.ID,
		OrderNo:             order.OrderNo,
		FulfillmentPluginID: fulfillmentPluginID,
		DeliveryType:        deliveryType,
		Status:              DeliveryStatusCreated,
		IdempotencyKey:      idempotencyKey,
	}
	return delivery, s.repo.GetDB().Create(delivery).Error
}

func (s *OrderKernelService) SubmitDeliveryFact(fact pluginapi.FulfillmentFact) (*model.OrderDeliveryFact, error) {
	order, err := s.repo.GetOrderByOrderNo(fact.OrderNo)
	if err != nil {
		return nil, errors.New("订单不存在")
	}
	record := &model.OrderDeliveryFact{
		FactID:              "df_" + uuid.NewString(),
		DeliveryNo:          fact.DeliveryNo,
		OrderNo:             fact.OrderNo,
		FulfillmentPluginID: fact.FulfillmentPluginID,
		FactType:            fact.FactType,
		Status:              fact.Status,
		ResultPayloadJSON:   jsonForDB(fact.ResultPayload),
		IdempotencyKey:      fact.IdempotencyKey,
		Retryable:           fact.Retryable,
		ErrorCode:           fact.ErrorCode,
		ErrorMessage:        fact.ErrorMessage,
		OccurredAt:          time.Now(),
	}
	err = s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
		var existing model.OrderDeliveryFact
		if fact.IdempotencyKey != "" {
			err := tx.Where("idempotency_key = ? AND order_no = ?", fact.IdempotencyKey, fact.OrderNo).First(&existing).Error
			if err == nil {
				*record = existing
				return nil
			}
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
		}
		if err := tx.Create(record).Error; err != nil {
			return err
		}
		fromStatus := order.OrderStatus
		if fact.Status == DeliveryStatusSuccess {
			now := time.Now()
			order.Status = model.OrderStatusCompleted
			order.OrderStatus = model.OrderKernelStatusCompleted
			order.DeliveryStatus = model.DeliveryStatusFulfilled
			order.Version++
			if err := tx.Model(&model.OrderDelivery{}).
				Where("delivery_no = ?", fact.DeliveryNo).
				Updates(map[string]any{
					"status":              DeliveryStatusSuccess,
					"result_payload_json": record.ResultPayloadJSON,
					"delivered_at":        &now,
				}).Error; err != nil {
				return err
			}
		} else {
			order.DeliveryStatus = model.DeliveryStatusDeliveryFailed
			if err := tx.Model(&model.OrderDelivery{}).
				Where("delivery_no = ?", fact.DeliveryNo).
				Updates(map[string]any{
					"status":        DeliveryStatusFailed,
					"error_code":    fact.ErrorCode,
					"error_message": fact.ErrorMessage,
					"retryable":     fact.Retryable,
				}).Error; err != nil {
				return err
			}
		}
		if err := tx.Save(order).Error; err != nil {
			return err
		}
		return s.createOrderEvent(tx, order, "delivery.fact.accepted", fromStatus, order.OrderStatus, fact.FulfillmentPluginID, record.ResultPayloadJSON)
	})
	if err != nil {
		return nil, err
	}
	_ = s.recordEvent("delivery.fact.accepted", "order", order.OrderNo, fact.FulfillmentPluginID, fact)
	return record, nil
}

func (s *OrderKernelService) SnapshotByOrderID(orderID uint) (OrderSnapshot, error) {
	var snapshot OrderSnapshot
	order, err := s.repo.GetOrderByID(orderID)
	if err != nil {
		return snapshot, err
	}
	return s.snapshot(*order)
}

func (s *OrderKernelService) SnapshotByOrderNo(orderNo string) (OrderSnapshot, error) {
	order, err := s.repo.GetOrderByOrderNo(orderNo)
	if err != nil {
		return OrderSnapshot{}, err
	}
	return s.snapshot(*order)
}

func (s *OrderKernelService) snapshot(order model.Order) (OrderSnapshot, error) {
	var snapshot OrderSnapshot
	snapshot.Order = order
	db := s.repo.GetDB()
	if err := db.Where("order_id = ?", order.ID).Order("id ASC").Find(&snapshot.Items).Error; err != nil {
		return snapshot, err
	}
	if err := db.Where("order_id = ?", order.ID).Order("id ASC").Find(&snapshot.PaymentAttempts).Error; err != nil {
		return snapshot, err
	}
	if err := db.Where("order_no = ?", order.OrderNo).Order("id ASC").Find(&snapshot.PaymentCallbacks).Error; err != nil {
		return snapshot, err
	}
	if err := db.Where("order_id = ?", order.ID).Order("id ASC").Find(&snapshot.Deliveries).Error; err != nil {
		return snapshot, err
	}
	if err := db.Where("order_no = ?", order.OrderNo).Order("id ASC").Find(&snapshot.DeliveryFacts).Error; err != nil {
		return snapshot, err
	}
	if err := db.Where("order_id = ?", order.ID).Order("id ASC").Find(&snapshot.StateEvents).Error; err != nil {
		return snapshot, err
	}
	return snapshot, nil
}

func (s *OrderKernelService) createOrderEvent(tx *gorm.DB, order *model.Order, eventType, fromStatus, toStatus, ownerPluginID, payloadJSON string) error {
	event := model.OrderStateEvent{
		OrderID:        order.ID,
		OrderNo:        order.OrderNo,
		EventType:      eventType,
		FromStatus:     fromStatus,
		ToStatus:       toStatus,
		PaymentStatus:  order.PaymentStatus,
		DeliveryStatus: order.DeliveryStatus,
		OwnerPluginID:  ownerPluginID,
		PayloadJSON:    payloadJSON,
	}
	return tx.Create(&event).Error
}

func (s *OrderKernelService) recordEvent(eventType, sourceType, sourceID, ownerPluginID string, payload any) error {
	if s.govSvc == nil {
		return nil
	}
	_, err := s.govSvc.RecordEvent(EventInput{
		EventType:     eventType,
		SourceType:    sourceType,
		SourceID:      sourceID,
		OwnerPluginID: ownerPluginID,
		Payload:       payload,
	})
	return err
}

func moneyToCents(value float64) int64 {
	return int64(math.Round(value * 100))
}

func jsonForDB(value any) string {
	if value == nil {
		return ""
	}
	data, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(data)
}

func (o *OrderSnapshot) HasDeliveryFacts() bool {
	return len(o.DeliveryFacts) > 0
}

func (o *OrderSnapshot) HasPaymentCallbacks() bool {
	return len(o.PaymentCallbacks) > 0
}

func (o *OrderSnapshot) HasStateEvents() bool {
	return len(o.StateEvents) > 0
}

func (o *OrderSnapshot) HasItems() bool {
	return len(o.Items) > 0
}
