-- Create payments table FIRST
CREATE TABLE payments (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    contribution_id BIGINT NULL, -- Link to the contribution this payment is for
    user_id BIGINT NOT NULL,     -- The user receiving/initiating the payment
    amount DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    status ENUM('pending', 'escrowed', 'released', 'failed', 'refunded') NOT NULL DEFAULT 'pending',
    type ENUM('bounty_payout', 'escrow_deposit', 'admin_transfer') NOT NULL,
    transaction_id VARCHAR(255) UNIQUE, -- External ID from payment gateway
    payment_gateway VARCHAR(255),       -- e.g., "Stripe", "PayPal"
    payment_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (contribution_id) REFERENCES contributions(id) ON DELETE SET NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB;

-- Then add BountyEscrowID to tasks table (no dependency on payments table)
ALTER TABLE tasks ADD COLUMN bounty_escrow_id VARCHAR(255) NULL;

-- Then add PaymentID to contributions table
ALTER TABLE contributions ADD COLUMN payment_id BIGINT NULL;

-- Finally, add the foreign key constraint to contributions (now payments table exists)
ALTER TABLE contributions ADD CONSTRAINT fk_contributions_payment FOREIGN KEY (payment_id) REFERENCES payments(id) ON DELETE SET NULL;

-- Add index to contribution_id in payments table for faster lookups
CREATE INDEX idx_payments_contribution_id ON payments (contribution_id);