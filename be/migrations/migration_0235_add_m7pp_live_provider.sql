-- BOAN uses BOANZZPP for both slot and live games. The local UI needs a
-- separate provider row because providers are assigned to one game type.
UPDATE game_providers
SET name = 'PP', status = 1
WHERE platform_code = 'M7PP'
  AND code = 'BOANZZPP';

INSERT INTO game_providers (platform_code, code, name, status, type_id, api_config, created_at)
SELECT
  'M7PP',
  'BOANZZPP_LIVE',
  'PP真人',
  1,
  gt.id,
  JSON_OBJECT(
    'wallet_mode', 'transfer',
    'api_url', 'https://test-gw.boan.games',
    'vendor_code', 'BOANZZPP'
  ),
  NOW()
FROM game_types AS gt
WHERE gt.code = 'CASINO'
LIMIT 1
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  status = 1,
  type_id = VALUES(type_id),
  api_config = VALUES(api_config);

UPDATE games AS game_row
JOIN game_providers AS live_provider
  ON live_provider.platform_code = 'M7PP'
 AND live_provider.code = 'BOANZZPP_LIVE'
SET game_row.provider_id = live_provider.id,
    game_row.game_type_id = live_provider.type_id,
    game_row.status = 1
WHERE game_row.platform_code = 'M7PP'
  AND game_row.game_code = 'BOANZZPP_ppt_live';
