CREATE TABLE IF NOT EXISTS songs(
    id UUID DEFAULT gen_random_uuid(),
    group_name VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    release_date DATE,
    text TEXT,
    link VARCHAR(2048),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY(id)
);

CREATE OR REPLACE FUNCTION update_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER songs_update_updated_at
BEFORE UPDATE ON songs
FOR EACH ROW
EXECUTE FUNCTION update_timestamp();
