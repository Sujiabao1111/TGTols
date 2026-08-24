-- Repair M7 cover URLs reintroduced by periodic game synchronization before
-- normalizeM7ImageURL learned to unwrap CDN hosts embedded below the API host.

UPDATE games AS g
JOIN game_providers AS p ON p.id = g.provider_id
SET g.img_url = CONCAT(
  'https://',
  SUBSTRING(g.img_url, LENGTH('https://api.ddhh.work/') + 1)
)
WHERE g.platform_code = 'M7'
  AND p.code IN ('ONE_PT_SLOT', 'ONE_PT_LIVE')
  AND g.img_url LIKE 'https://api.ddhh.work/%'
  AND SUBSTRING_INDEX(
    SUBSTRING(g.img_url, LENGTH('https://api.ddhh.work/') + 1),
    '/',
    1
  ) LIKE '%.%';
