-- +goose Up
ALTER TABLE users ADD hashed_password TEXT NOT NULL DEFAULT 'unset';

-- +goose Down

ALTER TABLE DROP COLUMN hashed_passwords;