-- Paper Agent — LLM 模型配置 + 工作流环节绑定
-- · paper_llm_model_config：Provider、model、Base URL、Key、超时等连接参数
-- · paper_llm_workflow_binding：module / stage / 任务角色 → 使用哪条 model_config
-- 依赖：20261005（及可选 20261007）
-- 说明：run 按 paper_llm_workflow_binding 解析 model_config，写入 paper_llm_call_logs.model_config_id
-- 引擎：MySQL 8.0+，utf8mb4

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ---------------------------------------------------------------------------
-- 1. 模型连接配置（Key、Host、路径、默认参数等）
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS `paper_llm_model_config` (
  `id`                   BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`              BIGINT UNSIGNED NOT NULL COMMENT 'users.id；0=平台预置模板（只读引用）',
  `label`                VARCHAR(128) NOT NULL COMMENT '展示名，如「Claude Sonnet · 写作」',
  `provider_code`        VARCHAR(32)  NOT NULL COMMENT 'openai|anthropic|azure_openai|openai_compatible|ollama|gateway|other',
  `model_name`           VARCHAR(128) NOT NULL COMMENT '上游 model 参数，如 claude-sonnet-4-6、gpt-5.4',
  `api_base_url`         VARCHAR(512) NOT NULL COMMENT 'Host / Base URL，如 https://api.openai.com/v1',
  `api_path_chat`        VARCHAR(128) DEFAULT NULL COMMENT '可选：/chat/completions 等相对路径，NULL=Provider 默认',
  `api_key_ciphertext`   VARBINARY(4096) DEFAULT NULL COMMENT '本地/自建 Key 密文；与 user_api_key_id 二选一或并存（优先商城 Key）',
  `api_key_header`       VARCHAR(64)  NOT NULL DEFAULT 'Authorization' COMMENT '如 Authorization、x-api-key',
  `api_key_prefix`       VARCHAR(32)  DEFAULT 'Bearer ' COMMENT 'Key 前缀，空字符串表示裸 Key',
  `user_api_key_id`      BIGINT UNSIGNED DEFAULT NULL COMMENT '商城 user_api_keys.id，走 Token 网关时优先',
  `default_headers`      JSON         DEFAULT NULL COMMENT '额外 HTTP 头',
  `default_params`       JSON         DEFAULT NULL COMMENT '默认 temperature、max_tokens、top_p 等',
  `timeout_ms`           INT UNSIGNED NOT NULL DEFAULT 120000,
  `max_retries`          TINYINT UNSIGNED NOT NULL DEFAULT 2,
  `supports_vision`      TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '读 PDF/图时使用',
  `context_window_hint`  INT UNSIGNED DEFAULT NULL COMMENT '上下文长度提示（token），用于路由',
  `extra`                JSON         DEFAULT NULL COMMENT '代理、组织 ID、Azure deployment 等杂项',
  `status`               VARCHAR(16)  NOT NULL DEFAULT 'active' COMMENT 'active|disabled',
  `created_at`           DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`           DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_paper_llm_model_cfg_user` (`user_id`, `status`),
  KEY `idx_paper_llm_model_cfg_provider` (`provider_code`, `model_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='LLM 端点配置（Key/Host/模型名）';

-- ---------------------------------------------------------------------------
-- 2. 工作流环节 → 模型配置（什么环节用哪个 AI）
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS `paper_llm_workflow_binding` (
  `id`                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`           BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '0=平台默认；否则用户覆盖',
  `manuscript_id`     BIGINT UNSIGNED DEFAULT NULL COMMENT '非空=仅本篇论文 override',
  `module_code`       VARCHAR(32)  NOT NULL COMMENT
    'topic_discovery|literature_review|experiment_planning|auto_review|paper_writing|figure_generation|manuscript_analysis|*',
  `stage_code`        VARCHAR(64)  DEFAULT NULL COMMENT
    'topic: retrieve|generate_ideas|novelty|audit；NULL=整模块默认',
  `llm_role`          VARCHAR(32)  NOT NULL COMMENT
    'executor|reviewer|ingest|system|citation_audit|claim_audit|kill_argument',
  `intensity_code`    VARCHAR(16)  DEFAULT NULL COMMENT 'fast|balanced|deep；NULL=全部强度',
  `audit_level_code`  VARCHAR(16)  DEFAULT NULL COMMENT 'standard|polished|strict；NULL=全部',
  `model_config_id`   BIGINT UNSIGNED NOT NULL COMMENT 'paper_llm_model_config.id',
  `priority`          INT          NOT NULL DEFAULT 100 COMMENT '同键多条时越小越优先；fallback 链',
  `status`            VARCHAR(16)  NOT NULL DEFAULT 'active' COMMENT 'active|disabled',
  `note`              VARCHAR(256) DEFAULT NULL,
  `created_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_paper_llm_bind_lookup` (
    `user_id`,
    `module_code`,
    `stage_code`,
    `llm_role`,
    `status`,
    `priority`
  ),
  KEY `idx_paper_llm_bind_ms` (`manuscript_id`, `module_code`),
  KEY `idx_paper_llm_bind_model` (`model_config_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='环节/阶段/角色 → LLM 配置';

-- ---------------------------------------------------------------------------
-- 3. LLM 调用日志表更名 + 配置溯源列（旧库曾用 paper_llm_call）
-- ---------------------------------------------------------------------------

RENAME TABLE `paper_llm_call` TO `paper_llm_call_logs`;

ALTER TABLE `paper_llm_call_logs`
  ADD COLUMN `model_config_id` BIGINT UNSIGNED DEFAULT NULL COMMENT 'paper_llm_model_config.id' AFTER `stage_id`,
  ADD COLUMN `workflow_binding_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '命中的 paper_llm_workflow_binding.id' AFTER `model_config_id`;

-- 若已是 paper_llm_call_logs 或列已存在，请跳过 RENAME / ALTER

-- ---------------------------------------------------------------------------
-- 4. 平台默认 binding 占位（model_config_id 需接入后替换为真实 id）
-- ---------------------------------------------------------------------------

-- 不在 migration 里 INSERT 具体 model_config_id，避免空库 FK 语义混乱。
-- 应用启动时：若 user_id=0 无 binding，则报错或回退环境变量默认 model_config。

SET FOREIGN_KEY_CHECKS = 1;
