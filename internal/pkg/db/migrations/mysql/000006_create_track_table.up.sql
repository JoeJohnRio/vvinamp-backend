CREATE TABLE tracks (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    title VARCHAR(100) NOT NULL,
    duration INT NOT NULL CHECK (duration > 0),
    audio_file_url TEXT NOT NULL,
    stream_count INT DEFAULT 0,
    album_id CHAR(36),
    explicit BOOLEAN DEFAULT FALSE,
    lyrics TEXT,
    release_date DATE,
    FOREIGN KEY (album_id) REFERENCES albums(id) ON DELETE SET NULL
) ENGINE=InnoDB;