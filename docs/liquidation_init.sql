-- =============================================================================
-- 质押清算执行引擎 — 新建库一次性初始化（表结构 + 默认配置 + 脚本调度 seed）
-- =============================================================================
-- 编号链路：loan_order_id → liquidation_id → batch_no（orders）→ client_order_id → exchange_trade_id
--
-- 说明：
-- 1. 仅用于空库初始化；已有库请勿直接执行（会 CREATE IF NOT EXISTS，不会删旧列）。
-- 2. 任务累计成交/剩余负债不落 liquidation_task，由 liquidation_traders / liquidation_orders 实时汇总。
-- 3. 无 liquidation_task_sell_batch 表；每轮 IOC 对应 liquidation_orders 一行。
-- 4. 执行前请修改 liquidation_platform_config：sign_key、callback_url、bn_api_key、bn_api_secret。
-- =============================================================================

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ---------------------------------------------------------------------------
-- 1. 强平任务
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `liquidation_task` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `liquidation_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '强平编号，BeTrust 生成，全局唯一',
  `loan_order_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '借款订单 ID',
  `uid` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '被强平用户 ID',
  `status` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'liquidating' COMMENT 'liquidating|paused|settled|bad_debt|cancelled|failed',
  `end_reason` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'debt_cleared|collateral_exhausted|cancelled|failed|manual 等',
  `original_ratio` decimal(20,8) NOT NULL COMMENT '初始质押率快照',
  `trigger_ratio` decimal(20,8) NOT NULL COMMENT '触发时质押率快照',
  `trigger_mark_price` decimal(36,18) NOT NULL COMMENT '触发标记价格 P',
  `trigger_collateral_asset` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '触发时抵押币种',
  `trigger_collateral_qty` decimal(36,18) NOT NULL COMMENT '触发时抵押数量 C',
  `trigger_collateral_due_amount` decimal(36,18) NOT NULL COMMENT '触发时负债总额 D',
  `exchange_code` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '路由交易所，如 binance',
  `account_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '执行账户',
  `symbol` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'Binance 现货交易对，如 BTCUSDT',
  `current_tier` varchar(8) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'P2' COMMENT '当前档位 P0|P1|P2，只升不降',
  `current_tier_reason_json` json DEFAULT NULL COMMENT '最近一次档位重算的判定输入与命中规则',
  `fee_rate` decimal(10,8) NOT NULL DEFAULT '0.00100000' COMMENT '手续费率 f，默认 0.1%',
  `last_mid_price` decimal(36,18) DEFAULT NULL COMMENT '上次 mid/bid 快照（档位跌幅判断）',
  `last_mid_price_at` datetime(3) DEFAULT NULL COMMENT '上次行情快照时间',
  `next_quote_at` datetime(3) DEFAULT NULL COMMENT '下次允许报价/发单时间（档位 quote_interval_sec，可选）',
  `order_fail_streak` int unsigned NOT NULL DEFAULT '0' COMMENT '连续下单失败次数',
  `cancel_reason` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'collateral_top_up|early_repayment|manual|other',
  `cancel_requested_at` datetime(3) DEFAULT NULL COMMENT '平台发起取消时间',
  `cancel_result` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'accepted|already_finished|rejected',
  `in_flight_orders_cleared` int NOT NULL DEFAULT '0' COMMENT '在途挂单是否已清理完成 0否 1是',
  `callback_url` varchar(512) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'BeTrust 回调地址快照',
  `callback_status` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'pending|success|failed',
  `callback_at` datetime(3) DEFAULT NULL COMMENT '最近回调时间',
  `signal_source` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'push' COMMENT 'push|pull',
  `started_at` datetime(3) DEFAULT NULL COMMENT '开始执行时间',
  `finished_at` datetime(3) DEFAULT NULL COMMENT '终态时间',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_liquidation_id` (`liquidation_id`),
  KEY `idx_loan_order_id` (`loan_order_id`),
  KEY `idx_uid` (`uid`),
  KEY `idx_status` (`status`),
  KEY `idx_symbol_status` (`symbol`,`status`),
  KEY `idx_trigger_collateral_asset_status` (`trigger_collateral_asset`,`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='强平任务主表';

-- ---------------------------------------------------------------------------
-- 2. 交易所子订单（IOC，每轮一条）
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `liquidation_orders` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `liquidation_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `batch_no` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '轮次号 {liquidation_id}-B{n}',
  `client_order_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '幂等发单',
  `exchange_code` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `exchange_order_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `symbol` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `side` varchar(8) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'SELL',
  `order_type` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'LIMIT_IOC',
  `tier` varchar(8) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `floor_price` decimal(36,18) DEFAULT NULL COMMENT '本轮保护价',
  `price` decimal(36,18) NOT NULL COMMENT '限价',
  `quantity` decimal(36,18) NOT NULL COMMENT '委托数量',
  `filled_qty` decimal(36,18) NOT NULL DEFAULT '0.000000000000000000' COMMENT '已成交数量',
  `avg_price` decimal(36,18) DEFAULT NULL,
  `status` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'new' COMMENT 'new|partially_filled|filled|cancelled|rejected|unknown',
  `failure_code` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `failure_reason` varchar(512) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `submit_raw` json DEFAULT NULL COMMENT '提交交易所请求摘要',
  `exec_snapshot` json DEFAULT NULL COMMENT '发单前 quote/plan/trigger 中间量快照',
  `last_report_raw` json DEFAULT NULL COMMENT '最近一次查单/回执 raw',
  `submitted_at` datetime(3) DEFAULT NULL,
  `finished_at` datetime(3) DEFAULT NULL,
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_client_order_id` (`client_order_id`),
  KEY `idx_liquidation_id` (`liquidation_id`),
  KEY `idx_batch_no` (`batch_no`),
  KEY `idx_exchange_order_id` (`exchange_code`,`exchange_order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='强平子订单（IOC）';

-- ---------------------------------------------------------------------------
-- 3. 成交明细
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `liquidation_traders` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `liquidation_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `batch_no` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `liquidation_order_id` bigint unsigned NOT NULL COMMENT 'liquidation_orders.id',
  `exchange_code` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `exchange_trade_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `symbol` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `price` decimal(36,18) NOT NULL,
  `quantity` decimal(36,18) NOT NULL,
  `amount` decimal(36,18) NOT NULL,
  `fee` decimal(36,18) NOT NULL DEFAULT '0.000000000000000000',
  `fee_asset` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `is_maker` tinyint(1) DEFAULT NULL,
  `trade_time` datetime(3) NOT NULL,
  `raw_payload` json NOT NULL,
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_exchange_trade` (`exchange_code`,`exchange_trade_id`),
  KEY `idx_liquidation_id` (`liquidation_id`),
  KEY `idx_liquidation_order_id` (`liquidation_order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='强平成交明细';

-- ---------------------------------------------------------------------------
-- 4. BeTrust 入站 API 留痕
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `liquidation_log_inbound` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `liquidation_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `loan_order_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `api_path` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `api_action` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'trigger|cancel',
  `event_time_ms` bigint NOT NULL COMMENT '请求方业务时间戳（毫秒）',
  `idempotency_key` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `request_body` json NOT NULL,
  `response_body` json DEFAULT NULL,
  `http_status` int DEFAULT NULL,
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_liquidation_id` (`liquidation_id`),
  KEY `idx_loan_order_action_time` (`loan_order_id`,`api_action`,`event_time_ms`),
  KEY `idx_liquidation_action_time` (`liquidation_id`,`api_action`,`event_time_ms`),
  KEY `idx_idempotency_key` (`idempotency_key`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='BeTrust → Mojo 入站日志';

-- ---------------------------------------------------------------------------
-- 5. BeTrust 回调 outbound
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `liquidation_log_callback` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `liquidation_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `callback_url` varchar(512) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `payload` json NOT NULL,
  `http_status` int DEFAULT NULL,
  `response_body` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
  `status` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'pending' COMMENT 'pending|success|failed',
  `retry_count` int unsigned NOT NULL DEFAULT '0',
  `next_retry_at` datetime(3) DEFAULT NULL,
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_liquidation_id` (`liquidation_id`),
  KEY `idx_status_next_retry` (`status`,`next_retry_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Mojo → BeTrust 回调日志';

-- ---------------------------------------------------------------------------
-- 6. 任务事件流水
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `liquidation_task_event` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `liquidation_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `event_type` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `from_status` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `to_status` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `message` varchar(512) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `extra` json DEFAULT NULL,
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_liquidation_id_created` (`liquidation_id`,`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='强平任务事件流水';

-- ---------------------------------------------------------------------------
-- 7. 全局执行策略（单行 id=1）
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `liquidation_config_strategy` (
  `id` int unsigned NOT NULL DEFAULT '1' COMMENT '固定 1',
  `global_depth_cap_ratio` decimal(10,8) NOT NULL DEFAULT '0.30000000' COMMENT '同 symbol 在途卖量+本笔 ≤ 全买盘×比例',
  `order_depth_ratio` decimal(10,8) NOT NULL DEFAULT '0.20000000' COMMENT '单笔卖量 ≤ 保护价以上买盘×比例',
  `max_concurrent_tasks_per_symbol` int unsigned NOT NULL DEFAULT '3' COMMENT '同 symbol 强平任务数告警阈值',
  `max_notional_usd_per_symbol` decimal(20,2) NOT NULL DEFAULT '500000.00' COMMENT '同 symbol 剩余抵押名义 USD 告警阈值',
  `max_single_order_duration_sec` int unsigned NOT NULL DEFAULT '300' COMMENT '挂单超时 order_stale（秒）',
  `max_orders_per_task` int unsigned NOT NULL DEFAULT '5' COMMENT '单任务 IOC 次数告警阈值',
  `consecutive_order_fail_pause` int unsigned NOT NULL DEFAULT '3' COMMENT '连续下单失败则暂停任务',
  `stale_quote_sec` int unsigned NOT NULL DEFAULT '3' COMMENT 'last_mid 超过本秒数未更新则跳过发单',
  `hard_boundary_btc_pct` decimal(10,8) NOT NULL DEFAULT '0.02000000',
  `hard_boundary_eth_pct` decimal(10,8) NOT NULL DEFAULT '0.02500000',
  `hard_boundary_default_pct` decimal(10,8) NOT NULL DEFAULT '0.02000000',
  `is_enabled` tinyint(1) NOT NULL DEFAULT '1',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='强平全局参数';

INSERT INTO `liquidation_config_strategy` (`id`)
SELECT 1 FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM `liquidation_config_strategy` WHERE `id` = 1);

-- ---------------------------------------------------------------------------
-- 8. 档位 P0/P1/P2
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `liquidation_config_tier` (
  `tier` varchar(8) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `tier_rank` tinyint unsigned NOT NULL,
  `safety_buffer_lt_pct` decimal(10,8) DEFAULT NULL,
  `safety_buffer_gte_pct` decimal(10,8) DEFAULT NULL,
  `safety_buffer_between_low_pct` decimal(10,8) DEFAULT NULL,
  `safety_buffer_between_high_pct` decimal(10,8) DEFAULT NULL,
  `exec_time_exceed_sec` int unsigned DEFAULT NULL,
  `exec_time_under_sec` int unsigned DEFAULT NULL,
  `consecutive_ioc_unfilled` int unsigned DEFAULT NULL,
  `price_drop_window_sec` int unsigned DEFAULT NULL,
  `price_drop_pct` decimal(10,8) DEFAULT NULL,
  `bid_offset_btc_pct` decimal(10,8) NOT NULL,
  `bid_offset_eth_pct` decimal(10,8) NOT NULL,
  `bid_offset_default_pct` decimal(10,8) NOT NULL,
  `quote_interval_sec` float NOT NULL DEFAULT '1',
  `is_enabled` tinyint(1) NOT NULL DEFAULT '1',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`tier`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='强平档位配置';

INSERT INTO `liquidation_config_tier` (
  `tier`, `tier_rank`, `safety_buffer_lt_pct`, `exec_time_exceed_sec`,
  `bid_offset_btc_pct`, `bid_offset_eth_pct`, `bid_offset_default_pct`, `quote_interval_sec`
) VALUES
  ('P0', 0, 0.08000000, 30, 0.01500000, 0.02000000, 0.01500000, 0.5)
ON DUPLICATE KEY UPDATE `updated_at` = CURRENT_TIMESTAMP(3);

INSERT INTO `liquidation_config_tier` (
  `tier`, `tier_rank`, `safety_buffer_between_low_pct`, `safety_buffer_between_high_pct`,
  `exec_time_exceed_sec`, `consecutive_ioc_unfilled`, `price_drop_window_sec`, `price_drop_pct`,
  `bid_offset_btc_pct`, `bid_offset_eth_pct`, `bid_offset_default_pct`, `quote_interval_sec`
) VALUES
  ('P1', 1, 0.08000000, 0.15000000, 15, 3, 5, 0.00500000, 0.00800000, 0.00800000, 0.00800000, 1)
ON DUPLICATE KEY UPDATE `updated_at` = CURRENT_TIMESTAMP(3);

INSERT INTO `liquidation_config_tier` (
  `tier`, `tier_rank`, `safety_buffer_gte_pct`, `exec_time_under_sec`,
  `bid_offset_btc_pct`, `bid_offset_eth_pct`, `bid_offset_default_pct`, `quote_interval_sec`
) VALUES
  ('P2', 2, 0.15000000, 15, 0.00300000, 0.00300000, 0.00300000, 1)
ON DUPLICATE KEY UPDATE `updated_at` = CURRENT_TIMESTAMP(3);

-- ---------------------------------------------------------------------------
-- 9. 对接平台（sign + 回调地址）
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `liquidation_platform_config` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `platform_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '对接平台名称',
  `sign_key` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'HMAC sign 密钥',
  `bn_api_key` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'Binance 现货 API Key',
  `bn_api_secret` varchar(512) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'Binance 现货 API Secret',
  `callback_url` varchar(512) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '终态 POST 回调',
  `is_enabled` tinyint(1) NOT NULL DEFAULT '1',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='清算对接平台配置';

INSERT INTO `liquidation_platform_config` (`platform_name`, `sign_key`, `callback_url`, `is_enabled`)
SELECT 'BeTrust', 'a8f2c91e04b3d7651f0e6a2c9b4d8f1e3c7a5926b0d4e8f2a6c1b9d3e7f5a0', 'https://betrust.example.com/liquidation/callback', 1 FROM DUAL
WHERE NOT EXISTS (
  SELECT 1 FROM `liquidation_platform_config` WHERE `platform_name` = 'BeTrust'
);

-- ---------------------------------------------------------------------------
-- 10. 脚本调度（module=liquidation）
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

INSERT INTO `bot_schedule_config` (`module`, `task_name`, `interval_seconds`, `description`, `is_enabled`, `is_strategy_enabled`, `is_primary_machine_run`)
SELECT 'liquidation', 'task_tier_calc', 1, '重算 P0/P1/P2、last_mid（task_exec 内也会调，可关独立调度）', 1, 1, 0 FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM `bot_schedule_config` WHERE `module` = 'liquidation' AND `task_name` = 'task_tier_calc');

INSERT INTO `bot_schedule_config` (`module`, `task_name`, `interval_seconds`, `description`, `is_enabled`, `is_strategy_enabled`, `is_primary_machine_run`)
SELECT 'liquidation', 'task_alert', 10, '内部告警：同币对并发/名义、单任务 IOC 次数', 1, 1, 0 FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM `bot_schedule_config` WHERE `module` = 'liquidation' AND `task_name` = 'task_alert');

