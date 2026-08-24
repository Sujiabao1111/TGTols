-- ============================================
-- Migration: migration_0219.sql
-- Description: add game platform isolation and seed M7 providers
-- ============================================

ALTER TABLE game_providers
  ADD COLUMN platform_code VARCHAR(32) NOT NULL DEFAULT 'HEDOC' COMMENT '游戏平台代码: HEDOC/M7' AFTER id;

ALTER TABLE games
  ADD COLUMN platform_code VARCHAR(32) NOT NULL DEFAULT 'HEDOC' COMMENT '游戏平台代码: HEDOC/M7' AFTER id;

UPDATE game_providers SET platform_code = 'HEDOC' WHERE platform_code IS NULL OR platform_code = '';
UPDATE games SET platform_code = 'HEDOC' WHERE platform_code IS NULL OR platform_code = '';

ALTER TABLE game_providers DROP INDEX idx_provider_code;
ALTER TABLE game_providers
  ADD UNIQUE KEY idx_provider_platform_code (platform_code, code),
  ADD KEY idx_game_providers_platform_status (platform_code, status);

ALTER TABLE games DROP INDEX idx_provider_game_code;
ALTER TABLE games DROP INDEX idx_provider_type;
ALTER TABLE games
  ADD UNIQUE KEY idx_platform_provider_game_code (platform_code, provider_id, game_code),
  ADD KEY idx_platform_provider_type (platform_code, provider_id, game_type_id),
  ADD KEY idx_games_platform_status (platform_code, status);

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



