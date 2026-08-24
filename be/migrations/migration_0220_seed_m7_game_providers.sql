-- ============================================
-- Migration: migration_0220_seed_m7_game_providers.sql
-- Description: seed/update M7 game providers for provider-level switches
-- ============================================

INSERT INTO game_providers (platform_code, code, name, status, type_id, api_config, created_at)
VALUES
  ('M7', 'JDB_SLOT', 'JDB', 1, 2, JSON_OBJECT('wallet_mode', 'transfer', 'api_url', 'https://api.ddhh.work/'), NOW()),
  ('M7', 'PP_SLOT', 'PP', 1, 2, JSON_OBJECT('wallet_mode', 'transfer', 'api_url', 'https://api.ddhh.work/'), NOW()),
  ('M7', 'PG_SLOT', 'PG', 1, 2, JSON_OBJECT('wallet_mode', 'transfer', 'api_url', 'https://api.ddhh.work/'), NOW()),
  ('M7', 'ONE_PT_SLOT', 'Playtech', 1, 2, JSON_OBJECT('wallet_mode', 'transfer', 'api_url', 'https://api.ddhh.work/'), NOW()),
  ('M7', 'ONE_MG_SLOT', 'Microgaming+', 1, 2, JSON_OBJECT('wallet_mode', 'transfer', 'api_url', 'https://api.ddhh.work/'), NOW()),
  ('M7', 'DNG_HS_SLOT', 'Hacksaw', 1, 2, JSON_OBJECT('wallet_mode', 'transfer', 'api_url', 'https://api.ddhh.work/'), NOW()),
  ('M7', 'DNG_STM_SLOT', 'Slotmill', 1, 2, JSON_OBJECT('wallet_mode', 'transfer', 'api_url', 'https://api.ddhh.work/'), NOW()),
  ('M7', 'PP_LIVE', 'PP真人', 1, 3, JSON_OBJECT('wallet_mode', 'transfer', 'api_url', 'https://api.ddhh.work/'), NOW()),
  ('M7', 'ONE_PT_LIVE', 'Playtech', 1, 3, JSON_OBJECT('wallet_mode', 'transfer', 'api_url', 'https://api.ddhh.work/'), NOW()),
  ('M7', 'EVO_LIVE', 'Evolution', 1, 3, JSON_OBJECT('wallet_mode', 'transfer', 'api_url', 'https://api.ddhh.work/'), NOW())
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  type_id = VALUES(type_id),
  api_config = VALUES(api_config);
