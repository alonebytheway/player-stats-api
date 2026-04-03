-- +goose Up
CREATE TABLE IF NOT EXISTS players(
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    score INTEGER NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS players;
