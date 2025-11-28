package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// QuotaType 配额类型
type QuotaType int32

const (
	QuotaTypeUnspecified       QuotaType = 0
	QuotaTypeMarketingCampaign QuotaType = 1 // 营销活动
	QuotaTypeRedeemCode        QuotaType = 2 // 兑换码
	QuotaTypeSMS               QuotaType = 3 // 短信
)

// LimitType 限制类型
type LimitType int32

const (
	LimitTypeUnspecified LimitType = 0
	LimitTypeDaily       LimitType = 1 // 每日
	LimitTypeMonthly     LimitType = 2 // 每月
	LimitTypeTotal       LimitType = 3 // 总量
	LimitTypeConcurrent  LimitType = 4 // 并发
)

// OperationType 操作类型
type OperationType int32

const (
	OperationTypeUnspecified OperationType = 0
	OperationTypeConsume     OperationType = 1 // 消费
	OperationTypeRelease     OperationType = 2 // 释放
	OperationTypeAdjust      OperationType = 3 // 调整
)

// QuotaInfo 配额信息
type QuotaInfo struct {
	QuotaID       int64      // 配额ID
	TenantID      string     // 租户ID
	QuotaType     QuotaType  // 配额类型
	LimitType     LimitType  // 限制类型
	HardLimit     int32      // 硬限制
	SoftLimit     int32      // 软限制
	UsedCount     int32      // 已使用数量
	ResetTime     time.Time  // 重置时间
	NextResetTime time.Time  // 下次重置时间
	EffectiveTime time.Time  // 生效时间
	ExpireTime    time.Time  // 过期时间
	IsGlobal      bool       // 是否全局
	ProductCodes  []string   // 产品代码列表
	ExtraConfig   string     // 额外配置
}

// QuotaUsageRecord 配额使用记录
type QuotaUsageRecord struct {
	RecordID      int64          // 记录ID
	QuotaID       int64          // 配额ID
	TenantID      string         // 租户ID
	OperationType OperationType  // 操作类型
	DeltaValue    int32          // 变更数值
	CurrentUsed   int32          // 当前已用量
	BizID         string         // 业务ID
	BizType       string         // 业务类型
	Operator      string         // 操作人
	OperationTime time.Time      // 操作时间
	ExpireTime    time.Time      // 过期时间
	Remark        string         // 备注
}

// QuotaRepo 配额仓储接口
type QuotaRepo interface {
	CreateQuota(ctx context.Context, quota *QuotaInfo) (*QuotaInfo, error)
	GetQuota(ctx context.Context, tenantID string, quotaType QuotaType, limitType LimitType, productCode string) (*QuotaInfo, error)
	UpdateQuota(ctx context.Context, quota *QuotaInfo) (*QuotaInfo, error)
	DeleteQuota(ctx context.Context, quotaID int64) error
	ListQuotas(ctx context.Context, tenantID string, quotaType QuotaType) ([]*QuotaInfo, error)
	ConsumeQuota(ctx context.Context, tenantID string, quotaType QuotaType, limitType LimitType, amount int32, productCode, bizID, bizType string) (bool, int32, error)
	ReleaseQuota(ctx context.Context, tenantID string, quotaType QuotaType, limitType LimitType, amount int32, productCode, bizID string) (bool, int32, error)
	ResetQuotas(ctx context.Context, limitType LimitType) error
}

// QuotaUsecase 配额用例
type QuotaUsecase struct {
	repo QuotaRepo
	log  *log.Helper
}

// NewQuotaUsecase 创建配额用例
func NewQuotaUsecase(repo QuotaRepo, logger log.Logger) *QuotaUsecase {
	return &QuotaUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

// CheckQuota 检查配额
func (uc *QuotaUsecase) CheckQuota(ctx context.Context, tenantID string, quotaType QuotaType, limitType LimitType, productCode string) (*QuotaInfo, bool, int32, error) {
	uc.log.WithContext(ctx).Infof("CheckQuota: tenantID=%v, quotaType=%v, limitType=%v", tenantID, quotaType, limitType)
	
	quota, err := uc.repo.GetQuota(ctx, tenantID, quotaType, limitType, productCode)
	if err != nil {
		return nil, false, 0, err
	}
	
	if quota == nil {
		return nil, false, 0, nil
	}
	
	available := quota.HardLimit - quota.UsedCount
	hasQuota := available > 0
	
	return quota, hasQuota, available, nil
}

// ConsumeQuota 消费配额
func (uc *QuotaUsecase) ConsumeQuota(ctx context.Context, tenantID string, quotaType QuotaType, limitType LimitType, amount int32, productCode, bizID, bizType string) (bool, int32, error) {
	uc.log.WithContext(ctx).Infof("ConsumeQuota: tenantID=%v, quotaType=%v, limitType=%v, amount=%v", tenantID, quotaType, limitType, amount)
	return uc.repo.ConsumeQuota(ctx, tenantID, quotaType, limitType, amount, productCode, bizID, bizType)
}

// ReleaseQuota 释放配额
func (uc *QuotaUsecase) ReleaseQuota(ctx context.Context, tenantID string, quotaType QuotaType, limitType LimitType, amount int32, productCode, bizID string) (bool, int32, error) {
	uc.log.WithContext(ctx).Infof("ReleaseQuota: tenantID=%v, quotaType=%v, limitType=%v, amount=%v", tenantID, quotaType, limitType, amount)
	return uc.repo.ReleaseQuota(ctx, tenantID, quotaType, limitType, amount, productCode, bizID)
}

// ResetQuotas 重置配额
func (uc *QuotaUsecase) ResetQuotas(ctx context.Context, limitType LimitType) error {
	uc.log.WithContext(ctx).Infof("ResetQuotas: limitType=%v", limitType)
	return uc.repo.ResetQuotas(ctx, limitType)
}
