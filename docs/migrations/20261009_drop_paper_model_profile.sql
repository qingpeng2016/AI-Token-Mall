-- 移除已废弃的 paper_model_profile，统一使用 paper_llm_model_config + paper_llm_workflow_binding
-- 依赖：20261005；若已执行 20261008，请先完成 paper_llm_call_logs 更名
-- 若从未创建 paper_model_profile（新基线 20261005），可跳过本脚本

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- paper_run：profile → model_config 快照字段
ALTER TABLE `paper_run`
  ADD COLUMN `executor_model_config_id` BIGINT UNSIGNED DEFAULT NULL COMMENT 'executor 快照' AFTER `input_params`,
  ADD COLUMN `reviewer_model_config_id` BIGINT UNSIGNED DEFAULT NULL COMMENT 'reviewer 快照' AFTER `executor_model_config_id`;

-- 若新基线已含上述列，请跳过 ADD；下面 DROP 旧列若不存在则跳过
ALTER TABLE `paper_run`
  DROP COLUMN `executor_profile_id`,
  DROP COLUMN `reviewer_profile_id`;

ALTER TABLE `paper_manuscript_runtime`
  DROP COLUMN `executor_profile_id`,
  DROP COLUMN `reviewer_profile_id`;

ALTER TABLE `paper_llm_call_logs`
  DROP COLUMN `profile_id`;

DROP TABLE IF EXISTS `paper_model_profile`;

SET FOREIGN_KEY_CHECKS = 1;
