CREATE EXTENSION IF NOT EXISTS unaccent;

CREATE TABLE trust_centers (
    id TEXT PRIMARY KEY,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    tenant_id TEXT NOT NULL,
    active BOOLEAN NOT NULL,
    slug TEXT NOT NULL CHECK (slug ~ '^[a-z0-9_-]+$'),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE(slug)
);

INSERT INTO trust_centers (
    id,
    organization_id,
    tenant_id,
    active,
    slug,
    created_at,
    updated_at
)
SELECT
    generate_gid(decode_base64_unpadded(o.tenant_id), 22),
    o.id,
    o.tenant_id,
    false,
    LOWER(
        REGEXP_REPLACE(
            REGEXP_REPLACE(
                unaccent(o.name),
                '[^a-zA-Z0-9\s]', '', 'g'
            ),
            '\s+', '-', 'g'
        )
    ),
    NOW(),
    NOW()
FROM organizations o;

ALTER TABLE documents ADD COLUMN show_on_trust_center BOOLEAN NOT NULL DEFAULT false;

ALTER TABLE audits ADD COLUMN show_on_trust_center BOOLEAN NOT NULL DEFAULT false;

ALTER TABLE vendors ADD COLUMN show_on_trust_center BOOLEAN NOT NULL DEFAULT false;

ALTER TABLE documents ALTER COLUMN show_on_trust_center DROP DEFAULT;

ALTER TABLE audits ALTER COLUMN show_on_trust_center DROP DEFAULT;

ALTER TABLE vendors ALTER COLUMN show_on_trust_center DROP DEFAULT;
