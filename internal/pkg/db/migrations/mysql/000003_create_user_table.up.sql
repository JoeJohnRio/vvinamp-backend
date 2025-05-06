CREATE TABLE users (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    profile_picture TEXT,
    join_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    subscription_type VARCHAR(20) DEFAULT 'Free' CHECK (subscription_type IN ('Free', 'Premium')),
    last_login TIMESTAMP NULL
) ENGINE=InnoDB;