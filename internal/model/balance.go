package model

import (
	"time"
)

// UserBalance 用户余额
type UserBalance struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"uniqueIndex" json:"user_id"`           // 用户ID
	Balance   float64   `gorm:"default:0;type:decimal(10,2)" json:"balance"`    // 可用余额
	Frozen    float64   `gorm:"default:0;type:decimal(10,2)" json:"frozen"`     // 冻结金额
	TotalIn   float64   `gorm:"default:0;type:decimal(10,2)" json:"total_in"`   // 累计入账
	TotalOut  float64   `gorm:"default:0;type:decimal(10,2)" json:"total_out"`  // 累计支出
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BalanceLog 余额变动记录
type BalanceLog struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"index" json:"user_id"`                          // 用户ID
	Type          string    `gorm:"size:20;index" json:"type"`                     // 类型：consume/refund/withdraw/freeze/unfreeze/adjust
	Amount        float64   `gorm:"type:decimal(10,2)" json:"amount"`              // 变动金额（正数增加，负数减少）
	BeforeBalance float64   `gorm:"type:decimal(10,2)" json:"before_balance"`      // 变动前余额
	AfterBalance  float64   `gorm:"type:decimal(10,2)" json:"after_balance"`       // 变动后余额
	OrderNo       string    `gorm:"size:64;index" json:"order_no"`                 // 关联订单号
	Remark        string    `gorm:"size:500" json:"remark"`                        // 备注
	OperatorID    uint      `gorm:"default:0" json:"operator_id"`                  // 操作者ID（管理员调整时）
	OperatorType  string    `gorm:"size:20;default:'user'" json:"operator_type"`   // 操作者类型：user/admin/system
	ClientIP      string    `gorm:"size:50" json:"client_ip"`                      // 客户端IP
	CreatedAt     time.Time `json:"created_at"`
}

// 余额变动类型常量
const (
	BalanceTypeConsume  = "consume"  // 消费
	BalanceTypeRefund   = "refund"   // 退款
	BalanceTypeWithdraw = "withdraw" // 提现
	BalanceTypeFreeze   = "freeze"   // 冻结
	BalanceTypeUnfreeze = "unfreeze" // 解冻
	BalanceTypeAdjust   = "adjust"   // 调整
)

// GetTypeText 获取余额变动类型文本
func (b *BalanceLog) GetTypeText() string {
	switch b.Type {
	case BalanceTypeConsume:
		return "消费"
	case BalanceTypeRefund:
		return "退款"
	case BalanceTypeWithdraw:
		return "提现"
	case BalanceTypeFreeze:
		return "冻结"
	case BalanceTypeUnfreeze:
		return "解冻"
	case BalanceTypeAdjust:
		return "调整"
	default:
		return "未知"
	}
}

// TableName 设置表名
func (UserBalance) TableName() string {
	return "user_balances"
}

// TableName 设置表名
func (BalanceLog) TableName() string {
	return "balance_logs"
}

// TableName 设置表名
