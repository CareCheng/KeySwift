package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	Username          string         `gorm:"type:varchar(100);uniqueIndex" json:"username"`
	Email             string         `gorm:"type:varchar(255);uniqueIndex" json:"email"`
	PasswordHash      string         `gorm:"type:varchar(255)" json:"-"`
	Phone             string         `gorm:"type:varchar(20)" json:"phone"`
	EmailVerified     bool           `gorm:"default:false" json:"email_verified"` // 邮箱是否已验证
	Enable2FA         bool           `gorm:"default:false" json:"enable_2fa"`     // 是否启用两步验证
	TOTPSecret        string         `gorm:"type:varchar(100)" json:"-"`
	PreferEmailAuth   bool           `gorm:"default:true" json:"prefer_email_auth"` // 登录时优先使用邮箱验证（否则使用TOTP）
	PayPassword       string         `gorm:"type:varchar(255)" json:"-"`            // 支付密码（bcrypt加密）
	PayPasswordSet    bool           `gorm:"default:false" json:"pay_password_set"` // 是否已设置支付密码
	PayPasswordErrors int            `gorm:"default:0" json:"-"`                    // 支付密码连续错误次数
	PayPasswordLockAt *time.Time     `json:"-"`                                     // 支付密码锁定时间
	Status            int            `gorm:"default:1" json:"status"`               // 1:正常 0:禁用
	LastLoginAt       *time.Time     `json:"last_login_at"`
	LastLoginIP       string         `gorm:"type:varchar(50)" json:"last_login_ip"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

// EmailVerifyCode 邮箱验证码
type EmailVerifyCode struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Email     string    `gorm:"type:varchar(255);index" json:"email"`
	Code      string    `gorm:"type:varchar(10)" json:"code"`
	Type      string    `gorm:"type:varchar(20)" json:"type"` // register, login, reset_password, enable_2fa
	Used      bool      `gorm:"default:false" json:"used"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// Product 商品模型
