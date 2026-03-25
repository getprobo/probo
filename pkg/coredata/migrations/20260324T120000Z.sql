ALTER TABLE trust_centers
ADD COLUMN search_engine_indexing TEXT NOT NULL DEFAULT 'NOT_INDEXABLE';

ALTER TABLE trust_centers
ALTER COLUMN search_engine_indexing DROP DEFAULT;
-- Add description to campaigns and decision audit trail

-- 1. Campaign description
ALTER TABLE access_review_campaigns ADD COLUMN description TEXT NOT NULL DEFAULT '';

-- 2. Decision audit trail: immutable log of every decision recorded
CREATE TABLE access_entry_decision_history (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    access_entry_id TEXT NOT NULL REFERENCES access_entries(id) ON DELETE CASCADE,
    decision access_entry_decision NOT NULL,
    decision_note TEXT,
    decided_by TEXT,
    decided_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX idx_access_entry_decision_history_tenant_id
    ON access_entry_decision_history(tenant_id);
CREATE INDEX idx_access_entry_decision_history_entry_id
    ON access_entry_decision_history(access_entry_id);
CREATE INDEX idx_access_entry_decision_history_entry_decided_at
    ON access_entry_decision_history(access_entry_id, decided_at);
