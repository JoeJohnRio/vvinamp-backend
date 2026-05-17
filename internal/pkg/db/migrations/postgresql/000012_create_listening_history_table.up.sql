CREATE TABLE IF NOT EXISTS listening_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL,
    content_id   UUID NOT NULL,
    content_type VARCHAR(16) NOT NULL, -- 'track', 'album', 'playlist', 'podcast', 'artist', etc
    played_at    TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    progress     INT NOT NULL CHECK (progress >= 0),
    -- Optionally any additional columns (device, source, etc)

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
    -- You manage the foreign key application in code based on content_type.
);