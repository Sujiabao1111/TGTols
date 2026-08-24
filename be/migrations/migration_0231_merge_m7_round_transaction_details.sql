-- ============================================
-- Migration: migration_0231_merge_m7_round_transaction_details.sql
-- Description: merge split M7 bet/payout detail rows into one row per round
-- ============================================

START TRANSACTION;

DROP TEMPORARY TABLE IF EXISTS tmp_m7_round_detail_merge;

CREATE TEMPORARY TABLE tmp_m7_round_detail_merge AS
SELECT
  d.user_id,
  d.stat_date,
  d.game_provider_code,
  d.round_id,
  COALESCE(
    MAX(CASE
      WHEN d.record_hash = SHA2(CONCAT(d.user_id, '|', d.stat_date, '|m7-round|', d.game_provider_code, '|', d.round_id), 256)
        THEN d.id
      ELSE NULL
    END),
    MAX(d.id)
  ) AS keep_id,
  SHA2(CONCAT(d.user_id, '|', d.stat_date, '|m7-round|', d.game_provider_code, '|', d.round_id), 256) AS merged_record_hash,
  COALESCE(
    SUBSTRING_INDEX(
      GROUP_CONCAT(NULLIF(d.provider_user_id, '') ORDER BY COALESCE(d.end_date, d.start_date, d.created_at) DESC, d.id DESC SEPARATOR ','),
      ',',
      1
    ),
    ''
  ) AS provider_user_id,
  COALESCE(
    SUBSTRING_INDEX(
      GROUP_CONCAT(NULLIF(d.login_id, '') ORDER BY COALESCE(d.end_date, d.start_date, d.created_at) DESC, d.id DESC SEPARATOR ','),
      ',',
      1
    ),
    ''
  ) AS login_id,
  COALESCE(
    SUBSTRING_INDEX(
      GROUP_CONCAT(NULLIF(d.game_code, '') ORDER BY COALESCE(d.end_date, d.start_date, d.created_at) DESC, d.id DESC SEPARATOR ','),
      ',',
      1
    ),
    ''
  ) AS game_code,
  COALESCE(
    SUBSTRING_INDEX(
      GROUP_CONCAT(NULLIF(d.external_id, '') ORDER BY COALESCE(d.end_date, d.start_date, d.created_at) DESC, d.id DESC SEPARATOR ','),
      ',',
      1
    ),
    ''
  ) AS external_id,
  MIN(d.type) AS type,
  MIN(d.start_date) AS start_date,
  MAX(d.end_date) AS end_date,
  COALESCE(
    SUBSTRING_INDEX(
      GROUP_CONCAT(d.start_balance ORDER BY COALESCE(d.start_date, d.end_date, d.created_at) ASC, d.id ASC SEPARATOR ','),
      ',',
      1
    ),
    0
  ) AS start_balance,
  COALESCE(
    SUBSTRING_INDEX(
      GROUP_CONCAT(d.end_balance ORDER BY COALESCE(d.end_date, d.start_date, d.created_at) DESC, d.id DESC SEPARATOR ','),
      ',',
      1
    ),
    0
  ) AS end_balance,
  COALESCE(SUM(d.deposit), 0) AS deposit,
  COALESCE(SUM(d.turnover), 0) AS turnover,
  COALESCE(SUM(d.bet), 0) AS bet,
  COALESCE(SUM(d.win), 0) AS win,
  COALESCE(SUM(d.winlose), 0) AS winlose,
  COALESCE(SUM(d.jp_share), 0) AS jp_share,
  COALESCE(SUM(d.jp_win), 0) AS jp_win,
  JSON_OBJECT(
    'source', 'M7_GAME_HISTORY_AGGREGATED',
    'roundId', d.round_id,
    'mergedDetailIds', GROUP_CONCAT(d.id ORDER BY d.id ASC SEPARATOR ',')
  ) AS remark
FROM user_game_transaction_details d
JOIN game_providers gp
  ON gp.id = d.game_provider_code
 AND gp.platform_code = 'M7'
WHERE COALESCE(d.round_id, '') <> ''
GROUP BY d.user_id, d.stat_date, d.game_provider_code, d.round_id
HAVING COUNT(*) > 1;

DELETE d
FROM user_game_transaction_details d
JOIN tmp_m7_round_detail_merge m
  ON m.user_id = d.user_id
 AND m.stat_date = d.stat_date
 AND m.game_provider_code = d.game_provider_code
 AND m.round_id = d.round_id
 AND d.id <> m.keep_id;

UPDATE user_game_transaction_details d
JOIN tmp_m7_round_detail_merge m
  ON m.keep_id = d.id
SET
  d.provider_user_id = m.provider_user_id,
  d.login_id = m.login_id,
  d.game_code = m.game_code,
  d.external_id = m.external_id,
  d.type = m.type,
  d.start_date = m.start_date,
  d.end_date = m.end_date,
  d.start_balance = m.start_balance,
  d.end_balance = m.end_balance,
  d.deposit = m.deposit,
  d.turnover = m.turnover,
  d.bet = m.bet,
  d.win = m.win,
  d.winlose = m.winlose,
  d.jp_share = m.jp_share,
  d.jp_win = m.jp_win,
  d.remark = m.remark,
  d.record_hash = m.merged_record_hash,
  d.updated_at = CURRENT_TIMESTAMP;

DROP TEMPORARY TABLE IF EXISTS tmp_m7_round_detail_merge;

COMMIT;
