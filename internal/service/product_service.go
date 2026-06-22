package service

import (
	"encoding/json"
	"errors"
	"log"

	"user-frontend/internal/cache"
	"user-frontend/internal/model"
	"user-frontend/internal/repository"
)

type ProductService struct {
	repo *repository.Repository
}

func NewProductService(repo *repository.Repository) *ProductService {
	return &ProductService{repo: repo}
}

// ==================== 缓存辅助方法 ====================

// cacheProduct 缓存商品信息
func (s *ProductService) cacheProduct(product *model.Product) {
	cm := cache.GetManager()
	if cm == nil {
		return
	}

	key := cache.ProductKey(product.ID)
	data, err := json.Marshal(product)
	if err != nil {
		log.Printf("[ProductService] 序列化商品缓存失败: %v", err)
		return
	}

	if err := cm.Set(key, string(data), cache.ProductTTL); err != nil {
		log.Printf("[ProductService] 缓存商品失败: %v", err)
	}
}

// getProductFromCache 从缓存获取商品
func (s *ProductService) getProductFromCache(productID uint) *model.Product {
	cm := cache.GetManager()
	if cm == nil {
		return nil
	}

	key := cache.ProductKey(productID)
	data, ok := cm.Get(key)
	if !ok {
		return nil
	}

	dataStr, ok := data.(string)
	if !ok {
		return nil
	}

	var product model.Product
	if err := json.Unmarshal([]byte(dataStr), &product); err != nil {
		log.Printf("[ProductService] 反序列化商品缓存失败: %v", err)
		return nil
	}

	return &product
}

// invalidateProductCache 使商品缓存失效
func (s *ProductService) invalidateProductCache(productID uint) {
	cm := cache.GetManager()
	if cm == nil {
		return
	}

	key := cache.ProductKey(productID)
	if err := cm.Delete(key); err != nil {
		log.Printf("[ProductService] 删除商品缓存失败: %v", err)
	}

	// 同时清除商品列表缓存（因为商品变更会影响列表）
	s.invalidateProductListCache()
}

// invalidateProductListCache 使商品列表缓存失效
func (s *ProductService) invalidateProductListCache() {
	cm := cache.GetManager()
	if cm == nil {
		return
	}

	// 清除常见的列表缓存键；列表缓存键包含分页参数，无法枚举所有组合。
	commonPages := []int{1, 2, 3, 4, 5}
	commonSizes := []int{10, 20, 50}
	for _, page := range commonPages {
		for _, size := range commonSizes {
			for _, active := range []bool{true, false} {
				key := cache.ProductListKey(page, size, active, 0)
				cm.Delete(key)
			}
		}
	}
}

// CreateProduct 创建商品
func (s *ProductService) CreateProduct(name, description string, price float64, duration int, durationUnit string, stock int) (*model.Product, error) {
	if name == "" {
		return nil, errors.New("商品名称不能为空")
	}
	if price < 0 {
		return nil, errors.New("价格不能为负数")
	}
	if duration <= 0 {
		return nil, errors.New("时长必须大于0")
	}

	product := &model.Product{
		Name:         name,
		Description:  description,
		Price:        price,
		Duration:     duration,
		DurationUnit: durationUnit,
		Stock:        stock,
		Status:       1,
		ProductType:  model.ProductTypeManual, // 默认手动卡密类型
	}

	if err := s.repo.CreateProduct(product); err != nil {
		return nil, err
	}

	return product, nil
}

// CreateProductFull 创建商品（完整字段）
func (s *ProductService) CreateProductFull(product *model.Product) error {
	if product.Name == "" {
		return errors.New("商品名称不能为空")
	}
	if product.Price < 0 {
		return errors.New("价格不能为负数")
	}
	if product.Duration <= 0 {
		return errors.New("时长必须大于0")
	}
	product.Status = 1
	return s.repo.CreateProduct(product)
}

// UpdateProduct 更新商品
func (s *ProductService) UpdateProduct(id uint, name, description string, price float64, duration int, durationUnit string, stock int, status int) (*model.Product, error) {
	product, err := s.repo.GetProductByID(id)
	if err != nil {
		return nil, errors.New("商品不存在")
	}

	if name != "" {
		product.Name = name
	}
	product.Description = description
	if price >= 0 {
		product.Price = price
	}
	if duration > 0 {
		product.Duration = duration
	}
	if durationUnit != "" {
		product.DurationUnit = durationUnit
	}
	// 手动卡密类型的商品库存由卡密数量决定，不允许手动修改
	if product.ProductType != model.ProductTypeManual {
		product.Stock = stock
	}
	product.Status = status

	if err := s.repo.UpdateProduct(product); err != nil {
		return nil, err
	}

	// 更新成功后使缓存失效
	s.invalidateProductCache(id)

	return product, nil
}

// UpdateProductFull 更新商品（完整字段）
func (s *ProductService) UpdateProductFull(product *model.Product) error {
	existing, err := s.repo.GetProductByID(product.ID)
	if err != nil {
		return errors.New("商品不存在")
	}
	// 手动卡密类型的商品库存由卡密数量决定，不允许手动修改
	if existing.ProductType == model.ProductTypeManual {
		product.Stock = existing.Stock
	}
	err = s.repo.UpdateProduct(product)
	if err == nil {
		s.invalidateProductCache(product.ID)
	}
	return err
}

// DeleteProduct 删除商品
func (s *ProductService) DeleteProduct(id uint) error {
	err := s.repo.DeleteProduct(id)
	if err == nil {
		s.invalidateProductCache(id)
	}
	return err
}

// GetProductByID 获取商品（支持缓存）
func (s *ProductService) GetProductByID(id uint) (*model.Product, error) {
	// 先从缓存获取
	if product := s.getProductFromCache(id); product != nil {
		return product, nil
	}

	// 从数据库获取
	product, err := s.repo.GetProductByID(id)
	if err != nil {
		return nil, err
	}

	// 缓存商品
	s.cacheProduct(product)

	return product, nil
}

// GetAllProducts 获取所有商品
func (s *ProductService) GetAllProducts(onlyActive bool) ([]model.Product, error) {
	return s.repo.GetAllProducts(onlyActive)
}

// GetProductsWithPagination 分页获取商品
func (s *ProductService) GetProductsWithPagination(page, pageSize int, onlyActive bool) ([]model.Product, int64, error) {
	return s.repo.GetProductsWithPagination(page, pageSize, onlyActive)
}

// UpdateProductStatus 更新商品状态
func (s *ProductService) UpdateProductStatus(id uint, status int) error {
	product, err := s.repo.GetProductByID(id)
	if err != nil {
		return errors.New("商品不存在")
	}

	product.Status = status
	err = s.repo.UpdateProduct(product)
	if err == nil {
		s.invalidateProductCache(id)
	}
	return err
}

// UpdateProductStock 更新商品库存
func (s *ProductService) UpdateProductStock(id uint, stock int) error {
	product, err := s.repo.GetProductByID(id)
	if err != nil {
		return errors.New("商品不存在")
	}

	product.Stock = stock
	err = s.repo.UpdateProduct(product)
	if err == nil {
		s.invalidateProductCache(id)
	}
	return err
}
