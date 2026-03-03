CREATE TABLE IF NOT EXISTS track_artist (
    track_id UUID,
    artist_id UUID,
    "role" VARCHAR(50) DEFAULT 'Main',
    PRIMARY KEY (track_id, artist_id),
    FOREIGN KEY (track_id) REFERENCES tracks(id) ON DELETE CASCADE,
    FOREIGN KEY (artist_id) REFERENCES artists(id) ON DELETE CASCADE
);
