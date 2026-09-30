-- 企业采购需求：关联登录用户（未登录留 NULL）
ALTER TABLE `enterprise_inquiry`
  ADD COLUMN `user_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '登录用户 ID' AFTER `id`,
  ADD KEY `idx_enterprise_inquiry_user` (`user_id`);
