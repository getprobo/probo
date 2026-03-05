ALTER TABLE emails
ADD COLUMN reply_to TEXT,
ADD COLUMN unsubscribe_url TEXT;

CREATE TYPE compliance_news_status AS ENUM ('DRAFT', 'SENT');

CREATE TABLE compliance_news (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON UPDATE CASCADE ON DELETE CASCADE,
    trust_center_id TEXT NOT NULL REFERENCES trust_centers(id) ON UPDATE CASCADE ON DELETE CASCADE,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    status compliance_news_status NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);
