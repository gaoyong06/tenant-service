# 使用示例场景

1. 初始化渠道商配额

```sql
-- 为渠道商CH_123设置每月兑换码生成限额
INSERT INTO tenant_quotas (
  tenant_id, quota_type, limit_type, hard_limit, 
  soft_limit, effective_time, product_codes
) VALUES (
  'CH_123', 'REDEEM_CODE', 'MONTHLY', 10000, 
  8000, '2023-08-01 00:00:00', 
  '["app_mall","wx_miniprogram"]'
);
```

2. 核销时检查配额（事务操作）

```sql

START TRANSACTION;

-- 检查配额
SELECT hard_limit, used_count 
FROM tenant_quotas 
WHERE tenant_id = 'CH_123' 
  AND quota_type = 'REDEEM_CODE'
  AND limit_type = 'MONTHLY'
FOR UPDATE;

-- 更新使用量（如果未超限）
UPDATE tenant_quotas 
SET used_count = used_count + 1 
WHERE quota_id = 123 
  AND used_count < hard_limit;

-- 记录流水
INSERT INTO quota_usage_records (
  quota_id, tenant_id, operation_type, 
  delta_value, current_used, biz_id, biz_type
) VALUES (
  123, 'CH_123', 'CONSUME', 
  1, (SELECT used_count FROM tenant_quotas WHERE quota_id = 123),
  'order_789', 'REDEMPTION'
);

COMMIT;

```

3. 自动配额重置（定时任务）


```sql

-- 每月1日重置用量
UPDATE tenant_quotas 
SET used_count = 0, 
    reset_time = NOW(),
    next_reset_time = DATE_ADD(NOW(), INTERVAL 1 MONTH)
WHERE limit_type = 'MONTHLY' 
  AND next_reset_time <= NOW();

```

4. 并发配额控制

```sql

-- 添加并发配额记录
ALTER TABLE tenant_quotas 
ADD COLUMN `concurrent_tokens` int(11) DEFAULT NULL COMMENT '并发令牌数';

-- 使用存储过程实现令牌桶算法
DELIMITER //
CREATE PROCEDURE acquire_quota_token(
  IN p_tenant_id VARCHAR(24),
  IN p_quota_type VARCHAR(32),
  OUT p_success BOOLEAN
)
BEGIN
  DECLARE v_available INT;
  
  START TRANSACTION;
  
  SELECT concurrent_tokens - used_count INTO v_available
  FROM tenant_quotas
  WHERE tenant_id = p_tenant_id 
    AND quota_type = p_quota_type
    AND limit_type = 'CONCURRENT'
  FOR UPDATE;
  
  IF v_available > 0 THEN
    UPDATE tenant_quotas 
    SET used_count = used_count + 1
    WHERE tenant_id = p_tenant_id 
      AND quota_type = p_quota_type;
    
    SET p_success = TRUE;
    COMMIT;
  ELSE
    SET p_success = FALSE;
    ROLLBACK;
  END IF;
END //
DELIMITER ;

```

5. 可视化监控视图

```sql

CREATE VIEW quota_usage_monitor AS
SELECT 
  t.tenant_id,
  t.tenant_name,
  q.quota_type,
  q.limit_type,
  q.hard_limit,
  q.used_count,
  ROUND(q.used_count/q.hard_limit*100,2) AS usage_rate,
  CASE 
    WHEN q.used_count >= q.hard_limit THEN 'EXCEEDED'
    WHEN q.soft_limit IS NOT NULL AND q.used_count >= q.soft_limit THEN 'WARNING'
    ELSE 'NORMAL'
  END AS status_level
FROM tenant_quotas q
JOIN tenants t ON q.tenant_id = t.tenant_id;

```

