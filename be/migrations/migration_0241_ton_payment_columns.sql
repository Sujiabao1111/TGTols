-- Add TON payment columns required by the PaymentOrder model.
-- Safe for installations where the TON migration was not applied yet.
ALTER TABLE `payment_orders`
  ADD COLUMN `ton_amount` DECIMAL(20,9) NOT NULL DEFAULT 0 AFTER `stars_amount`,
  ADD COLUMN `ton_rate` DECIMAL(20,8) NOT NULL DEFAULT 0 AFTER `ton_amount`,
  ADD COLUMN `ton_tx_hash` VARCHAR(128) NOT NULL DEFAULT '' AFTER `ton_rate`,
  ADD COLUMN `ton_sender` VARCHAR(128) NOT NULL DEFAULT '' AFTER `ton_tx_hash`,
  ADD COLUMN `rate_snapshot` TEXT NULL AFTER `ton_sender`;

ALTER TABLE `payment_orders`
  ADD UNIQUE INDEX `idx_payment_ton_tx` (`ton_tx_hash`);
