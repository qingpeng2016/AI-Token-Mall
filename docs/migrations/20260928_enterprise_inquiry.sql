-- 企业采购 / 需求提交

CREATE TABLE IF NOT EXISTS `enterprise_inquiry` (
  `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`       BIGINT UNSIGNED DEFAULT NULL COMMENT '登录用户 ID',
  `company_name`  VARCHAR(256) NOT NULL COMMENT '公司名称',
  `contact_name`  VARCHAR(128) NOT NULL COMMENT '联系人',
  `phone`         VARCHAR(32)  NOT NULL COMMENT '手机',
  `email`         VARCHAR(255) DEFAULT NULL COMMENT '邮箱',
  `status`        VARCHAR(32)  NOT NULL DEFAULT 'pending' COMMENT 'pending|contacted|closed',
  `created_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_enterprise_inquiry_user` (`user_id`),
  KEY `idx_enterprise_inquiry_status` (`status`),
  KEY `idx_enterprise_inquiry_created` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='企业采购需求';
