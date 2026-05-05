-- +goose Up
-- Single column indexes for fast individual filtering
CREATE INDEX IF NOT EXISTS idx_profiles_country_name ON profiles (country_name);
CREATE INDEX IF NOT EXISTS idx_profiles_country_id ON profiles (country_id);
CREATE INDEX IF NOT EXISTS idx_profiles_gender ON profiles (gender);
CREATE INDEX IF NOT EXISTS idx_profiles_age ON profiles (age);

-- Composite index covering all three for complex queries
CREATE INDEX IF NOT EXISTS idx_profiles_country_gender_age ON profiles (country_id, gender, age);

-- +goose Down
DROP INDEX IF EXISTS idx_profiles_country_gender_age;
DROP INDEX IF EXISTS idx_profiles_age;
DROP INDEX IF EXISTS idx_profiles_gender;
DROP INDEX IF EXISTS idx_profiles_country_id;
DROP INDEX IF EXISTS idx_profiles_country_name;
