package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"user-frontend/internal/config"
	"user-frontend/internal/model"
	"user-frontend/internal/repository"
	"user-frontend/internal/utils"
)

type OrderService struct {
	repo          *repository.Repository
	cfg           *config.Config
	configSvc     *ConfigService
	manualKamiSvc *ManualKamiService
}

func NewOrderService(repo *repository.Repository, cfg *config.Config) *OrderService {
	return &OrderService{
		repo: repo,
		cfg:  cfg,
	}
}

// SetConfigService 设置配置服务
func (s *OrderService) SetConfigService(configSvc *ConfigService) {
	s.configSvc = configSvc
}

// SetManualKamiService 设置手动卡密服务
func (s *OrderService) SetManualKamiService(manualKamiSvc *ManualKamiService) {
	s.manualKamiSvc = manualKamiSvc
}

// CreateOrderParams 创建订单参数
type CreateOrderParams struct {
	UserID    uint
	Username  string
	ProductID uint
	Quantity  int
	ClientIP  string
	Remark    string
}

// CreateOrder 创建单数量订单。
// 使用手动卡密模式：本地生成订单号，从卡密池分配卡密
func (s *OrderService) CreateOrder(userID uint, username string, productID uint, clientIP string) (*model.Order, error) {
	return s.CreateOrderWithParams(&CreateOrderParams{
		UserID:    userID,
		Username:  username,
		ProductID: productID,
		Quantity:  1,
		ClientIP:  clientIP,
	})
}

// CreateOrderWithQuantity 创建订单（支持数量）
func (s *OrderService) CreateOrderWithQuantity(userID uint, username string, productID uint, quantity int, clientIP string) (*model.Order, error) {
	if quantity < 1 {
		quantity = 1
	}
	return s.CreateOrderWithParams(&CreateOrderParams{
		UserID:    userID,
		Username:  username,
		ProductID: productID,
		Quantity:  quantity,
		ClientIP:  clientIP,
	})
}

// CreateOrderWithRemark 创建带备注的订单
func (s *OrderService) CreateOrderWithRemark(userID uint, username string, productID uint, clientIP string, remark string) (*model.Order, error) {
	return s.CreateOrderWithParams(&CreateOrderParams{
		UserID:    userID,
		Username:  username,
		ProductID: productID,
		Quantity:  1,
		ClientIP:  clientIP,
		Remark:    remark,
	})
}

// CreateOrderWithParams 创建订单（完整参数）
// 安全特性：
//   - 锁定商品原价到订单
//   - 支持多数量购买
func (s *OrderService) CreateOrderWithParams(params *CreateOrderParams) (*model.Order, error) {
	// 获取商品信息
	product, err := s.repo.GetProductByID(params.ProductID)
	if err != nil {
		return nil, errors.New("商品不存在")
	}

	if product.Status != 1 {
		return nil, errors.New("商品已下架")
	}

	// 数量校验
	quantity := params.Quantity
	if quantity < 1 {
		quantity = 1
	}

	// 检查库存（-1表示无限库存）
	if product.Stock == 0 {
		return nil, errors.New("商品库存不足")
	}
	if product.Stock != -1 && product.Stock < quantity {
		return nil, fmt.Errorf("商品库存不足，当前库存: %d", product.Stock)
	}

	// 生成订单号（本地生成）
	orderNo := utils.GenerateLocalOrderNo()

	// 计算应付金额（单价 * 数量）
	unitPrice := product.Price
	originalPrice := unitPrice * float64(quantity)

	// 创建订单，锁定价格
	order := &model.Order{
		OrderNo:       orderNo,
		UserID:        params.UserID,
		Username:      params.Username,
		ProductID:     params.ProductID,
		ProductName:   product.Name,
		Quantity:      quantity,
		OriginalPrice: originalPrice,
		Price:         originalPrice,
		Duration:      product.Duration,
		DurationUnit:  product.DurationUnit,
		Status:        model.OrderStatusPending,
		ClientIP:      params.ClientIP,
		Remark:        params.Remark,
	}

	if err := s.repo.CreateOrder(order); err != nil {
		return nil, err
	}

	return order, nil
}

// ProcessPaymentParams 处理支付参数
type ProcessPaymentParams struct {
	OrderNo       string
	PaymentMethod string
	PaymentNo     string
	PaidAmount    float64 // 实际支付金额（用于验证）
}

// ProcessPayment 处理支付（正式订单）
// 安全特性：验证支付金额、原子库存扣减
func (s *OrderService) ProcessPayment(orderNo, paymentMethod, paymentNo string) (*model.Order, error) {
	return s.ProcessPaymentWithAmount(orderNo, paymentMethod, paymentNo, 0)
}

