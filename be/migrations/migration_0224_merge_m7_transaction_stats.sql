-- ============================================
-- Migration: migration_0224_merge_m7_transaction_stats.sql
-- Description: merge stored M7 transaction details into user_game_transaction_stats
-- ============================================

DROP TEMPORARY TABLE IF EXISTS tmp_m7_daily_transaction_stats;
DROP TEMPORARY TABLE IF EXISTS tmp_weekly_transaction_stats;
DROP TEMPORARY TABLE IF EXISTS tmp_total_transaction_stats;

CREATE TEMPORARY TABLE tmp_m7_daily_transaction_stats AS
SELECT
  d.user_id,
  DATE_FORMAT(d.stat_date, '%Y-%m-%d') AS period_key,
  MIN(COALESCE(d.login_id, '')) AS login_id,
  COALESCE(SUM(CASE
    WHEN d.turnover <> 0 OR d.bet <> 0 OR d.win <> 0 OR d.winlose <> 0 THEN 1
    ELSE 0
  END), 0) AS count,
  COALESCE(SUM(d.turnover), 0) AS turnover,
  COALESCE(SUM(d.bet), 0) AS bet,
  COALESCE(SUM(d.win), 0) AS win,
  COALESCE(SUM(d.winlose), 0) AS winlose,
  COALESCE(SUM(d.jp_share), 0) AS jp_share,
  COALESCE(SUM(d.jp_win), 0) AS jp_win
FROM user_game_transaction_details d
JOIN game_providers gp
  ON gp.id = d.game_provider_code
 AND gp.platform_code = 'M7'
GROUP BY d.user_id, DATE_FORMAT(d.stat_date, '%Y-%m-%d')
HAVING COALESCE(SUM(CASE
  WHEN d.turnover <> 0 OR d.bet <> 0 OR d.win <> 0 OR d.winlose <> 0 THEN 1
  ELSE 0
END), 0) > 0;

INSERT INTO user_game_transaction_stats (
  user_id,
  period_type,
  period_key,
  period_start,
  period_end,
  login_id,
  amount_unit,
  count,
  turnover,
  bet,
  win,
  winlose,
  jp_share,
  jp_win,
  synced_at,
  created_at,
  updated_at
)
SELECT
  user_id,
  'daily',
  period_key,
  STR_TO_DATE(period_key, '%Y-%m-%d'),
  DATE_ADD(STR_TO_DATE(period_key, '%Y-%m-%d'), INTERVAL 1 DAY) - INTERVAL 1 MICROSECOND,
  login_id,
  'U',
  count,
  turnover,
  bet,
  win,
  winlose,
  jp_share,
  jp_win,
  NOW(),
  NOW(),
  NOW()
FROM tmp_m7_daily_transaction_stats
ON DUPLICATE KEY UPDATE
  count = IF(
    user_game_transaction_stats.count >= VALUES(count)
    AND user_game_transaction_stats.turnover + 0.0001 >= VALUES(turnover)
    AND user_game_transaction_stats.bet + 0.0001 >= VALUES(bet),
    user_game_transaction_stats.count,
    user_game_transaction_stats.count + VALUES(count)
  ),
  turnover = IF(
    user_game_transaction_stats.count >= VALUES(count)
    AND user_game_transaction_stats.turnover + 0.0001 >= VALUES(turnover)
    AND user_game_transaction_stats.bet + 0.0001 >= VALUES(bet),
    user_game_transaction_stats.turnover,
    user_game_transaction_stats.turnover + VALUES(turnover)
  ),
  bet = IF(
    user_game_transaction_stats.count >= VALUES(count)
    AND user_game_transaction_stats.turnover + 0.0001 >= VALUES(turnover)
    AND user_game_transaction_stats.bet + 0.0001 >= VALUES(bet),
    user_game_transaction_stats.bet,
    user_game_transaction_stats.bet + VALUES(bet)
  ),
  win = IF(
    user_game_transaction_stats.count >= VALUES(count)
    AND user_game_transaction_stats.turnover + 0.0001 >= VALUES(turnover)
    AND user_game_transaction_stats.bet + 0.0001 >= VALUES(bet),
    user_game_transaction_stats.win,
    user_game_transaction_stats.win + VALUES(win)
  ),
  winlose = IF(
    user_game_transaction_stats.count >= VALUES(count)
    AND user_game_transaction_stats.turnover + 0.0001 >= VALUES(turnover)
    AND user_game_transaction_stats.bet + 0.0001 >= VALUES(bet),
    user_game_transaction_stats.winlose,
    user_game_transaction_stats.winlose + VALUES(winlose)
  ),
  jp_share = IF(
    user_game_transaction_stats.count >= VALUES(count)
    AND user_game_transaction_stats.turnover + 0.0001 >= VALUES(turnover)
    AND user_game_transaction_stats.bet + 0.0001 >= VALUES(bet),
    user_game_transaction_stats.jp_share,
    user_game_transaction_stats.jp_share + VALUES(jp_share)
  ),
  jp_win = IF(
    user_game_transaction_stats.count >= VALUES(count)
    AND user_game_transaction_stats.turnover + 0.0001 >= VALUES(turnover)
    AND user_game_transaction_stats.bet + 0.0001 >= VALUES(bet),
    user_game_transaction_stats.jp_win,
    user_game_transaction_stats.jp_win + VALUES(jp_win)
  ),
  login_id = IF(user_game_transaction_stats.login_id = '', VALUES(login_id), user_game_transaction_stats.login_id),
  amount_unit = 'U',
  synced_at = NOW(),
  updated_at = NOW();

