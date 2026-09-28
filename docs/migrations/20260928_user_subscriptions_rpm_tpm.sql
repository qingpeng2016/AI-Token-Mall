-- user_subscriptions：开通/升档时从 products 快照 RPM/TPM，网关限流读履约表
ALTER TABLE `user_subscriptions`
  ADD COLUMN `rpm_limit` INT NOT NULL DEFAULT 0 COMMENT '用户 RPM 快照，0=不限' AFTER `used_tokens`,
  ADD COLUMN `tpm_limit` INT DEFAULT NULL COMMENT '用户 TPM 快照（可选）' AFTER `rpm_limit`;

-- 已有订阅：从当前 SKU 回填（无 product 或已删 SKU 则保持 0 / NULL）
UPDATE `user_subscriptions` us
INNER JOIN `products` p ON p.`id` = us.`product_id`
SET
  us.`rpm_limit` = p.`rpm_limit`,
  us.`tpm_limit` = p.`tpm_limit`;
