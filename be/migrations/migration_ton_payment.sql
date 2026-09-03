-- TON native payment support. Orders reuse payment_orders and remain idempotent.
-- Configure TON_TREASURY_ADDRESS and TON_USD_RATE in the server environment.
ALTER TABLE payment_orders
  ADD COLUMN ton_amount DECIMAL(20,9) NOT NULL DEFAULT 0,
  ADD COLUMN ton_rate DECIMAL(20,8) NOT NULL DEFAULT 0,
  ADD COLUMN ton_tx_hash VARCHAR(128) NOT NULL DEFAULT '',
  ADD COLUMN ton_sender VARCHAR(128) NOT NULL DEFAULT '',
  ADD INDEX idx_payment_ton_dst (dst_code, status),
  ADD INDEX idx_payment_ton_tx (ton_tx_hash);
