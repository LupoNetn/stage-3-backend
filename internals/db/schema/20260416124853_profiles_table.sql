-- +goose Up
CREATE TABLE profiles (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    gender VARCHAR(50),
    gender_probability DOUBLE PRECISION,
    sample_size INTEGER,
    age INTEGER,
    age_group VARCHAR(50),
    country_id VARCHAR(10),
    country_probability DOUBLE PRECISION,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE IF EXISTS profiles;
