package service

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"user-frontend/internal/model"
	"user-frontend/internal/repository"
)

// ProductImageService 商品图片服务
type ProductImageService struct {
	repo *repository.Repository
}

// NewProductImageService 创建商品图片服务
func NewProductImageService(repo *repository.Repository) *ProductImageService {
	return &ProductImageService{repo: repo}
}

// GetProductImages 获取商品的所有图片
func (s *ProductImageService) GetProductImages(productID uint) ([]model.ProductImage, error) {
	var images []model.ProductImage
	err := s.repo.GetDB().Where("product_id = ?", productID).Order("sort_order ASC, id ASC").Find(&images).Error
	return images, err
}

// GetPrimaryImage 获取商品主图
func (s *ProductImageService) GetPrimaryImage(productID uint) (*model.ProductImage, error) {
	var image model.ProductImage
	err := s.repo.GetDB().Where("product_id = ? AND is_primary = ?", productID, true).First(&image).Error
	if err != nil {
		// 如果没有主图，返回第一张图片
		err = s.repo.GetDB().Where("product_id = ?", productID).Order("sort_order ASC, id ASC").First(&image).Error
	}
	return &image, err
}

// AddProductImage 添加商品图片
func (s *ProductImageService) AddProductImage(productID uint, url string, isPrimary bool) (*model.ProductImage, error) {
	db := s.repo.GetDB()

	// 如果设置为主图，先取消其他主图
	if isPrimary {
		db.Model(&model.ProductImage{}).Where("product_id = ?", productID).Update("is_primary", false)
	}

	// 获取当前最大排序值
	var maxSort int
	db.Model(&model.ProductImage{}).Where("product_id = ?", productID).Select("COALESCE(MAX(sort_order), 0)").Scan(&maxSort)

	image := &model.ProductImage{
		ProductID: productID,
		URL:       url,
		SortOrder: maxSort + 1,
		IsPrimary: isPrimary,
	}

	err := db.Create(image).Error
	return image, err
}

// DeleteProductImage 删除商品图片
func (s *ProductImageService) DeleteProductImage(imageID uint) error {
	var image model.ProductImage
	if err := s.repo.GetDB().First(&image, imageID).Error; err != nil {
		return errors.New("图片不存在")
	}

	// 删除文件
	if image.URL != "" {
		filePath := "." + image.URL // URL 格式为 /product/xxx.jpg
		os.Remove(filePath)
	}

	return s.repo.GetDB().Delete(&image).Error
}

// SetPrimaryImage 设置主图
func (s *ProductImageService) SetPrimaryImage(productID, imageID uint) error {
	db := s.repo.GetDB()

	// 先取消所有主图
	db.Model(&model.ProductImage{}).Where("product_id = ?", productID).Update("is_primary", false)

	// 设置新主图
	return db.Model(&model.ProductImage{}).Where("id = ? AND product_id = ?", imageID, productID).Update("is_primary", true).Error
}

// UpdateImageOrder 更新图片排序
func (s *ProductImageService) UpdateImageOrder(productID uint, imageIDs []uint) error {
	db := s.repo.GetDB()

	for i, id := range imageIDs {
		if err := db.Model(&model.ProductImage{}).Where("id = ? AND product_id = ?", id, productID).Update("sort_order", i).Error; err != nil {
			return err
		}
	}

	return nil
}

// UploadProductImage 上传商品图片
func (s *ProductImageService) UploadProductImage(productID uint, filename string, file io.Reader, isPrimary bool) (*model.ProductImage, error) {
	// 验证文件扩展名
	ext := strings.ToLower(filepath.Ext(filename))
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	if !allowedExts[ext] {
		return nil, errors.New("不支持的图片格式，仅支持 jpg/jpeg/png/gif/webp")
	}

	// 创建目录
	uploadDir := fmt.Sprintf("./Product/%d", productID)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, errors.New("创建目录失败")
	}

	// 生成文件名
	newFilename := fmt.Sprintf("%d_%d%s", time.Now().UnixNano(), productID, ext)
	filePath := filepath.Join(uploadDir, newFilename)

	// 保存文件
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, errors.New("创建文件失败")
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return nil, errors.New("保存文件失败")
	}

	// 生成URL
	url := fmt.Sprintf("/product-files/%d/%s", productID, newFilename)

	// 添加到数据库
	return s.AddProductImage(productID, url, isPrimary)
}

// DeleteAllProductImages 删除商品的所有图片
func (s *ProductImageService) DeleteAllProductImages(productID uint) error {
	images, err := s.GetProductImages(productID)
	if err != nil {
		return err
	}

	for _, img := range images {
		if img.URL != "" {
			filePath := "." + img.URL
			os.Remove(filePath)
		}
	}

	return s.repo.GetDB().Where("product_id = ?", productID).Delete(&model.ProductImage{}).Error
}

// GetImageCount 获取商品图片数量
func (s *ProductImageService) GetImageCount(productID uint) int64 {
	var count int64
	s.repo.GetDB().Model(&model.ProductImage{}).Where("product_id = ?", productID).Count(&count)
	return count
}
