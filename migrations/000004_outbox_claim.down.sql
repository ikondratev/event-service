DROP INDEX IF EXISTS idx_outbox_claimable;

CREATE INDEX IF NOT EXISTS idx_outbox_pending
    ON outbox (status, created_at)
    WHERE status = 'pending';

ALTER TABLE outbox
    DROP CONSTRAINT IF EXISTS outbox_status_check;
    
ALTER TABLE outbox
    DROP COLUMN IF EXISTS locked_by,
    DROP COLUMN IF EXISTS locked_at,
    DROP COLUMN IF EXISTS attempts,
    DROP COLUMN IF EXISTS next_retry_at,
    DROP COLUMN IF EXISTS last_error;