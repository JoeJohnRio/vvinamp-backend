CREATE TABLE IF NOT EXISTS artists (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    bio TEXT,
    profile_image TEXT,
    verified BOOLEAN DEFAULT FALSE,
    monthly_listeners INT DEFAULT 0
);
