CREATE TABLE IF NOT EXISTS `activity_wager_configs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `config_key` VARCHAR(50) NOT NULL,
  `deposit_wager_multiplier` DECIMAL(10,2) NOT NULL DEFAULT 2.00 COMMENT '充值本金打码倍数',
  `reward_wager_multiplier` DECIMAL(10,2) NOT NULL DEFAULT 20.00 COMMENT '赠送奖励打码倍数',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_activity_wager_config_key` (`config_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='活动打码倍数配置';

INSERT INTO `activity_wager_configs`
  (`config_key`, `deposit_wager_multiplier`, `reward_wager_multiplier`, `created_at`, `updated_at`)
VALUES
  ('default', 2.00, 20.00, NOW(), NOW())
ON DUPLICATE KEY UPDATE
  `config_key` = VALUES(`config_key`);
