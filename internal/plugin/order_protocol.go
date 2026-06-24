package plugin

// ProductOrderMaterial 是商品插件可提交给宿主的订单素材。
// 插件只能提供商品、数量、金额和交付要求，不能生成正式订单号或裁决订单状态。
type ProductOrderMaterial struct {
	ProductRef          string         `json:"productRef"`
	SKURef              string         `json:"skuRef"`
	ItemType            string         `json:"itemType"`
	Quantity            int            `json:"quantity"`
	UnitPriceCents      int64          `json:"unitPriceCents"`
	Currency            string         `json:"currency"`
	ItemSnapshot        map[string]any `json:"itemSnapshot"`
	DeliveryRequirement map[string]any `json:"deliveryRequirement"`
	OwnerPluginID       string         `json:"ownerPluginId"`
	IdempotencyKey      string         `json:"idempotencyKey"`
}

// PaymentFact 是支付插件提交给宿主裁决的支付事实。
type PaymentFact struct {
	AttemptNo             string         `json:"attemptNo"`
	OrderNo               string         `json:"orderNo"`
	PaymentPluginID       string         `json:"paymentPluginId"`
	PaymentChannel        string         `json:"paymentChannel"`
	ProviderTransactionID string         `json:"providerTransactionId"`
	ProviderStatus        string         `json:"providerStatus"`
	AmountCents           int64          `json:"amountCents"`
	Currency              string         `json:"currency"`
	Verified              bool           `json:"verified"`
	IdempotencyKey        string         `json:"idempotencyKey"`
	CallbackPayload       map[string]any `json:"callbackPayload"`
}

// FulfillmentFact 是交付插件提交给宿主裁决的交付事实。
type FulfillmentFact struct {
	DeliveryNo          string         `json:"deliveryNo"`
	OrderNo             string         `json:"orderNo"`
	FulfillmentPluginID string         `json:"fulfillmentPluginId"`
	FactType            string         `json:"factType"`
	Status              string         `json:"status"`
	ResultPayload       map[string]any `json:"resultPayload"`
	IdempotencyKey      string         `json:"idempotencyKey"`
	Retryable           bool           `json:"retryable"`
	ErrorCode           string         `json:"errorCode"`
	ErrorMessage        string         `json:"errorMessage"`
}
