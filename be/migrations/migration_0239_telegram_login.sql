-- Telegram Mini App login fields.
-- Apply this migration to existing databases before enabling Telegram login.
ALTER TABLE `users`
  ADD COLUMN `telegram_user_id` VARCHAR(32) NULL,
  ADD COLUMN `telegram_username` VARCHAR(64) NOT NULL DEFAULT '',
  ADD COLUMN `telegram_first_name` VARCHAR(128) NOT NULL DEFAULT '',
  ADD COLUMN `telegram_last_name` VARCHAR(128) NOT NULL DEFAULT '',
  ADD COLUMN `telegram_photo_url` VARCHAR(512) NOT NULL DEFAULT '',
  ADD COLUMN `telegram_bound_at` DATETIME NULL,
  ADD UNIQUE INDEX `idx_user_telegram_id` (`telegram_user_id`);
