-- Access review hardening: decisions, lifecycle, idempotency, audit, evidence

DO $$
BEGIN
    ALTER TYPE access_entry_decision ADD VALUE IF NOT EXISTS 'DEFER';
EXCEPTION
    WHEN duplicate_object THEN NULL;
END$$;

DO $$
BEGIN
    ALTER TYPE access_entry_decision ADD VALUE IF NOT EXISTS 'ESCALATE';
EXCEPTION
    WHEN duplicate_object THEN NULL;
END$$;

DO $$
BEGIN
    ALTER TYPE access_review_campaign_status ADD VALUE IF NOT EXISTS 'FAILED';
EXCEPTION
    WHEN duplicate_object THEN NULL;
END$$;

DO $$
BEGIN
    CREATE TYPE access_entry_incremental_tag AS ENUM (
        'NEW',
        'REMOVED',
        'UNCHANGED'
    );
EXCEPTION
    WHEN duplicate_object THEN NULL;
END$$;

DO $$
BEGIN
    CREATE TYPE access_review_remediation_task_status AS ENUM (
        'OPEN',
        'ACKNOWLEDGED',
        'COMPLETED',
        'CANCELLED'
    );
EXCEPTION
    WHEN duplicate_object THEN NULL;
END$$;

ALTER TABLE access_entries
    ADD COLUMN IF NOT EXISTS account_key TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS incremental_tag access_entry_incremental_tag NOT NULL DEFAULT 'NEW';

-- Backfill and normalize old data.
UPDATE access_entries
SET
    account_key = CASE
        WHEN external_id IS NOT NULL AND trim(external_id) <> '' THEN lower(trim(email)) || '|' || trim(external_id)
        ELSE lower(trim(email))
    END
WHERE account_key = '';

CREATE UNIQUE INDEX IF NOT EXISTS idx_access_entries_campaign_source_account_key
ON access_entries (access_review_campaign_id, access_source_id, account_key);

CREATE INDEX IF NOT EXISTS idx_access_entries_campaign_decision
ON access_entries (access_review_campaign_id, decision);

CREATE INDEX IF NOT EXISTS idx_access_entries_campaign_source
ON access_entries (access_review_campaign_id, access_source_id);

CREATE INDEX IF NOT EXISTS idx_access_entries_campaign_decided_at
ON access_entries (access_review_campaign_id, decided_at);

CREATE TABLE IF NOT EXISTS access_entry_decision_events (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    access_entry_id TEXT NOT NULL REFERENCES access_entries(id) ON DELETE CASCADE,
    access_review_campaign_id TEXT NOT NULL REFERENCES access_review_campaigns(id) ON DELETE CASCADE,
    previous_decision access_entry_decision NOT NULL,
    new_decision access_entry_decision NOT NULL,
    decision_note TEXT,
    decided_by TEXT REFERENCES peoples(id) ON DELETE SET NULL,
    decided_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_access_entry_decision_events_tenant_id
ON access_entry_decision_events (tenant_id);

CREATE INDEX IF NOT EXISTS idx_access_entry_decision_events_campaign_id
ON access_entry_decision_events (access_review_campaign_id);

CREATE INDEX IF NOT EXISTS idx_access_entry_decision_events_entry_id
ON access_entry_decision_events (access_entry_id);

CREATE TABLE IF NOT EXISTS access_review_remediation_tasks (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    access_review_campaign_id TEXT NOT NULL REFERENCES access_review_campaigns(id) ON DELETE CASCADE,
    access_entry_id TEXT NOT NULL REFERENCES access_entries(id) ON DELETE CASCADE,
    decided_by TEXT REFERENCES peoples(id) ON DELETE SET NULL,
    status access_review_remediation_task_status NOT NULL DEFAULT 'OPEN',
    status_note TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE (access_entry_id)
);

CREATE INDEX IF NOT EXISTS idx_access_review_remediation_tasks_tenant_id
ON access_review_remediation_tasks (tenant_id);

CREATE INDEX IF NOT EXISTS idx_access_review_remediation_tasks_campaign_id
ON access_review_remediation_tasks (access_review_campaign_id);

CREATE INDEX IF NOT EXISTS idx_access_review_remediation_tasks_status
ON access_review_remediation_tasks (status);

CREATE TABLE IF NOT EXISTS access_review_campaign_validation_checkpoints (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    access_review_campaign_id TEXT NOT NULL REFERENCES access_review_campaigns(id) ON DELETE CASCADE,
    validated_by TEXT REFERENCES peoples(id) ON DELETE SET NULL,
    note TEXT,
    validated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE (access_review_campaign_id)
);

CREATE INDEX IF NOT EXISTS idx_access_review_campaign_validation_checkpoints_tenant_id
ON access_review_campaign_validation_checkpoints (tenant_id);

CREATE TABLE IF NOT EXISTS access_review_campaign_evidence_snapshots (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    access_review_campaign_id TEXT NOT NULL REFERENCES access_review_campaigns(id) ON DELETE CASCADE,
    payload JSONB NOT NULL,
    checksum_sha256 TEXT NOT NULL,
    signature TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE (access_review_campaign_id)
);

CREATE INDEX IF NOT EXISTS idx_access_review_campaign_evidence_snapshots_tenant_id
ON access_review_campaign_evidence_snapshots (tenant_id);
