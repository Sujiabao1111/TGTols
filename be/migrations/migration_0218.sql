-- ============================================
-- Migration: migration_0218.sql
-- Description: create user game launch error table
-- ============================================

CREATE TABLE IF NOT EXISTS user_game_launch_errors (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  username VARCHAR(100) NOT NULL DEFAULT '',
  game_code VARCHAR(100) NOT NULL DEFAULT '',
  game_name VARCHAR(255) NOT NULL DEFAULT '',
  provider_code VARCHAR(50) NOT NULL DEFAULT '',
  provider_name VARCHAR(100) NOT NULL DEFAULT '',
  is_lobby TINYINT(1) NOT NULL DEFAULT 0,
  is_mobile TINYINT(1) NOT NULL DEFAULT 0,
  language VARCHAR(20) NOT NULL DEFAULT '',
  error_type VARCHAR(64) NOT NULL,
  error_message TEXT NULL,
  page_url TEXT NULL,
  game_url_host VARCHAR(255) NOT NULL DEFAULT '',
  client_ip VARCHAR(64) NOT NULL DEFAULT '',
  user_agent TEXT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_ugle_user_created (user_id, created_at),
  KEY idx_ugle_user_type_created (user_id, error_type, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
