-- ============================================
-- Migration: migration_0221_fix_m7_live_casino_type.sql
-- Description: ensure M7 live providers and games belong to Live Casino
-- ============================================

SET @casino_type_id := (
  SELECT id
  FROM game_types
  WHERE code = 'CASINO'
  LIMIT 1
);

UPDATE game_providers
SET type_id = @casino_type_id
WHERE @casino_type_id IS NOT NULL
  AND platform_code = 'M7'
  AND code IN ('PP_LIVE', 'ONE_PT_LIVE', 'EVO_LIVE');

UPDATE games g
JOIN game_providers p
  ON p.id = g.provider_id
 AND p.platform_code = g.platform_code
SET g.game_type_id = @casino_type_id
WHERE @casino_type_id IS NOT NULL
  AND g.platform_code = 'M7'
  AND p.code IN ('PP_LIVE', 'ONE_PT_LIVE', 'EVO_LIVE');
