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

-- Create approval quorum status enum
CREATE TYPE document_version_approval_quorum_status AS ENUM ('PENDING', 'APPROVED', 'REJECTED');

-- Create approval decision state enum
CREATE TYPE document_version_approval_decision_state AS ENUM ('PENDING', 'APPROVED', 'REJECTED');

-- Create approval quorum table: groups approval decisions for a document version
CREATE TABLE document_version_approval_quorums (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL,
    version_id TEXT NOT NULL REFERENCES document_versions(id) ON DELETE CASCADE,
    status document_version_approval_quorum_status NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

-- Ensure only one PENDING quorum per document version
CREATE UNIQUE INDEX document_one_pending_quorum_idx
    ON document_version_approval_quorums (version_id)
    WHERE status = 'PENDING';

-- Create approval decision table: one row per approver per quorum
CREATE TABLE document_version_approval_decisions (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL,
    quorum_id TEXT NOT NULL REFERENCES document_version_approval_quorums(id) ON DELETE CASCADE,
    approver_id TEXT NOT NULL REFERENCES iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    state document_version_approval_decision_state NOT NULL,
    comment TEXT,
    electronic_signature_id TEXT REFERENCES electronic_signatures(id) ON DELETE RESTRICT,
    decided_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE (quorum_id, approver_id)
);

-- Add document types to electronic_signature_document_type enum
ALTER TYPE electronic_signature_document_type ADD VALUE 'GOVERNANCE';
ALTER TYPE electronic_signature_document_type ADD VALUE 'POLICY';
ALTER TYPE electronic_signature_document_type ADD VALUE 'PROCEDURE';
ALTER TYPE electronic_signature_document_type ADD VALUE 'PLAN';
ALTER TYPE electronic_signature_document_type ADD VALUE 'REGISTER';
ALTER TYPE electronic_signature_document_type ADD VALUE 'RECORD';
ALTER TYPE electronic_signature_document_type ADD VALUE 'REPORT';
ALTER TYPE electronic_signature_document_type ADD VALUE 'TEMPLATE';

-- Add document_name to electronic_signatures for email subject
ALTER TABLE electronic_signatures ADD COLUMN document_name TEXT;

-- Backfill: create APPROVED quorums for existing published versions that have approvers
INSERT INTO document_version_approval_quorums (id, tenant_id, organization_id, version_id, status, created_at, updated_at)
SELECT DISTINCT
    generate_gid(decode_base64_unpadded(dv.tenant_id), 69),
    dv.tenant_id,
    dv.organization_id,
    dv.id,
    'APPROVED'::document_version_approval_quorum_status,
    dv.published_at,
    dv.published_at
FROM document_versions dv
WHERE dv.status = 'PUBLISHED'
    AND dv.published_at IS NOT NULL
    AND (
        EXISTS (SELECT 1 FROM document_version_approvers dva WHERE dva.document_version_id = dv.id)
        OR EXISTS (SELECT 1 FROM document_approvers da WHERE da.document_id = dv.document_id)
    );

-- Backfill: create APPROVED decisions linked to the quorums
INSERT INTO document_version_approval_decisions (
    id, tenant_id, organization_id, quorum_id,
    approver_id, state, decided_at, created_at, updated_at
)
SELECT
    generate_gid(decode_base64_unpadded(dv.tenant_id), 70),
    dv.tenant_id,
    dv.organization_id,
    q.id,
    COALESCE(dva.approver_profile_id, da.approver_profile_id),
    'APPROVED'::document_version_approval_decision_state,
    dv.published_at,
    dv.published_at,
    dv.published_at
FROM document_versions dv
JOIN document_version_approval_quorums q ON q.version_id = dv.id
LEFT JOIN document_version_approvers dva ON dva.document_version_id = dv.id
LEFT JOIN document_approvers da ON da.document_id = dv.document_id
    AND dva.approver_profile_id IS NULL
WHERE dv.status = 'PUBLISHED'
    AND COALESCE(dva.approver_profile_id, da.approver_profile_id) IS NOT NULL
ON CONFLICT (quorum_id, approver_id) DO NOTHING;

-- TODO: DROP TABLE document_version_approvers once confirmed safe
-- TODO: DROP TABLE document_approvers once confirmed safe
