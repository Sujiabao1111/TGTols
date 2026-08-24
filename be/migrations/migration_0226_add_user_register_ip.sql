-- ============================================
-- Migration: migration_0226_add_user_register_ip.sql
-- Description: add user registration IP for admin duplicate-IP checks
-- ============================================

SET @add_register_ip = (
  SELECT IF(
    EXISTS(
      SELECT 1
      FROM INFORMATION_SCHEMA.COLUMNS
      WHERE TABLE_SCHEMA = DATABASE()
        AND TABLE_NAME = 'users'
        AND COLUMN_NAME = 'register_ip'
    ),
    'SELECT 1',
    'ALTER TABLE users ADD COLUMN register_ip VARCHAR(64) NOT NULL DEFAULT '''' AFTER last_login_ip'
  )
);
PREPARE stmt FROM @add_register_ip;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

UPDATE users
SET register_ip = last_login_ip
WHERE (register_ip IS NULL OR register_ip = '')
  AND last_login_ip IS NOT NULL
  AND last_login_ip <> '';

SET @add_register_ip_index = (
  SELECT IF(
    EXISTS(
      SELECT 1
      FROM INFORMATION_SCHEMA.STATISTICS
      WHERE TABLE_SCHEMA = DATABASE()
        AND TABLE_NAME = 'users'
        AND INDEX_NAME = 'idx_users_register_ip'
    ),
    'SELECT 1',
    'ALTER TABLE users ADD INDEX idx_users_register_ip (register_ip)'
  )
);
PREPARE stmt FROM @add_register_ip_index;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
