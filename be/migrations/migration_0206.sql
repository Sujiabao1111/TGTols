-- ============================================
-- 迁移脚本: migration_0206.sql
-- 说明: 2026-05-25 扩大 payment_orders.pay_url 字段，兼容菲律宾 QuantixCore 长 payInfo
-- ============================================

ALTER TABLE `payment_orders`
  MODIFY COLUMN `pay_url` TEXT DEFAULT NULL COMMENT '支付链接';

-- 执行命令: mysql -u root -p tols < be/migrations/migration_0206.sql
