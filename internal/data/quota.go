package data

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"tenant-service/internal/biz"
)

// QuotaModel 配额数据模型
type QuotaModel struct {
	QuotaID       int64     `gorm:"column:quota_id;primaryKey;autoIncrement"`
	TenantID      string    `gorm:"column:tenant_id;index;not null"`
	QuotaType     string    `gorm:"column:quota_type;not null"`
	LimitType     string    `gorm:"column:limit_type;not null"`
	HardLimit     int32     `gorm:"column:hard_limit;not null"`
	SoftLimit     int32     `gorm:"column:soft_limit"`
	UsedCount     int32     `gorm:"column:used_count;default:0"`
	ResetTime     time.Time `gorm:"column:reset_time"`
	NextResetTime time.Time `gorm:"column:next_reset_time;index"`
	EffectiveTime time.Time `gorm:"column:effective_time;not null"`
	ExpireTime    time.Time `gorm:"column:expire_time"`
	IsGlobal      bool      `gorm:"column:is_global;default:0;index"`
	ProductCodes  string    `gorm:"column:product_codes;type:json"`
	ExtraConfig   string    `gorm:"column:extra_config;type:json"`
	CreatedBy     string    `gorm:"column:created_by"`
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 表名
func (QuotaModel) TableName() string {
	return "tenant_quotas"
}

// QuotaUsageModel 配额使用记录数据模型
type QuotaUsageModel struct {
	RecordID      int64     `gorm:"column:record_id;primaryKey;autoIncrement"`
	QuotaID       int64     `gorm:"column:quota_id;index;not null"`
	TenantID      string    `gorm:"column:tenant_id;index;not null"`
	OperationType string    `gorm:"column:operation_type;not null"`
	DeltaValue    int32     `gorm:"column:delta_value;not null"`
	CurrentUsed   int32     `gorm:"column:current_used;not null"`
	BizID         string    `gorm:"column:biz_id;index"`
	BizType       string    `gorm:"column:biz_type;index"`
	Operator      string    `gorm:"column:operator"`
	OperationTime time.Time `gorm:"column:operation_time;autoCreateTime;index"`
	ExpireTime    time.Time `gorm:"column:expire_time"`
	Remark        string    `gorm:"column:remark"`
}

// TableName 表名
func (QuotaUsageModel) TableName() string {
	return "quota_usage_records"
}

// quotaRepo 配额仓库实现
type quotaRepo struct {
	data *Data
	log  *log.Helper
}

// NewQuotaRepo 创建配额仓库
func NewQuotaRepo(data *Data, logger log.Logger) biz.QuotaRepo {
	return &quotaRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// convertQuotaTypeToEnum 转换配额类型为枚举
func convertQuotaTypeToEnum(quotaType string) biz.QuotaType {
	switch quotaType {
	case "MARKETING_CAMPAIGN":
		return biz.QuotaTypeMarketingCampaign
	case "REDEEM_CODE":
		return biz.QuotaTypeRedeemCode
	case "SMS":
		return biz.QuotaTypeSMS
	default:
		return biz.QuotaTypeUnspecified
	}
}

// convertQuotaTypeToString 转换配额类型为字符串
func convertQuotaTypeToString(quotaType biz.QuotaType) string {
	switch quotaType {
	case biz.QuotaTypeMarketingCampaign:
		return "MARKETING_CAMPAIGN"
	case biz.QuotaTypeRedeemCode:
		return "REDEEM_CODE"
	case biz.QuotaTypeSMS:
		return "SMS"
	default:
		return "UNSPECIFIED"
	}
}

// convertLimitTypeToEnum 转换限制类型为枚举
func convertLimitTypeToEnum(limitType string) biz.LimitType {
	switch limitType {
	case "DAILY":
		return biz.LimitTypeDaily
	case "MONTHLY":
		return biz.LimitTypeMonthly
	case "TOTAL":
		return biz.LimitTypeTotal
	case "CONCURRENT":
		return biz.LimitTypeConcurrent
	default:
		return biz.LimitTypeUnspecified
	}
}

// convertLimitTypeToString 转换限制类型为字符串
func convertLimitTypeToString(limitType biz.LimitType) string {
	switch limitType {
	case biz.LimitTypeDaily:
		return "DAILY"
	case biz.LimitTypeMonthly:
		return "MONTHLY"
	case biz.LimitTypeTotal:
		return "TOTAL"
	case biz.LimitTypeConcurrent:
		return "CONCURRENT"
	default:
		return "UNSPECIFIED"
	}
}

// convertOperationTypeToEnum 转换操作类型为枚举
func convertOperationTypeToEnum(operationType string) biz.OperationType {
	switch operationType {
	case "CONSUME":
		return biz.OperationTypeConsume
	case "RELEASE":
		return biz.OperationTypeRelease
	case "ADJUST":
		return biz.OperationTypeAdjust
	default:
		return biz.OperationTypeUnspecified
	}
}

// convertOperationTypeToString 转换操作类型为字符串
func convertOperationTypeToString(operationType biz.OperationType) string {
	switch operationType {
	case biz.OperationTypeConsume:
		return "CONSUME"
	case biz.OperationTypeRelease:
		return "RELEASE"
	case biz.OperationTypeAdjust:
		return "ADJUST"
	default:
		return "UNSPECIFIED"
	}
}

// convertModelToBiz 转换数据模型到业务模型
func (r *quotaRepo) convertModelToBiz(model *QuotaModel) (*biz.QuotaInfo, error) {
	if model == nil {
		return nil, nil
	}

	// 解析产品代码JSON
	var productCodes []string
	if model.ProductCodes != "" {
		if err := json.Unmarshal([]byte(model.ProductCodes), &productCodes); err != nil {
			return nil, err
		}
	}

	return &biz.QuotaInfo{
		QuotaID:       model.QuotaID,
		TenantID:      model.TenantID,
		QuotaType:     convertQuotaTypeToEnum(model.QuotaType),
		LimitType:     convertLimitTypeToEnum(model.LimitType),
		HardLimit:     model.HardLimit,
		SoftLimit:     model.SoftLimit,
		UsedCount:     model.UsedCount,
		ResetTime:     model.ResetTime,
		NextResetTime: model.NextResetTime,
		EffectiveTime: model.EffectiveTime,
		ExpireTime:    model.ExpireTime,
		IsGlobal:      model.IsGlobal,
		ProductCodes:  productCodes,
		ExtraConfig:   model.ExtraConfig,
	}, nil
}

// CreateQuota 创建配额
func (r *quotaRepo) CreateQuota(ctx context.Context, quota *biz.QuotaInfo) (*biz.QuotaInfo, error) {
	// 序列化产品代码为JSON
	productCodesJSON, err := json.Marshal(quota.ProductCodes)
	if err != nil {
		return nil, err
	}

	// 创建配额模型
	model := &QuotaModel{
		TenantID:      quota.TenantID,
		QuotaType:     convertQuotaTypeToString(quota.QuotaType),
		LimitType:     convertLimitTypeToString(quota.LimitType),
		HardLimit:     quota.HardLimit,
		SoftLimit:     quota.SoftLimit,
		UsedCount:     quota.UsedCount,
		ResetTime:     quota.ResetTime,
		NextResetTime: quota.NextResetTime,
		EffectiveTime: quota.EffectiveTime,
		ExpireTime:    quota.ExpireTime,
		IsGlobal:      quota.IsGlobal,
		ProductCodes:  string(productCodesJSON),
		ExtraConfig:   quota.ExtraConfig,
	}

	// 创建配额记录
	if err := r.data.db.Create(model).Error; err != nil {
		return nil, err
	}

	return r.convertModelToBiz(model)
}

// GetQuota 获取配额
func (r *quotaRepo) GetQuota(ctx context.Context, tenantID string, quotaType biz.QuotaType, limitType biz.LimitType, productCode string) (*biz.QuotaInfo, error) {
	var model QuotaModel

	// 构建查询条件
	query := r.data.db.Where("tenant_id = ? AND quota_type = ? AND limit_type = ?",
		tenantID, convertQuotaTypeToString(quotaType), convertLimitTypeToString(limitType))

	// 如果指定了产品代码，则添加产品代码条件
	if productCode != "" {
		query = query.Where("JSON_CONTAINS(product_codes, ?)", fmt.Sprintf("\"%s\"", productCode))
	}

	// 查询配额
	err := query.First(&model).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 尝试查找全局默认配额
			err = r.data.db.Where("is_global = ? AND quota_type = ? AND limit_type = ?",
				true, convertQuotaTypeToString(quotaType), convertLimitTypeToString(limitType)).First(&model).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					return nil, nil
				}
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return r.convertModelToBiz(&model)
}

// UpdateQuota 更新配额
func (r *quotaRepo) UpdateQuota(ctx context.Context, quota *biz.QuotaInfo) (*biz.QuotaInfo, error) {
	// 查询配额是否存在
	var model QuotaModel
	err := r.data.db.Where("quota_id = ?", quota.QuotaID).First(&model).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("quota not found: %d", quota.QuotaID)
		}
		return nil, err
	}

	// 序列化产品代码为JSON
	productCodesJSON, err := json.Marshal(quota.ProductCodes)
	if err != nil {
		return nil, err
	}

	// 更新配额信息
	model.HardLimit = quota.HardLimit
	model.SoftLimit = quota.SoftLimit
	model.UsedCount = quota.UsedCount
	model.ResetTime = quota.ResetTime
	model.NextResetTime = quota.NextResetTime
	model.EffectiveTime = quota.EffectiveTime
	model.ExpireTime = quota.ExpireTime
	model.IsGlobal = quota.IsGlobal
	model.ProductCodes = string(productCodesJSON)
	model.ExtraConfig = quota.ExtraConfig

	err = r.data.db.Save(&model).Error
	if err != nil {
		return nil, err
	}

	return r.convertModelToBiz(&model)
}

