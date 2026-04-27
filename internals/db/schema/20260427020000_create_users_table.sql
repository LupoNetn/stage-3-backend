-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    github_id VARCHAR UNIQUE NOT NULL,
    username VARCHAR NOT NULL,
    email VARCHAR NOT NULL,
    avatar_url VARCHAR,
    role VARCHAR NOT NULL DEFAULT 'analyst',
    is_active BOOLEAN NOT NULL DEFAULT true,
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users;
-- +goose StatementEnd
