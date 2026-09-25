-- 聚合 AI 平台 — 表结构（依据 docs/ai-platform-design.md）
-- 引擎：MySQL 8.0+，字符集 utf8mb4
-- 说明：Key / 上游 Key 仅存哈希或密文；金额以「分」为单位存储

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ---------------------------------------------------------------------------
-- 会员与账户
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS `users` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '用户 ID',
  `email`           VARCHAR(255) DEFAULT NULL COMMENT '邮箱（登录）',
  `phone`           VARCHAR(32)  DEFAULT NULL COMMENT '手机号（登录）',
  `password_hash`   VARCHAR(255) NOT NULL COMMENT '密码哈希',
  `nickname`        VARCHAR(64)  DEFAULT NULL COMMENT '昵称',
  `status`          VARCHAR(32)  NOT NULL DEFAULT 'active' COMMENT 'active|disabled|banned',
  `last_login_at`   DATETIME     DEFAULT NULL,
  `created_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_users_email` (`email`),
  UNIQUE KEY `uk_users_phone` (`phone`),
  KEY `idx_users_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='注册用户';

CREATE TABLE IF NOT EXISTS `user_invoice_config` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`         BIGINT UNSIGNED NOT NULL COMMENT '用户 ID',
  `profile_type`    VARCHAR(16)  NOT NULL DEFAULT 'enterprise' COMMENT 'enterprise|personal',
  `title`           VARCHAR(256) NOT NULL COMMENT '发票抬头',
  `tax_no`          VARCHAR(64)  DEFAULT NULL COMMENT '税号',
  `bank_name`       VARCHAR(128) DEFAULT NULL,
  `bank_account`    VARCHAR(64)  DEFAULT NULL,
  `address`         VARCHAR(512) DEFAULT NULL,
  `phone`           VARCHAR(32)  DEFAULT NULL,
  `is_default`      TINYINT(1)   NOT NULL DEFAULT 0,
  `created_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_invoice_config_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户发票抬头配置（含企业 VAT）';

-- ---------------------------------------------------------------------------
-- 目录与 SKU（quota / 池路由引用）
-- ---------------------------------------------------------------------------

-- 上游信息：逻辑总池 + 单条上游账号/凭证；容量颗粒度到每一条记录
CREATE TABLE IF NOT EXISTS `upstream_info` (
  `id`                    BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `upstream_name`         VARCHAR(32)  NOT NULL COMMENT '上游厂商/协议线：openai|anthropic|xai|gemini|perplexity',
  `upstream_product`      VARCHAR(128) NOT NULL COMMENT '上游采购档位，如 GPT PRO 5X',
  `account_label`         VARCHAR(128) NOT NULL COMMENT '上游账号/条目运营标签',
  `api_key_ciphertext`    VARBINARY(1024) NOT NULL COMMENT '官方 API Key 密文',
  `api_secret_ciphertext` VARBINARY(1024) DEFAULT NULL COMMENT '部分厂商 Secret 密文',
  `procurement_cost_note` VARCHAR(256) DEFAULT NULL COMMENT '采购成本备注',
  `expires_at`            DATETIME     DEFAULT NULL COMMENT '凭证到期日',
  -- 容量快照：remain=cap_tokens-used_tokens；used>=cap 时停选路由（status=over_cap）
  `cap_tokens`            BIGINT       DEFAULT NULL COMMENT 'token 配额上限，NULL=不限或未录入',
  `used_tokens`           BIGINT       NOT NULL DEFAULT 0 COMMENT '当前周期已消耗 token',
  `quota_reset_at`        DATETIME     DEFAULT NULL COMMENT '下次清零 used_tokens（自然月/采购周期由业务写入）',
  `weight`                INT          NOT NULL DEFAULT 100 COMMENT '同 upstream_name+upstream_product 组内路由权重',
  `status`                VARCHAR(32)  NOT NULL DEFAULT 'active' COMMENT 'active=参与路由；disabled|unhealthy|over_cap=不参与',
  `fail_rate`             DECIMAL(5,4) DEFAULT NULL COMMENT '近期失败率',
  `created_at`            DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`            DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_upstream_info_product` (`upstream_name`, `upstream_product`),
  KEY `idx_upstream_info_name_status` (`upstream_name`, `status`),
  KEY `idx_upstream_info_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='上游信息（厂商+档位 + 凭证 + 单条容量）';

