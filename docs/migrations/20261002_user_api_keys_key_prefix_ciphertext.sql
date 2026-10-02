-- 展示前缀 + 加密存储明文（会员中心复制 Key）

ALTER TABLE `user_api_keys`
  ADD COLUMN `key_prefix` VARCHAR(16) NOT NULL DEFAULT ''
    COMMENT 'Key 前缀用于脱敏展示' AFTER `key_hash`,
  ADD COLUMN `key_ciphertext` VARBINARY(512) DEFAULT NULL
    COMMENT 'AES-GCM 加密的完整 Key，供归属用户复制' AFTER `key_prefix`;

UPDATE `user_api_keys`
SET `key_prefix` = CONCAT('ap-', LEFT(`key_hash`, 4))
WHERE `key_prefix` = '' OR `key_prefix` IS NULL;
