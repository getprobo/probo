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

CREATE TABLE mailing_lists (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON UPDATE CASCADE ON DELETE CASCADE,
    reply_to CITEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TYPE mailing_list_subscriber_status AS ENUM (
    'PENDING',
    'CONFIRMED'
);

CREATE TABLE mailing_list_subscribers (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON UPDATE CASCADE ON DELETE CASCADE,
    mailing_list_id TEXT NOT NULL REFERENCES mailing_lists(id) ON UPDATE CASCADE ON DELETE CASCADE,
    full_name TEXT NOT NULL,
    email CITEXT NOT NULL,
    status mailing_list_subscriber_status NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE (mailing_list_id, email)
);

ALTER TABLE trust_centers
    ADD COLUMN mailing_list_id TEXT REFERENCES mailing_lists(id) ON UPDATE CASCADE ON DELETE RESTRICT;

INSERT INTO mailing_lists (
    id,
    tenant_id,
    organization_id,
    created_at,
    updated_at
)
SELECT
    generate_gid(decode_base64_unpadded(tc.tenant_id), 64),
    tc.tenant_id,
    tc.organization_id,
    NOW(),
    NOW()
FROM trust_centers tc;

UPDATE trust_centers tc
SET mailing_list_id = ml.id
FROM mailing_lists ml
WHERE ml.organization_id = tc.organization_id;
