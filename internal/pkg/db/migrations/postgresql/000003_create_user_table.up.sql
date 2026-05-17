CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) NOT NULL CONSTRAINT uni_users_username UNIQUE,
    email VARCHAR(100) NOT NULL CONSTRAINT uni_users_email UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    profile_picture TEXT,
    join_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    subscription_type VARCHAR(20) DEFAULT 'Free' CHECK (subscription_type IN ('Free', 'Premium')),
    last_login TIMESTAMP NULL
);