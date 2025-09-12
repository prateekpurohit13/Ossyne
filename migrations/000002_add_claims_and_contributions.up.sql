CREATE TABLE claims (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    task_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    claim_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status ENUM('pending', 'accepted', 'rejected', 'withdrawn') NOT NULL DEFAULT 'pending',
    mentor_id BIGINT NULL, -- Optional mentor for this claim
    notes TEXT,
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (mentor_id) REFERENCES users(id) ON DELETE SET NULL, -- Mentor can be deleted, but claim remains
    UNIQUE (task_id, user_id) -- A user can only claim a specific task once
) ENGINE=InnoDB;

CREATE TABLE contributions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    task_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    pr_url VARCHAR(512) NOT NULL,
    pr_commit_hashes JSON, -- Store an array of commit hashes for verification
    submitted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    verification_status ENUM('unverified', 'auto_verified', 'manual_verified', 'rejected') NOT NULL DEFAULT 'unverified',
    accepted_at TIMESTAMP NULL,
    payout_amount DECIMAL(10, 2) DEFAULT 0.00, -- Copied from task bounty upon acceptance
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE (task_id, user_id) -- A user can only submit one contribution for a specific task
) ENGINE=InnoDB;