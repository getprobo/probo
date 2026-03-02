CREATE TYPE webhook_event_type AS ENUM (
    'meeting:created',
    'meeting:updated',
    'meeting:deleted',
    'vendor:created',
    'vendor:updated',
    'vendor:deleted'
);

CREATE TABLE webhook_subscriptions (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON UPDATE CASCADE ON DELETE CASCADE,
    endpoint_url TEXT NOT NULL,
    selected_events webhook_event_type[] NOT NULL,
    encrypted_signing_secret BYTEA NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE webhook_data (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON UPDATE CASCADE ON DELETE CASCADE,
    event_type webhook_event_type NOT NULL,
    data JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    processed_at TIMESTAMP WITH TIME ZONE
);

CREATE TYPE webhook_event_status AS ENUM (
    'PENDING',
    'SUCCEEDED',
    'FAILED'
);

CREATE TABLE webhook_events (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    webhook_data_id TEXT NOT NULL REFERENCES webhook_data(id) ON UPDATE CASCADE ON DELETE CASCADE,
    webhook_subscription_id TEXT NOT NULL REFERENCES webhook_subscriptions(id) ON UPDATE CASCADE ON DELETE CASCADE,
    status webhook_event_status NOT NULL,
    response JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);
