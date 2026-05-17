CREATE TABLE IF NOT EXISTS genres (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT
);

CREATE TABLE IF NOT EXISTS albums (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(100),
    artist_id UUID,
    release_date DATE,
    cover_art TEXT,
    genre_id UUID,
    FOREIGN KEY (artist_id) REFERENCES artists(id) ON DELETE CASCADE,
    FOREIGN KEY (genre_id) REFERENCES genres(id)
);

CREATE TABLE IF NOT EXISTS podcast_series (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(150) NOT NULL,
    description TEXT,
    cover_art TEXT,
    genre_id UUID,
    publisher VARCHAR(100),
    FOREIGN KEY (genre_id) REFERENCES genres(id)
);

CREATE TABLE IF NOT EXISTS podcast_episodes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    series_id UUID NOT NULL,
    title VARCHAR(200) NOT NULL,
    episode_number INT,
    season_number INT,
    description TEXT,
    release_date DATE,
    duration INT NOT NULL CHECK (duration > 0),
    audio_file_url TEXT NOT NULL,
    cover_art TEXT,
    FOREIGN KEY (series_id) REFERENCES podcast_series(id) ON DELETE CASCADE
);