INSERT INTO `bot_schedule_config` (`module`, `task_name`, `interval_seconds`, `description`, `is_enabled`, `is_strategy_enabled`, `is_primary_machine_run`)
SELECT 'liquidation', 'task_exec', 2, '强平执行：按档位优先级发 IOC', 1, 1, 0 FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM `bot_schedule_config` WHERE `module` = 'liquidation' AND `task_name` = 'task_exec');

INSERT INTO `bot_schedule_config` (`module`, `task_name`, `interval_seconds`, `description`, `is_enabled`, `is_strategy_enabled`, `is_primary_machine_run`)
SELECT 'liquidation', 'result_callback', 5, '终态任务回调 BeTrust', 1, 1, 0 FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM `bot_schedule_config` WHERE `module` = 'liquidation' AND `task_name` = 'result_callback');

INSERT INTO `bot_schedule_config` (`module`, `task_name`, `interval_seconds`, `description`, `is_enabled`, `is_strategy_enabled`, `is_primary_machine_run`)
SELECT 'liquidation', 'order_updater', 10, 'unknown/超时 new 查单对齐 orders+traders', 1, 1, 0 FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM `bot_schedule_config` WHERE `module` = 'liquidation' AND `task_name` = 'order_updater');

INSERT INTO `bot_schedule_config` (`module`, `task_name`, `interval_seconds`, `description`, `is_enabled`, `is_strategy_enabled`, `is_primary_machine_run`)
SELECT 'liquidation', 'task_cancel', 5, 'paused/cancelled：撤在途 IOC', 1, 1, 0 FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM `bot_schedule_config` WHERE `module` = 'liquidation' AND `task_name` = 'task_cancel');

