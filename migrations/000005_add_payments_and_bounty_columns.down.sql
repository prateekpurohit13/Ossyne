ALTER TABLE contributions DROP FOREIGN KEY fk_contributions_payment;
ALTER TABLE contributions DROP COLUMN payment_id;
DROP TABLE IF EXISTS payments;
ALTER TABLE tasks DROP COLUMN bounty_escrow_id;