CREATE TABLE IF NOT EXISTS `products` (
  `id`                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `sku_code`          VARCHAR(64)  NOT NULL COMMENT 'SKU 编码',
  `marketing_tier`    VARCHAR(128) NOT NULL COMMENT '对外档位名，如 Pro 20X',
  `upstream_name`     VARCHAR(32)  NOT NULL,
  `upstream_product`  VARCHAR(128) NOT NULL COMMENT '绑定的上游档位，如 GPT PRO 5X',
  `limit_tokens`      BIGINT       NOT NULL COMMENT '每计费周期 token 额度（售卖给用户）',
  `rpm_limit`         INT          NOT NULL DEFAULT 0 COMMENT '用户 RPM，0=不限',
  `tpm_limit`         INT          DEFAULT NULL COMMENT '用户 TPM（可选）',
  `allowed_models`    JSON         NOT NULL COMMENT '允许 model 列表',
  `product_type`      VARCHAR(32)  NOT NULL DEFAULT 'subscription' COMMENT 'subscription|token_topup',
  `billing_period`    VARCHAR(16)  NOT NULL DEFAULT 'month' COMMENT 'month|year|once',
  `price_cents`       BIGINT       NOT NULL COMMENT '售价（分）',
  `currency`          CHAR(3)      NOT NULL DEFAULT 'CNY',
  `compare_at_price_cents` BIGINT  DEFAULT NULL COMMENT '对比官方价（可选）',
  `highlights_json`   JSON         DEFAULT NULL COMMENT '卖点',
  `is_hot`            TINYINT(1)   NOT NULL DEFAULT 0,
  `is_api_enabled`    TINYINT(1)   NOT NULL DEFAULT 1 COMMENT '是否走 API 通道',
  `topup_token_amount` BIGINT      DEFAULT NULL COMMENT '加购包 token 数（product_type=token_topup）',
  `sort_order`        INT          NOT NULL DEFAULT 0,
  `status`            VARCHAR(16)  NOT NULL DEFAULT 'on_sale' COMMENT 'on_sale|off_sale',
  `created_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_products_sku` (`sku_code`),
  KEY `idx_products_upstream_name` (`upstream_name`),
  KEY `idx_products_upstream` (`upstream_name`, `upstream_product`),
  KEY `idx_products_status_sort` (`status`, `sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商城 SKU（含品牌筛选：upstream_name）';

-- ---------------------------------------------------------------------------
-- 交易与订单
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS `user_orders` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `order_no`        VARCHAR(64)  NOT NULL COMMENT '业务订单号',
  `user_id`         BIGINT UNSIGNED NOT NULL,
  `product_id`      BIGINT UNSIGNED NOT NULL COMMENT '购买的 SKU（MVP 一单一件商品）',
  `quantity`        INT          NOT NULL DEFAULT 1 COMMENT '购买数量',
  `unit_price_cents` BIGINT       NOT NULL COMMENT '下单时单价快照（分）',
  `status`          VARCHAR(32)  NOT NULL COMMENT 'pending_payment|paid|fulfilling|completed|failed|cancelled|refunding',
  `total_amount_cents` BIGINT    NOT NULL COMMENT '应付总额（分），一般=unit_price_cents*quantity',
  `currency`        CHAR(3)      NOT NULL DEFAULT 'CNY',
  `enterprise_invoice` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否企业开票',
  `paid_at`         DATETIME     DEFAULT NULL,
  `expire_at`       DATETIME     NOT NULL COMMENT '待支付关单时间',
  `closed_at`       DATETIME     DEFAULT NULL,
  `fail_reason`     VARCHAR(512) DEFAULT NULL,
  `created_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_orders_order_no` (`order_no`),
  KEY `idx_user_orders_user_status` (`user_id`, `status`),
  KEY `idx_user_orders_expire` (`expire_at`),
  KEY `idx_user_orders_product` (`product_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户订单（MVP：一单对应一个 product）';

CREATE TABLE IF NOT EXISTS `user_payments` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `order_id`        BIGINT UNSIGNED NOT NULL,
  `channel`         VARCHAR(32)  NOT NULL COMMENT 'alipay|wechat|stripe|paypal',
  `out_trade_no`    VARCHAR(64)  NOT NULL COMMENT '平台支付单号',
  `third_trade_no`  VARCHAR(128) DEFAULT NULL COMMENT '渠道流水号',
  `amount_cents`    BIGINT       NOT NULL,
  `status`          VARCHAR(32)  NOT NULL COMMENT 'pending|success|failed|closed',
  `paid_at`         DATETIME     DEFAULT NULL,
  `raw_request_json`  JSON       DEFAULT NULL,
  `raw_notify_json`   JSON       DEFAULT NULL,
  `created_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_payments_out_trade_no` (`out_trade_no`),
  KEY `idx_user_payments_order` (`order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户支付记录';

CREATE TABLE IF NOT EXISTS `user_payment_callbacks` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `channel`         VARCHAR(32)  NOT NULL,
  `idempotency_key` VARCHAR(128) NOT NULL COMMENT '幂等键',
  `payload_json`    JSON         NOT NULL,
  `signature_ok`    TINYINT(1)   NOT NULL DEFAULT 0,
  `process_result`  VARCHAR(32)  NOT NULL COMMENT 'success|ignored|failed',
  `processed_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_payment_callbacks_idem` (`channel`, `idempotency_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户支付回调幂等与对账';

CREATE TABLE IF NOT EXISTS `user_refunds` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `order_id`        BIGINT UNSIGNED NOT NULL,
  `payment_id`      BIGINT UNSIGNED DEFAULT NULL,
  `refund_no`       VARCHAR(64)  NOT NULL,
  `amount_cents`    BIGINT       NOT NULL,
  `reason`          VARCHAR(512) DEFAULT NULL,
  `status`          VARCHAR(32)  NOT NULL COMMENT 'pending|success|failed',
  `refunded_at`     DATETIME     DEFAULT NULL,
  `created_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_refunds_refund_no` (`refund_no`),
  KEY `idx_user_refunds_order` (`order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户退款';

CREATE TABLE IF NOT EXISTS `user_invoices` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `order_id`        BIGINT UNSIGNED NOT NULL,
  `user_id`         BIGINT UNSIGNED NOT NULL,
  `invoice_type`    VARCHAR(32)  NOT NULL COMMENT 'personal|electronic|enterprise_vat',
  `title`           VARCHAR(256) NOT NULL,
  `tax_no`          VARCHAR(64)  DEFAULT NULL,
  `amount_cents`    BIGINT       NOT NULL,
  `status`          VARCHAR(32)  NOT NULL COMMENT 'pending|issued|failed',
  `file_url`        VARCHAR(512) DEFAULT NULL,
  `issued_at`       DATETIME     DEFAULT NULL,
  `created_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_invoices_user` (`user_id`),
  KEY `idx_user_invoices_order` (`order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户发票';

-- ---------------------------------------------------------------------------
-- 履约：用户订阅 + API Key
-- 履约约定：started_at/expires_at=整段服务有效期（续费延长 expires_at）；
-- period_start/period_end=当前 token 计费周期（滚动时 used=0, limit=base）；新开 base=limit=products.limit_tokens；
-- 永久升档改 base_limit_tokens；本周期加购 limit_tokens+=N，user_orders 追加订单 id
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS `user_subscriptions` (
  `id`                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`           BIGINT UNSIGNED NOT NULL,
  `product_id`        BIGINT UNSIGNED NOT NULL,
  `orders`            JSON         NOT NULL DEFAULT (JSON_ARRAY()) COMMENT '关联 user_orders.id 数组（同一 product 的开通、续费等）',
  `upstream_name`     VARCHAR(32)  NOT NULL,
  `upstream_product`  VARCHAR(128) NOT NULL,
  `base_limit_tokens` BIGINT       NOT NULL COMMENT '套餐基准上限；续费升档改此值；周期重置时 limit_tokens 回到此值',
  `limit_tokens`      BIGINT       NOT NULL COMMENT '当前周期有效上限（=base+本周期临时加购，可>base）',
  `used_tokens`       BIGINT       NOT NULL DEFAULT 0 COMMENT '当前周期已用；剩余=limit_tokens-used_tokens',
  `started_at`        DATETIME     NOT NULL COMMENT '服务整体开始时间（首开）',
  `expires_at`        DATETIME     NOT NULL COMMENT '服务整体到期（续费延长；网关校验是否仍可调用）',
  `period_start`      DATETIME     NOT NULL COMMENT '当前 token 计费周期起（如自然月）',
  `period_end`        DATETIME     NOT NULL COMMENT '当前 token 计费周期止（到期滚动，与 expires_at 可不同）',
  `status`            VARCHAR(32)  NOT NULL DEFAULT 'active' COMMENT 'active|expired|suspended|cancelled',
  `created_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_subscriptions_user` (`user_id`),
  KEY `idx_user_subscriptions_status_expires` (`status`, `expires_at`),
  KEY `idx_user_subscriptions_period_end` (`period_end`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户订阅（额度+周期+新开/续费/升档/加购）';

CREATE TABLE IF NOT EXISTS `user_api_keys` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`         BIGINT UNSIGNED NOT NULL,
  `user_subscription_id` BIGINT UNSIGNED NOT NULL COMMENT '归属的用户订阅；同订阅下可有多条 Key',
  `key_hash`        CHAR(64)     NOT NULL COMMENT '平台 Key 的 SHA-256（明文仅创建时展示一次，不入库）',
  `upstream_name`   VARCHAR(32)  NOT NULL COMMENT '网关 Path 隔离，如 openai',
  `limit_tokens`      BIGINT       NOT NULL COMMENT '该 Key 本周期 token 上限（多条 Key 分配之和不超过 user_subscriptions.limit_tokens）',
  `used_tokens`       BIGINT       NOT NULL DEFAULT 0 COMMENT '该 Key 本周期已用；剩余=limit_tokens-used_tokens',
  `status`          VARCHAR(32)  NOT NULL DEFAULT 'active' COMMENT 'active|disabled|rotated',
  `rotated_at`      DATETIME     DEFAULT NULL,
  `created_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_api_keys_hash` (`key_hash`),
  KEY `idx_user_api_keys_subscription` (`user_subscription_id`),
  KEY `idx_user_api_keys_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户 API Key（同 user_subscriptions 下多条平级）';

CREATE TABLE IF NOT EXISTS `user_notifications` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`         BIGINT UNSIGNED NOT NULL,
  `user_subscription_id` BIGINT UNSIGNED DEFAULT NULL,
  `api_key_id`      BIGINT UNSIGNED DEFAULT NULL,
  `channel`         VARCHAR(32)  NOT NULL COMMENT 'email|in_app|sms',
  `template_code`   VARCHAR(64)  NOT NULL COMMENT 'key_issued|renew_reminder 等',
  `status`          VARCHAR(32)  NOT NULL COMMENT 'pending|sent|failed',
  `sent_at`         DATETIME     DEFAULT NULL,
  `created_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_notifications_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户通知（站内/邮件等，开通 Key、续费提醒）';

-- ---------------------------------------------------------------------------
-- 用户 API 网关访问日志
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS `user_access_logs` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `request_id`      VARCHAR(64)  NOT NULL,
  `user_id`         BIGINT UNSIGNED DEFAULT NULL,
  `user_subscription_id` BIGINT UNSIGNED DEFAULT NULL,
  `api_key_id`      BIGINT UNSIGNED DEFAULT NULL,
  `upstream_name`   VARCHAR(32)  NOT NULL,
  `upstream_product` VARCHAR(128) DEFAULT NULL COMMENT '实际路由档位（冗余便于检索）',
  `upstream_info_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '实际选用的上游条目',
  `model`           VARCHAR(128) DEFAULT NULL,
  `http_method`     VARCHAR(16)  NOT NULL,
  `path`            VARCHAR(512) NOT NULL,
  `client_ip`       VARCHAR(64)  DEFAULT NULL,
  `gateway_status`  SMALLINT     NOT NULL COMMENT '网关响应码',
  `upstream_status` SMALLINT     DEFAULT NULL,
  `latency_ms`      INT          NOT NULL DEFAULT 0,
  `tokens_prompt`   INT          DEFAULT NULL,
  `tokens_completion` INT        DEFAULT NULL,
  `tokens_total`    INT          DEFAULT NULL,
  `is_stream`       TINYINT(1)   NOT NULL DEFAULT 0,
  `error_code`      VARCHAR(32)  DEFAULT NULL COMMENT '401|402|429|502 等',
  `created_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_access_logs_request_id` (`request_id`),
  KEY `idx_user_access_logs_user_time` (`user_id`, `created_at`),
  KEY `idx_user_access_logs_key_time` (`api_key_id`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户 API 网关访问日志（默认不存完整 prompt）';

-- ---------------------------------------------------------------------------
-- 脚本调度配置
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS `bot_schedule_config` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `module` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `task_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `interval_seconds` float NOT NULL DEFAULT '60',
  `exe_sort` int DEFAULT NULL,
  `concurrency` int DEFAULT NULL,
  `description` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `is_enabled` tinyint(1) NOT NULL DEFAULT '1',
  `is_strategy_enabled` tinyint(1) DEFAULT '1',
  `last_live_time` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `last_live_time_by_machine` json DEFAULT NULL,
  `is_primary_machine_run` int NOT NULL DEFAULT '0',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_module_task` (`module`,`task_name`),
  KEY `idx_module` (`module`),
  KEY `idx_bot_schedule_module_enabled` (`module`,`is_enabled`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='脚本调度配置';

SET FOREIGN_KEY_CHECKS = 1;
