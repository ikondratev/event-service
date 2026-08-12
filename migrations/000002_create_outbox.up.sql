CREATE TABLE IF NOT EXISTS outbox (
    id         BIGSERIAL PRIMARY KEY,
    event_id   BIGINT      NOT NULL REFERENCES events(id),
    topic      TEXT        NOT NULL,
    payload    JSONB       NOT NULL,
    status     TEXT        NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    sent_at    TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_outbox_pending
    ON outbox (status, created_at)
    WHERE status = 'pending';