package api

import (
	"strconv"

	"user-frontend/internal/model"

	"github.com/gin-gonic/gin"
)

// ==================== 商品管理相关 API ====================

// AdminGetProducts 获取商品列表
func AdminGetProducts(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	if ProductSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	products, err := ProductSvc.GetAllProducts(false)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "products": buildProductResponses(products)})
}

// AdminCreateProduct 创建商品
func AdminCreateProduct(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	if ProductSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	var req struct {
		Name         string  `json:"name" binding:"required"`
		Description  string  `json:"description"`
		Detail       string  `json:"detail"`   // 详细介绍（Markdown/HTML）
		Specs        string  `json:"specs"`    // 规格参数（JSON格式）
		Features     string  `json:"features"` // 特性/卖点列表（JSON格式）
		Tags         string  `json:"tags"`     // 商品标签（逗号分隔）
		Price        float64 `json:"price"`
		Duration     int     `json:"duration" binding:"required"`
		DurationUnit string  `json:"duration_unit"`
		Stock        int     `json:"stock"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if req.DurationUnit == "" {
		req.DurationUnit = "天"
	}

	// 手动卡密类型初始库存为0，由导入的卡密数量决定
	stock := 0

	// 使用完整字段创建商品
	product := &model.Product{
		Name:         req.Name,
		Description:  req.Description,
		Detail:       req.Detail,
		Specs:        req.Specs,
		Features:     req.Features,
		Tags:         req.Tags,
		Price:        req.Price,
		Duration:     req.Duration,
		DurationUnit: req.DurationUnit,
		Stock:        stock,
		ProductType:  model.ProductTypeManual,
	}

	if err := ProductSvc.CreateProductFull(product); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "product": buildProductResponse(product)})
}

// AdminUpdateProduct 更新商品
func AdminUpdateProduct(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	if ProductSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	var req struct {
		Name         string  `json:"name"`
		Description  string  `json:"description"`
		Detail       string  `json:"detail"`   // 详细介绍（Markdown/HTML）
		Specs        string  `json:"specs"`    // 规格参数（JSON格式）
		Features     string  `json:"features"` // 特性/卖点列表（JSON格式）
		Tags         string  `json:"tags"`     // 商品标签（逗号分隔）
		Price        float64 `json:"price"`
		Duration     int     `json:"duration"`
		DurationUnit string  `json:"duration_unit"`
		Stock        int     `json:"stock"`
		Status       int     `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	// 获取现有商品
	existing, err := ProductSvc.GetProductByID(uint(id))
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "商品不存在"})
		return
	}

	// 更新字段
	if req.Name != "" {
		existing.Name = req.Name
	}
	existing.Description = req.Description
	existing.Detail = req.Detail
	existing.Specs = req.Specs
	existing.Features = req.Features
	existing.Tags = req.Tags
	if req.Price >= 0 {
		existing.Price = req.Price
	}
	if req.Duration > 0 {
		existing.Duration = req.Duration
	}
	if req.DurationUnit != "" {
		existing.DurationUnit = req.DurationUnit
	}
	existing.Stock = req.Stock
	existing.Status = req.Status

	if err := ProductSvc.UpdateProductFull(existing); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "product": buildProductResponse(existing)})
}

// AdminDeleteProduct 删除商品
func AdminDeleteProduct(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	if ProductSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	if err := ProductSvc.DeleteProduct(uint(id)); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "商品已删除"})
}
