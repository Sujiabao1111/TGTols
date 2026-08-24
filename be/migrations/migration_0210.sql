-- 迁移脚本: migration_0210.sql
-- 说明: 为提现订单增加后台审核字段

SET @reviewed_by_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'withdraw_orders'
    AND COLUMN_NAME = 'reviewed_by'
);
SET @sql := IF(
  @reviewed_by_exists = 0,
  'ALTER TABLE `withdraw_orders` ADD COLUMN `reviewed_by` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT "审核人ID" AFTER `ref_msg`',
  'SELECT "reviewed_by already exists"'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @reviewer_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'withdraw_orders'
    AND COLUMN_NAME = 'reviewer'
);
SET @sql := IF(
  @reviewer_exists = 0,
  'ALTER TABLE `withdraw_orders` ADD COLUMN `reviewer` VARCHAR(100) NOT NULL DEFAULT "" COMMENT "审核人账号" AFTER `reviewed_by`',
  'SELECT "reviewer already exists"'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @review_remark_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'withdraw_orders'
    AND COLUMN_NAME = 'review_remark'
);
SET @sql := IF(
  @review_remark_exists = 0,
  'ALTER TABLE `withdraw_orders` ADD COLUMN `review_remark` VARCHAR(255) NOT NULL DEFAULT "" COMMENT "审核备注" AFTER `reviewer`',
  'SELECT "review_remark already exists"'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @reviewed_at_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'withdraw_orders'
    AND COLUMN_NAME = 'reviewed_at'
);
SET @sql := IF(
  @reviewed_at_exists = 0,
  'ALTER TABLE `withdraw_orders` ADD COLUMN `reviewed_at` DATETIME NULL COMMENT "审核时间" AFTER `review_remark`',
  'SELECT "reviewed_at already exists"'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 执行命令: mysql -u root -p tols < be/migrations/migration_0210.sql
