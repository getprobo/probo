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

CREATE TABLE malaysia_pdpa_profiles (
    organization_id TEXT PRIMARY KEY REFERENCES organizations(id) ON DELETE CASCADE,
    tenant_id TEXT NOT NULL,
    total_data_subjects BIGINT NOT NULL,
    sensitive_data_subjects BIGINT NOT NULL,
    regular_systematic_monitoring BOOLEAN NOT NULL,
    dpo_required BOOLEAN NOT NULL,
    dpo_requirement_reasons TEXT[] NOT NULL,
    assessed_by_profile_id TEXT REFERENCES iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    assessed_at TIMESTAMP WITH TIME ZONE,
    dpo_profile_id TEXT REFERENCES iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    dpo_appointed_at TIMESTAMP WITH TIME ZONE,
    commissioner_notified_at TIMESTAMP WITH TIME ZONE,
    commissioner_notification_reference TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    CONSTRAINT malaysia_pdpa_profiles_total_count_nonnegative
        CHECK (total_data_subjects >= 0),
    CONSTRAINT malaysia_pdpa_profiles_sensitive_count_nonnegative
        CHECK (sensitive_data_subjects >= 0),
    CONSTRAINT malaysia_pdpa_profiles_sensitive_within_total
        CHECK (sensitive_data_subjects <= total_data_subjects),
    CONSTRAINT malaysia_pdpa_profiles_assessment_complete
        CHECK ((assessed_by_profile_id IS NULL) = (assessed_at IS NULL)),
    CONSTRAINT malaysia_pdpa_profiles_dpo_appointment_complete
        CHECK ((dpo_profile_id IS NULL) = (dpo_appointed_at IS NULL)),
    CONSTRAINT malaysia_pdpa_profiles_notification_after_appointment
        CHECK (
            commissioner_notified_at IS NULL
            OR (
                dpo_appointed_at IS NOT NULL
                AND commissioner_notified_at >= dpo_appointed_at
            )
        ),
    CONSTRAINT malaysia_pdpa_profiles_reference_with_notification
        CHECK (
            commissioner_notification_reference IS NULL
            OR commissioner_notified_at IS NOT NULL
        )
);

INSERT INTO malaysia_pdpa_profiles (
    organization_id,
    tenant_id,
    total_data_subjects,
    sensitive_data_subjects,
    regular_systematic_monitoring,
    dpo_required,
    dpo_requirement_reasons,
    assessed_by_profile_id,
    assessed_at,
    dpo_profile_id,
    dpo_appointed_at,
    commissioner_notified_at,
    commissioner_notification_reference,
    created_at,
    updated_at
)
SELECT
    id,
    tenant_id,
    0,
    0,
    FALSE,
    FALSE,
    ARRAY[]::TEXT[],
    NULL,
    NULL,
    NULL,
    NULL,
    NULL,
    NULL,
    NOW(),
    NOW()
FROM organizations
ON CONFLICT (organization_id) DO NOTHING;
