DROP TRIGGER IF EXISTS songs_update_updated_at ON songs;
DROP TRIGGER IF EXISTS groups_update_updated_at ON groups;

DROP FUNCTION IF EXISTS update_timestamp();

DROP TABLE IF EXISTS songs;
DROP TABLE IF EXISTS groups;
