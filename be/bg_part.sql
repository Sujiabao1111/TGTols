SET NAMES utf8mb4;

-- ==========================================
-- 1. stats_retention
-- ==========================================
DROP TABLE IF EXISTS `stats_retention`;
CREATE TABLE `stats_retention` (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT,
  `stat_date` date NOT NULL COMMENT 'stat date (registration date)',
  `new_users` int(11) DEFAULT 0 COMMENT 'daily new users',
  `deposit_users` int(11) DEFAULT 0 COMMENT 'daily deposit users',
  `deposit_amount` decimal(15,2) DEFAULT 0.00 COMMENT 'daily deposit amount',
  `game_winlose_users` int(11) DEFAULT 0 COMMENT 'daily players with non-zero winlose',
  `retention_1` int(11) DEFAULT 0 COMMENT 'next-day retention users',
  `retention_3` int(11) DEFAULT 0 COMMENT '3-day retention users',
  `retention_7` int(11) DEFAULT 0 COMMENT '7-day retention users',
  `retention_30` int(11) DEFAULT 0 COMMENT '30-day retention users',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_date` (`stat_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='user retention daily stats';

-- ==========================================
-- 2. agent_rebate_configs
-- ==========================================
DROP TABLE IF EXISTS `agent_rebate_configs`;
CREATE TABLE `agent_rebate_configs` (
  `level` int(11) NOT NULL COMMENT 'rebate level distance',
  `rate` decimal(5,4) NOT NULL COMMENT 'rebate rate',
  PRIMARY KEY (`level`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='agent rebate config';

INSERT INTO `agent_rebate_configs` (`level`, `rate`) VALUES
(1, 0.0100), (2, 0.0050), (3, 0.0030), (4, 0.0020), (5, 0.0010),
(6, 0.0008), (7, 0.0006), (8, 0.0004), (9, 0.0002);
