-- Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
--
-- Permission to use, copy, modify, and/or distribute this software for any
-- purpose with or without fee is hereby granted, provided that the above
-- copyright notice and this permission notice appear in all copies.
--
-- THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
-- REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
-- AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
-- INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
-- LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
-- OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
-- PERFORMANCE OF THIS SOFTWARE.

ALTER TABLE
    iam_membership_profiles
ALTER COLUMN
    kind DROP NOT NULL;

ALTER TABLE
    trust_center_accesses
ADD
    COLUMN identity_id TEXT REFERENCES identities(id);

-- Create missing identities
WITH tca AS (
    SELECT
        DISTINCT ON (email) name,
        email
    FROM
        trust_center_accesses
)
INSERT INTO
    identities (
        id,
        created_at,
        updated_at,
        email_address,
        email_address_verified,
        full_name
    )
SELECT
    generate_gid('\x0000000000000000' :: bytea, 11),
    NOW(),
    NOW(),
    tca.email,
    FALSE,
    tca.name
FROM
    tca ON CONFLICT (email_address) DO NOTHING;

-- Update trust_center_accesses with identity_id
WITH tca_identities AS (
    SELECT
        tca.id,
        tca.tenant_id,
        tca.name AS full_name,
        tca.organization_id,
        tca.state,
        i.id AS identity_id
    FROM
        trust_center_accesses tca
        JOIN identities i ON i.email_address = tca.email
)
UPDATE
    trust_center_accesses tca
SET
    identity_id = tcai.identity_id
FROM
    tca_identities tcai
WHERE
    tca.id = tcai.id;

ALTER TABLE
    trust_center_accesses
ALTER COLUMN
    identity_id
SET
    NOT NULL,
ALTER COLUMN
    email DROP NOT NULL,
ALTER COLUMN
    name DROP NOT NULL,
ALTER COLUMN
    state DROP NOT NULL,
ALTER COLUMN
    has_accepted_non_disclosure_agreement DROP NOT NULL;

CREATE UNIQUE INDEX idx_trust_center_accesses_identity_id_organization_id ON trust_center_accesses(identity_id, organization_id);

-- Create missing profiles
WITH tca_identities AS (
    SELECT
        tca.tenant_id,
        tca.name AS full_name,
        tca.organization_id,
        tca.state,
        i.id AS identity_id
    FROM
        trust_center_accesses tca
        JOIN identities i ON i.email_address = tca.email
)
INSERT INTO
    iam_membership_profiles (
        id,
        tenant_id,
        identity_id,
        organization_id,
        full_name,
        kind,
        additional_email_addresses,
        source,
        state,
        created_at,
        updated_at
    )
SELECT
    generate_gid(decode_base64_unpadded(tcai.tenant_id), 51),
    tcai.tenant_id,
    tcai.identity_id,
    tcai.organization_id,
    tcai.full_name,
    NULL,
    '{}' :: CITEXT [],
    'MANUAL',
    CASE
        WHEN tcai.state = 'ACTIVE' THEN 'ACTIVE' :: membership_state
        ELSE 'INACTIVE' :: membership_state
    END,
    NOW(),
    NOW()
FROM
    tca_identities tcai ON CONFLICT DO NOTHING;