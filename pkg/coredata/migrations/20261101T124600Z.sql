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

-- A notification that was already read is resolved. Unread rows stay pending.
UPDATE notifications
SET status = 'RESOLVED'
WHERE read_at IS NOT NULL;

-- A signature or approval the recipient already finished is resolved.
-- An open request on the same document stays pending.
UPDATE notifications
SET status = 'RESOLVED'
WHERE type = 'DOCUMENT_SIGNATURE_REQUESTED'
    AND status = 'PENDING'
    AND EXISTS (
        SELECT 1
        FROM document_version_signatures
        JOIN document_versions
            ON document_versions.id = document_version_signatures.document_version_id
        WHERE document_versions.document_id = notifications.subject_id
            AND document_versions.organization_id = notifications.organization_id
            AND document_version_signatures.signed_by_profile_id = notifications.recipient_profile_id
            AND document_version_signatures.state = 'SIGNED'
    )
    AND NOT EXISTS (
        SELECT 1
        FROM document_version_signatures
        JOIN document_versions
            ON document_versions.id = document_version_signatures.document_version_id
        WHERE document_versions.document_id = notifications.subject_id
            AND document_versions.organization_id = notifications.organization_id
            AND document_version_signatures.signed_by_profile_id = notifications.recipient_profile_id
            AND document_version_signatures.state = 'REQUESTED'
    );

UPDATE notifications
SET status = 'RESOLVED'
WHERE type = 'DOCUMENT_APPROVAL_REQUESTED'
    AND status = 'PENDING'
    AND EXISTS (
        SELECT 1
        FROM document_version_approval_decisions
        JOIN document_version_approval_quorums
            ON document_version_approval_quorums.id = document_version_approval_decisions.quorum_id
        JOIN document_versions
            ON document_versions.id = document_version_approval_quorums.version_id
        WHERE document_versions.document_id = notifications.subject_id
            AND document_versions.organization_id = notifications.organization_id
            AND document_version_approval_decisions.approver_id = notifications.recipient_profile_id
            AND document_version_approval_decisions.state = 'APPROVED'
    )
    AND NOT EXISTS (
        SELECT 1
        FROM document_version_approval_decisions
        JOIN document_version_approval_quorums
            ON document_version_approval_quorums.id = document_version_approval_decisions.quorum_id
        JOIN document_versions
            ON document_versions.id = document_version_approval_quorums.version_id
        WHERE document_versions.document_id = notifications.subject_id
            AND document_versions.organization_id = notifications.organization_id
            AND document_version_approval_decisions.approver_id = notifications.recipient_profile_id
            AND document_version_approval_decisions.state = 'PENDING'
    );

ALTER TABLE notifications
    ALTER COLUMN status DROP DEFAULT,
    ALTER COLUMN email_count DROP DEFAULT;

CREATE TABLE notification_preferences (
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    recipient_profile_id TEXT NOT NULL REFERENCES iam_membership_profiles(id) ON DELETE CASCADE,
    notification_type notification_type NOT NULL,
    email_delivery notification_email_delivery NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    PRIMARY KEY (recipient_profile_id, notification_type)
);

CREATE TABLE notification_weekly_sends (
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
