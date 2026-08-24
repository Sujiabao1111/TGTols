SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;
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