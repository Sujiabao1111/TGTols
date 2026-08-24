-- ============================================
-- Migration: migration_0213.sql
-- Description: add readable win/lose summary columns and refresh Chinese comments
-- ============================================

SET @result_status_exists := (
  SELECT COUNT(*)
  FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'user_game_transaction_stats'
    AND COLUMN_NAME = 'result_status'
);

SET @add_result_status_sql := IF(
  @result_status_exists = 0,
  'ALTER TABLE `user_game_transaction_stats`
     ADD COLUMN `result_status` VARCHAR(16)
       GENERATED ALWAYS AS (
         CASE
           WHEN `winlose` > 0 THEN ''win''
           WHEN `winlose` < 0 THEN ''lose''
           ELSE ''draw''
         END
       ) STORED
       COMMENT ''总结果: win=赢, lose=输, draw=平'' AFTER `winlose`,
     ADD COLUMN `result_amount` DECIMAL(15,2)
       GENERATED ALWAYS AS (ABS(`winlose`)) STORED
       COMMENT ''输赢金额绝对值(U)'' AFTER `result_status`',
  'SELECT "result_status already exists"'
);

PREPARE stmt FROM @add_result_status_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

ALTER TABLE `user_game_transaction_stats`
  MODIFY COLUMN `user_id` BIGINT UNSIGNED NOT NULL COMMENT '站内用户ID',
  MODIFY COLUMN `period_type` VARCHAR(16) NOT NULL COMMENT '统计周期类型: daily=每日, weekly=每周, total=总计',
  MODIFY COLUMN `period_key` VARCHAR(32) NOT NULL COMMENT '统计键值: 日=YYYY-MM-DD, 周=周起始日, 总计=all',
  MODIFY COLUMN `period_start` DATETIME NOT NULL COMMENT '统计开始时间(UTC)',
  MODIFY COLUMN `period_end` DATETIME NOT NULL COMMENT '统计结束时间(UTC)',
  MODIFY COLUMN `login_id` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '第三方登录账号',
  MODIFY COLUMN `amount_unit` VARCHAR(8) NOT NULL DEFAULT '' COMMENT '金额单位: U',
  MODIFY COLUMN `count` INT NOT NULL DEFAULT 0 COMMENT '注单数',
  MODIFY COLUMN `turnover` DECIMAL(15,2) NOT NULL DEFAULT 0.00 COMMENT '流水/打码量(U)',
  MODIFY COLUMN `bet` DECIMAL(15,2) NOT NULL DEFAULT 0.00 COMMENT '下注金额(U)',
  MODIFY COLUMN `win` DECIMAL(15,2) NOT NULL DEFAULT 0.00 COMMENT '派彩金额(U)',
  MODIFY COLUMN `winlose` DECIMAL(15,2) NOT NULL DEFAULT 0.00 COMMENT '玩家输赢(U): 正数=赢, 负数=输',
  MODIFY COLUMN `jp_share` DECIMAL(15,2) NOT NULL DEFAULT 0.00 COMMENT 'JP分摊(U)',
  MODIFY COLUMN `jp_win` DECIMAL(15,2) NOT NULL DEFAULT 0.00 COMMENT 'JP中奖(U)',
  MODIFY COLUMN `synced_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '最近同步时间',
  MODIFY COLUMN `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  MODIFY COLUMN `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间';
