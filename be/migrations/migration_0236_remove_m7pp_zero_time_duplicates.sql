-- Remove M7PP rows written by the pre-fix timestamp parser. The old parser
-- also used a different record hash (betId), so corrected rows were inserted
-- beside the bad 1970 rows instead of updating them.
DELETE bad
FROM user_game_transaction_details bad
JOIN user_game_transaction_details good
  ON good.user_id = bad.user_id
 AND good.game_provider_code = bad.game_provider_code
 AND good.external_id = bad.external_id
 AND good.id <> bad.id
WHERE bad.record_hash IS NOT NULL
  AND bad.start_date <= '1970-01-01 00:00:01'
  AND bad.end_date <= '1970-01-01 00:00:01'
  AND good.start_date > '1970-01-01 00:00:01';
