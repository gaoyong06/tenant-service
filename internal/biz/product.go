package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

// Product u4ea7u54c1u4fe1u606f
type Product struct {
	ProductCode string // u4ea7u54c1u4ee3u7801
	ProductName string // u4ea7u54c1u540du79f0
	Description string // u63cfu8ff0
}

// TenantProduct u79dfu6237u4ea7u54c1u5173u8054
type TenantProduct struct {
	ID          int64  // ID
	TenantID    string // u79dfu6237ID
	ProductCode string // u4ea7u54c1u4ee3u7801
}

// ProductRepo u4ea7u54c1u4ed3u50a8u63a5u53e3
type ProductRepo interface {
	CreateProduct(ctx context.Context, product *Product) (*Product, error)
	GetProduct(ctx context.Context, code string) (*Product, error)
	UpdateProduct(ctx context.Context, product *Product) (*Product, error)
	DeleteProduct(ctx context.Context, code string) error
	ListProducts(ctx context.Context) ([]*Product, error)
	ListProductsByTenant(ctx context.Context, tenantID string) ([]*Product, error)
	AssociateProductToTenant(ctx context.Context, tenantID, productCode string) error
	DisassociateProductFromTenant(ctx context.Context, tenantID, productCode string) error
}

// ProductUsecase u4ea7u54c1u7528u4f8b
type ProductUsecase struct {
	repo ProductRepo
	log  *log.Helper
}

// NewProductUsecase u521bu5efau4ea7u54c1u7528u4f8b
func NewProductUsecase(repo ProductRepo, logger log.Logger) *ProductUsecase {
	return &ProductUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

// ListProducts u5217u51fau4ea7u54c1
func (uc *ProductUsecase) ListProducts(ctx context.Context) ([]*Product, error) {
	uc.log.WithContext(ctx).Info("ListProducts")
	return uc.repo.ListProducts(ctx)
}

// ListProductsByTenant u6839u636eu79dfu6237u5217u51fau4ea7u54c1
func (uc *ProductUsecase) ListProductsByTenant(ctx context.Context, tenantID string) ([]*Product, error) {
	uc.log.WithContext(ctx).Infof("ListProductsByTenant: tenantID=%v", tenantID)
	return uc.repo.ListProductsByTenant(ctx, tenantID)
}