type Product struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Name         string         `gorm:"type:varchar(200)" json:"name"`
	Description  string         `gorm:"type:text" json:"description"`  // 简短描述
	Detail       string         `gorm:"type:text" json:"detail"`       // 详细介绍（Markdown/HTML）
	Specs        string         `gorm:"type:text" json:"specs"`        // 规格参数（JSON格式）
	Features     string         `gorm:"type:text" json:"features"`     // 特性/卖点列表（JSON格式）
	Tags         string         `gorm:"type:varchar(500)" json:"tags"` // 商品标签（逗号分隔）
	Price        float64        `json:"price"`
	Duration     int            `json:"duration"`                                          // 时长数值
	DurationUnit string         `gorm:"type:varchar(20);default:'天'" json:"duration_unit"` // 天/周/月/年
	Stock        int            `json:"stock"`                                             // 库存，-1表示无限
	Status       int            `gorm:"default:1" json:"status"`                           // 1:上架 0:下架
	SortOrder    int            `gorm:"default:0" json:"sort_order"`
	CategoryID   uint           `gorm:"default:0" json:"category_id"`  // 分类ID
	ProductType  int            `gorm:"default:1" json:"product_type"` // 商品类型：1手动卡密
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// Order 订单模型
type Order struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	OrderNo       string         `gorm:"type:varchar(64);uniqueIndex" json:"order_no"`
	PaymentNo     string         `gorm:"type:varchar(100)" json:"payment_no"` // 支付订单号
	UserID        uint           `gorm:"index" json:"user_id"`
	Username      string         `gorm:"type:varchar(100)" json:"username"`
	ProductID     uint           `json:"product_id"`
	ProductName   string         `gorm:"type:varchar(200)" json:"product_name"`
	Quantity      int            `gorm:"default:1" json:"quantity"`    // 购买数量
	OriginalPrice float64        `json:"original_price"`               // 原价（锁定商品价格，单价*数量）
	Price         float64        `json:"price"`                        // 实际应付金额
	PaidAmount    float64        `gorm:"default:0" json:"paid_amount"` // 实际支付金额（用于验证）
	Duration      int            `json:"duration"`
	DurationUnit  string         `gorm:"type:varchar(20)" json:"duration_unit"`
	Status        int            `gorm:"default:0" json:"status"` // 0:待支付 1:已支付 2:已完成 3:已取消 4:已退款
	PaymentMethod string         `gorm:"type:varchar(50)" json:"payment_method"`
	PaymentTime   *time.Time     `json:"payment_time"`
	KamiCode      string         `gorm:"type:text" json:"kami_code"` // 生成的卡密（多个用换行分隔）
	Remark        string         `gorm:"type:text" json:"remark"`
	ClientIP      string         `gorm:"type:varchar(50)" json:"client_ip"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// OrderStatus 订单状态常量
const (
	OrderStatusPending   = 0 // 待支付
	OrderStatusPaid      = 1 // 已支付
	OrderStatusCompleted = 2 // 已完成
	OrderStatusCancelled = 3 // 已取消
	OrderStatusRefunded  = 4 // 已退款
)

// ValidatePaymentAmount 验证支付金额
// 参数：paidAmount 实际支付金额
// 返回：是否有效，误差不超过0.01
func (o *Order) ValidatePaymentAmount(paidAmount float64) bool {
	diff := o.Price - paidAmount
	if diff < 0 {
		diff = -diff
	}
	// 允许0.01的误差（处理浮点精度问题）
	return diff <= 0.01
}

// SystemSetting 系统设置
type SystemSetting struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Key       string    `gorm:"type:varchar(100);uniqueIndex" json:"key"`
	Value     string    `gorm:"type:text" json:"value"`
	Remark    string    `gorm:"type:varchar(255)" json:"remark"`
	UpdatedAt time.Time `json:"updated_at"`
}

// EmailConfigDB 邮箱配置（数据库存储）
type EmailConfigDB struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Enabled      bool      `gorm:"default:false" json:"enabled"`
	SMTPHost     string    `gorm:"type:varchar(255)" json:"smtp_host"`
	SMTPPort     int       `gorm:"default:465" json:"smtp_port"`
	SMTPUser     string    `gorm:"type:varchar(255)" json:"smtp_user"`
	SMTPPassword string    `gorm:"type:varchar(255)" json:"smtp_password"`
	FromName     string    `gorm:"type:varchar(100)" json:"from_name"`
	FromEmail    string    `gorm:"type:varchar(255)" json:"from_email"`
	Encryption   string    `gorm:"type:varchar(20);default:'ssl'" json:"encryption"` // 加密方式：none/ssl/starttls
	CodeLength   int       `gorm:"default:6" json:"code_length"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// SystemConfigDB 系统配置（数据库存储）
type SystemConfigDB struct {
	ID                           uint      `gorm:"primaryKey" json:"id"`
	SystemTitle                  string    `gorm:"type:varchar(200)" json:"system_title"`                // 系统标题
	AdminSuffix                  string    `gorm:"type:varchar(100)" json:"admin_suffix"`                // 管理后台路径后缀
	EnableLogin                  bool      `gorm:"default:true" json:"enable_login"`                     // 是否启用后台登录验证
	EnableCaptcha                bool      `gorm:"default:true" json:"enable_captcha"`                   // 后台登录是否要求图形验证码
	AdminUsername                string    `gorm:"type:varchar(100)" json:"admin_username"`              // 管理员用户名
	AdminPassword                string    `gorm:"type:varchar(255)" json:"admin_password"`              // 管理员密码
	AdminPasswordInitialized     bool      `gorm:"default:false" json:"admin_password_initialized"`      // 初始管理员密码是否已由用户完成设置
	Enable2FA                    bool      `gorm:"default:false" json:"enable_2fa"`                      // 后台是否启用两步验证
	TOTPSecret                   string    `gorm:"type:varchar(100)" json:"totp_secret"`                 // 后台TOTP密钥
	EnableSessionTimeout         bool      `gorm:"default:true" json:"enable_session_timeout"`           // 后台是否启用会话超时
	SessionTimeout               int       `gorm:"default:60" json:"session_timeout"`                    // 后台会话超时分钟数
	UserAllowRegister            bool      `gorm:"default:true" json:"user_allow_register"`              // 是否允许用户注册
	UserEnableCaptcha            bool      `gorm:"default:true" json:"user_enable_captcha"`              // 用户登录/注册是否要求图形验证码
	UserEnable2FA                bool      `gorm:"default:true" json:"user_enable_2fa"`                  // 是否允许用户使用两步验证
	UserRequireEmailVerification bool      `gorm:"default:false" json:"user_require_email_verification"` // 注册是否要求邮箱验证码
	UserEnableSessionTimeout     bool      `gorm:"default:true" json:"user_enable_session_timeout"`      // 用户侧是否启用会话超时
	UserSessionTimeout           int       `gorm:"default:120" json:"user_session_timeout"`              // 用户侧会话超时分钟数
	EnableWhitelist              bool      `gorm:"default:false" json:"enable_whitelist"`                // 是否启用IP白名单
	IPWhitelist                  string    `gorm:"type:text" json:"ip_whitelist"`                        // IP白名单（JSON数组格式）
	CreatedAt                    time.Time `json:"created_at"`
	UpdatedAt                    time.Time `json:"updated_at"`
}

