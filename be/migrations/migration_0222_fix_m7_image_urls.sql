-- Migration 0222: fix M7 image URLs whose CDN host was treated as an API-relative path
-- Example:
--   https://api.ddhh.work/dluqiiaw.cnwzhy.com/jdb-assetsv3/games/14006/14006_en.png
-- becomes:
--   https://dluqiiaw.cnwzhy.com/jdb-assetsv3/games/14006/14006_en.png

UPDATE games
SET img_url = CONCAT('https://', SUBSTRING(img_url, LENGTH('https://api.ddhh.work/') + 1))
WHERE platform_code = 'M7'
  AND img_url LIKE 'https://api.ddhh.work/%'
  AND SUBSTRING_INDEX(SUBSTRING(img_url, LENGTH('https://api.ddhh.work/') + 1), '/', 1) LIKE '%.%';
