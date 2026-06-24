-- Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