// DeleteQuota 删除配额
func (r *quotaRepo) DeleteQuota(ctx context.Context, quotaID int64) error {
	return r.data.db.Where("quota_id = ?", quotaID).Delete(&QuotaModel{}).Error
}

// ListQuotas 列出配额
func (r *quotaRepo) ListQuotas(ctx context.Context, tenantID string, quotaType biz.QuotaType) ([]*biz.QuotaInfo, error) {
	var models []*QuotaModel

	query := r.data.db.Model(&QuotaModel{})

	// 添加查询条件
	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}

	if quotaType != biz.QuotaTypeUnspecified {
		query = query.Where("quota_type = ?", convertQuotaTypeToString(quotaType))
	}

	// 查询配额列表
	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	// 转换为业务模型
	quotas := make([]*biz.QuotaInfo, 0, len(models))
	for _, model := range models {
		quota, err := r.convertModelToBiz(model)
		if err != nil {
			return nil, err
		}
		quotas = append(quotas, quota)
	}

	return quotas, nil
}

// ConsumeQuota 消费配额
func (r *quotaRepo) ConsumeQuota(ctx context.Context, tenantID string, quotaType biz.QuotaType, limitType biz.LimitType, amount int32, productCode, bizID, bizType string) (bool, int32, error) {
	// 开启事务
	var success bool
	var remainingQuota int32

	err := r.data.db.Transaction(func(tx *gorm.DB) error {
		// 查询配额并锁定
		var model QuotaModel
		query := tx.Where("tenant_id = ? AND quota_type = ? AND limit_type = ?",
			tenantID, convertQuotaTypeToString(quotaType), convertLimitTypeToString(limitType)).Clauses(clause.Locking{Strength: "UPDATE"})

		// 如果指定了产品代码，则添加产品代码条件
		if productCode != "" {
			query = query.Where("JSON_CONTAINS(product_codes, ?)", fmt.Sprintf("\"%s\"", productCode))
		}

		// 查询配额
		err := query.First(&model).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				// 尝试查找全局默认配额
				err = tx.Where("is_global = ? AND quota_type = ? AND limit_type = ?",
					true, convertQuotaTypeToString(quotaType), convertLimitTypeToString(limitType)).Clauses(clause.Locking{Strength: "UPDATE"}).First(&model).Error
				if err != nil {
					if err == gorm.ErrRecordNotFound {
						success = false
						remainingQuota = 0
						return fmt.Errorf("quota not found")
					}
					return err
				}
			} else {
				return err
			}
		}

		// 检查配额是否足够
		if model.UsedCount+amount > model.HardLimit {
			success = false
			remainingQuota = model.HardLimit - model.UsedCount
			return fmt.Errorf("quota exceeded: %d/%d", model.UsedCount, model.HardLimit)
		}

		// 更新使用量
		model.UsedCount += amount
		if err := tx.Save(&model).Error; err != nil {
			return err
		}

		// 记录使用记录
		usageRecord := &QuotaUsageModel{
			QuotaID:       model.QuotaID,
			TenantID:      tenantID,
			OperationType: convertOperationTypeToString(biz.OperationTypeConsume),
			DeltaValue:    amount,
			CurrentUsed:   model.UsedCount,
			BizID:         bizID,
			BizType:       bizType,
		}

		if err := tx.Create(usageRecord).Error; err != nil {
			return err
		}

		success = true
		remainingQuota = model.HardLimit - model.UsedCount
		return nil
	})

	return success, remainingQuota, err
}

