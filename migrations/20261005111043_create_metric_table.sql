-- +goose Up
CREATE TABLE metric (
    id VARCHAR(255) NOT NULL,
    mtype metric_type NOT NULL,
    delta BIGINT DEFAULT NULL,
    value double precision DEFAULT NULL,
    hash varchar(32) NOT NULL,
    PRIMARY KEY (mtype, hash)
);

-- +goose Down
DROP TABLE metric;
