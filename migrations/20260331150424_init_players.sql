-- +goose Up
CREATE TABLE IF NOT EXISTS players(
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    kills INTEGER DEFAULT 0,
    deaths INTEGER DEFAULT 0,
    matches INTEGER DEFAULT 0
);

-- +goose Down
DROP TABLE IF EXISTS players;
