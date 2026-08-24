-- Migration 0223: prefer English M7 game cover images for existing synced rows.
-- Future syncs use image2 first; this updates rows already stored from image1.

UPDATE games
SET img_url = REPLACE(img_url, '_cn.', '_en.')
WHERE platform_code = 'M7'
  AND img_url LIKE '%_cn.%';
