-- Electronic signature enums
CREATE TYPE electronic_signature_document_type AS ENUM (
  'NDA', 'DPA', 'MSA', 'SOW', 'SLA', 'TOS', 'PRIVACY_POLICY', 'OTHER'
);

CREATE TYPE electronic_signature_status AS ENUM (
  'PENDING',
  'ACCEPTED',
  'PROCESSING',
  'COMPLETED',
  'FAILED'
);

CREATE TYPE electronic_signature_event_type AS ENUM (
  'DOCUMENT_VIEWED', 'CONSENT_GIVEN', 'FULL_NAME_TYPED',
  'SIGNATURE_ACCEPTED',
  'SIGNATURE_COMPLETED', 'SEAL_COMPUTED', 'TIMESTAMP_REQUESTED',
  'CERTIFICATE_GENERATED',
  'PROCESSING_ERROR'
);

CREATE TYPE electronic_signature_event_source AS ENUM ('CLIENT', 'SERVER');

-- Electronic signatures table
CREATE TABLE electronic_signatures (
  id                       TEXT PRIMARY KEY,
  tenant_id                TEXT NOT NULL,
  organization_id          TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  status                   electronic_signature_status NOT NULL DEFAULT 'PENDING',
  document_type            electronic_signature_document_type NOT NULL,
  file_id                  TEXT NOT NULL REFERENCES files(id),
  signer_email             CITEXT NOT NULL,

  -- Set at creation time (PENDING):
  consent_text             TEXT NOT NULL,

  -- Set when status transitions to ACCEPTED (signer submits):
  signer_full_name         TEXT,
  signer_ip_address        TEXT,
  signer_user_agent        TEXT,

  -- Set when status transitions to COMPLETED (worker finishes):
  file_hash                TEXT,
  seal                     TEXT,
  seal_version             INT NOT NULL DEFAULT 1,
  tsa_token                BYTEA,
  signed_at                TIMESTAMPTZ,

  -- Set by certificate worker after COMPLETED:
  certificate_file_id      TEXT REFERENCES files(id),
  certificate_processing_started_at TIMESTAMPTZ,

  -- Async processing state:
  attempt_count            INT NOT NULL DEFAULT 0,
  max_attempts             INT NOT NULL DEFAULT 10,
  last_attempted_at        TIMESTAMPTZ,
  last_error               TEXT,
  processing_started_at    TIMESTAMPTZ,

  created_at               TIMESTAMPTZ NOT NULL,
  updated_at               TIMESTAMPTZ NOT NULL,

  -- file_id in the constraint allows re-signing when the org replaces the NDA file.
  UNIQUE(organization_id, signer_email, document_type, file_id)
);

-- Electronic signature events table
CREATE TABLE electronic_signature_events (
  id                       TEXT PRIMARY KEY,
  tenant_id                TEXT NOT NULL,
  electronic_signature_id  TEXT NOT NULL REFERENCES electronic_signatures(id) ON DELETE CASCADE,
  event_type               electronic_signature_event_type NOT NULL,
  event_source             electronic_signature_event_source NOT NULL,
  actor_email              CITEXT NOT NULL,
  actor_ip_address         TEXT NOT NULL,
  actor_user_agent         TEXT NOT NULL,
  occurred_at              TIMESTAMPTZ NOT NULL,
  created_at               TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_esig_events_signature
  ON electronic_signature_events (electronic_signature_id, occurred_at ASC);

-- Email attachments table (generic, not esign-specific)
CREATE TABLE email_attachments (
  id           TEXT PRIMARY KEY,
  email_id     TEXT NOT NULL REFERENCES emails(id) ON DELETE CASCADE,
  file_id      TEXT NOT NULL REFERENCES files(id),
  filename     TEXT NOT NULL,
  content_type TEXT NOT NULL,
  created_at   TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_email_attachments_email_id ON email_attachments (email_id);
