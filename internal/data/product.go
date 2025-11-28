package data

import (
	"context"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
	"tenant-service/internal/biz"
)

// ProductModel 产品数据模型
type ProductModel struct {
	ProductCode string `gorm:"column:product_code;primaryKey"`
	ProductName string `gorm:"column:product_name;not null"`
	Description string `gorm:"column:description"`
}

// TableName 表名
func (ProductModel) TableName() string {
	return "products"
}

// TenantProductModel 租户产品关联数据模型
type TenantProductModel struct {
	ID          int64  `gorm:"column:id;primaryKey;autoIncrement"`
	TenantID    string `gorm:"column:tenant_id;uniqueIndex:idx_tenant_product;not null"`
	ProductCode string `gorm:"column:product_code;uniqueIndex:idx_tenant_product;not null"`
}

// TableName 表名
func (TenantProductModel) TableName() string {
	return "tenant_products"
}

// productRepo 产品仓库实现
type productRepo struct {
	data *Data
	log  *log.Helper
}

// NewProductRepo 创建产品仓库
func NewProductRepo(data *Data, logger log.Logger) biz.ProductRepo {
	return &productRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// convertModelToBiz 转换数据模型到业务模型
func (r *productRepo) convertModelToBiz(model *ProductModel) *biz.Product {
	if model == nil {
		return nil
	}

	return &biz.Product{
		ProductCode: model.ProductCode,
		ProductName: model.ProductName,
		Description: model.Description,
	}
}

// CreateProduct 创建产品
func (r *productRepo) CreateProduct(ctx context.Context, product *biz.Product) (*biz.Product, error) {
	// 创建产品模型
	model := &ProductModel{
		ProductCode: product.ProductCode,
		ProductName: product.ProductName,
		Description: product.Description,
	}

	// 创建产品记录
	if err := r.data.db.Create(model).Error; err != nil {
		return nil, err
	}

	return r.convertModelToBiz(model), nil
}

// GetProduct 获取产品
func (r *productRepo) GetProduct(ctx context.Context, code string) (*biz.Product, error) {
	var model ProductModel
	err := r.data.db.Where("product_code = ?", code).First(&model).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return r.convertModelToBiz(&model), nil
}

// UpdateProduct 更新产品
func (r *productRepo) UpdateProduct(ctx context.Context, product *biz.Product) (*biz.Product, error) {
	// 查询产品是否存在
	var model ProductModel
	err := r.data.db.Where("product_code = ?", product.ProductCode).First(&model).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("product not found: %s", product.ProductCode)
		}
		return nil, err
	}

	// 更新产品信息
	model.ProductName = product.ProductName
	model.Description = product.Description

	err = r.data.db.Save(&model).Error
	if err != nil {
		return nil, err
	}

	return r.convertModelToBiz(&model), nil
}

// DeleteProduct 删除产品
func (r *productRepo) DeleteProduct(ctx context.Context, code string) error {
	// 开启事务
	return r.data.db.Transaction(func(tx *gorm.DB) error {
		// 删除租户产品关联
		if err := tx.Where("product_code = ?", code).Delete(&TenantProductModel{}).Error; err != nil {
			return err
		}

		// 删除产品
		if err := tx.Where("product_code = ?", code).Delete(&ProductModel{}).Error; err != nil {
			return err
		}

		return nil
	})
}

// ListProducts 列出产品
func (r *productRepo) ListProducts(ctx context.Context) ([]*biz.Product, error) {
	var models []*ProductModel

	// 查询产品列表
	if err := r.data.db.Find(&models).Error; err != nil {
		return nil, err
	}

	// 转换为业务模型
	products := make([]*biz.Product, 0, len(models))
	for _, model := range models {
		products = append(products, r.convertModelToBiz(model))
	}

	return products, nil
}

// ListProductsByTenant 根据租户列出产品
func (r *productRepo) ListProductsByTenant(ctx context.Context, tenantID string) ([]*biz.Product, error) {
	var products []*biz.Product

	// 查询租户关联的产品
	err := r.data.db.Raw(
		`SELECT p.product_code, p.product_name, p.description 
		FROM products p 
		JOIN tenant_products tp ON p.product_code = tp.product_code 
		WHERE tp.tenant_id = ?`,
		tenantID,
	).Scan(&products).Error

	if err != nil {
		return nil, err
	}

	return products, nil
}

// AssociateProductToTenant 关联产品到租户
func (r *productRepo) AssociateProductToTenant(ctx context.Context, tenantID, productCode string) error {
	// 检查产品是否存在
	var productCount int64
	err := r.data.db.Model(&ProductModel{}).Where("product_code = ?", productCode).Count(&productCount).Error
	if err != nil {
		return err
	}

	if productCount == 0 {
		return fmt.Errorf("product not found: %s", productCode)
	}

	// 检查租户是否存在
	var tenantCount int64
	err = r.data.db.Model(&TenantModel{}).Where("tenant_id = ?", tenantID).Count(&tenantCount).Error
	if err != nil {
		return err
	}

	if tenantCount == 0 {
		return fmt.Errorf("tenant not found: %s", tenantID)
	}

	// 创建关联
	association := &TenantProductModel{
		TenantID:    tenantID,
		ProductCode: productCode,
	}

	// 检查是否已存在关联
	var count int64
	err = r.data.db.Model(&TenantProductModel{}).Where("tenant_id = ? AND product_code = ?", tenantID, productCode).Count(&count).Error
	if err != nil {
		return err
	}

	if count > 0 {
		return nil // 已存在关联，不需要重复创建
	}

	return r.data.db.Create(association).Error
}

// DisassociateProductFromTenant 解除产品与租户的关联
func (r *productRepo) DisassociateProductFromTenant(ctx context.Context, tenantID, productCode string) error {
	return r.data.db.Where("tenant_id = ? AND product_code = ?", tenantID, productCode).Delete(&TenantProductModel{}).Error
}
