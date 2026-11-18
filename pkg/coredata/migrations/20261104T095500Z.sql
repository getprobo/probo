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

-- Inbox rows are always created. Email delivery is chosen per organization
-- and membership profile. The email worker is the only reader of that choice,
-- and it sends from pending inbox rows.

CREATE TYPE notification_status AS ENUM (
    'PENDING',
    'RESOLVED'
);

CREATE TYPE notification_email_delivery AS ENUM (
    'NONE',
    'EACH',
    'REMINDERS',
    'WEEKLY'
);

ALTER TABLE notifications
    ADD COLUMN status notification_status NOT NULL DEFAULT 'PENDING',
    ADD COLUMN email_count INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN last_emailed_at TIMESTAMP WITH TIME ZONE;

-- Keep the reminder progress that already lived on the request rows.
-- A document can have several versions, so the furthest sequence wins.
UPDATE notifications AS n
SET
    email_count = src.notification_count,
    last_emailed_at = src.last_notified_at
FROM (
    SELECT
        document_versions.document_id,
        document_versions.organization_id,
        document_version_signatures.signed_by_profile_id AS recipient_profile_id,
        MAX(document_version_signatures.notification_count) AS notification_count,
        MAX(document_version_signatures.last_notified_at) AS last_notified_at
    FROM document_version_signatures
    JOIN document_versions
        ON document_versions.id = document_version_signatures.document_version_id
    GROUP BY
        document_versions.document_id,
        document_versions.organization_id,
        document_version_signatures.signed_by_profile_id
) AS src
WHERE n.type = 'DOCUMENT_SIGNATURE_REQUESTED'
    AND n.subject_id = src.document_id
    AND n.organization_id = src.organization_id
    AND n.recipient_profile_id = src.recipient_profile_id;

UPDATE notifications AS n
SET
    email_count = src.notification_count,
    last_emailed_at = src.last_notified_at
FROM (
    SELECT
        document_versions.document_id,
        document_versions.organization_id,
        document_version_approval_decisions.approver_id AS recipient_profile_id,
        MAX(document_version_approval_decisions.notification_count) AS notification_count,
        MAX(document_version_approval_decisions.last_notified_at) AS last_notified_at
    FROM document_version_approval_decisions
    JOIN document_version_approval_quorums
        ON document_version_approval_quorums.id = document_version_approval_decisions.quorum_id
    JOIN document_versions
        ON document_versions.id = document_version_approval_quorums.version_id
    GROUP BY
        document_versions.document_id,
        document_versions.organization_id,
        document_version_approval_decisions.approver_id
) AS src
WHERE n.type = 'DOCUMENT_APPROVAL_REQUESTED'
    AND n.subject_id = src.document_id
    AND n.organization_id = src.organization_id
    AND n.recipient_profile_id = src.recipient_profile_id;

-- A notification that was already read is resolved. Unread rows stay pending.
UPDATE notifications
SET status = 'RESOLVED'
WHERE read_at IS NOT NULL;

-- Only the latest version this person can still open counts. An older
-- open request does not keep the row pending. That version is the newest
-- one by creation time, matching the employee portal.
UPDATE notifications
SET status = 'RESOLVED'
WHERE type = 'DOCUMENT_SIGNATURE_REQUESTED'
    AND status = 'PENDING'
    AND (
        SELECT document_version_signatures.state
        FROM document_version_signatures
        JOIN document_versions
            ON document_versions.id = document_version_signatures.document_version_id
        WHERE document_versions.document_id = notifications.subject_id
            AND document_versions.organization_id = notifications.organization_id
            AND document_version_signatures.signed_by_profile_id = notifications.recipient_profile_id
            AND document_version_signatures.state IN ('REQUESTED', 'SIGNED')
        ORDER BY
            document_versions.created_at DESC,
            document_versions.major DESC,
            document_versions.minor DESC,
            document_versions.id DESC
        LIMIT 1
    ) IS DISTINCT FROM 'REQUESTED';

UPDATE notifications
SET status = 'RESOLVED'
WHERE type = 'DOCUMENT_APPROVAL_REQUESTED'
    AND status = 'PENDING'
    AND (
        SELECT document_version_approval_decisions.state
        FROM document_version_approval_decisions
        JOIN document_version_approval_quorums
            ON document_version_approval_quorums.id = document_version_approval_decisions.quorum_id
        JOIN document_versions
            ON document_versions.id = document_version_approval_quorums.version_id
        WHERE document_versions.document_id = notifications.subject_id
            AND document_versions.organization_id = notifications.organization_id
            AND document_version_approval_decisions.approver_id = notifications.recipient_profile_id
            AND NOT (
                document_version_approval_decisions.state = 'APPROVED'
                AND document_version_approval_decisions.electronic_signature_id IS NULL
            )
        ORDER BY
            document_versions.created_at DESC,
            document_versions.major DESC,
            document_versions.minor DESC,
            document_versions.id DESC,
            document_version_approval_decisions.id DESC
        LIMIT 1
    ) IS DISTINCT FROM 'PENDING';

ALTER TABLE notifications
    ALTER COLUMN status DROP DEFAULT,
    ALTER COLUMN email_count DROP DEFAULT;

CREATE TABLE notification_preferences (
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    recipient_profile_id TEXT NOT NULL REFERENCES iam_membership_profiles(id) ON DELETE CASCADE,
    notification_type notification_type NOT NULL,
    email_delivery notification_email_delivery NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    PRIMARY KEY (recipient_profile_id, notification_type)
);

CREATE TABLE notification_weekly_sends (
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    recipient_profile_id TEXT NOT NULL REFERENCES iam_membership_profiles(id) ON DELETE CASCADE,
    week_start DATE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    PRIMARY KEY (recipient_profile_id, week_start)
);

ALTER TABLE document_version_signatures
    DROP COLUMN notification_count,
    DROP COLUMN last_notified_at;

ALTER TABLE document_version_approval_decisions
    DROP COLUMN notification_count,
    DROP COLUMN last_notified_at;
