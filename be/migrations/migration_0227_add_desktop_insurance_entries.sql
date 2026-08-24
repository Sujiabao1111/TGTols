-- Migration: migration_0227_add_desktop_insurance_entries.sql
-- Add Desktop reward is now an insurance coupon. This table records game entry/exit settlement.

CREATE TABLE IF NOT EXISTS `add_desktop_game_entries` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL,
  `platform_code` VARCHAR(32) NOT NULL DEFAULT 'HEDOC',
  `reference_id` VARCHAR(100) NOT NULL,
  `entry_amount_u` DECIMAL(15,2) NOT NULL DEFAULT 0.00,
  `exit_amount_u` DECIMAL(15,2) NOT NULL DEFAULT 0.00,
  `loss_amount_u` DECIMAL(15,2) NOT NULL DEFAULT 0.00,
  `compensation_amount_u` DECIMAL(15,2) NOT NULL DEFAULT 0.00,
  `coupon_id` BIGINT UNSIGNED NULL DEFAULT NULL,
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '0:pending 1:settled 2:insured',
  `entered_at` DATETIME NOT NULL,
  `settled_at` DATETIME NULL DEFAULT NULL,
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_add_desktop_entry_reference` (`reference_id`),
  KEY `idx_add_desktop_entry_user_status` (`user_id`, `platform_code`, `status`),
  KEY `idx_add_desktop_entry_coupon` (`coupon_id`),
  KEY `idx_add_desktop_entry_entered_at` (`entered_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Add Desktop insurance game entry records';
