package service

import (
	"context"
	"time"

	pb "github.com/gaoyong06/middleground/proto-repo/gen/go/platform/tenant_service/v1"
	"github.com/gaoyong06/middleground/tenant-service/internal/biz"
	"github.com/go-kratos/kratos/v2/log"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// TenantService is a tenant service.
type TenantService struct {
	pb.UnimplementedTenantServiceServer

	tu  *biz.TenantUsecase
	qu  *biz.QuotaUsecase
	pu  *biz.ProductUsecase
	log *log.Helper
}

// NewTenantService new a tenant service.
func NewTenantService(tu *biz.TenantUsecase, qu *biz.QuotaUsecase, pu *biz.ProductUsecase, logger log.Logger) *TenantService {
	return &TenantService{
		tu:  tu,
		qu:  qu,
		pu:  pu,
		log: log.NewHelper(logger),
	}
}

// convertTenantTypeToEnum converts tenant type from proto to biz enum
func convertTenantTypeToEnum(tenantType pb.TenantType) biz.TenantType {
	switch tenantType {
	case pb.TenantType_TENANT_TYPE_PLATFORM:
		return biz.TenantTypePlatform
	case pb.TenantType_TENANT_TYPE_CHANNEL:
		return biz.TenantTypeChannel
	case pb.TenantType_TENANT_TYPE_ENTERPRISE:
		return biz.TenantTypeEnterprise
	default:
		return biz.TenantTypeUnspecified
	}
}

// convertTenantTypeToProto converts tenant type from biz enum to proto
func convertTenantTypeToProto(tenantType biz.TenantType) pb.TenantType {
	switch tenantType {
	case biz.TenantTypePlatform:
		return pb.TenantType_TENANT_TYPE_PLATFORM
	case biz.TenantTypeChannel:
		return pb.TenantType_TENANT_TYPE_CHANNEL
	case biz.TenantTypeEnterprise:
		return pb.TenantType_TENANT_TYPE_ENTERPRISE
	default:
		return pb.TenantType_TENANT_TYPE_UNSPECIFIED
	}
}

// convertQuotaTypeToEnum converts quota type from proto to biz enum
func convertQuotaTypeToEnum(quotaType pb.QuotaType) biz.QuotaType {
	switch quotaType {
	case pb.QuotaType_QUOTA_TYPE_MARKETING_CAMPAIGN:
		return biz.QuotaTypeMarketingCampaign
	case pb.QuotaType_QUOTA_TYPE_REDEEM_CODE:
		return biz.QuotaTypeRedeemCode
	case pb.QuotaType_QUOTA_TYPE_SMS:
		return biz.QuotaTypeSMS
	default:
		return biz.QuotaTypeUnspecified
	}
}

// convertLimitTypeToEnum converts limit type from proto to biz enum
func convertLimitTypeToEnum(limitType pb.LimitType) biz.LimitType {
	switch limitType {
	case pb.LimitType_LIMIT_TYPE_DAILY:
		return biz.LimitTypeDaily
	case pb.LimitType_LIMIT_TYPE_MONTHLY:
		return biz.LimitTypeMonthly
	case pb.LimitType_LIMIT_TYPE_TOTAL:
		return biz.LimitTypeTotal
	case pb.LimitType_LIMIT_TYPE_CONCURRENT:
		return biz.LimitTypeConcurrent
	default:
		return biz.LimitTypeUnspecified
	}
}

// convertTenantToPB converts tenant from biz to proto
func convertTenantToPB(tenant *biz.Tenant) *pb.Tenant {
	if tenant == nil {
		return nil
	}

	return &pb.Tenant{
		TenantId:       tenant.TenantID,
		TenantName:     tenant.TenantName,
		TenantType:     convertTenantTypeToProto(tenant.TenantType),
		ParentTenantId: tenant.ParentTenantID,
		Status:         tenant.Status,
		QuotaConfig:    tenant.QuotaConfig,
		CreatedAt:      tenant.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      tenant.UpdatedAt.Format(time.RFC3339),
	}
}

// convertQuotaInfoToPB converts quota info from biz to proto
func convertQuotaInfoToPB(quota *biz.QuotaInfo) *pb.QuotaInfo {
	if quota == nil {
		return nil
	}

	return &pb.QuotaInfo{
		QuotaId:       quota.QuotaID,
		TenantId:      quota.TenantID,
		QuotaType:     pb.QuotaType(quota.QuotaType),
		LimitType:     pb.LimitType(quota.LimitType),
		HardLimit:     quota.HardLimit,
		SoftLimit:     quota.SoftLimit,
		UsedCount:     quota.UsedCount,
		ResetTime:     quota.ResetTime.Format(time.RFC3339),
		NextResetTime: quota.NextResetTime.Format(time.RFC3339),
		EffectiveTime: quota.EffectiveTime.Format(time.RFC3339),
		ExpireTime:    quota.ExpireTime.Format(time.RFC3339),
		IsGlobal:      quota.IsGlobal,
		ProductCodes:  quota.ProductCodes,
	}
}

// convertProductToPB converts product from biz to proto
func convertProductToPB(product *biz.Product) *pb.Product {
	if product == nil {
		return nil
	}

	return &pb.Product{
		ProductCode: product.ProductCode,
		ProductName: product.ProductName,
		Description: product.Description,
	}
}

// CreateTenant implements tenant.CreateTenant
func (s *TenantService) CreateTenant(ctx context.Context, req *pb.CreateTenantRequest) (*pb.CreateTenantReply, error) {
	s.log.WithContext(ctx).Infof("CreateTenant: %v", req.GetTenantName())

	// Convert request to biz model
	tenant := &biz.Tenant{
		TenantName:     req.GetTenantName(),
		TenantType:     convertTenantTypeToEnum(req.GetTenantType()),
		ParentTenantID: req.GetParentTenantId(),
		Status:         true, // Default to active
		QuotaConfig:    req.GetQuotaConfig(),
	}

	// Call business logic
	createdTenant, err := s.tu.CreateTenant(ctx, tenant)
	if err != nil {
		return nil, err
	}

	return &pb.CreateTenantReply{
		Tenant: convertTenantToPB(createdTenant),
	}, nil
}

// GetTenant implements tenant.GetTenant
func (s *TenantService) GetTenant(ctx context.Context, req *pb.GetTenantRequest) (*pb.GetTenantReply, error) {
	s.log.WithContext(ctx).Infof("GetTenant: %v", req.GetTenantId())

	// Call business logic
	tenant, err := s.tu.GetTenant(ctx, req.GetTenantId())
	if err != nil {
		return nil, err
	}

	return &pb.GetTenantReply{
		Tenant: convertTenantToPB(tenant),
	}, nil
}

// ListTenants implements tenant.ListTenants
func (s *TenantService) ListTenants(ctx context.Context, req *pb.ListTenantsRequest) (*pb.ListTenantsReply, error) {
	s.log.WithContext(ctx).Info("ListTenants")

	// Set default pagination values if not provided
	pageSize := req.GetPageSize()
	if pageSize <= 0 {
		pageSize = 10
	}

	pageNum := req.GetPageNum()
	if pageNum <= 0 {
		pageNum = 1
	}

	// Call business logic
	tenants, total, err := s.tu.ListTenants(
		ctx,
		convertTenantTypeToEnum(req.GetTenantType()),
		req.GetParentTenantId(),
		req.GetStatus(),
		pageNum,
		pageSize,
	)
	if err != nil {
		return nil, err
	}

	// Convert to proto response
	pbTenants := make([]*pb.Tenant, 0, len(tenants))
	for _, tenant := range tenants {
		pbTenants = append(pbTenants, convertTenantToPB(tenant))
	}

	return &pb.ListTenantsReply{
		Tenants: pbTenants,
		Total:   total,
	}, nil
}

// UpdateTenant implements tenant.UpdateTenant
func (s *TenantService) UpdateTenant(ctx context.Context, req *pb.UpdateTenantRequest) (*pb.UpdateTenantReply, error) {
	s.log.WithContext(ctx).Infof("UpdateTenant: %v", req.GetTenantId())

	// Get existing tenant
	existingTenant, err := s.tu.GetTenant(ctx, req.GetTenantId())
	if err != nil {
		return nil, err
	}

	if existingTenant == nil {
		return nil, status.Errorf(codes.NotFound, "tenant not found: %s", req.GetTenantId())
	}

	// Update fields
	existingTenant.TenantName = req.GetTenantName()
	existingTenant.Status = req.GetStatus()
	existingTenant.QuotaConfig = req.GetQuotaConfig()

	// Call business logic
	updatedTenant, err := s.tu.UpdateTenant(ctx, existingTenant)
	if err != nil {
		return nil, err
	}

	return &pb.UpdateTenantReply{
		Tenant: convertTenantToPB(updatedTenant),
	}, nil
}

// DeleteTenant implements tenant.DeleteTenant
func (s *TenantService) DeleteTenant(ctx context.Context, req *pb.DeleteTenantRequest) (*pb.DeleteTenantReply, error) {
	s.log.WithContext(ctx).Infof("DeleteTenant: %v", req.GetTenantId())

	// Call business logic
	err := s.tu.DeleteTenant(ctx, req.GetTenantId())
	if err != nil {
		return nil, err
	}

	return &pb.DeleteTenantReply{
		Success: true,
	}, nil
}

// CheckQuota implements tenant.CheckQuota
func (s *TenantService) CheckQuota(ctx context.Context, req *pb.CheckQuotaRequest) (*pb.CheckQuotaReply, error) {
	s.log.WithContext(ctx).Infof("CheckQuota: tenantID=%v, quotaType=%v", req.GetTenantId(), req.GetQuotaType())

	// Call business logic
	quota, hasQuota, available, err := s.qu.CheckQuota(
		ctx,
		req.GetTenantId(),
		convertQuotaTypeToEnum(req.GetQuotaType()),
		convertLimitTypeToEnum(req.GetLimitType()),
		req.GetProductCode(),
	)
	if err != nil {
		return nil, err
	}

	return &pb.CheckQuotaReply{
		Quota:          convertQuotaInfoToPB(quota),
		HasQuota:       hasQuota,
		AvailableQuota: available,
	}, nil
}

// ConsumeQuota implements tenant.ConsumeQuota
func (s *TenantService) ConsumeQuota(ctx context.Context, req *pb.ConsumeQuotaRequest) (*pb.ConsumeQuotaReply, error) {
	s.log.WithContext(ctx).Infof("ConsumeQuota: tenantID=%v, quotaType=%v, amount=%v",
		req.GetTenantId(), req.GetQuotaType(), req.GetAmount())

	// Call business logic
	success, remaining, err := s.qu.ConsumeQuota(
		ctx,
		req.GetTenantId(),
		convertQuotaTypeToEnum(req.GetQuotaType()),
		convertLimitTypeToEnum(req.GetLimitType()),
		req.GetAmount(),
		req.GetProductCode(),
		req.GetBizId(),
		req.GetBizType(),
	)

	message := ""
	if err != nil {
		message = err.Error()
	}

	return &pb.ConsumeQuotaReply{
		Success:        success,
		RemainingQuota: remaining,
		Message:        message,
	}, nil
}

// ReleaseQuota implements tenant.ReleaseQuota
func (s *TenantService) ReleaseQuota(ctx context.Context, req *pb.ReleaseQuotaRequest) (*pb.ReleaseQuotaReply, error) {
	s.log.WithContext(ctx).Infof("ReleaseQuota: tenantID=%v, quotaType=%v, amount=%v",
		req.GetTenantId(), req.GetQuotaType(), req.GetAmount())

	// Call business logic
	success, remaining, err := s.qu.ReleaseQuota(
		ctx,
		req.GetTenantId(),
		convertQuotaTypeToEnum(req.GetQuotaType()),
		convertLimitTypeToEnum(req.GetLimitType()),
		req.GetAmount(),
		req.GetProductCode(),
		req.GetBizId(),
	)

	message := ""
	if err != nil {
		message = err.Error()
	}

	return &pb.ReleaseQuotaReply{
		Success:        success,
		RemainingQuota: remaining,
		Message:        message,
	}, nil
}

// ListProducts implements tenant.ListProducts
func (s *TenantService) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsReply, error) {
	s.log.WithContext(ctx).Info("ListProducts")

	var products []*biz.Product
	var err error

	// If tenant ID is provided, list products for that tenant
	if req.GetTenantId() != "" {
		products, err = s.pu.ListProductsByTenant(ctx, req.GetTenantId())
	} else {
		// Otherwise, list all products
		products, err = s.pu.ListProducts(ctx)
	}

	if err != nil {
		return nil, err
	}

	// Convert to proto response
	pbProducts := make([]*pb.Product, 0, len(products))
	for _, product := range products {
		pbProducts = append(pbProducts, convertProductToPB(product))
	}

	return &pb.ListProductsReply{
		Products: pbProducts,
	}, nil
}
