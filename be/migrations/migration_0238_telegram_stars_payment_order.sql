-- Telegram Stars payment order fields.
--
-- This migration is intentionally numbered after the existing migrations because
-- migration_0219.sql is already used for an unrelated schema change.  The Stars
-- migration was previously stored as migration_0219_telegram_stars.sql and could
-- therefore be skipped by deployment tooling that tracks numeric versions.
ALTER TABLE `payment_orders`
  ADD COLUMN `telegram_user_id` VARCHAR(32) NOT NULL DEFAULT '',
  ADD COLUMN `stars_amount` BIGINT NOT NULL DEFAULT 0,
  ADD COLUMN `invoice_payload` VARCHAR(128) NULL,
  ADD COLUMN `telegram_charge_id` VARCHAR(128) NULL,
  ADD COLUMN `provider_charge_id` VARCHAR(128) NOT NULL DEFAULT '',
  ADD COLUMN `rate_snapshot` TEXT NULL,
  ADD INDEX `idx_payment_telegram_user` (`telegram_user_id`),
  ADD UNIQUE INDEX `idx_payment_invoice_payload` (`invoice_payload`),
  ADD UNIQUE INDEX `idx_payment_telegram_charge` (`telegram_charge_id`);
