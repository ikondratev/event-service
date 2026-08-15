CREATE TABLE IF NOT EXISTS idempotency_keys (
    key          TEXT        PRIMARY KEY,
    request_hash TEXT        NOT NULL,
    event_id     BIGINT      REFERENCES events(id),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_outbox_event_id ON outbox (event_id);