ALTER TABLE tracks ADD COLUMN IF NOT EXISTS youtube_id text;
CREATE INDEX IF NOT EXISTS idx_tracks_youtube_id ON tracks(youtube_id);
