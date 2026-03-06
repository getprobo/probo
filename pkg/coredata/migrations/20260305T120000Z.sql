ALTER TABLE emails
ADD COLUMN reply_to TEXT,
ADD COLUMN unsubscribe_url TEXT,
ADD COLUMN mailing_list_update_id TEXT REFERENCES mailing_list_updates(id) ON DELETE SET NULL;

CREATE TYPE mailing_list_update_status AS ENUM ('DRAFT', 'ENQUEUED', 'PROCESSING', 'SENT');

CREATE TABLE mailing_list_updates (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON UPDATE CASCADE ON DELETE CASCADE,
    mailing_list_id TEXT NOT NULL REFERENCES mailing_lists(id) ON UPDATE CASCADE ON DELETE CASCADE,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    status mailing_list_update_status NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);
