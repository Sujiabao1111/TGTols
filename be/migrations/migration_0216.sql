-- ============================================
-- Migration: migration_0216.sql
-- Description: backfill daily deposit stats from successful deposit transactions
-- ============================================

INSERT INTO daily_user_stats (
  user_id,
  stat_date,
  bet_amount,
  win_amount,
  deposit_amount,
  withdraw_amount,
  login_count,
  created_at,
  updated_at
)
SELECT
  t.user_id,
  DATE(t.created_at) AS stat_date,
  0 AS bet_amount,
  0 AS win_amount,
  SUM(t.amount) AS deposit_amount,
  0 AS withdraw_amount,
  0 AS login_count,
  NOW() AS created_at,
  NOW() AS updated_at
FROM transactions t
WHERE t.type = 1
  AND t.amount > 0
GROUP BY t.user_id, DATE(t.created_at)
ON DUPLICATE KEY UPDATE
  deposit_amount = GREATEST(COALESCE(daily_user_stats.deposit_amount, 0), VALUES(deposit_amount)),
  updated_at = CURRENT_TIMESTAMP;

UPDATE stats_retention AS sr
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
  sr.deposit_users = COALESCE(d.deposit_users, 0),
  sr.deposit_amount = COALESCE(d.deposit_amount, 0),
  sr.game_winlose_users = COALESCE(d.game_winlose_users, 0);
