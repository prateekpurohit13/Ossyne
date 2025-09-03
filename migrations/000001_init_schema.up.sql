CREATE TABLE users (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    username VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    avatar_url VARCHAR(255),
    github_id VARCHAR(255) UNIQUE,
    reputation_score INT NOT NULL DEFAULT 0,
    roles JSON
) ENGINE=InnoDB;

CREATE TABLE projects (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    owner_id BIGINT NOT NULL,
    title VARCHAR(255) NOT NULL,
    short_desc TEXT,
    repo_url VARCHAR(255) UNIQUE,
    tags JSON,
    visibility ENUM('public', 'private') NOT NULL DEFAULT 'public',
    FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB;

CREATE TABLE tasks (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    project_id BIGINT NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    difficulty_level ENUM('easy', 'medium', 'hard') NOT NULL DEFAULT 'easy',
    estimated_hours INT,
    tags JSON,
    skills_required JSON,
    bounty_amount DECIMAL(10, 2) DEFAULT 0.00,
    status ENUM('open', 'claimed', 'in_progress', 'submitted', 'completed', 'archived') NOT NULL DEFAULT 'open',
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
) ENGINE=InnoDB;