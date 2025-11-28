package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// TenantType 租户类型
type TenantType int32

const (
	TenantTypeUnspecified TenantType = 0
	TenantTypePlatform    TenantType = 1 // 平台
	TenantTypeChannel     TenantType = 2 // 渠道
	TenantTypeEnterprise  TenantType = 3 // 企业
)

// Tenant 租户领域模型
type Tenant struct {
	TenantID       string            // 租户ID
	TenantName     string            // 租户名称
	TenantType     TenantType        // 租户类型
	ParentTenantID string            // 父租户ID
	Status         bool              // 状态
	QuotaConfig    map[string]string // 配额配置
	CreatedAt      time.Time         // 创建时间
	UpdatedAt      time.Time         // 更新时间
}

// TenantRepo 租户仓储接口
type TenantRepo interface {
	Create(ctx context.Context, tenant *Tenant) (*Tenant, error)
	Get(ctx context.Context, id string) (*Tenant, error)
	Update(ctx context.Context, tenant *Tenant) (*Tenant, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, tenantType TenantType, parentID string, status bool, pageNum, pageSize int32) ([]*Tenant, int32, error)
}

// TenantUsecase 租户用例
type TenantUsecase struct {
	repo TenantRepo
	log  *log.Helper
}

// NewTenantUsecase 创建租户用例
func NewTenantUsecase(repo TenantRepo, logger log.Logger) *TenantUsecase {
	return &TenantUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

// CreateTenant 创建租户
func (uc *TenantUsecase) CreateTenant(ctx context.Context, tenant *Tenant) (*Tenant, error) {
	uc.log.WithContext(ctx).Infof("CreateTenant: %v", tenant.TenantName)
	return uc.repo.Create(ctx, tenant)
}

// GetTenant 获取租户
func (uc *TenantUsecase) GetTenant(ctx context.Context, id string) (*Tenant, error) {
	uc.log.WithContext(ctx).Infof("GetTenant: %v", id)
	return uc.repo.Get(ctx, id)
}

// UpdateTenant 更新租户
func (uc *TenantUsecase) UpdateTenant(ctx context.Context, tenant *Tenant) (*Tenant, error) {
	uc.log.WithContext(ctx).Infof("UpdateTenant: %v", tenant.TenantID)
	return uc.repo.Update(ctx, tenant)
}

// DeleteTenant 删除租户
func (uc *TenantUsecase) DeleteTenant(ctx context.Context, id string) error {
	uc.log.WithContext(ctx).Infof("DeleteTenant: %v", id)
	return uc.repo.Delete(ctx, id)
}

// ListTenants 列出租户
func (uc *TenantUsecase) ListTenants(ctx context.Context, tenantType TenantType, parentID string, status bool, pageNum, pageSize int32) ([]*Tenant, int32, error) {
	uc.log.WithContext(ctx).Infof("ListTenants: type=%v, parentID=%v", tenantType, parentID)
	return uc.repo.List(ctx, tenantType, parentID, status, pageNum, pageSize)
}