CREATE TEMPORARY TABLE tmp_weekly_transaction_stats AS
SELECT
  user_id,
  DATE_FORMAT(
    DATE_SUB(STR_TO_DATE(period_key, '%Y-%m-%d'), INTERVAL WEEKDAY(STR_TO_DATE(period_key, '%Y-%m-%d')) DAY),
    '%Y-%m-%d'
  ) AS period_key,
  MIN(period_start) AS period_start,
  MAX(period_end) AS period_end,
  MIN(login_id) AS login_id,
  SUM(count) AS count,
  SUM(turnover) AS turnover,
  SUM(bet) AS bet,
  SUM(win) AS win,
  SUM(winlose) AS winlose,
  SUM(jp_share) AS jp_share,
  SUM(jp_win) AS jp_win
FROM user_game_transaction_stats
WHERE period_type = 'daily'
GROUP BY
  user_id,
  DATE_FORMAT(
    DATE_SUB(STR_TO_DATE(period_key, '%Y-%m-%d'), INTERVAL WEEKDAY(STR_TO_DATE(period_key, '%Y-%m-%d')) DAY),
    '%Y-%m-%d'
  );

INSERT INTO user_game_transaction_stats (
  user_id,
  period_type,
  period_key,
  period_start,
  period_end,
  login_id,
  amount_unit,
  count,
  turnover,
  bet,
  win,
  winlose,
  jp_share,
  jp_win,
  synced_at,
  created_at,
  updated_at
)
SELECT
  user_id,
  'weekly',
  period_key,
  period_start,
  period_end,
  login_id,
  'U',
  count,
  turnover,
  bet,
  win,
  winlose,
  jp_share,
  jp_win,
  NOW(),
  NOW(),
  NOW()
FROM tmp_weekly_transaction_stats
ON DUPLICATE KEY UPDATE
  period_start = VALUES(period_start),
  period_end = VALUES(period_end),
  login_id = VALUES(login_id),
  amount_unit = 'U',
  count = VALUES(count),
  turnover = VALUES(turnover),
  bet = VALUES(bet),
  win = VALUES(win),
  winlose = VALUES(winlose),
  jp_share = VALUES(jp_share),
  jp_win = VALUES(jp_win),
  synced_at = NOW(),
  updated_at = NOW();

CREATE TEMPORARY TABLE tmp_total_transaction_stats AS
SELECT
  user_id,
  MIN(period_start) AS period_start,
  MAX(period_end) AS period_end,
  MIN(login_id) AS login_id,
  SUM(count) AS count,
  SUM(turnover) AS turnover,
  SUM(bet) AS bet,
  SUM(win) AS win,
  SUM(winlose) AS winlose,
  SUM(jp_share) AS jp_share,
  SUM(jp_win) AS jp_win
FROM user_game_transaction_stats
WHERE period_type = 'daily'
GROUP BY user_id;

INSERT INTO user_game_transaction_stats (
  user_id,
  period_type,
  period_key,
  period_start,
  period_end,
  login_id,
  amount_unit,
  count,
  turnover,
  bet,
  win,
  winlose,
  jp_share,
  jp_win,
  synced_at,
  created_at,
  updated_at
)
SELECT
  user_id,
  'total',
  'all',
  period_start,
  period_end,
  login_id,
  'U',
  count,
  turnover,
  bet,
  win,
  winlose,
  jp_share,
  jp_win,
  NOW(),
  NOW(),
  NOW()
FROM tmp_total_transaction_stats
ON DUPLICATE KEY UPDATE
  period_start = VALUES(period_start),
  period_end = VALUES(period_end),
  login_id = VALUES(login_id),
  amount_unit = 'U',
  count = VALUES(count),
  turnover = VALUES(turnover),
  bet = VALUES(bet),
  win = VALUES(win),
  winlose = VALUES(winlose),
  jp_share = VALUES(jp_share),
  jp_win = VALUES(jp_win),
  synced_at = NOW(),
  updated_at = NOW();

DROP TEMPORARY TABLE IF EXISTS tmp_m7_daily_transaction_stats;
DROP TEMPORARY TABLE IF EXISTS tmp_weekly_transaction_stats;
DROP TEMPORARY TABLE IF EXISTS tmp_total_transaction_stats;
