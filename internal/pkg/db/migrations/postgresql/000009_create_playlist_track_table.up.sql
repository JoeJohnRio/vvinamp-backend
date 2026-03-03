CREATE TABLE IF NOT EXISTS playlist_track (
    playlist_id UUID NOT NULL,
    track_id UUID NOT NULL,
    position INT NOT NULL,
    added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (playlist_id, track_id),
    FOREIGN KEY (playlist_id) REFERENCES playlists(id) ON DELETE CASCADE,
    FOREIGN KEY (track_id) REFERENCES tracks(id) ON DELETE CASCADE
);

-- Create index for faster lookups by playlist (to get songs in order)
CREATE INDEX IF NOT EXISTS idx_playlist_track_playlist_id ON playlist_track(playlist_id, position);