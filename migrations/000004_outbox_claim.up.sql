ALTER TABLE outbox
    ADD COLUMN IF NOT EXISTS locked_by    TEXT,
    ADD COLUMN IF NOT EXISTS locked_at    TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS attempts     INT          NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS next_retry_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS last_error   TEXT;

ALTER TABLE outbox
    ADD CONSTRAINT outbox_status_check
    CHECK (status IN ('pending', 'processing', 'sent', 'failed'));

DROP INDEX IF EXISTS idx_outbox_pending;

CREATE INDEX IF NOT EXISTS idx_outbox_claimable
    ON outbox (created_at)
    WHERE status IN ('pending', 'processing', 'failed');    