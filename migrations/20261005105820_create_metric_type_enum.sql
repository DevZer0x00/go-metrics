-- +goose Up
CREATE TYPE metric_type AS ENUM('counter', 'gauge');

-- +goose Down
SELECT 'DROP TYPE metric_type';