// ReleaseQuota 释放配额
func (r *quotaRepo) ReleaseQuota(ctx context.Context, tenantID string, quotaType biz.QuotaType, limitType biz.LimitType, amount int32, productCode, bizID string) (bool, int32, error) {
	// 开启事务
	var success bool
	var remainingQuota int32

	err := r.data.db.Transaction(func(tx *gorm.DB) error {
		// 查询配额并锁定
		var model QuotaModel
		query := tx.Where("tenant_id = ? AND quota_type = ? AND limit_type = ?",
			tenantID, convertQuotaTypeToString(quotaType), convertLimitTypeToString(limitType)).Clauses(clause.Locking{Strength: "UPDATE"})

		// 如果指定了产品代码，则添加产品代码条件
		if productCode != "" {
			query = query.Where("JSON_CONTAINS(product_codes, ?)", fmt.Sprintf("\"%s\"", productCode))
		}

		// 查询配额
		err := query.First(&model).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				success = false
				remainingQuota = 0
				return fmt.Errorf("quota not found")
			}
			return err
		}

		// 更新使用量（不能小于0）
		if model.UsedCount < amount {
			model.UsedCount = 0
		} else {
			model.UsedCount -= amount
		}

		if err := tx.Save(&model).Error; err != nil {
			return err
		}

		// 记录使用记录
		usageRecord := &QuotaUsageModel{
			QuotaID:       model.QuotaID,
			TenantID:      tenantID,
			OperationType: convertOperationTypeToString(biz.OperationTypeRelease),
			DeltaValue:    -amount, // 负数表示释放
			CurrentUsed:   model.UsedCount,
			BizID:         bizID,
		}

		if err := tx.Create(usageRecord).Error; err != nil {
			return err
		}

		success = true
		remainingQuota = model.HardLimit - model.UsedCount
		return nil
	})

	return success, remainingQuota, err
}

