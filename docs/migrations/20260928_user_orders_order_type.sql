ALTER TABLE `user_orders`
  ADD COLUMN `order_type` VARCHAR(32) NOT NULL DEFAULT 'purchase'
    COMMENT 'purchase=新购套餐 renewal=续费'
    AFTER `product_id`;

UPDATE `user_orders` SET `order_type` = 'purchase' WHERE `order_type` = '' OR `order_type` IS NULL;
