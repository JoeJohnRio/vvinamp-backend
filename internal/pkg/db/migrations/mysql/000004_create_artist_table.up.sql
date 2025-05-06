CREATE TABLE artists (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    name VARCHAR(100) NOT NULL,
    bio TEXT,
    profile_image TEXT,
    verified BOOLEAN DEFAULT FALSE,
    monthly_listeners INT DEFAULT 0
) ENGINE=InnoDB;