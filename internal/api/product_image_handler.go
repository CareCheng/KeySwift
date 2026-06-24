package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// ==================== 商品图片 API ====================

// GetProductImages 获取商品图片列表
func GetProductImages(c *gin.Context) {
	if ProductImageSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	productID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "无效的商品ID"})
		return
	}

	images, err := ProductImageSvc.GetProductImages(uint(productID))
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取图片失败"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": images})
}

// ==================== 管理员商品图片 API ====================

// AdminUploadProductImages 管理员上传商品图片（多图片）
func AdminUploadProductImages(c *gin.Context) {
	if ProductImageSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	productID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "无效的商品ID"})
		return
	}

	// 检查图片数量限制
	count := ProductImageSvc.GetImageCount(uint(productID))
	if count >= 10 {
		c.JSON(400, gin.H{"success": false, "error": "最多只能上传10张图片"})
		return
	}

	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "请选择图片文件"})
		return
	}
	defer file.Close()

	// 检查文件大小（最大5MB）
	if header.Size > 5*1024*1024 {
		c.JSON(400, gin.H{"success": false, "error": "图片大小不能超过5MB"})
		return
	}

	isPrimary := c.PostForm("is_primary") == "true"

	image, err := ProductImageSvc.UploadProductImage(uint(productID), header.Filename, file, isPrimary)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 记录操作日志
	adminUsername, _ := c.Get("admin_username")
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(adminUsername.(string), "上传商品图片", "product_image", c.Param("id"), gin.H{"filename": header.Filename}, GetClientIP(c), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "data": image})
}

// AdminDeleteProductImages 管理员删除商品图片（多图片）
func AdminDeleteProductImages(c *gin.Context) {
	if ProductImageSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	imageID, err := strconv.ParseUint(c.Param("image_id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "无效的图片ID"})
		return
	}

	if err := ProductImageSvc.DeleteProductImage(uint(imageID)); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 记录操作日志
	adminUsername, _ := c.Get("admin_username")
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(adminUsername.(string), "删除商品图片", "product_image", c.Param("image_id"), nil, GetClientIP(c), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "message": "删除成功"})
}

// AdminSetPrimaryImage 管理员设置主图
func AdminSetPrimaryImage(c *gin.Context) {
	if ProductImageSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	productID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "无效的商品ID"})
		return
	}

	var req struct {
		ImageID uint `json:"image_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if err := ProductImageSvc.SetPrimaryImage(uint(productID), req.ImageID); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 记录操作日志
	adminUsername, _ := c.Get("admin_username")
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(adminUsername.(string), "设置商品主图", "product_image", c.Param("id"), req, GetClientIP(c), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "message": "设置成功"})
}

// AdminUpdateImageOrder 管理员更新图片排序
func AdminUpdateImageOrder(c *gin.Context) {
	if ProductImageSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	productID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "无效的商品ID"})
		return
	}

	var req struct {
		ImageIDs []uint `json:"image_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if err := ProductImageSvc.UpdateImageOrder(uint(productID), req.ImageIDs); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 记录操作日志
	adminUsername, _ := c.Get("admin_username")
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(adminUsername.(string), "更新图片排序", "product_image", c.Param("id"), req, GetClientIP(c), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "message": "排序更新成功"})
}

// AdminGetProductImages 管理员获取商品图片列表
func AdminGetProductImages(c *gin.Context) {
	if ProductImageSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	productID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "无效的商品ID"})
		return
	}

	images, err := ProductImageSvc.GetProductImages(uint(productID))
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取图片失败"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": images})
}
