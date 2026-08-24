-- ============================================
-- Migration: migration_0215.sql
-- Description: deduplicate stats_retention by calendar day and enforce one row per day
-- ============================================

DROP TEMPORARY TABLE IF EXISTS tmp_stats_retention_keep;
CREATE TEMPORARY TABLE tmp_stats_retention_keep AS
SELECT MAX(id) AS keep_id
FROM stats_retention
GROUP BY DATE(stat_date);

UPDATE stats_retention AS target
INNER JOIN (
  SELECT
    DATE(stat_date) AS stat_day,
    MAX(id) AS keep_id,
    MAX(new_users) AS new_users,
    MAX(deposit_users) AS deposit_users,
    MAX(deposit_amount) AS deposit_amount,
    MAX(game_winlose_users) AS game_winlose_users,
    MAX(retention_1) AS retention_1,
    MAX(retention_3) AS retention_3,
    MAX(retention_7) AS retention_7,
    MAX(retention_30) AS retention_30
  FROM stats_retention
  GROUP BY DATE(stat_date)
) AS merged
  ON target.id = merged.keep_id
SET
  target.stat_date = merged.stat_day,
  target.new_users = merged.new_users,
  target.deposit_users = merged.deposit_users,
  target.deposit_amount = merged.deposit_amount,
  target.game_winlose_users = merged.game_winlose_users,
  target.retention_1 = merged.retention_1,
  target.retention_3 = merged.retention_3,
  target.retention_7 = merged.retention_7,
  target.retention_30 = merged.retention_30;

DELETE FROM stats_retention
WHERE id NOT IN (
  SELECT keep_id
  FROM tmp_stats_retention_keep
);

ALTER TABLE `stats_retention`
  MODIFY COLUMN `stat_date` DATE NOT NULL COMMENT 'stat date (registration date)';

UPDATE stats_retention AS sr
LEFT JOIN (
  SELECT
    DATE(created_at) AS stat_day,
    COUNT(*) AS new_users
  FROM users
  GROUP BY DATE(created_at)
) AS u
  ON u.stat_day = sr.stat_date
LEFT JOIN (
  SELECT
    stat_date AS stat_day,
    COALESCE(SUM(CASE WHEN deposit_amount > 0 THEN 1 ELSE 0 END), 0) AS deposit_users,
    COALESCE(SUM(CASE WHEN deposit_amount > 0 THEN deposit_amount ELSE 0 END), 0) AS deposit_amount,
    COALESCE(SUM(CASE WHEN bet_amount > 0 AND win_amount <> 0 THEN 1 ELSE 0 END), 0) AS game_winlose_users
  FROM daily_user_stats
  GROUP BY stat_date
) AS d
  ON d.stat_day = sr.stat_date
SET
  sr.new_users = COALESCE(u.new_users, 0),
  sr.deposit_users = COALESCE(d.deposit_users, 0),
  sr.deposit_amount = COALESCE(d.deposit_amount, 0),
  sr.game_winlose_users = COALESCE(d.game_winlose_users, 0);

SET @idx_date_exists := (
  SELECT COUNT(*)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'stats_retention'
    AND INDEX_NAME = 'idx_date'
);

SET @add_idx_date_sql := IF(
  @idx_date_exists = 0,
  'ALTER TABLE `stats_retention` ADD UNIQUE KEY `idx_date` (`stat_date`)',
  'SELECT ''idx_date already exists'''
);

PREPARE stmt FROM @add_idx_date_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

DROP TEMPORARY TABLE IF EXISTS tmp_stats_retention_keep;
