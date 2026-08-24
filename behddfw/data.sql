-- 设置字符集和排序规则
SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- 1. 用户表 (Users) - 核心分销体系
-- ----------------------------
DROP TABLE IF EXISTS `users`;
CREATE TABLE `users` (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT,
  `username` varchar(32) NOT NULL COMMENT '用户名',
  `password` varchar(255) NOT NULL COMMENT '加密后的密码',
  `balance` decimal(15,2) DEFAULT 0.00 COMMENT '用户余额/积分',
  `vip_level` int(11) DEFAULT 0 COMMENT 'VIP等级',
  
  -- 分销核心字段
  `parent_id` bigint(20) UNSIGNED DEFAULT 0 COMMENT '直属上级ID, 0表示无上级',
  `path` varchar(255) DEFAULT '' COMMENT '物化路径, 格式: 0/1/5/ (根ID/一级ID/二级ID/)',
  `level` int(11) DEFAULT 1 COMMENT '层级深度, 根用户为1',
  `invite_code` varchar(32) DEFAULT NULL COMMENT '个人邀请码',
  
  -- 状态与时间
  `status` tinyint(1) DEFAULT 1 COMMENT '状态 1:正常 0:禁用',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_username` (`username`),
  UNIQUE KEY `idx_invite_code` (`invite_code`),
  KEY `idx_parent_id` (`parent_id`),
  -- 核心优化: 路径索引，用于快速查询团队业绩 WHERE path LIKE '0/1/5/%'
  KEY `idx_path` (`path`) 
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';

-- ----------------------------
-- 2. VIP 配置表 (Vip Configs)
-- ----------------------------
DROP TABLE IF EXISTS `vip_configs`;
CREATE TABLE `vip_configs` (
  `level` int(11) NOT NULL COMMENT 'VIP等级',
  `title` varchar(50) NOT NULL COMMENT '等级名称 (青铜, 白银...)',
  `min_deposit` decimal(15,2) DEFAULT 0.00 COMMENT '升级所需最小累计充值',
  `rebate_rate` decimal(5,4) DEFAULT 0.0000 COMMENT '自身返水比例 (如 0.0050 代表 0.5%)',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`level`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='VIP等级配置';

-- ----------------------------
-- 3. 游戏厂商表 (Game Providers)
-- ----------------------------
DROP TABLE IF EXISTS `game_providers`;
CREATE TABLE `game_providers` (
  `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT,
  `code` varchar(50) NOT NULL COMMENT '厂商代码 (如 PG, AG)',
  `name` varchar(100) NOT NULL COMMENT '厂商显示名称',
  `api_config` json DEFAULT NULL COMMENT 'JSON格式存储API Key, Secret, Endpoint等敏感信息',
  `status` tinyint(1) DEFAULT 1 COMMENT '1:开启 0:维护',
  `type_id` int(10) DEFAULT 1 COMMENT '游戏类型id',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_provider_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='游戏供应商';

-- ----------------------------
-- 4. 游戏分类表 (Game Types)
-- ----------------------------
DROP TABLE IF EXISTS `game_types`;
CREATE TABLE `game_types` (
  `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` varchar(50) NOT NULL COMMENT '分类名称 (电子, 真人, 体育)',
  `sort` int(11) DEFAULT 0 COMMENT '排序权重',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='游戏分类';

-- ----------------------------
-- Table structure for game_types
-- ----------------------------
DROP TABLE IF EXISTS `game_types`;
CREATE TABLE `game_types`  (
  `id` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` varchar(50) NOT NULL COMMENT '分类名称 (电子, 真人, 体育)',
  `sort` int NULL DEFAULT 0 COMMENT '排序权重',
  `code` varchar(32) COMMENT '代码字符串',
  `status` tinyint NULL DEFAULT 1 COMMENT '1:启用 0:禁用',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='游戏分类';

-- ----------------------------
-- Records of game_types
-- ----------------------------
INSERT INTO `game_types` VALUES (1, 'Other(其他)', 1, 'OTHER', 0);
INSERT INTO `game_types` VALUES (2, 'Slots (老虎机)', 2, 'SLOT', 1);
INSERT INTO `game_types` VALUES (3, 'Live Casino (真人)', 3, 'CASINO', 1);
INSERT INTO `game_types` VALUES (4, 'Sports (体育)', 4, 'SPORTBOOK', 1);
INSERT INTO `game_types` VALUES (5, 'Keno', 5, 'KENO', 0);
INSERT INTO `game_types` VALUES (6, 'Lottery', 6, 'LOTTERY', 0);
INSERT INTO `game_types` VALUES (7, 'Fishing(捕鱼)', 7, 'FISHING', 0);
INSERT INTO `game_types` VALUES (8, 'Esport(电竞)', 8, 'ESPORT', 0);

-- ----------------------------
-- 5. 游戏表 (Games) - 最新版
-- ----------------------------
DROP TABLE IF EXISTS `games`;
CREATE TABLE `games` (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT,
  
  `provider_id` int(10) UNSIGNED NOT NULL COMMENT '关联厂商ID (game_providers.id)',
  `game_type_id` int(10) UNSIGNED NOT NULL COMMENT '关联分类ID (game_types.id)',
  `game_code` varchar(100) NOT NULL COMMENT '厂商侧的游戏ID',
  
  -- 修改: 使用 JSON 类型存储多语言名称
  -- 示例: {"CN": "麻将胡了", "EN": "Mahjong Ways"}
  `name` json DEFAULT NULL COMMENT '游戏名称(多语言)',
  
  `img_url` varchar(255) DEFAULT '' COMMENT '封面图片URL',
  `views` bigint(20) DEFAULT 0 COMMENT '点击/热度',
  `sort` int(11) DEFAULT 0 COMMENT '排序, 越大越前',
  `status` tinyint(1) DEFAULT 1 COMMENT '1:上架 0:下架',
  
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  
  PRIMARY KEY (`id`),
  
  -- 核心修改: 联合唯一索引
  -- 解决不同厂商可能有相同 game_code 的问题，确保同一厂商下 code 唯一
  UNIQUE KEY `idx_provider_game_code` (`provider_id`, `game_code`),
  
  -- 常用筛选索引
  KEY `idx_provider_type` (`provider_id`, `game_type_id`),
  
  -- 排序索引
  KEY `idx_sort_views` (`sort`, `views`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='游戏库';

-- ----------------------------
-- 6. 广告位表 (Banners)
-- ----------------------------
DROP TABLE IF EXISTS `banners`;
CREATE TABLE `banners` (
  `id` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `title` varchar(255) NOT NULL COMMENT '对应前端 key, 如 hero.slide1.title',
  `subtitle` varchar(255) DEFAULT '' COMMENT '副标题',
  `tag` varchar(100) DEFAULT '' COMMENT '标签',
  `image` varchar(500) NOT NULL COMMENT '图片URL',
  `color` varchar(255) DEFAULT '' COMMENT 'Tailwind 渐变色类名',
  
  -- 差异化字段
  `jump_link` varchar(255) DEFAULT NULL COMMENT '整图跳转链接 (形式2)',
  `buttons` json DEFAULT NULL COMMENT '按钮配置数组 (形式1)',
  
  `sort` int DEFAULT 0 COMMENT '排序',
  `status` tinyint(1) DEFAULT 1 COMMENT '1:启用 0:禁用',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='首页轮播图';

-- ----------------------------
-- 插入测试数据 (对应你的前端样例)
-- ----------------------------

-- 样例 1: 带按钮
INSERT INTO `banners` (`id`, `title`, `subtitle`, `tag`, `image`, `color`, `buttons`, `jump_link`, `sort`) 
VALUES (1, 'hero.slide1.title', 'hero.slide1.subtitle', 'hero.slide1.tag', 'https://picsum.photos/seed/bearlucky/600/600', 'from-lucky-gold to-orange-400', 
'[
  {"label": "hero.slide1.btn1", "action": "/wallet", "primary": true},
  {"label": "hero.slide1.btn2", "action": "/all-games", "primary": false}
]', NULL, 10);

-- 样例 2: 带跳转链接
INSERT INTO `banners` (`id`, `title`, `subtitle`, `tag`, `image`, `color`, `buttons`, `jump_link`, `sort`) 
VALUES (2, 'hero.slide2.title', 'hero.slide2.subtitle', 'hero.slide2.tag', 'https://picsum.photos/seed/poker/600/600', 'from-blue-400 to-purple-500', 
NULL, '/activity', 9);

-- 样例 3: 纯展示
INSERT INTO `banners` (`id`, `title`, `subtitle`, `tag`, `image`, `color`, `buttons`, `jump_link`, `sort`) 
VALUES (3, 'hero.slide3.title', 'hero.slide3.subtitle', 'hero.slide3.tag', 'https://picsum.photos/seed/vip/600/600', 'from-lucky-pink to-red-500', 
NULL, NULL, 8);

-- ----------------------------
-- 7. 活动表 (Activities) - 预留扩展
-- ----------------------------
DROP TABLE IF EXISTS `activities`;
CREATE TABLE `activities` (
  `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT,
  `type` varchar(50) NOT NULL COMMENT '活动类型 (sign_in, recharge_bonus)',
  `name` varchar(100) NOT NULL,
  `config` json DEFAULT NULL COMMENT 'JSON存储规则: {reward:10, min:100}',
  `start_time` datetime DEFAULT NULL,
  `end_time` datetime DEFAULT NULL,
  `status` tinyint(1) DEFAULT 1,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='运营活动配置';

-- ----------------------------
-- 8. 用户每日签到记录 (User Sign Ins)
-- ----------------------------
DROP TABLE IF EXISTS `user_sign_ins`;
CREATE TABLE `user_sign_ins` (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` bigint(20) UNSIGNED NOT NULL,
  `sign_date` date NOT NULL COMMENT '签到日期',
  `reward` decimal(10,2) DEFAULT 0.00 COMMENT '获得奖励',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  -- 保证用户每天只能签到一次
  UNIQUE KEY `idx_user_date` (`user_id`, `sign_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='签到日志';

-- ----------------------------
-- 9. 资金账变记录 (Transactions) - 必须存在
-- ----------------------------
DROP TABLE IF EXISTS `transactions`;
CREATE TABLE `transactions` (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` bigint(20) UNSIGNED NOT NULL,
  `type` tinyint(4) NOT NULL COMMENT '1:充值 2:提现 3:游戏下注 4:游戏派彩 5:分销返点 6:活动赠送',
  `amount` decimal(15,2) NOT NULL COMMENT '变动金额 (正负)',
  `before_balance` decimal(15,2) NOT NULL COMMENT '变动前余额',
  `after_balance` decimal(15,2) NOT NULL COMMENT '变动后余额',
  `reference_id` varchar(100) DEFAULT '' COMMENT '关联业务ID (订单号/游戏局号)',
  `remark` varchar(255) DEFAULT '',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_created` (`user_id`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='资金流水表';

-- ----------------------------
-- 10. 用户每日统计 (Daily User Stats)
-- ----------------------------
DROP TABLE IF EXISTS `daily_user_stats`;
CREATE TABLE `daily_user_stats` (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` bigint(20) UNSIGNED NOT NULL COMMENT '用户ID',
  `stat_date` date NOT NULL COMMENT '统计日期',
  
  -- 资金相关 (用于返佣计算)
  `bet_amount` decimal(15,2) DEFAULT 0.00 COMMENT '当日总流水/打码量',
  `win_amount` decimal(15,2) DEFAULT 0.00 COMMENT '当日总输赢(派彩-下注)',
  `deposit_amount` decimal(15,2) DEFAULT 0.00 COMMENT '当日总充值',
  `withdraw_amount` decimal(15,2) DEFAULT 0.00 COMMENT '当日总提现',
  
  -- 活跃相关 (用于留存分析)
  `login_count` int(11) DEFAULT 0 COMMENT '当日登录次数',
  
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  
  PRIMARY KEY (`id`),
  -- 核心索引: 确保每个用户每天只有一条记录，用于 Upsert
  UNIQUE KEY `idx_user_date` (`user_id`, `stat_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户每日数据聚合表';

-- ----------------------------
-- 插入测试数据 (Seed Data)
-- ----------------------------

-- 1. 插入一个顶层管理员用户/根代理
INSERT INTO `users` (`id`, `username`, `password`, `balance`, `vip_level`, `parent_id`, `path`, `level`, `invite_code`, `status`) 
VALUES (1, 'admin_root', 'hashed_password_here', 0.00, 99, 0, '', 1, '888888', 1);

-- 2. 插入一些游戏分类
INSERT INTO `game_types` (`id`, `name`, `sort`) VALUES 
(1, 'Slots (老虎机)', 10),
(2, 'Live Casino (真人)', 9),
(3, 'Sports (体育)', 8);

-- 3. 插入一个游戏厂商
INSERT INTO `game_providers` (`id`, `code`, `name`, `api_config`, `status`) VALUES 
(1, 'PG', 'PG Soft', '{"api_key": "123456", "endpoint": "https://api.pgsoft.com"}', 1);

-- 4. 插入 VIP 配置
INSERT INTO `vip_configs` (`level`, `title`, `min_deposit`, `rebate_rate`) VALUES
(0, '普通会员', 0, 0.0000),
(1, 'VIP1', 1000, 0.0030),
(2, 'VIP2', 5000, 0.0050);

-- ==========================================
-- 5. 站内信系统 (Mail System)
-- ==========================================

-- 信件内容 (支持群发)
DROP TABLE IF EXISTS `sys_mails`;
CREATE TABLE `sys_mails` (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT,
  `type` tinyint(4) DEFAULT 1 COMMENT '1:个人 2:公告',
  `title` varchar(150) NOT NULL,
  `content` text COMMENT '正文',
  `attachments` json DEFAULT NULL COMMENT '附件奖励 [{"type":"balance","value":100}]',
  `admin_id` bigint(20) DEFAULT 0,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='站内信内容库';

-- 用户收件箱 (关联状态)
DROP TABLE IF EXISTS `user_inbox`;
CREATE TABLE `user_inbox` (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` bigint(20) UNSIGNED NOT NULL,
  `sys_mail_id` bigint(20) UNSIGNED NOT NULL,
  `is_read` tinyint(1) DEFAULT 0 COMMENT '0:未读 1:已读',
  `is_claimed` tinyint(1) DEFAULT 0 COMMENT '0:未领 1:已领 2:无附件',
  `received_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `claimed_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_unique_mail` (`user_id`, `sys_mail_id`),
  KEY `idx_user_list` (`user_id`, `id` DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户收件箱';

SET FOREIGN_KEY_CHECKS = 1;