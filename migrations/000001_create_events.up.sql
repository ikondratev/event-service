CREATE TABLE IF NOT EXISTS events (
    id         BIGSERIAL PRIMARY KEY,
    kind       TEXT        NOT NULL,
    data       TEXT        NOT NULL,
    status     TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);