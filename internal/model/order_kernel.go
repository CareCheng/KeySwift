package model

import "time"

const (
	OrderKernelStatusCreated        = "created"
	OrderKernelStatusPendingPayment = "pending_payment"
	OrderKernelStatusCompleted      = "completed"
	OrderKernelStatusCancelled      = "cancelled"
	OrderKernelStatusFailed         = "failed"

	PaymentStatusUnpaid    = "unpaid"
	PaymentStatusPaying    = "paying"
	PaymentStatusPaid      = "paid"
	PaymentStatusPayFailed = "pay_failed"

	DeliveryStatusUnfulfilled    = "unfulfilled"
	DeliveryStatusFulfilling     = "fulfilling"
	DeliveryStatusFulfilled      = "fulfilled"
	DeliveryStatusDeliveryFailed = "delivery_failed"
)

// OrderItem 记录订单明细和商品快照。
type OrderItem struct {
	ID                      uint      `gorm:"primaryKey" json:"id"`
	OrderID                 uint      `gorm:"index" json:"order_id"`
	ItemType                string    `gorm:"size:80;default:host_product" json:"item_type"`
	ProductRef              string    `gorm:"size:120;index" json:"product_ref"`
	SKURef                  string    `gorm:"size:120" json:"sku_ref"`
	Quantity                int       `gorm:"default:1" json:"quantity"`
	UnitPriceCents          int64     `json:"unit_price_cents"`
	ItemSnapshotJSON        string    `gorm:"type:text" json:"item_snapshot_json"`
	DeliveryRequirementJSON string    `gorm:"type:text" json:"delivery_requirement_json"`
	OwnerPluginID           string    `gorm:"size:120;index" json:"owner_plugin_id"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}

func (OrderItem) TableName() string { return "order_items" }

// OrderPaymentAttempt 记录一次支付尝试。
type OrderPaymentAttempt struct {
	ID                    uint      `gorm:"primaryKey" json:"id"`
	AttemptNo             string    `gorm:"size:120;uniqueIndex" json:"attempt_no"`
	OrderID               uint      `gorm:"index;uniqueIndex:idx_payment_attempt_idempotency" json:"order_id"`
	OrderNo               string    `gorm:"size:120;index" json:"order_no"`
	PaymentPluginID       string    `gorm:"size:120;index" json:"payment_plugin_id"`
	PaymentChannel        string    `gorm:"size:80" json:"payment_channel"`
	AmountCents           int64     `json:"amount_cents"`
	Currency              string    `gorm:"size:20;default:CNY" json:"currency"`
	Status                string    `gorm:"size:50;default:created;index" json:"status"`
	ProviderTransactionID string    `gorm:"size:180" json:"provider_transaction_id"`
	IdempotencyKey        string    `gorm:"size:180;uniqueIndex:idx_payment_attempt_idempotency" json:"idempotency_key"`
	CallbackToken         string    `gorm:"size:180" json:"callback_token"`
	ErrorCode             string    `gorm:"size:120" json:"error_code"`
	ErrorMessage          string    `gorm:"type:text" json:"error_message"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

func (OrderPaymentAttempt) TableName() string { return "order_payment_attempts" }

// OrderPaymentCallback 记录支付渠道回调原始事实。
type OrderPaymentCallback struct {
	ID                    uint      `gorm:"primaryKey" json:"id"`
	CallbackID            string    `gorm:"size:120;uniqueIndex" json:"callback_id"`
	AttemptNo             string    `gorm:"size:120;index" json:"attempt_no"`
	OrderNo               string    `gorm:"size:120;index" json:"order_no"`
	ProviderTransactionID string    `gorm:"size:180" json:"provider_transaction_id"`
	ProviderStatus        string    `gorm:"size:80" json:"provider_status"`
	AmountCents           int64     `json:"amount_cents"`
	Currency              string    `gorm:"size:20;default:CNY" json:"currency"`
	Verified              bool      `gorm:"default:false" json:"verified"`
	IdempotencyKey        string    `gorm:"size:180" json:"idempotency_key"`
	CallbackPayloadJSON   string    `gorm:"type:text" json:"callback_payload_json"`
	Status                string    `gorm:"size:50;default:received;index" json:"status"`
	ErrorMessage          string    `gorm:"type:text" json:"error_message"`
	ReceivedAt            time.Time `json:"received_at"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

func (OrderPaymentCallback) TableName() string { return "order_payment_callbacks" }

// OrderDelivery 记录宿主创建的交付任务。
type OrderDelivery struct {
	ID                  uint       `gorm:"primaryKey" json:"id"`
	DeliveryNo          string     `gorm:"size:120;uniqueIndex" json:"delivery_no"`
	OrderID             uint       `gorm:"index;uniqueIndex:idx_order_delivery_idempotency" json:"order_id"`
	OrderNo             string     `gorm:"size:120;index" json:"order_no"`
	FulfillmentPluginID string     `gorm:"size:120;index" json:"fulfillment_plugin_id"`
	DeliveryType        string     `gorm:"size:80" json:"delivery_type"`
	Status              string     `gorm:"size:50;default:created;index" json:"status"`
	ResultPayloadJSON   string     `gorm:"type:text" json:"result_payload_json"`
	IdempotencyKey      string     `gorm:"size:180;uniqueIndex:idx_order_delivery_idempotency" json:"idempotency_key"`
	DeliveredAt         *time.Time `json:"delivered_at"`
	ErrorCode           string     `gorm:"size:120" json:"error_code"`
	ErrorMessage        string     `gorm:"type:text" json:"error_message"`
	Retryable           bool       `gorm:"default:false" json:"retryable"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

func (OrderDelivery) TableName() string { return "order_deliveries" }

// OrderDeliveryFact 记录交付插件提交的交付事实。
type OrderDeliveryFact struct {
	ID                  uint      `gorm:"primaryKey" json:"id"`
	FactID              string    `gorm:"size:120;uniqueIndex" json:"fact_id"`
	DeliveryNo          string    `gorm:"size:120;index" json:"delivery_no"`
	OrderNo             string    `gorm:"size:120;index" json:"order_no"`
	FulfillmentPluginID string    `gorm:"size:120;index" json:"fulfillment_plugin_id"`
	FactType            string    `gorm:"size:80" json:"fact_type"`
	Status              string    `gorm:"size:50" json:"status"`
	ResultPayloadJSON   string    `gorm:"type:text" json:"result_payload_json"`
	IdempotencyKey      string    `gorm:"size:180" json:"idempotency_key"`
	Retryable           bool      `gorm:"default:false" json:"retryable"`
	ErrorCode           string    `gorm:"size:120" json:"error_code"`
	ErrorMessage        string    `gorm:"type:text" json:"error_message"`
	OccurredAt          time.Time `json:"occurred_at"`
	CreatedAt           time.Time `json:"created_at"`
}

func (OrderDeliveryFact) TableName() string { return "order_delivery_facts" }

// OrderStateEvent 记录订单状态轨迹。
type OrderStateEvent struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	OrderID        uint      `gorm:"index" json:"order_id"`
	OrderNo        string    `gorm:"size:120;index" json:"order_no"`
	EventType      string    `gorm:"size:120;index" json:"event_type"`
	FromStatus     string    `gorm:"size:80" json:"from_status"`
	ToStatus       string    `gorm:"size:80" json:"to_status"`
	PaymentStatus  string    `gorm:"size:80" json:"payment_status"`
	DeliveryStatus string    `gorm:"size:80" json:"delivery_status"`
	ActorSubjectID string    `gorm:"size:120" json:"actor_subject_id"`
	OwnerPluginID  string    `gorm:"size:120" json:"owner_plugin_id"`
	PayloadJSON    string    `gorm:"type:text" json:"payload_json"`
	CreatedAt      time.Time `json:"created_at"`
}

func (OrderStateEvent) TableName() string { return "order_state_events" }
