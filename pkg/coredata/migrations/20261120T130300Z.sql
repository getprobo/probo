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

-- Agreement tables: sites, agreements, subjects, acceptances.
-- This is the M1 clickwrap data plane.

CREATE TYPE agreement_site_mode AS ENUM ('LIVE', 'TEST');

CREATE TYPE agreement_kind AS ENUM (
    'TERMS_OF_SERVICE',
    'ACCEPTABLE_USE_POLICY',
    'OTHER'
);

-- A site scopes a set of agreements to one customer-facing product.
-- hashed_secret_key stores SHA-256(sk_live_... | sk_test_...) so we can look
-- up the site on each API request without storing the plaintext key.
CREATE TABLE agreement_sites (
    id                TEXT PRIMARY KEY,
    tenant_id         TEXT NOT NULL,
    organization_id   TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name              TEXT NOT NULL,
    mode              agreement_site_mode NOT NULL,
    hashed_secret_key BYTEA NOT NULL,
    allowed_origins   TEXT[] NOT NULL DEFAULT '{}',
    created_at        TIMESTAMPTZ NOT NULL,
    updated_at        TIMESTAMPTZ NOT NULL
);

CREATE UNIQUE INDEX agreement_sites_hashed_secret_key_idx
    ON agreement_sites (hashed_secret_key);

-- An agreement binds a document to a site and declares its kind.
-- UNIQUE(document_id) enforces one agreement per document.
CREATE TABLE agreements (
    id                 TEXT PRIMARY KEY,
    tenant_id          TEXT NOT NULL,
    organization_id    TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    site_id            TEXT NOT NULL REFERENCES agreement_sites(id) ON DELETE CASCADE,
    kind               agreement_kind NOT NULL,
    document_id        TEXT NOT NULL REFERENCES documents(id) ON DELETE RESTRICT,
    current_version_id TEXT REFERENCES document_versions(id),
    created_at         TIMESTAMPTZ NOT NULL,
    updated_at         TIMESTAMPTZ NOT NULL,
    UNIQUE (document_id)
);

-- A subject is an end-user known to Probo via their external_user_id on a site.
-- external_user_id is unique per site (not globally).
CREATE TABLE agreement_subjects (
    id               TEXT PRIMARY KEY,
    tenant_id        TEXT NOT NULL,
    organization_id  TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    site_id          TEXT NOT NULL REFERENCES agreement_sites(id) ON DELETE CASCADE,
    external_user_id TEXT NOT NULL,
    account_id       TEXT,
    email            CITEXT NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL,
    updated_at       TIMESTAMPTZ NOT NULL,
    UNIQUE (site_id, external_user_id)
);

-- An acceptance records one subject's acceptance of one agreement version.
-- The acceptance is backed by an electronic_signature for cryptographic evidence.
CREATE TABLE agreement_acceptances (
    id                      TEXT PRIMARY KEY,
    tenant_id               TEXT NOT NULL,
    organization_id         TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    site_id                 TEXT NOT NULL REFERENCES agreement_sites(id) ON DELETE CASCADE,
    agreement_id            TEXT NOT NULL REFERENCES agreements(id) ON DELETE CASCADE,
    subject_id              TEXT NOT NULL REFERENCES agreement_subjects(id) ON DELETE CASCADE,
    document_version_id     TEXT NOT NULL REFERENCES document_versions(id),
    content_sha256          TEXT NOT NULL,
    electronic_signature_id TEXT NOT NULL REFERENCES electronic_signatures(id),
    ip_address              TEXT,
    user_agent              TEXT,
    sdk_version             TEXT,
    presentation            JSONB,
    accepted_at             TIMESTAMPTZ NOT NULL,
    created_at              TIMESTAMPTZ NOT NULL,
    updated_at              TIMESTAMPTZ NOT NULL
);

CREATE INDEX agreement_acceptances_subject_agreement_idx
    ON agreement_acceptances (subject_id, agreement_id, accepted_at DESC);

CREATE INDEX agreement_acceptances_site_idx
    ON agreement_acceptances (site_id, accepted_at DESC);
