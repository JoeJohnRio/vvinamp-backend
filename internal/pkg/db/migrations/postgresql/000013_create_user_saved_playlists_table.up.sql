CREATE TABLE IF NOT EXISTS user_saved_playlists (
    user_id UUID NOT NULL,
    playlist_id UUID NOT NULL,
    saved_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, playlist_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (playlist_id) REFERENCES playlists(id) ON DELETE CASCADE
);

-- Create index for faster lookups by playlist (to count saves)
CREATE INDEX IF NOT EXISTS idx_user_saved_playlists_playlist_id ON user_saved_playlists(playlist_id);

-- Create index for faster lookups by user (to get user's saved playlists)
CREATE INDEX IF NOT EXISTS idx_user_saved_playlists_user_id ON user_saved_playlists(user_id);
