-- +goose Up
-- +goose StatementBegin
CREATE TABLE lock
(
    name                  CHARACTER VARYING(255) PRIMARY KEY,
    record_version_number BIGINT,
    data                  BYTEA,
    owner                 CHARACTER VARYING(255)
);
CREATE SEQUENCE lock_rvn CYCLE OWNED BY lock.record_version_number;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP SEQUENCE lock_rvn;

DROP TABLE lock;
-- +goose StatementEnd