// LoginAttempt 登录尝试记录（用于登录失败锁定）
type LoginAttempt struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"type:varchar(100);index" json:"username"`
	IP        string    `gorm:"type:varchar(50);index" json:"ip"`
	Success   bool      `gorm:"default:false" json:"success"`
	CreatedAt time.Time `json:"created_at"`
}

// ProductCategory 商品分类
type ProductCategory struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"type:varchar(100)" json:"name"`
	Icon      string         `gorm:"type:varchar(50)" json:"icon"`
	SortOrder int            `gorm:"default:0" json:"sort_order"`
	Status    int            `gorm:"default:1" json:"status"` // 1:启用 0:禁用
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// UserSession 用户会话（数据库持久化）
type UserSession struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	SessionID string    `gorm:"type:varchar(64);uniqueIndex" json:"session_id"`
	UserID    uint      `gorm:"index" json:"user_id"`
	Username  string    `gorm:"type:varchar(100)" json:"username"`
	IP        string    `gorm:"type:varchar(50)" json:"ip"`
	UserAgent string    `gorm:"type:varchar(500)" json:"user_agent"`
	ExpiresAt time.Time `gorm:"index" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AdminSession 管理员会话（数据库持久化）
type AdminSession struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	SessionID string    `gorm:"type:varchar(64);uniqueIndex" json:"session_id"`
	Username  string    `gorm:"type:varchar(100)" json:"username"`
	Role      string    `gorm:"type:varchar(50)" json:"role"`
	IP        string    `gorm:"type:varchar(50)" json:"ip"`
	UserAgent string    `gorm:"type:varchar(500)" json:"user_agent"`
	Verified  bool      `gorm:"default:false" json:"verified"` // 2FA验证状态
	ExpiresAt time.Time `gorm:"index" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// LoginFailureRecord 登录失败记录（数据库持久化）
type LoginFailureRecord struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	Key          string     `gorm:"type:varchar(100);uniqueIndex" json:"key"` // IP或用户名
	FailureCount int        `gorm:"default:0" json:"failure_count"`
	FirstFailAt  time.Time  `json:"first_fail_at"`
	LockedAt     *time.Time `json:"locked_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// TableName 设置表名
func (User) TableName() string {
	return "users"
}

func (Product) TableName() string {
	return "products"
}

func (Order) TableName() string {
	return "orders"
}

func (SystemSetting) TableName() string {
	return "system_settings"
}

func (EmailVerifyCode) TableName() string {
	return "email_verify_codes"
}

func (EmailConfigDB) TableName() string {
	return "email_configs"
}

func (SystemConfigDB) TableName() string {
	return "system_configs"
}

func (LoginAttempt) TableName() string {
	return "login_attempts"
}

func (ProductCategory) TableName() string {
	return "product_categories"
}

func (UserSession) TableName() string {
	return "user_sessions"
}

func (AdminSession) TableName() string {
	return "admin_sessions"
}

func (LoginFailureRecord) TableName() string {
	return "login_failure_records"
}
