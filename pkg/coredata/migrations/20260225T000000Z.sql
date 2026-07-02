-- Copyright (c) 2026 Probo Inc <hello@probo.com>.
--
-- Permission is hereby granted, free of charge, to any person obtaining a copy
-- of this software and associated documentation files (the "Software"), to deal
-- in the Software without restriction, including without limitation the rights
-- to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
-- copies of the Software, and to permit persons to whom the Software is
-- furnished to do so, subject to the following conditions:
--
-- The above copyright notice and this permission notice shall be included in
-- all copies or substantial portions of the Software.
--
-- THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
-- IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
-- FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
-- AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
-- LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
-- OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
-- SOFTWARE.

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
