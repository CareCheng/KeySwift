// Package service 提供主程序核心业务服务。
//
// 服务层承接 API 层和 Repository 层之间的业务编排，当前只保留用户、会话、商品、分类、订单、余额、卡密、配置、邮箱、安全、日志和后台角色权限等核心能力。
package service

// OrderStatus 订单状态常量。
const (
	OrderStatusPending   = 0 // 待支付
	OrderStatusPaid      = 1 // 已支付，等待发卡
	OrderStatusCompleted = 2 // 已完成
	OrderStatusCancelled = 3 // 已取消
	OrderStatusRefunded  = 4 // 已退款
)

// UserStatus 用户状态常量。
const (
	UserStatusActive   = 1 // 正常
	UserStatusInactive = 0 // 未激活
	UserStatusBanned   = 2 // 已禁用
)

// PaymentMethod 支付方式常量。
const (
	PaymentMethodBalance = "balance" // 余额支付
)

// BalanceType 余额变动类型常量。
const (
	BalanceTypePayment = "payment" // 支付
	BalanceTypeRefund  = "refund"  // 退款
	BalanceTypeAdmin   = "admin"   // 管理员调整
)
