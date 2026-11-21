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

-- Signable Documents / Clickwrap feature — M1 schema

-- ─────────────────────────────────────────────
-- 1. electronic_signatures additions
-- ─────────────────────────────────────────────

CREATE TYPE electronic_signature_flow AS ENUM ('ESIGN');

ALTER TABLE electronic_signatures
  ADD COLUMN flow electronic_signature_flow NOT NULL DEFAULT 'ESIGN';

ALTER TABLE electronic_signatures
  ALTER COLUMN flow DROP DEFAULT;

-- Drop the global unique constraint; a partial unique index scoped to ESIGN
-- replaces it so CLICKWRAP rows in a future table are not constrained here.
ALTER TABLE electronic_signatures
  DROP CONSTRAINT IF EXISTS electronic_signatures_organization_id_signer_email_document_type_file_id_key;

CREATE UNIQUE INDEX uq_esig_esign_per_signer_doc
  ON electronic_signatures (organization_id, signer_email, document_type, file_id)
  WHERE flow = 'ESIGN';

-- consent_text column already existed on electronic_signatures (confirmed by
-- pkg/coredata/electronic_signature.go). No addition needed.

-- ─────────────────────────────────────────────
-- 2. document_versions additions
-- ─────────────────────────────────────────────

ALTER TABLE document_versions
  ADD COLUMN content_sha256  TEXT,
  ADD COLUMN presentable_at  TIMESTAMPTZ;

-- ─────────────────────────────────────────────
-- 3. document_signing_configs
-- ─────────────────────────────────────────────

CREATE TYPE document_signing_config_state AS ENUM ('DRAFT', 'ACTIVE', 'ARCHIVED');
CREATE TYPE document_signing_config_mode  AS ENUM ('LIVE', 'TEST');
CREATE TYPE document_signing_config_kind  AS ENUM (
  'TERMS_OF_SERVICE',
  'ACCEPTABLE_USE_POLICY',
  'OTHER'
);
CREATE TYPE document_signing_reacceptance_policy AS ENUM (
  'ON_MAJOR',
  'ON_ANY',
  'NEVER'
);

CREATE TABLE document_signing_configs (
  id               TEXT PRIMARY KEY,
  tenant_id        TEXT NOT NULL,
  organization_id  TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  document_id      TEXT NOT NULL UNIQUE REFERENCES documents(id) ON DELETE RESTRICT,

  publishable_key  TEXT NOT NULL UNIQUE,

  state            document_signing_config_state        NOT NULL DEFAULT 'DRAFT',
  mode             document_signing_config_mode         NOT NULL DEFAULT 'LIVE',
  kind             document_signing_config_kind         NOT NULL DEFAULT 'OTHER',
  reacceptance_policy document_signing_reacceptance_policy NOT NULL DEFAULT 'ON_MAJOR',

  allowed_origins  TEXT[],

  created_at       TIMESTAMPTZ NOT NULL,
  updated_at       TIMESTAMPTZ NOT NULL
);

ALTER TABLE document_signing_configs
  ALTER COLUMN state DROP DEFAULT,
  ALTER COLUMN mode  DROP DEFAULT,
  ALTER COLUMN kind  DROP DEFAULT,
  ALTER COLUMN reacceptance_policy DROP DEFAULT;

-- ─────────────────────────────────────────────
-- 4. organization_signing_secrets
-- ─────────────────────────────────────────────

CREATE TABLE organization_signing_secrets (
  id               TEXT PRIMARY KEY,
  tenant_id        TEXT NOT NULL,
  organization_id  TEXT NOT NULL UNIQUE REFERENCES organizations(id) ON DELETE CASCADE,

  secret_key_hash  TEXT NOT NULL,

  mode             document_signing_config_mode NOT NULL DEFAULT 'LIVE',
  created_at       TIMESTAMPTZ NOT NULL,
  updated_at       TIMESTAMPTZ NOT NULL
);

ALTER TABLE organization_signing_secrets
  ALTER COLUMN mode DROP DEFAULT;

-- ─────────────────────────────────────────────
-- 5. clickwrap_subjects
-- ─────────────────────────────────────────────

