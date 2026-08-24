-- Deduplicate M7PP/BOANZZPP rows created by an older migration and the
-- automatic sync running against a database without the expected unique key.
CREATE TEMPORARY TABLE tmp_m7pp_provider_keep AS
SELECT MIN(id) AS keep_id
FROM game_providers
WHERE platform_code = 'M7PP'
  AND code = 'BOANZZPP';

-- If both provider rows already contain the same game, keep the game attached
-- to the canonical provider before moving the remaining games.
DELETE duplicate_game
FROM games AS duplicate_game
JOIN game_providers AS duplicate_provider
  ON duplicate_provider.id = duplicate_game.provider_id
 AND duplicate_provider.platform_code = 'M7PP'
 AND duplicate_provider.code = 'BOANZZPP'
JOIN tmp_m7pp_provider_keep AS keeper
  ON duplicate_provider.id <> keeper.keep_id
JOIN games AS canonical_game
  ON canonical_game.platform_code = 'M7PP'
 AND canonical_game.provider_id = keeper.keep_id
 AND canonical_game.game_code = duplicate_game.game_code;

UPDATE games AS game_row
JOIN game_providers AS duplicate_provider
  ON duplicate_provider.id = game_row.provider_id
 AND duplicate_provider.platform_code = 'M7PP'
 AND duplicate_provider.code = 'BOANZZPP'
JOIN tmp_m7pp_provider_keep AS keeper
  ON duplicate_provider.id <> keeper.keep_id
SET game_row.provider_id = keeper.keep_id;

DELETE duplicate_provider
FROM game_providers AS duplicate_provider
JOIN tmp_m7pp_provider_keep AS keeper
  ON duplicate_provider.id <> keeper.keep_id
WHERE duplicate_provider.platform_code = 'M7PP'
  AND duplicate_provider.code = 'BOANZZPP';

UPDATE game_providers
SET name = 'PP', status = 1, type_id = 2
WHERE id = (SELECT keep_id FROM tmp_m7pp_provider_keep);

DROP TEMPORARY TABLE tmp_m7pp_provider_keep;