// ProcessPaymentWithAmount 处理支付（带金额验证）
// 参数：
//   - orderNo: 订单号
//   - paymentMethod: 支付方式
//   - paymentNo: 支付流水号
//   - paidAmount: 实际支付金额（0表示跳过验证，用于无法获取金额的支付方式）
func (s *OrderService) ProcessPaymentWithAmount(orderNo, paymentMethod, paymentNo string, paidAmount float64) (*model.Order, error) {
	order, err := s.repo.GetOrderByOrderNo(orderNo)
	if err != nil {
		return nil, errors.New("订单不存在")
	}

	if order.Status != model.OrderStatusPending {
		// 如果订单已完成，直接返回（幂等处理）
		if order.Status == model.OrderStatusCompleted {
			return order, nil
		}
		return nil, errors.New("订单状态异常")
	}

	// 验证支付金额（如果提供了金额）
	if paidAmount > 0 && !order.ValidatePaymentAmount(paidAmount) {
		return nil, fmt.Errorf("支付金额不匹配，应付: %.2f, 实付: %.2f", order.Price, paidAmount)
	}

	// 获取商品信息
	product, _ := s.repo.GetProductByID(order.ProductID)
	if product == nil {
		return nil, errors.New("商品不存在")
	}

	quantity := order.Quantity
	if quantity < 1 {
		quantity = 1
	}

	// 原子扣减库存（非无限库存）
	if product.Stock != -1 {
		affected, err := s.repo.DecrementProductStock(product.ID, quantity)
		if err != nil {
			return nil, errors.New("库存扣减失败")
		}
		if affected == 0 {
			return nil, errors.New("商品库存不足，请联系管理员处理")
		}
	}

	var kamiCodes []string

	// 从本地卡密池获取多个卡密
	for i := 0; i < quantity; i++ {
		kamiCode, err := s.getManualKami(product.ID, order.ID, order.OrderNo)
		if err != nil {
			// 回滚库存
			if product.Stock != -1 {
				s.repo.IncrementProductStock(product.ID, quantity)
			}
			return nil, err
		}
		kamiCodes = append(kamiCodes, kamiCode)
	}

	// 更新订单状态
	now := time.Now()
	order.Status = model.OrderStatusCompleted
	order.PaymentMethod = paymentMethod
	order.PaymentNo = paymentNo
	order.PaymentTime = &now
	order.PaidAmount = paidAmount
	// 多个卡密用换行符分隔
	order.KamiCode = strings.Join(kamiCodes, "\n")

	if err := s.repo.UpdateOrder(order); err != nil {
		return nil, err
	}

	return order, nil
}

// getManualKami 从手动卡密池获取卡密
func (s *OrderService) getManualKami(productID uint, orderID uint, orderNo string) (string, error) {
	if s.manualKamiSvc == nil {
		return "", errors.New("手动卡密服务未初始化")
	}

	// 获取一个可用的卡密
	kami, err := s.manualKamiSvc.GetAvailableKami(productID)
	if err != nil {
		return "", errors.New("卡密库存不足")
	}

	// 标记卡密为已售出
	if err := s.manualKamiSvc.MarkKamiSold(kami.ID, orderID, orderNo); err != nil {
		return "", errors.New("卡密分配失败")
	}

	return kami.KamiCode, nil
}

// ValidateOrderOwnership 验证订单归属
// 用于支付接口验证用户是否有权操作订单
func (s *OrderService) ValidateOrderOwnership(orderNo string, userID uint) (*model.Order, error) {
	order, err := s.repo.GetOrderByOrderNo(orderNo)
	if err != nil {
		return nil, errors.New("订单不存在")
	}

	if order.UserID != userID {
		return nil, errors.New("无权操作此订单")
	}

	return order, nil
}

// GetOrderByOrderNo 获取订单
func (s *OrderService) GetOrderByOrderNo(orderNo string) (*model.Order, error) {
	return s.repo.GetOrderByOrderNo(orderNo)
}

// GetUserOrders 获取用户订单列表
func (s *OrderService) GetUserOrders(userID uint, page, pageSize int) ([]model.Order, int64, error) {
	return s.repo.GetOrdersByUserID(userID, page, pageSize)
}

// CancelOrder 取消订单
func (s *OrderService) CancelOrder(orderNo string, userID uint) error {
	order, err := s.repo.GetOrderByOrderNo(orderNo)
	if err != nil {
		return errors.New("订单不存在")
	}

	if order.UserID != userID {
		return errors.New("无权操作此订单")
	}

	if order.Status != model.OrderStatusPending {
		return errors.New("只能取消待支付订单")
	}

	order.Status = model.OrderStatusCancelled
	return s.repo.UpdateOrder(order)
}

// GetOrderStats 获取订单统计
func (s *OrderService) GetOrderStats() (map[string]interface{}, error) {
	return s.repo.GetOrderStats()
}

// GetAllOrders 获取所有订单
func (s *OrderService) GetAllOrders(page, pageSize int, status *int) ([]model.Order, int64, error) {
	return s.repo.GetAllOrders(page, pageSize, status)
}

// GetOrderByID 获取订单
func (s *OrderService) GetOrderByID(id uint) (*model.Order, error) {
	return s.repo.GetOrderByID(id)
}

// CancelExpiredOrders 取消过期订单
func (s *OrderService) CancelExpiredOrders(expireMinutes int) (int64, error) {
	return s.repo.CancelExpiredOrders(expireMinutes)
}

// SearchOrders 搜索订单
func (s *OrderService) SearchOrders(params *repository.OrderSearchParams, page, pageSize int) ([]model.Order, int64, error) {
	return s.repo.SearchOrders(params, page, pageSize)
}

// GetOrderByOrderNoAndEmail 通过订单号和邮箱查询订单
func (s *OrderService) GetOrderByOrderNoAndEmail(orderNo, email string) (*model.Order, error) {
	return s.repo.GetOrderByOrderNoAndEmail(orderNo, email)
}

// GetOrderStatsByDateRange 获取日期范围内的订单统计
func (s *OrderService) GetOrderStatsByDateRange(startDate, endDate time.Time) ([]map[string]interface{}, error) {
	return s.repo.GetOrderStatsByDateRange(startDate, endDate)
}
