SET NAMES utf8mb4;

-- ==========================================
-- 1. 留存统计表 (Stats Retention)
-- ==========================================
-- 每天凌晨跑定时任务，计算前几日的注册用户在今日的留存情况
DROP TABLE IF EXISTS `stats_retention`;
CREATE TABLE `stats_retention` (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT,
  `stat_date` date NOT NULL COMMENT '统计日期(注册日期)',
  `new_users` int(11) DEFAULT 0 COMMENT '当日新增注册人数',
  
  -- 留存数 (有多少人在第N天登录/活跃)
  `retention_1` int(11) DEFAULT 0 COMMENT '次日留存数',
  `retention_3` int(11) DEFAULT 0 COMMENT '3日留存数',
  `retention_7` int(11) DEFAULT 0 COMMENT '7日留存数',
  `retention_30` int(11) DEFAULT 0 COMMENT '月留存数',
  
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_date` (`stat_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户留存统计表';

-- ==========================================
-- 2. 代理返佣配置表 (Rebate Configs)
-- ==========================================
-- 定义 1-9 级每一级的返点比例
DROP TABLE IF EXISTS `agent_rebate_configs`;
CREATE TABLE `agent_rebate_configs` (
  `level` int(11) NOT NULL COMMENT '层级距离 (1表示直属上级, 2表示上上级...)',
  `rate` decimal(5,4) NOT NULL COMMENT '返点比例 (如 0.0050 表示 0.5%)',
  PRIMARY KEY (`level`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='分销返点比例配置';

-- 初始化 9 级配置 (示例数据)
INSERT INTO `agent_rebate_configs` (`level`, `rate`) VALUES 
(1, 0.0100), (2, 0.0050), (3, 0.0030), (4, 0.0020), (5, 0.0010),
(6, 0.0008), (7, 0.0006), (8, 0.0004), (9, 0.0002);