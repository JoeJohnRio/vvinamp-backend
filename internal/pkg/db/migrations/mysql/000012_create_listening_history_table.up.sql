CREATE TABLE listening_history (
    user_id CHAR(36),
    track_id CHAR(36),
    played_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    progress INT NOT NULL CHECK (progress >= 0),
    PRIMARY KEY (user_id, track_id, played_at),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (track_id) REFERENCES tracks(id) ON DELETE CASCADE
) ENGINE=InnoDB;