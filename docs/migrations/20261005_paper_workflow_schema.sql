-- 科研工作流（Paper Agent）— 多学科 / 多 venue / 多文献源
-- 模块：选题发现、文献综述、实验规划、结果审查、论文写作、图表管理、环境配置
-- 引擎：MySQL 8.0+，utf8mb4；表前缀 paper_
-- user_id 与商城 users.id 对齐，不设 FK 便于独立部署
--
-- 表名规范：paper_{类型}[_{子实体}]
--   · 域前缀 paper_ = Paper Agent 全家桶
--   · {类型}       = 业务归类（见下表「类型」列），同类一眼可扫
--   · {子实体}     = 归属某类型的从表，如 run_stage 归属 run
--
-- | 类型       | 表名 |
-- |------------|------|
-- | ref        | paper_ref_discipline, paper_ref_venue, paper_ref_literature_source, paper_ref_execution_intensity, paper_ref_audit_level, paper_ref_prompt_template |
-- | manuscript | paper_manuscript, paper_manuscript_runtime, paper_manuscript_upload, paper_manuscript_milestone |
-- | user       | paper_user_preference |
-- | model      | paper_model_profile |
-- | run        | paper_run, paper_run_stage, paper_run_checkpoint, paper_run_literature_hit |
-- | literature | paper_literature_record |
-- | citation   | paper_citation_gate |
-- | artifact   | paper_artifact, paper_artifact_idea |
-- | llm        | paper_llm_call |
-- | audit      | paper_operation_log |

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ---------------------------------------------------------------------------
-- 1. ref 字典：学科、venue、文献源、强度与审计档位（七模块 code 见前端 agent/paper/types.ts）
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS `paper_ref_discipline` (
  `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `code`        VARCHAR(64)  NOT NULL COMMENT '学科键，如 cs_ai|medicine|law|economics',
  `name`        VARCHAR(128) NOT NULL COMMENT '展示名',
  `name_en`     VARCHAR(128) DEFAULT NULL,
  `sort`        INT          NOT NULL DEFAULT 0,
  `literature_source_codes` JSON         NOT NULL DEFAULT (JSON_ARRAY()) COMMENT '该学科推荐文献源 paper_ref_literature_source.code 列表，顺序即 UI 默认',
  `status`      VARCHAR(16)  NOT NULL DEFAULT 'active' COMMENT 'active|hidden',
  `created_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_paper_ref_discipline_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='学科';

CREATE TABLE IF NOT EXISTS `paper_ref_venue` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `code`            VARCHAR(64)  NOT NULL COMMENT 'venue 键，如 neurips|iclr|nature|ssci_q1',
  `name`            VARCHAR(256) NOT NULL COMMENT 'NeurIPS 2026 / Nature 等',
  `venue_type`      VARCHAR(16)  NOT NULL COMMENT 'conference|journal|workshop|other',
  `discipline_id`   BIGINT UNSIGNED DEFAULT NULL COMMENT 'NULL=跨学科通用',
  `publisher`       VARCHAR(128) DEFAULT NULL,
  `website_url`     VARCHAR(512) DEFAULT NULL,
  `rubric`          JSON         NOT NULL COMMENT 'venue 规则：contribution_types|experiment_bar|writing_style 等 prompt/结构化约束',
  `status`          VARCHAR(16)  NOT NULL DEFAULT 'active',
  `created_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_paper_ref_venue_code` (`code`),
  KEY `idx_paper_ref_venue_discipline` (`discipline_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='目标会议/期刊';

CREATE TABLE IF NOT EXISTS `paper_ref_literature_source` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `code`            VARCHAR(64)  NOT NULL COMMENT 'arxiv|openalex|semantic_scholar|pubmed|crossref|cnki|…',
  `name`            VARCHAR(128) NOT NULL,
  `api_kind`        VARCHAR(32)  NOT NULL COMMENT 'rest|oai|custom',
  `base_url`        VARCHAR(512) NOT NULL,
  `auth_type`       VARCHAR(32)  NOT NULL DEFAULT 'none' COMMENT 'none|api_key|oauth',
  `config_schema`   JSON         DEFAULT NULL COMMENT '连接参数 JSON Schema（Key 名、必填项）',
  `default_config`  JSON         DEFAULT NULL COMMENT '默认非密钥配置',
  `rate_limit_hint` VARCHAR(256) DEFAULT NULL,
  `status`          VARCHAR(16)  NOT NULL DEFAULT 'active',
  `created_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_paper_ref_literature_source_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文献/API 数据源';

CREATE TABLE IF NOT EXISTS `paper_ref_execution_intensity` (
  `code`        VARCHAR(16)  NOT NULL COMMENT 'fast|balanced|deep',
  `name`        VARCHAR(64)  NOT NULL,
  `multiplier`  DECIMAL(4,2) NOT NULL DEFAULT 1.00 COMMENT '相对检索量/迭代轮数系数',
  `config`      JSON         DEFAULT NULL,
  PRIMARY KEY (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='执行强度档位';

CREATE TABLE IF NOT EXISTS `paper_ref_audit_level` (
  `code`        VARCHAR(16)  NOT NULL COMMENT 'standard|polished|strict',
  `name`        VARCHAR(64)  NOT NULL,
  `config`      JSON         NOT NULL COMMENT 'citation|claim|kill_argument 强度与轮数',
  PRIMARY KEY (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='审计等级';

-- 模块 × venue × 学科 的 Prompt / rubric 模板（可版本化）
CREATE TABLE IF NOT EXISTS `paper_ref_prompt_template` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `module_code`     VARCHAR(32)  NOT NULL COMMENT 'topic_discovery|literature_review|…',
  `venue_id`        BIGINT UNSIGNED DEFAULT NULL COMMENT 'NULL=venue 无关通用模板',
  `discipline_id`   BIGINT UNSIGNED DEFAULT NULL COMMENT 'NULL=学科无关',
  `version`         INT          NOT NULL DEFAULT 1,
  `role`            VARCHAR(16)  NOT NULL DEFAULT 'system' COMMENT 'system|reviewer|retrieve_query',
  `template_body`   MEDIUMTEXT   NOT NULL,
  `variables`       JSON         DEFAULT NULL COMMENT '可替换变量说明',
  `status`          VARCHAR(16)  NOT NULL DEFAULT 'active',
  `created_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_paper_ref_prompt_lookup` (`module_code`, `venue_id`, `discipline_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='LLM 模板（venue 规则落地）';

-- ---------------------------------------------------------------------------
-- 2. 稿件线与用户默认（工作台「当前论文」= paper_manuscript）
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS `paper_manuscript` (
  `id`                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`           BIGINT UNSIGNED NOT NULL COMMENT 'users.id',
  `title`             VARCHAR(256) NOT NULL COMMENT '工作标题 / 暂定篇名',
  `description`       TEXT         DEFAULT NULL,
  `manuscript_kind`   VARCHAR(32)  DEFAULT NULL COMMENT 'conference|journal|thesis|course|report',
  `deadline_at`       DATETIME     DEFAULT NULL,
  `target_words`      INT UNSIGNED DEFAULT NULL,
  `citation_style`    VARCHAR(64)  DEFAULT NULL COMMENT 'apa|ieee|acm|…',
  `discipline_id`     BIGINT UNSIGNED DEFAULT NULL,
  `default_venue_id`  BIGINT UNSIGNED DEFAULT NULL,
  `reference_gate_enabled` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '参考文献门禁',
  `workspace_uri`     VARCHAR(512) DEFAULT NULL COMMENT '平台 provision 的工作区根路径',
  `forked_from_run_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '从选题等 run fork 出新稿时链接',
  `status`            VARCHAR(16)  NOT NULL DEFAULT 'active' COMMENT 'active|archived',
  `created_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_paper_manuscript_user` (`user_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='单篇论文稿件线（数据隔离单元）';

CREATE TABLE IF NOT EXISTS `paper_user_preference` (
  `user_id`                   BIGINT UNSIGNED NOT NULL,
  `default_discipline_id`     BIGINT UNSIGNED DEFAULT NULL,
  `default_venue_id`          BIGINT UNSIGNED DEFAULT NULL,
  `default_intensity_code`    VARCHAR(16)  NOT NULL DEFAULT 'balanced',
  `default_audit_level_code`  VARCHAR(16)  NOT NULL DEFAULT 'polished',
  `default_human_checkpoint`  TINYINT(1)   NOT NULL DEFAULT 1,
  `default_literature_source_ids` JSON     DEFAULT NULL COMMENT '默认文献源 id 数组',
  `updated_at`                DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户全局默认（环境配置模块可写）';

-- ---------------------------------------------------------------------------
-- 3. 平台运行时（模型、算力、文献源凭证 — 不对用户暴露）
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS `paper_model_profile` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`         BIGINT UNSIGNED NOT NULL,
  `label`           VARCHAR(128) NOT NULL COMMENT '如「公司网关 Codex」',
  `role`            VARCHAR(16)  NOT NULL COMMENT 'executor|reviewer|both',
  `provider_code`   VARCHAR(32)  NOT NULL COMMENT 'openai|anthropic|gateway|ollama|…',
  `model_name`      VARCHAR(128) NOT NULL,
  `base_url`        VARCHAR(512) DEFAULT NULL,
  `api_key_ciphertext` VARBINARY(2048) DEFAULT NULL COMMENT '或走 user_api_keys 仅存 id',
  `user_api_key_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '商城平台 Key，可选',
  `extra`           JSON         DEFAULT NULL,
  `status`          VARCHAR(16)  NOT NULL DEFAULT 'active',
  `created_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_paper_model_profile_user` (`user_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='模型 Provider 配置';

CREATE TABLE IF NOT EXISTS `paper_manuscript_runtime` (
  `id`                    BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `manuscript_id`         BIGINT UNSIGNED NOT NULL,
  `label`                 VARCHAR(128) NOT NULL DEFAULT 'default',
  `executor_profile_id`   BIGINT UNSIGNED DEFAULT NULL,
  `reviewer_profile_id`   BIGINT UNSIGNED DEFAULT NULL,
  `literature_credentials` JSON        DEFAULT NULL COMMENT 'source_id -> 密钥密文/引用 id',
  `literature_source_ids` JSON        NOT NULL COMMENT '本稿件启用的 source id 列表',
  `gpu_profile`           JSON         DEFAULT NULL COMMENT 'SSH/集群/本地',
  `wiki_uri`              VARCHAR(512) DEFAULT NULL COMMENT 'Research Wiki 路径',
  `is_active`             TINYINT(1)   NOT NULL DEFAULT 1,
  `created_at`            DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`            DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_paper_manuscript_runtime_ms` (`manuscript_id`, `is_active`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='稿件平台运行时绑定（内部）';

-- ---------------------------------------------------------------------------
-- 4. 运行：一次「运行」= 某模块一条流水线
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS `paper_run` (
  `id`                      BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `manuscript_id`           BIGINT UNSIGNED NOT NULL,
  `user_id`                 BIGINT UNSIGNED NOT NULL,
  `module_code`             VARCHAR(32)  NOT NULL COMMENT 'topic_discovery|literature_review|experiment_planning|auto_review|paper_writing|figure_generation|manuscript_analysis',
  `parent_run_id`           BIGINT UNSIGNED DEFAULT NULL COMMENT '上游模块产出触发下游时链接',
  `status`                  VARCHAR(32)  NOT NULL DEFAULT 'pending'
    COMMENT 'pending|running|checkpoint|completed|failed|cancelled',
  `discipline_id`           BIGINT UNSIGNED DEFAULT NULL,
  `venue_id`                BIGINT UNSIGNED DEFAULT NULL,
  `venue_snapshot`          JSON         DEFAULT NULL COMMENT '运行时 rubric 快照',
  `intensity_code`          VARCHAR(16)  NOT NULL DEFAULT 'balanced',
  `audit_level_code`        VARCHAR(16)  NOT NULL DEFAULT 'polished',
  `human_checkpoint_enabled` TINYINT(1)  NOT NULL DEFAULT 1,
  `literature_source_ids`   JSON         NOT NULL COMMENT '本次运行使用的文献源',
  `input_params`            JSON         NOT NULL COMMENT '模块表单全量：研究方向等，schema 由代码/学科配置约束',
  `executor_profile_id`     BIGINT UNSIGNED DEFAULT NULL,
  `reviewer_profile_id`     BIGINT UNSIGNED DEFAULT NULL,
  `error_code`              VARCHAR(64)  DEFAULT NULL,
  `error_message`           TEXT         DEFAULT NULL,
  `started_at`              DATETIME     DEFAULT NULL,
  `finished_at`             DATETIME     DEFAULT NULL,
  `created_at`              DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`              DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_paper_run_manuscript` (`manuscript_id`, `created_at`),
  KEY `idx_paper_run_user_module` (`user_id`, `module_code`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工作流运行实例';

CREATE TABLE IF NOT EXISTS `paper_run_stage` (
  `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `run_id`      BIGINT UNSIGNED NOT NULL,
  `stage_code`  VARCHAR(64)  NOT NULL COMMENT 'topic_discovery: retrieve|generate_ideas|novelty|audit；其他 module 自定义',
  `status`      VARCHAR(32)  NOT NULL DEFAULT 'pending' COMMENT 'pending|running|completed|failed|skipped',
  `sort`        INT          NOT NULL DEFAULT 0,
  `meta`        JSON         DEFAULT NULL COMMENT 'token 用量、耗时等',
  `started_at`  DATETIME     DEFAULT NULL,
  `finished_at` DATETIME     DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_paper_run_stage_run` (`run_id`, `sort`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='运行内阶段（RAG/生成/审计）';

CREATE TABLE IF NOT EXISTS `paper_run_checkpoint` (
  `id`                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `run_id`            BIGINT UNSIGNED NOT NULL,
  `stage_id`          BIGINT UNSIGNED DEFAULT NULL,
  `checkpoint_key`    VARCHAR(64)  NOT NULL COMMENT 'topic: ideas_ready；experiment_planning: plan_ready（可选）',
  `status`            VARCHAR(16)  NOT NULL DEFAULT 'waiting' COMMENT 'waiting|approved|rejected|skipped',
  `payload`           JSON         NOT NULL COMMENT '展示给用户的中間产物摘要/全文引用',
  `user_note`         TEXT         DEFAULT NULL,
  `resolved_by_user_id` BIGINT UNSIGNED DEFAULT NULL,
  `resolved_at`       DATETIME     DEFAULT NULL,
  `created_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_paper_run_checkpoint_run` (`run_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='人工检查点';

-- ---------------------------------------------------------------------------
-- 5. RAG：检索文献与引用门禁
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS `paper_literature_record` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `source_id`       BIGINT UNSIGNED NOT NULL,
  `external_key`    VARCHAR(256) NOT NULL COMMENT 'arxiv:2401.12345 / doi:… / s2:…',
  `title`           VARCHAR(1024) NOT NULL DEFAULT '',
  `authors`         JSON         DEFAULT NULL,
  `abstract`        TEXT         DEFAULT NULL,
  `published_year`  SMALLINT     DEFAULT NULL,
  `doi`             VARCHAR(128) DEFAULT NULL,
  `url`             VARCHAR(512) DEFAULT NULL,
  `raw_payload`     JSON         DEFAULT NULL,
  `fetched_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_paper_literature_record` (`source_id`, `external_key`),
  KEY `idx_paper_literature_record_doi` (`doi`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文献库缓存（跨 run 复用）';

CREATE TABLE IF NOT EXISTS `paper_run_literature_hit` (
  `run_id`              BIGINT UNSIGNED NOT NULL,
  `literature_record_id` BIGINT UNSIGNED NOT NULL,
  `stage_id`            BIGINT UNSIGNED DEFAULT NULL,
  `relevance_score`     DECIMAL(6,4) DEFAULT NULL,
  `query_text`          VARCHAR(512) DEFAULT NULL,
  `created_at`          DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`run_id`, `literature_record_id`),
  KEY `idx_paper_run_lit_record` (`literature_record_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='某次运行检索命中的文献';

CREATE TABLE IF NOT EXISTS `paper_citation_gate` (
  `id`                  BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `run_id`              BIGINT UNSIGNED NOT NULL,
  `artifact_id`         BIGINT UNSIGNED DEFAULT NULL,
  `literature_record_id` BIGINT UNSIGNED DEFAULT NULL,
  `cited_key`           VARCHAR(256) DEFAULT NULL COMMENT '正文中的 cite key',
  `gate_type`           VARCHAR(32)  NOT NULL COMMENT 'reference|citation_audit|claim_audit|kill_argument',
  `status`              VARCHAR(16)  NOT NULL COMMENT 'verified|rejected|pending|waived',
  `detail`              JSON         DEFAULT NULL,
  `created_at`          DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_paper_citation_gate_run` (`run_id`, `gate_type`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='参考文献/审计门禁记录';

-- ---------------------------------------------------------------------------
-- 6. 产出物（idea 报告、综述、计划、审稿意见、稿、图）
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS `paper_artifact` (
  `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `run_id`        BIGINT UNSIGNED NOT NULL,
  `artifact_type` VARCHAR(32)  NOT NULL COMMENT
    'idea_report|novelty_report|literature_review|experiment_plan|result_review_report|manuscript_analysis_report|review_report|manuscript|figure|bibtex|other',
  `format`        VARCHAR(16)  NOT NULL COMMENT 'md|json|tex|pdf|png|svg|bib',
  `title`         VARCHAR(256) DEFAULT NULL,
  `storage_uri`   VARCHAR(1024) DEFAULT NULL COMMENT '对象存储或本地相对路径',
  `content_medium` MEDIUMTEXT   DEFAULT NULL COMMENT '小文件可内联',
  `content_hash`  CHAR(64)     DEFAULT NULL COMMENT 'SHA-256',
  `meta`          JSON         DEFAULT NULL COMMENT 'idea 条数、venue、版本号等',
  `created_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_paper_artifact_run` (`run_id`, `artifact_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='运行产出';

CREATE TABLE IF NOT EXISTS `paper_artifact_idea` (
  `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `run_id`        BIGINT UNSIGNED NOT NULL,
  `artifact_id`   BIGINT UNSIGNED DEFAULT NULL,
  `rank_no`       INT          NOT NULL DEFAULT 0,
  `title`         VARCHAR(512) NOT NULL,
  `problem`       TEXT         DEFAULT NULL,
  `approach`      TEXT         DEFAULT NULL,
  `contribution`  TEXT         DEFAULT NULL,
  `novelty_summary` TEXT       DEFAULT NULL,
  `novelty_risk`  VARCHAR(16)  DEFAULT NULL COMMENT 'low|medium|high',
  `status`        VARCHAR(16)  NOT NULL DEFAULT 'candidate' COMMENT 'candidate|selected|rejected',
  `meta`          JSON         DEFAULT NULL,
  `created_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_paper_artifact_idea_run` (`run_id`, `rank_no`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='选题发现：结构化 idea 行';

-- ---------------------------------------------------------------------------
-- 7. LLM 调用日志（计费、排错；不存完整 prompt 时可只存 hash）
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS `paper_llm_call` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `run_id`          BIGINT UNSIGNED NOT NULL,
  `stage_id`        BIGINT UNSIGNED DEFAULT NULL,
  `profile_id`      BIGINT UNSIGNED DEFAULT NULL,
  `role`            VARCHAR(16)  NOT NULL COMMENT 'executor|reviewer',
  `model_name`      VARCHAR(128) NOT NULL,
  `prompt_tokens`   INT          DEFAULT NULL,
  `completion_tokens` INT        DEFAULT NULL,
  `latency_ms`      INT          NOT NULL DEFAULT 0,
  `status`          SMALLINT     NOT NULL COMMENT 'HTTP 或业务码',
  `created_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_paper_llm_call_run` (`run_id`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='LLM 调用明细';

-- ---------------------------------------------------------------------------
-- 7b. 上传、里程碑、操作日志（与 20261007 增量一致；新库一次建全）
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS `paper_manuscript_upload` (
  `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `manuscript_id` BIGINT UNSIGNED NOT NULL,
  `user_id`       BIGINT UNSIGNED NOT NULL COMMENT 'users.id',
  `upload_scope`  VARCHAR(16)  NOT NULL COMMENT 'reference|figure',
  `file_kind`     VARCHAR(16)  NOT NULL COMMENT 'pdf|bib|txt|image|vector|other',
  `file_name`     VARCHAR(512) NOT NULL,
  `title`         VARCHAR(512) NOT NULL,
  `note`          TEXT         DEFAULT NULL,
  `size_bytes`    BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `storage_uri`   VARCHAR(1024) DEFAULT NULL,
  `content_hash`  CHAR(64)     DEFAULT NULL,
  `status`        VARCHAR(16)  NOT NULL DEFAULT 'active' COMMENT 'active|deleted',
  `created_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_paper_ms_upload_ms` (`manuscript_id`, `upload_scope`, `status`),
  KEY `idx_paper_ms_upload_user` (`user_id`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='本篇上传文件（文献库/figure 资产）';

CREATE TABLE IF NOT EXISTS `paper_manuscript_milestone` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `manuscript_id`   BIGINT UNSIGNED NOT NULL,
  `milestone_code`  VARCHAR(64)  NOT NULL COMMENT
    'topic_discovery|literature_review|experiment_planning|auto_review|paper_writing|figure_generation|manuscript_analysis',
  `run_id`          BIGINT UNSIGNED DEFAULT NULL,
  `completed_at`    DATETIME     NOT NULL,
  `meta`            JSON         DEFAULT NULL,
  `created_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_paper_ms_milestone` (`manuscript_id`, `milestone_code`),
  KEY `idx_paper_ms_milestone_run` (`run_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='稿件工作流里程碑';

CREATE TABLE IF NOT EXISTS `paper_operation_log` (
  `id`                 BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`            BIGINT UNSIGNED NOT NULL,
  `manuscript_id`      BIGINT UNSIGNED DEFAULT NULL,
  `manuscript_title`   VARCHAR(256) DEFAULT NULL,
  `module_code`        VARCHAR(32)  NOT NULL,
  `module_label`       VARCHAR(64)  NOT NULL,
  `action`             VARCHAR(512) NOT NULL,
  `tokens_prompt`      INT UNSIGNED NOT NULL DEFAULT 0,
  `tokens_completion`  INT UNSIGNED NOT NULL DEFAULT 0,
  `tokens_total`       INT UNSIGNED NOT NULL DEFAULT 0,
  `status`             VARCHAR(16)  NOT NULL COMMENT 'success|failed|cancelled',
  `note`               TEXT         DEFAULT NULL,
  `occurred_at`        DATETIME     NOT NULL,
  `created_at`         DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_paper_oplog_user` (`user_id`, `occurred_at`),
  KEY `idx_paper_oplog_ms` (`manuscript_id`, `occurred_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Paper Agent 操作日志';

-- ---------------------------------------------------------------------------
-- 8. 种子数据（强度、审计、示例学科与 venue、文献源）
-- ---------------------------------------------------------------------------

INSERT INTO `paper_ref_execution_intensity` (`code`, `name`, `multiplier`, `config`) VALUES
  ('fast',     '更快', 0.60, JSON_OBJECT('max_papers', 30,  'max_ideas', 6)),
  ('balanced', 'Balanced（平衡）', 1.00, JSON_OBJECT('max_papers', 80,  'max_ideas', 12)),
  ('deep',     '更深', 1.80, JSON_OBJECT('max_papers', 200, 'max_ideas', 20))
ON DUPLICATE KEY UPDATE `name` = VALUES(`name`);

INSERT INTO `paper_ref_audit_level` (`code`, `name`, `config`) VALUES
  ('standard', 'Standard', JSON_OBJECT('citation', 1, 'claim', 1, 'kill_argument', 0, 'rounds', 1)),
  ('polished', 'Polished（精修）', JSON_OBJECT('citation', 2, 'claim', 2, 'kill_argument', 1, 'rounds', 2)),
  ('strict',   'Strict', JSON_OBJECT('citation', 3, 'claim', 3, 'kill_argument', 2, 'rounds', 3))
ON DUPLICATE KEY UPDATE `name` = VALUES(`name`);

INSERT INTO `paper_ref_discipline` (`code`, `name`, `name_en`, `sort`, `literature_source_codes`) VALUES
  ('cs_ai', '计算机/人工智能', 'Computer Science & AI', 10,
   JSON_ARRAY('arxiv', 'openalex', 'semantic_scholar')),
  ('general', '跨学科通用', 'General', 0, JSON_ARRAY('openalex', 'crossref'))
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `literature_source_codes` = VALUES(`literature_source_codes`);

INSERT INTO `paper_ref_venue` (`code`, `name`, `venue_type`, `discipline_id`, `rubric`) 
SELECT 'ml_top3', 'NeurIPS/ICLR/ICML', 'conference', d.id,
  JSON_OBJECT(
    'contribution_types', JSON_ARRAY('method', 'theory', 'empirical'),
    'experiment_bar', 'strong_baselines_ablations',
    'writing_style', 'ml_conference'
  )
FROM `paper_ref_discipline` d WHERE d.code = 'cs_ai' LIMIT 1
ON DUPLICATE KEY UPDATE `name` = VALUES(`name`);

INSERT INTO `paper_ref_literature_source` (`code`, `name`, `api_kind`, `base_url`, `auth_type`) VALUES
  ('arxiv', 'arXiv', 'rest', 'https://export.arxiv.org/api/query', 'none'),
  ('openalex', 'OpenAlex', 'rest', 'https://api.openalex.org', 'none'),
  ('semantic_scholar', 'Semantic Scholar', 'rest', 'https://api.semanticscholar.org/graph/v1', 'api_key')
ON DUPLICATE KEY UPDATE `name` = VALUES(`name`);

SET FOREIGN_KEY_CHECKS = 1;
