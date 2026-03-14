-- Access review campaign source fetch lifecycle tracking

DO $$
BEGIN
    CREATE TYPE access_review_campaign_source_fetch_status AS ENUM (
        'QUEUED',
        'FETCHING',
        'SUCCESS',
        'FAILED'
    );
EXCEPTION
    WHEN duplicate_object THEN NULL;
END$$;

CREATE TABLE IF NOT EXISTS access_review_campaign_source_fetches (
    tenant_id TEXT NOT NULL,
    access_review_campaign_id TEXT NOT NULL REFERENCES access_review_campaigns(id) ON DELETE CASCADE,
    access_source_id TEXT NOT NULL REFERENCES access_sources(id) ON DELETE CASCADE,
    status access_review_campaign_source_fetch_status NOT NULL DEFAULT 'QUEUED',
    fetched_accounts_count INTEGER NOT NULL DEFAULT 0,
    attempt_count INTEGER NOT NULL DEFAULT 0,
    last_error TEXT,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    PRIMARY KEY (access_review_campaign_id, access_source_id)
);

CREATE INDEX IF NOT EXISTS idx_arc_source_fetches_tenant_id
ON access_review_campaign_source_fetches (tenant_id);

CREATE INDEX IF NOT EXISTS idx_arc_source_fetches_status_updated_at
ON access_review_campaign_source_fetches (status, updated_at);

CREATE INDEX IF NOT EXISTS idx_arc_source_fetches_campaign_id
ON access_review_campaign_source_fetches (access_review_campaign_id);
