-- BOAN exposes Pragmatic Play as BOANZZPP. Older M7PP deployments seeded
-- the display abbreviation PP as the API provider code, so game sync could
-- not find a matching provider and left the platform empty.
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
  status = 1,
  type_id = VALUES(type_id),
  api_config = VALUES(api_config);

-- Hide the legacy empty provider so the frontend does not display two PP tabs.
UPDATE game_providers
SET status = 0
WHERE platform_code = 'M7PP'
  AND code = 'PP';
