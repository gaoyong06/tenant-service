-- 这个设计支持以下业务场景：
-- 多维度配额控制（按租户/产品线/时间维度）
-- 硬性限制和软性告警
-- 配额预占和释放
-- 完整的审计追踪
-- 并发流量控制

CREATE DATABASE `tenant_service` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 该库包含：
-- tenants (租户核心表)
-- channels (渠道扩展表)
-- tenant_products (租户-产品线关联表)
-- tenant_quotas (租户配额表)
-- quota_usage_records (配额使用记录表)

-- 租户表（tenants）
CREATE TABLE `tenants` (
  `tenant_id` varchar(24) NOT NULL COMMENT '租户唯一标识',
  `tenant_name` varchar(64) NOT NULL COMMENT '租户名称',
  `tenant_type` enum('PLATFORM','CHANNEL','ENTERPRISE') NOT NULL COMMENT '租户类型：平台/渠道/企业',
  `parent_tenant_id` varchar(24) DEFAULT NULL COMMENT '父租户ID',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态：0-禁用 1-启用',
  `quota_config` json DEFAULT NULL COMMENT '配额配置',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`tenant_id`),
  KEY `idx_parent_tenant` (`parent_tenant_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='租户信息表';

-- 渠道扩展表（channels）
CREATE TABLE `channels` (
  `channel_id` bigint(20) NOT NULL AUTO_INCREMENT,
  `tenant_id` varchar(24) NOT NULL COMMENT '关联租户ID',
  `channel_code` varchar(32) NOT NULL COMMENT '渠道编码',
  `channel_name` varchar(64) NOT NULL COMMENT '渠道名称',
  `contact_name` varchar(32) DEFAULT NULL COMMENT '联系人',
  `contact_phone` varchar(20) DEFAULT NULL COMMENT '联系电话',
  `commission_rate` decimal(5,2) DEFAULT NULL COMMENT '佣金比例',
  `sales_target` decimal(12,2) DEFAULT NULL COMMENT '销售目标',
  `extra_data` json DEFAULT NULL COMMENT '扩展字段',
  PRIMARY KEY (`channel_id`),
  UNIQUE KEY `uk_tenant_id` (`tenant_id`),
  UNIQUE KEY `uk_channel_code` (`channel_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='渠道商扩展信息表';



-- 租户-产品线关联表
CREATE TABLE tenant_products (
    id BIGINT PRIMARY KEY,
    tenant_id VARCHAR(24),
    product_code VARCHAR(16),
    UNIQUE KEY (tenant_id, product_code)
);

-- 租户配额表
CREATE TABLE `tenant_quotas` (
  `quota_id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '配额ID',
  `tenant_id` varchar(24) NOT NULL COMMENT '关联租户ID',
  `quota_type` varchar(32) NOT NULL COMMENT '配额类型：MARKETING_CAMPAIGN-营销活动 REDEEM_CODE-兑换码 SMS-短信等',
  `limit_type` enum('DAILY','MONTHLY','TOTAL','CONCURRENT') NOT NULL COMMENT '限制类型：日/月/总量/并发',
  `hard_limit` int(11) NOT NULL COMMENT '硬性上限',
  `soft_limit` int(11) DEFAULT NULL COMMENT '软性上限（告警阈值）',
  `used_count` int(11) NOT NULL DEFAULT '0' COMMENT '已使用量',
  `reset_time` datetime DEFAULT NULL COMMENT '上次重置时间',
  `next_reset_time` datetime DEFAULT NULL COMMENT '下次计划重置时间',
  `effective_time` datetime NOT NULL COMMENT '生效时间',
  `expire_time` datetime DEFAULT NULL COMMENT '过期时间',
  `is_global` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否全局默认配额',
  `product_codes` json DEFAULT NULL COMMENT '适用产品线["app1","web2"]，null表示全部',
  `extra_config` json DEFAULT NULL COMMENT '扩展配置',
  `created_by` varchar(64) DEFAULT NULL COMMENT '创建人',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`quota_id`),
  UNIQUE KEY `uk_tenant_quota_type` (`tenant_id`, `quota_type`, `limit_type`),
  KEY `idx_reset_time` (`next_reset_time`),
  KEY `idx_global_quota` (`is_global`, `quota_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='租户配额表';


-- 配额使用记录表
CREATE TABLE `quota_usage_records` (
  `record_id` bigint(20) NOT NULL AUTO_INCREMENT,
  `quota_id` bigint(20) NOT NULL COMMENT '关联配额ID',
  `tenant_id` varchar(24) NOT NULL COMMENT '租户ID',
  `operation_type` enum('CONSUME','RELEASE','ADJUST') NOT NULL COMMENT '操作类型',
  `delta_value` int(11) NOT NULL COMMENT '变更数值（正数增加，负数消耗）',
  `current_used` int(11) NOT NULL COMMENT '变更后已用量',
  `biz_id` varchar(64) DEFAULT NULL COMMENT '关联业务ID',
  `biz_type` varchar(32) DEFAULT NULL COMMENT '业务类型',
  `operator` varchar(64) DEFAULT NULL COMMENT '操作人',
  `operation_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `expire_time` datetime DEFAULT NULL COMMENT '预占过期时间（针对临时配额）',
  `remark` varchar(255) DEFAULT NULL COMMENT '备注',
  PRIMARY KEY (`record_id`),
  KEY `idx_quota_tenant` (`quota_id`, `tenant_id`),
  KEY `idx_biz_reference` (`biz_type`, `biz_id`),
  KEY `idx_operation_time` (`operation_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='配额使用记录表';