CREATE TABLE clickwrap_subjects (
  id                   TEXT PRIMARY KEY,
  tenant_id            TEXT NOT NULL,
  organization_id      TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

  external_user_id     TEXT NOT NULL,
  email                CITEXT NOT NULL,
  full_name            TEXT,
  external_account_id  TEXT,

  created_at           TIMESTAMPTZ NOT NULL,
  updated_at           TIMESTAMPTZ NOT NULL,

  UNIQUE (organization_id, external_user_id)
);

CREATE INDEX idx_clickwrap_subjects_email
  ON clickwrap_subjects (organization_id, email);

CREATE INDEX idx_clickwrap_subjects_account
  ON clickwrap_subjects (organization_id, external_account_id)
  WHERE external_account_id IS NOT NULL;

-- ─────────────────────────────────────────────
-- 6. clickwrap_acceptances
-- ─────────────────────────────────────────────

CREATE TABLE clickwrap_acceptances (
  id                      TEXT PRIMARY KEY,
  tenant_id               TEXT NOT NULL,
  organization_id         TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

  signing_config_id       TEXT NOT NULL REFERENCES document_signing_configs(id)
                            ON DELETE RESTRICT,
  document_version_id     TEXT NOT NULL REFERENCES document_versions(id)
                            ON DELETE RESTRICT,
  subject_id              TEXT NOT NULL REFERENCES clickwrap_subjects(id)
                            ON DELETE RESTRICT,

  accepted_email          CITEXT NOT NULL,
  accepted_full_name      TEXT,
  accepted_account_id     TEXT,

  consent_text            TEXT NOT NULL,
  content_sha256          TEXT NOT NULL,
  ip_address              TEXT NOT NULL,
  user_agent              TEXT NOT NULL,
  sdk_version             TEXT,
  country_code            TEXT,

  mode                    document_signing_config_mode NOT NULL DEFAULT 'LIVE',

  idempotency_key         TEXT,
  UNIQUE (organization_id, signing_config_id, idempotency_key)
    DEFERRABLE INITIALLY DEFERRED,

  revoked_at              TIMESTAMPTZ,
  revocation_reason       TEXT,

  seal                    TEXT,
  seal_version            INT NOT NULL DEFAULT 2,
  tsa_token               BYTEA,
  sealed_at               TIMESTAMPTZ,

  seal_attempt_count      INT NOT NULL DEFAULT 0,
  seal_max_attempts       INT NOT NULL DEFAULT 10,
  seal_last_attempted_at  TIMESTAMPTZ,
  seal_last_error         TEXT,
  seal_processing_started_at TIMESTAMPTZ,

  accepted_at             TIMESTAMPTZ NOT NULL,
  created_at              TIMESTAMPTZ NOT NULL,
  updated_at              TIMESTAMPTZ NOT NULL
);

ALTER TABLE clickwrap_acceptances
  ALTER COLUMN mode         DROP DEFAULT,
  ALTER COLUMN seal_version DROP DEFAULT,
  ALTER COLUMN seal_attempt_count DROP DEFAULT,
  ALTER COLUMN seal_max_attempts  DROP DEFAULT;

CREATE INDEX idx_clickwrap_acc_subject_config
  ON clickwrap_acceptances (organization_id, signing_config_id, subject_id, accepted_at DESC)
  WHERE revoked_at IS NULL;

CREATE INDEX idx_clickwrap_acc_pending_seal
  ON clickwrap_acceptances (created_at ASC)
  WHERE seal IS NULL AND revoked_at IS NULL;

-- ─────────────────────────────────────────────
-- 7. data_plane_request_log
-- ─────────────────────────────────────────────

CREATE TABLE data_plane_request_log (
  id               TEXT PRIMARY KEY,
  tenant_id        TEXT NOT NULL,
  organization_id  TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  signing_config_id TEXT REFERENCES document_signing_configs(id) ON DELETE CASCADE,

  request_path     TEXT NOT NULL,
  method           TEXT NOT NULL,
  status_code      INT  NOT NULL,
  error_code       TEXT NOT NULL,
  request_id       TEXT NOT NULL,
  occurred_at      TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_dp_log_config_time
  ON data_plane_request_log (signing_config_id, occurred_at DESC)
  WHERE signing_config_id IS NOT NULL;