// ResetQuotas 重置配额
func (r *quotaRepo) ResetQuotas(ctx context.Context, limitType biz.LimitType) error {
	// 查询需要重置的配额
	var models []*QuotaModel
	err := r.data.db.Where("limit_type = ? AND next_reset_time <= ?",
		convertLimitTypeToString(limitType), time.Now()).Find(&models).Error
	if err != nil {
		return err
	}

	// 开启事务
	return r.data.db.Transaction(func(tx *gorm.DB) error {
		for _, model := range models {
			// 重置使用量
			model.UsedCount = 0
			model.ResetTime = time.Now()

			// 计算下次重置时间
			switch limitType {
			case biz.LimitTypeDaily:
				model.NextResetTime = time.Now().AddDate(0, 0, 1)
			case biz.LimitTypeMonthly:
				model.NextResetTime = time.Now().AddDate(0, 1, 0)
			}

			// 保存更新
			if err := tx.Save(model).Error; err != nil {
				return err
			}

			// 记录重置操作
			usageRecord := &QuotaUsageModel{
				QuotaID:       model.QuotaID,
				TenantID:      model.TenantID,
				OperationType: "RESET",
				DeltaValue:    -model.UsedCount, // 负数表示重置
				CurrentUsed:   0,
				Remark:        fmt.Sprintf("Scheduled reset for %s quota", convertLimitTypeToString(limitType)),
			}

			if err := tx.Create(usageRecord).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
