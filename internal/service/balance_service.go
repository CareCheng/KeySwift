package service

import (
	"errors"

	"user-frontend/internal/model"
	"user-frontend/internal/repository"

	"gorm.io/gorm"
)

// BalanceService 余额服务，宿主仅保留余额支付所需的基础能力。
type BalanceService struct {
	repo *repository.Repository
}

// OperatorInfo 操作者信息，用于余额变动日志。
type OperatorInfo struct {
	OperatorID   uint
	OperatorType string
	ClientIP     string
}

// NewBalanceService 创建余额服务。
func NewBalanceService(repo *repository.Repository) *BalanceService {
	return &BalanceService{repo: repo}
}

// GetUserBalance 获取用户余额；不存在时创建零余额账户。
func (s *BalanceService) GetUserBalance(userID uint) (*model.UserBalance, error) {
	db := s.repo.GetDB()
	var balance model.UserBalance
	err := db.Where("user_id = ?", userID).First(&balance).Error
	if err == gorm.ErrRecordNotFound {
		balance = model.UserBalance{
			UserID:  userID,
			Balance: 0,
			Frozen:  0,
		}
		if err := db.Create(&balance).Error; err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	return &balance, nil
}

// GetBalanceLogs 获取用户余额变动记录。
func (s *BalanceService) GetBalanceLogs(userID uint, page, pageSize int, logType string) ([]model.BalanceLog, int64, error) {
	db := s.repo.GetDB()
	var logs []model.BalanceLog
	var total int64

	query := db.Model(&model.BalanceLog{}).Where("user_id = ?", userID)
	if logType != "" {
		query = query.Where("type = ?", logType)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error
	return logs, total, err
}

// Consume 消费并扣减可用余额。
func (s *BalanceService) Consume(userID uint, amount float64, orderNo, remark string, operator *OperatorInfo) error {
	if amount <= 0 {
		return errors.New("消费金额必须大于0")
	}

	db := s.repo.GetDB()
	return db.Transaction(func(tx *gorm.DB) error {
		balance, err := s.lockBalance(tx, userID)
		if err != nil {
			return errors.New("余额记录不存在")
		}
		if balance.Balance < amount {
			return errors.New("余额不足")
		}

		beforeBalance := balance.Balance
		result := tx.Model(&model.UserBalance{}).
			Where("user_id = ? AND balance >= ?", userID, amount).
			Updates(map[string]interface{}{
				"balance":   gorm.Expr("balance - ?", amount),
				"total_out": gorm.Expr("total_out + ?", amount),
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("余额不足或更新失败")
		}

		return s.createLog(tx, &model.BalanceLog{
			UserID:        userID,
			Type:          model.BalanceTypeConsume,
			Amount:        -amount,
			BeforeBalance: beforeBalance,
			AfterBalance:  beforeBalance - amount,
			OrderNo:       orderNo,
			Remark:        remark,
		}, operator, "user")
	})
}

// Refund 退款到可用余额。
func (s *BalanceService) Refund(userID uint, amount float64, orderNo, remark string, operator *OperatorInfo) error {
	if amount <= 0 {
		return errors.New("退款金额必须大于0")
	}

	db := s.repo.GetDB()
	return db.Transaction(func(tx *gorm.DB) error {
		balance, err := s.getOrCreateBalance(tx, userID)
		if err != nil {
			return err
		}

		beforeBalance := balance.Balance
		if err := tx.Model(&model.UserBalance{}).
			Where("user_id = ?", userID).
			Updates(map[string]interface{}{
				"balance":   gorm.Expr("balance + ?", amount),
				"total_out": gorm.Expr("total_out - ?", amount),
			}).Error; err != nil {
			return err
		}

		return s.createLog(tx, &model.BalanceLog{
			UserID:        userID,
			Type:          model.BalanceTypeRefund,
			Amount:        amount,
			BeforeBalance: beforeBalance,
			AfterBalance:  beforeBalance + amount,
			OrderNo:       orderNo,
			Remark:        remark,
		}, operator, "system")
	})
}

// Freeze 冻结可用余额。
func (s *BalanceService) Freeze(userID uint, amount float64, orderNo, remark string, operator *OperatorInfo) error {
	if amount <= 0 {
		return errors.New("冻结金额必须大于0")
	}

	db := s.repo.GetDB()
	return db.Transaction(func(tx *gorm.DB) error {
		balance, err := s.lockBalance(tx, userID)
		if err != nil {
			return errors.New("余额记录不存在")
		}
		if balance.Balance < amount {
			return errors.New("可用余额不足")
		}

		beforeBalance := balance.Balance
		result := tx.Model(&model.UserBalance{}).
			Where("user_id = ? AND balance >= ?", userID, amount).
			Updates(map[string]interface{}{
				"balance": gorm.Expr("balance - ?", amount),
				"frozen":  gorm.Expr("frozen + ?", amount),
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("可用余额不足或更新失败")
		}

		return s.createLog(tx, &model.BalanceLog{
			UserID:        userID,
			Type:          model.BalanceTypeFreeze,
			Amount:        -amount,
			BeforeBalance: beforeBalance,
			AfterBalance:  beforeBalance - amount,
			OrderNo:       orderNo,
			Remark:        remark,
		}, operator, "user")
	})
}

// Unfreeze 解冻余额。
func (s *BalanceService) Unfreeze(userID uint, amount float64, orderNo, remark string, operator *OperatorInfo) error {
	if amount <= 0 {
		return errors.New("解冻金额必须大于0")
	}

	db := s.repo.GetDB()
	return db.Transaction(func(tx *gorm.DB) error {
		balance, err := s.lockBalance(tx, userID)
		if err != nil {
			return errors.New("余额记录不存在")
		}
		if balance.Frozen < amount {
			return errors.New("冻结余额不足")
		}

		beforeBalance := balance.Balance
		result := tx.Model(&model.UserBalance{}).
			Where("user_id = ? AND frozen >= ?", userID, amount).
			Updates(map[string]interface{}{
				"balance": gorm.Expr("balance + ?", amount),
				"frozen":  gorm.Expr("frozen - ?", amount),
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("冻结余额不足或更新失败")
		}

		return s.createLog(tx, &model.BalanceLog{
			UserID:        userID,
			Type:          model.BalanceTypeUnfreeze,
			Amount:        amount,
			BeforeBalance: beforeBalance,
			AfterBalance:  beforeBalance + amount,
			OrderNo:       orderNo,
			Remark:        remark,
		}, operator, "user")
	})
}

// DeductFrozen 扣除冻结金额。
func (s *BalanceService) DeductFrozen(userID uint, amount float64, orderNo, remark string, operator *OperatorInfo) error {
	if amount <= 0 {
		return errors.New("扣除金额必须大于0")
	}

	db := s.repo.GetDB()
	return db.Transaction(func(tx *gorm.DB) error {
		balance, err := s.lockBalance(tx, userID)
		if err != nil {
			return errors.New("余额记录不存在")
		}
		if balance.Frozen < amount {
			return errors.New("冻结余额不足")
		}

		currentBalance := balance.Balance
		result := tx.Model(&model.UserBalance{}).
			Where("user_id = ? AND frozen >= ?", userID, amount).
			Updates(map[string]interface{}{
				"frozen":    gorm.Expr("frozen - ?", amount),
				"total_out": gorm.Expr("total_out + ?", amount),
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("冻结余额不足或更新失败")
		}

		return s.createLog(tx, &model.BalanceLog{
			UserID:        userID,
			Type:          model.BalanceTypeConsume,
			Amount:        -amount,
			BeforeBalance: currentBalance,
			AfterBalance:  currentBalance,
			OrderNo:       orderNo,
			Remark:        remark + "（从冻结扣除）",
		}, operator, "user")
	})
}

// AdjustBalance 管理员调整用户余额。
func (s *BalanceService) AdjustBalance(userID uint, amount float64, remark string, operator *OperatorInfo) error {
	db := s.repo.GetDB()
	return db.Transaction(func(tx *gorm.DB) error {
		balance, err := s.getOrCreateBalance(tx, userID)
		if err != nil {
			return err
		}

		beforeBalance := balance.Balance
		newBalance := balance.Balance + amount
		if newBalance < 0 {
			return errors.New("调整后余额不能为负数")
		}

		updates := map[string]interface{}{
			"balance": gorm.Expr("balance + ?", amount),
		}
		if amount > 0 {
			updates["total_in"] = gorm.Expr("total_in + ?", amount)
		} else {
			updates["total_out"] = gorm.Expr("total_out - ?", amount)
		}

		if err := tx.Model(&model.UserBalance{}).
			Where("user_id = ?", userID).
			Updates(updates).Error; err != nil {
			return err
		}

		return s.createLog(tx, &model.BalanceLog{
			UserID:        userID,
			Type:          model.BalanceTypeAdjust,
			Amount:        amount,
			BeforeBalance: beforeBalance,
			AfterBalance:  newBalance,
			Remark:        remark,
		}, operator, "admin")
	})
}

// AdminGetAllBalances 管理员获取所有用户余额。
func (s *BalanceService) AdminGetAllBalances(page, pageSize int, keyword string) ([]map[string]interface{}, int64, error) {
	db := s.repo.GetDB()
	var total int64

	query := db.Table("user_balances").
		Select("user_balances.*, users.username, users.email").
		Joins("LEFT JOIN users ON users.id = user_balances.user_id")

	if keyword != "" {
		query = query.Where("users.username LIKE ? OR users.email LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	query.Count(&total)

	var results []map[string]interface{}
	offset := (page - 1) * pageSize
	err := query.Order("user_balances.balance DESC").Offset(offset).Limit(pageSize).Find(&results).Error
	return results, total, err
}

// AdminGetBalanceLogs 管理员获取余额变动记录。
func (s *BalanceService) AdminGetBalanceLogs(page, pageSize int, userID uint, logType string) ([]model.BalanceLog, int64, error) {
	db := s.repo.GetDB()
	var logs []model.BalanceLog
	var total int64

	query := db.Model(&model.BalanceLog{})
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if logType != "" {
		query = query.Where("type = ?", logType)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error
	return logs, total, err
}

// GetBalanceStats 获取基础余额统计。
func (s *BalanceService) GetBalanceStats() (map[string]interface{}, error) {
	db := s.repo.GetDB()
	var stats struct {
		TotalBalance float64
		TotalFrozen  float64
		TotalIn      float64
		TotalOut     float64
		UserCount    int64
	}

	db.Model(&model.UserBalance{}).Select("SUM(balance) as total_balance, SUM(frozen) as total_frozen, SUM(total_in) as total_in, SUM(total_out) as total_out, COUNT(*) as user_count").Scan(&stats)

	return map[string]interface{}{
		"total_balance": stats.TotalBalance,
		"total_frozen":  stats.TotalFrozen,
		"total_in":      stats.TotalIn,
		"total_out":     stats.TotalOut,
		"user_count":    stats.UserCount,
	}, nil
}

func (s *BalanceService) lockBalance(tx *gorm.DB, userID uint) (*model.UserBalance, error) {
	var balance model.UserBalance
	err := tx.Set("gorm:query_option", "FOR UPDATE").Where("user_id = ?", userID).First(&balance).Error
	return &balance, err
}

func (s *BalanceService) getOrCreateBalance(tx *gorm.DB, userID uint) (*model.UserBalance, error) {
	balance, err := s.lockBalance(tx, userID)
	if err == gorm.ErrRecordNotFound {
		balance = &model.UserBalance{
			UserID:  userID,
			Balance: 0,
			Frozen:  0,
		}
		if err := tx.Create(balance).Error; err != nil {
			return nil, err
		}
		return balance, nil
	}
	if err != nil {
		return nil, err
	}
	return balance, nil
}

func (s *BalanceService) createLog(tx *gorm.DB, log *model.BalanceLog, operator *OperatorInfo, defaultOperatorType string) error {
	if operator != nil {
		log.OperatorID = operator.OperatorID
		log.OperatorType = operator.OperatorType
		log.ClientIP = operator.ClientIP
	} else {
		log.OperatorType = defaultOperatorType
	}
	return tx.Create(log).Error
}
