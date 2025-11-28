package data

import (
	"context"
	"fmt"
	"time"

	"github.com/gaoyong06/middleground/tenant-service/internal/biz"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TenantModel 租户数据模型
type TenantModel struct {
	TenantID       string    `gorm:"column:tenant_id;primaryKey"`
	TenantName     string    `gorm:"column:tenant_name;not null"`
	TenantType     string    `gorm:"column:tenant_type;not null"`
	ParentTenantID string    `gorm:"column:parent_tenant_id"`
	Status         bool      `gorm:"column:status;default:1"`
	QuotaConfig    string    `gorm:"column:quota_config;type:json"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt      time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 表名
func (TenantModel) TableName() string {
	return "tenants"
}

// ChannelModel 渠道扩展数据模型
type ChannelModel struct {
	ChannelID     int64     `gorm:"column:channel_id;primaryKey;autoIncrement"`
	TenantID      string    `gorm:"column:tenant_id;uniqueIndex;not null"`
	ChannelCode   string    `gorm:"column:channel_code;uniqueIndex;not null"`
	ChannelName   string    `gorm:"column:channel_name;not null"`
	ContactName   string    `gorm:"column:contact_name"`
	ContactPhone  string    `gorm:"column:contact_phone"`
	CommissionRate float64   `gorm:"column:commission_rate"`
	SalesTarget   float64   `gorm:"column:sales_target"`
	ExtraData     string    `gorm:"column:extra_data;type:json"`
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 表名
func (ChannelModel) TableName() string {
	return "channels"
}

// tenantRepo 租户仓储实现
type tenantRepo struct {
	data *Data
	log  *log.Helper
}

// NewTenantRepo 创建租户仓储
func NewTenantRepo(data *Data, logger log.Logger) biz.TenantRepo {
	return &tenantRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// convertTenantTypeToEnum 转换租户类型为枚举
func convertTenantTypeToEnum(tenantType string) biz.TenantType {
	switch tenantType {
	case "PLATFORM":
		return biz.TenantTypePlatform
	case "CHANNEL":
		return biz.TenantTypeChannel
	case "ENTERPRISE":
		return biz.TenantTypeEnterprise
	default:
		return biz.TenantTypeUnspecified
	}
}

// convertTenantTypeToString 转换租户类型为字符串
func convertTenantTypeToString(tenantType biz.TenantType) string {
	switch tenantType {
	case biz.TenantTypePlatform:
		return "PLATFORM"
	case biz.TenantTypeChannel:
		return "CHANNEL"
	case biz.TenantTypeEnterprise:
		return "ENTERPRISE"
	default:
		return "UNSPECIFIED"
	}
}

// convertModelToBiz 转换数据模型到业务模型
func (r *tenantRepo) convertModelToBiz(model *TenantModel) (*biz.Tenant, error) {
	if model == nil {
		return nil, nil
	}

	// 解析配额配置JSON
	quotaConfig := make(map[string]string)
	// 这里应该实现JSON解析，简化处理

	return &biz.Tenant{
		TenantID:       model.TenantID,
		TenantName:     model.TenantName,
		TenantType:     convertTenantTypeToEnum(model.TenantType),
		ParentTenantID: model.ParentTenantID,
		Status:         model.Status,
		QuotaConfig:    quotaConfig,
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      model.UpdatedAt,
	}, nil
}

// Create 创建租户
func (r *tenantRepo) Create(ctx context.Context, tenant *biz.Tenant) (*biz.Tenant, error) {
	// 生成租户ID
	tenantID := fmt.Sprintf("TN_%s", uuid.New().String()[:8])
	if tenant.TenantType == biz.TenantTypeChannel {
		tenantID = fmt.Sprintf("CH_%s", uuid.New().String()[:8])
	} else if tenant.TenantType == biz.TenantTypeEnterprise {
		tenantID = fmt.Sprintf("EN_%s", uuid.New().String()[:8])
	}

	// 创建租户模型
	model := &TenantModel{
		TenantID:       tenantID,
		TenantName:     tenant.TenantName,
		TenantType:     convertTenantTypeToString(tenant.TenantType),
		ParentTenantID: tenant.ParentTenantID,
		Status:         tenant.Status,
		QuotaConfig:    "", // 应该序列化为JSON
	}

	// 开启事务
	err := r.data.db.Transaction(func(tx *gorm.DB) error {
		// 创建租户记录
		if err := tx.Create(model).Error; err != nil {
			return err
		}

		// 如果是渠道类型，创建渠道扩展信息
		if tenant.TenantType == biz.TenantTypeChannel {
			channel := &ChannelModel{
				TenantID:    tenantID,
				ChannelCode: fmt.Sprintf("CH%s", uuid.New().String()[:6]),
				ChannelName: tenant.TenantName,
			}
			if err := tx.Create(channel).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 查询创建的租户
	return r.Get(ctx, tenantID)
}

// Get 获取租户
func (r *tenantRepo) Get(ctx context.Context, id string) (*biz.Tenant, error) {
	var model TenantModel
	err := r.data.db.Where("tenant_id = ?", id).First(&model).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return r.convertModelToBiz(&model)
}

// Update 更新租户
func (r *tenantRepo) Update(ctx context.Context, tenant *biz.Tenant) (*biz.Tenant, error) {
	// 查询租户是否存在
	var model TenantModel
	err := r.data.db.Where("tenant_id = ?", tenant.TenantID).First(&model).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("tenant not found: %s", tenant.TenantID)
		}
		return nil, err
	}

	// 更新租户信息
	model.TenantName = tenant.TenantName
	model.Status = tenant.Status
	model.QuotaConfig = "" // 应该序列化为JSON

	err = r.data.db.Save(&model).Error
	if err != nil {
		return nil, err
	}

	return r.convertModelToBiz(&model)
}

// Delete 删除租户
func (r *tenantRepo) Delete(ctx context.Context, id string) error {
	// 开启事务
	return r.data.db.Transaction(func(tx *gorm.DB) error {
		// 删除渠道扩展信息
		if err := tx.Where("tenant_id = ?", id).Delete(&ChannelModel{}).Error; err != nil {
			return err
		}

		// 删除租户
		if err := tx.Where("tenant_id = ?", id).Delete(&TenantModel{}).Error; err != nil {
			return err
		}

		return nil
	})
}

// List 列出租户
func (r *tenantRepo) List(ctx context.Context, tenantType biz.TenantType, parentID string, status bool, pageNum, pageSize int32) ([]*biz.Tenant, int32, error) {
	var models []*TenantModel
	var count int64

	query := r.data.db.Model(&TenantModel{})

	// 添加查询条件
	if tenantType != biz.TenantTypeUnspecified {
		query = query.Where("tenant_type = ?", convertTenantTypeToString(tenantType))
	}

	if parentID != "" {
		query = query.Where("parent_tenant_id = ?", parentID)
	}

	// 查询总数
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (pageNum - 1) * pageSize
	if err := query.Offset(int(offset)).Limit(int(pageSize)).Find(&models).Error; err != nil {
		return nil, 0, err
	}

	// 转换为业务模型
	tenants := make([]*biz.Tenant, 0, len(models))
	for _, model := range models {
		tenant, err := r.convertModelToBiz(model)
		if err != nil {
			return nil, 0, err
		}
		tenants = append(tenants, tenant)
	}

	return tenants, int32(count), nil
}
