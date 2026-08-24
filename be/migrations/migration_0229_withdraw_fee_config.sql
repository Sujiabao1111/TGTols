CREATE TABLE IF NOT EXISTS `withdraw_fee_configs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `config_key` VARCHAR(50) NOT NULL DEFAULT 'default',
  `enabled` TINYINT(1) NOT NULL DEFAULT 1,
  `fee_rate` DECIMAL(10,6) NOT NULL DEFAULT 0.003000,
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_withdraw_fee_config_key` (`config_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='提现手续费配置';

INSERT INTO `withdraw_fee_configs` (`config_key`, `enabled`, `fee_rate`)
VALUES ('default', 1, 0.003000)
ON DUPLICATE KEY UPDATE
  `config_key` = VALUES(`config_key`);
