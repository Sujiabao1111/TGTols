-- Seed the BOAN Pragmatic Play transfer-wallet platform.
INSERT INTO game_providers (platform_code, code, name, status, type_id, api_config, created_at)
VALUES (
  'M7PP',
  'BOANZZPP',
  'PP',
  1,
  2,
  JSON_OBJECT('wallet_mode', 'transfer', 'api_url', 'https://test-gw.boan.games'),
  NOW()
)
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  type_id = VALUES(type_id),
  api_config = VALUES(api_config);
