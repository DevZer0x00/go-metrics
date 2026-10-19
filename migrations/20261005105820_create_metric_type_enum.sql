-- +goose Up
CREATE TYPE metric_type AS ENUM('counter', 'gauge');

-- +goose Down
DROP TYPE metric_type;
