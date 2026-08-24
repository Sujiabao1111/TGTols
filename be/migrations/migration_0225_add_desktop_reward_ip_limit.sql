-- ============================================
-- Migration: migration_0225_add_desktop_reward_ip_limit.sql
-- Description: limit Add Desktop reward to one claim per client IP
-- ============================================

SET @add_add_desktop_claim_ip = (
  SELECT IF(
    EXISTS(
      SELECT 1
      FROM INFORMATION_SCHEMA.COLUMNS
      WHERE TABLE_SCHEMA = DATABASE()
        AND TABLE_NAME = 'add_desktop_reward_claims'
        AND COLUMN_NAME = 'claim_ip'
    ),
    'SELECT 1',
    'ALTER TABLE add_desktop_reward_claims ADD COLUMN claim_ip VARCHAR(64) NULL DEFAULT NULL COMMENT ''claim IP'' AFTER user_id'
  )
);
PREPARE stmt FROM @add_add_desktop_claim_ip;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @add_add_desktop_claim_ip_unique = (
  SELECT IF(
    EXISTS(
      SELECT 1
      FROM INFORMATION_SCHEMA.STATISTICS
      WHERE TABLE_SCHEMA = DATABASE()
        AND TABLE_NAME = 'add_desktop_reward_claims'
        AND INDEX_NAME = 'uk_add_desktop_claim_ip'
    ),
    'SELECT 1',
    'ALTER TABLE add_desktop_reward_claims ADD UNIQUE KEY uk_add_desktop_claim_ip (claim_ip)'
  )
);
PREPARE stmt FROM @add_add_desktop_claim_ip_unique;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
