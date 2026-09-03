ALTER TABLE `payment_orders`
  MODIFY COLUMN `telegram_charge_id` VARCHAR(512) NULL,
  MODIFY COLUMN `provider_charge_id` VARCHAR(512) NOT NULL DEFAULT '';
