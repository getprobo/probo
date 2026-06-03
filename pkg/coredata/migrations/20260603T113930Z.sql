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

ALTER TABLE iam_membership_profiles
    ADD COLUMN activated_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN deactivated_at TIMESTAMP WITH TIME ZONE;

ALTER TABLE iam_memberships
    ALTER COLUMN state DROP DEFAULT;

ALTER TABLE iam_membership_profiles
    ALTER COLUMN state DROP DEFAULT;

ALTER TYPE membership_state RENAME TO membership_state_old;
CREATE TYPE membership_state AS ENUM ('PENDING', 'ACTIVE', 'DEACTIVATED');

ALTER TABLE iam_memberships
    ALTER COLUMN state TYPE membership_state
    USING CASE
        WHEN state::text = 'ACTIVE' THEN 'ACTIVE'::membership_state
        ELSE 'DEACTIVATED'::membership_state
    END;

ALTER TABLE iam_membership_profiles
    ALTER COLUMN state TYPE membership_state
    USING CASE
        WHEN state::text = 'ACTIVE' THEN 'ACTIVE'::membership_state
        WHEN source = 'SCIM' THEN 'DEACTIVATED'::membership_state
        WHEN EXISTS (
            SELECT 1
            FROM iam_invitations i
            WHERE i.user_id = iam_membership_profiles.id
              AND i.accepted_at IS NULL
              AND (
                  i.expires_at >= NOW()
                  OR i.created_at >= NOW() - INTERVAL '7 days'
              )
        ) THEN 'PENDING'::membership_state
        ELSE 'DEACTIVATED'::membership_state
    END;

DROP TYPE membership_state_old;

UPDATE iam_membership_profiles
SET activated_at = updated_at
WHERE state = 'ACTIVE';

UPDATE iam_membership_profiles
SET deactivated_at = NOW()
WHERE state = 'DEACTIVATED';

ALTER TABLE iam_memberships
    ALTER COLUMN state SET DEFAULT 'ACTIVE';
