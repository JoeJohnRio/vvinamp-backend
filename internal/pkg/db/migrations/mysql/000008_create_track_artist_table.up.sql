CREATE TABLE track_artist (
    track_id CHAR(36),
    artist_id CHAR(36),
    role VARCHAR(50) DEFAULT 'Main',
    PRIMARY KEY (track_id, artist_id),
    FOREIGN KEY (track_id) REFERENCES tracks(id) ON DELETE CASCADE,
    FOREIGN KEY (artist_id) REFERENCES artists(id) ON DELETE CASCADE
) ENGINE=InnoDB;