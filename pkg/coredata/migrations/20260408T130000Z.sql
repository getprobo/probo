-- Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

CREATE TABLE document_default_approvers (
    document_id         text                     NOT NULL,
    approver_profile_id text                     NOT NULL,
    tenant_id           text                     NOT NULL,
    organization_id     text                     NOT NULL,
    created_at          timestamp with time zone NOT NULL,
    updated_at          timestamp with time zone NOT NULL,
    PRIMARY KEY (document_id, approver_profile_id),
    FOREIGN KEY (document_id) REFERENCES documents(id) ON UPDATE CASCADE ON DELETE CASCADE,
    FOREIGN KEY (approver_profile_id) REFERENCES iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE CASCADE
);

-- Backfill default approvers from the last quorum of the last version of each document.
INSERT INTO document_default_approvers (document_id, approver_profile_id, tenant_id, organization_id, created_at, updated_at)
SELECT DISTINCT
    dv.document_id,
    d.approver_id,
    d.tenant_id,
    d.organization_id,
    d.created_at,
    d.created_at
FROM document_version_approval_decisions d
JOIN document_version_approval_quorums q ON q.id = d.quorum_id
JOIN document_versions dv ON dv.id = q.version_id
WHERE q.id = (
    SELECT q2.id
    FROM document_version_approval_quorums q2
    JOIN document_versions dv2 ON dv2.id = q2.version_id
    WHERE dv2.document_id = dv.document_id
    ORDER BY dv2.created_at DESC, q2.created_at DESC
    LIMIT 1
)
ON CONFLICT (document_id, approver_profile_id) DO NOTHING;
