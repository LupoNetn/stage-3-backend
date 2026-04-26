-- +goose Up
-- add unique constraint to name
ALTER TABLE profiles ADD CONSTRAINT profiles_name_unique UNIQUE (name);

-- drop sample_size
ALTER TABLE profiles DROP COLUMN IF EXISTS sample_size;

-- add country_name
ALTER TABLE profiles ADD COLUMN IF NOT EXISTS country_name VARCHAR;

-- alter country_id
ALTER TABLE profiles ALTER COLUMN country_id TYPE VARCHAR(2);

-- +goose Down
ALTER TABLE profiles ALTER COLUMN country_id TYPE VARCHAR(10);
ALTER TABLE profiles DROP COLUMN IF EXISTS country_name;
ALTER TABLE profiles ADD COLUMN IF NOT EXISTS sample_size INTEGER;
ALTER TABLE profiles DROP CONSTRAINT IF EXISTS profiles_name_unique;