-- ---------------------------------------------------------------------------
-- 已有库：若缺 Binance 凭证列，按需执行（新库 CREATE 已含列则跳过）
-- ---------------------------------------------------------------------------
-- ALTER TABLE `liquidation_platform_config` DROP INDEX `uk_platform_code`;
-- ALTER TABLE `liquidation_platform_config` DROP COLUMN `platform_code`;
-- ALTER TABLE `liquidation_platform_config`
--   CHANGE COLUMN `api_key` `sign_key` varchar(255) NOT NULL COMMENT 'HMAC sign 密钥';
-- ALTER TABLE `liquidation_platform_config`
--   ADD COLUMN `bn_api_key` varchar(255) NOT NULL DEFAULT '' COMMENT 'Binance 现货 API Key' AFTER `sign_key`,
--   ADD COLUMN `bn_api_secret` varchar(512) NOT NULL DEFAULT '' COMMENT 'Binance 现货 API Secret' AFTER `bn_api_key`;
-- ALTER TABLE `liquidation_task`
--   ADD COLUMN `current_tier_reason_json` json DEFAULT NULL COMMENT '最近一次档位重算的判定输入与命中规则' AFTER `current_tier`;
-- ALTER TABLE `liquidation_task`
--   ADD COLUMN `in_flight_orders_cleared` int NOT NULL DEFAULT '0' COMMENT '在途挂单是否已清理完成 0否 1是' AFTER `cancel_result`;

SET FOREIGN_KEY_CHECKS = 1;
