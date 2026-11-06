DROP INDEX IF EXISTS idx_tracks_youtube_id;
ALTER TABLE tracks DROP COLUMN IF EXISTS youtube_id;
