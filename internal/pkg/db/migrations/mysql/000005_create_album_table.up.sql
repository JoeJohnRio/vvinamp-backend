CREATE TABLE genres (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT
) ENGINE=InnoDB;

CREATE TABLE albums (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    title VARCHAR(100),
    artist_id CHAR(36),
    release_date DATE,
    cover_art TEXT,
    genre_id CHAR(36),
    FOREIGN KEY (artist_id) REFERENCES artists(id) ON DELETE CASCADE,
    FOREIGN KEY (genre_id) REFERENCES genres(id)
);
