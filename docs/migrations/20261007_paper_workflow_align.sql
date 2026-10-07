-- Paper Agent  schema 对齐（产品 2026-10-07）
-- · 选题 run 阶段不含 experiment_plan；检查点 ideas_ready（plan 在 experiment_planning run）
-- · 续跑链：topic_discovery → literature_review → experiment_planning（parent_run_id）
-- · 上传文献 / 图表、稿件里程碑、操作日志落库
-- 依赖：已执行 20261005_paper_workflow_schema.sql
-- 引擎：MySQL 8.0+，utf8mb4

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ---------------------------------------------------------------------------
-- 1. 注释与约定（与 agent/paper/types.ts、ARIS 产品说明一致）
-- ---------------------------------------------------------------------------

ALTER TABLE `paper_run`
  MODIFY COLUMN `module_code` VARCHAR(32) NOT NULL COMMENT
    'topic_discovery|literature_review|experiment_planning|auto_review|paper_writing|figure_generation|manuscript_analysis';

ALTER TABLE `paper_run_stage`
  MODIFY COLUMN `stage_code` VARCHAR(64) NOT NULL COMMENT
    'topic_discovery: retrieve|generate_ideas|novelty|audit；其他 module 自定义 stage';

ALTER TABLE `paper_run_checkpoint`
  MODIFY COLUMN `checkpoint_key` VARCHAR(64) NOT NULL COMMENT
    'topic_discovery: ideas_ready；experiment_planning: plan_ready（可选）；其他模块自定义';

ALTER TABLE `paper_ref_prompt_template`
  MODIFY COLUMN `module_code` VARCHAR(32) NOT NULL COMMENT
    'topic_discovery|literature_review|experiment_planning|auto_review|paper_writing|figure_generation|manuscript_analysis';

ALTER TABLE `paper_artifact`
  MODIFY COLUMN `artifact_type` VARCHAR(32) NOT NULL COMMENT
    'idea_report|novelty_report|literature_review|experiment_plan|result_review_report|manuscript_analysis_report|manuscript|figure|bibtex|other';

-- review_report 保留兼容旧数据；新写入：auto_review → result_review_report，manuscript_analysis → manuscript_analysis_report

-- ---------------------------------------------------------------------------
-- 2. 稿件元数据（任务边界 · 可选）
-- ---------------------------------------------------------------------------

ALTER TABLE `paper_manuscript`
  ADD COLUMN `manuscript_kind` VARCHAR(32) DEFAULT NULL COMMENT 'conference|journal|thesis|course|report' AFTER `description`,
  ADD COLUMN `deadline_at` DATETIME DEFAULT NULL COMMENT '截止/答辩节点（可选）' AFTER `manuscript_kind`,
  ADD COLUMN `target_words` INT UNSIGNED DEFAULT NULL COMMENT '目标字数（可选）' AFTER `deadline_at`,
  ADD COLUMN `citation_style` VARCHAR(64) DEFAULT NULL COMMENT 'apa|ieee|acm|…（可选）' AFTER `target_words`;

-- 若列已存在，请跳过本节或手工合并

-- ---------------------------------------------------------------------------
-- 3. 用户上传：上传文献 + 图表管理·上传图表
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
  `storage_uri`   VARCHAR(1024) DEFAULT NULL COMMENT '对象存储或工作区相对路径',
  `content_hash`  CHAR(64)     DEFAULT NULL COMMENT 'SHA-256，可选',
  `status`        VARCHAR(16)  NOT NULL DEFAULT 'active' COMMENT 'active|deleted',
  `created_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_paper_ms_upload_ms` (`manuscript_id`, `upload_scope`, `status`),
  KEY `idx_paper_ms_upload_user` (`user_id`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='本篇上传文件（文献库/figure 资产）';

-- ---------------------------------------------------------------------------
-- 4. 主流程里程碑（替代前端 localStorage 完成标记）
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS `paper_manuscript_milestone` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `manuscript_id`   BIGINT UNSIGNED NOT NULL,
  `milestone_code`  VARCHAR(64)  NOT NULL COMMENT
    'topic_discovery|literature_review|experiment_planning|auto_review|paper_writing|figure_generation|manuscript_analysis',
  `run_id`          BIGINT UNSIGNED DEFAULT NULL COMMENT '完成时关联 paper_run.id',
  `completed_at`    DATETIME     NOT NULL,
  `meta`            JSON         DEFAULT NULL,
  `created_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_paper_ms_milestone` (`manuscript_id`, `milestone_code`),
  KEY `idx_paper_ms_milestone_run` (`run_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='稿件工作流里程碑（顶栏续跑状态）';

-- ---------------------------------------------------------------------------
-- 5. 操作日志（个人中心 · 操作日志 Tab）
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS `paper_operation_log` (
  `id`                 BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`            BIGINT UNSIGNED NOT NULL,
  `manuscript_id`      BIGINT UNSIGNED DEFAULT NULL,
  `manuscript_title`   VARCHAR(256) DEFAULT NULL COMMENT '快照，便于列表展示',
  `module_code`        VARCHAR(32)  NOT NULL COMMENT '同 paper_run.module_code 或 system|environment',
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Paper Agent 操作与 Token 消耗日志';

SET FOREIGN_KEY_CHECKS = 1;
