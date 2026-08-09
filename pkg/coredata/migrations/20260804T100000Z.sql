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

CREATE TYPE malaysia_pdpa_dpia_recommendation AS ENUM (
    'NOT_INDICATED',
    'DPO_REVIEW_REQUIRED',
    'REQUIRED'
);

ALTER TABLE processing_activities
    ADD COLUMN malaysia_dpia_total_data_subjects BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN malaysia_dpia_sensitive_data_subjects BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN malaysia_dpia_legal_or_significant_effects BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN malaysia_dpia_systematic_monitoring BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN malaysia_dpia_innovative_technology BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN malaysia_dpia_denial_or_restriction_of_rights BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN malaysia_dpia_location_or_behaviour_tracking BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN malaysia_dpia_children_or_vulnerable_data_subjects BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN malaysia_dpia_high_risk_automated_decision_making BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN malaysia_dpia_other_high_risk_factors TEXT,
    ADD COLUMN malaysia_dpia_recommendation malaysia_pdpa_dpia_recommendation NOT NULL DEFAULT 'NOT_INDICATED',
    ADD COLUMN malaysia_dpia_reasons TEXT[] NOT NULL DEFAULT '{}',
    ADD COLUMN malaysia_dpia_assessed_by_profile_id TEXT REFERENCES iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    ADD COLUMN malaysia_dpia_assessed_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN malaysia_dpia_rule_version TEXT,
    ADD COLUMN malaysia_dpia_rule_source TEXT,
    ADD CONSTRAINT processing_activities_malaysia_dpia_total_nonnegative
        CHECK (malaysia_dpia_total_data_subjects >= 0),
    ADD CONSTRAINT processing_activities_malaysia_dpia_sensitive_nonnegative
        CHECK (malaysia_dpia_sensitive_data_subjects >= 0),
    ADD CONSTRAINT processing_activities_malaysia_dpia_sensitive_within_total
        CHECK (malaysia_dpia_sensitive_data_subjects <= malaysia_dpia_total_data_subjects),
    ADD CONSTRAINT processing_activities_malaysia_dpia_assessment_complete
        CHECK (
            (malaysia_dpia_assessed_by_profile_id IS NULL) = (malaysia_dpia_assessed_at IS NULL)
            AND (malaysia_dpia_assessed_at IS NULL) = (malaysia_dpia_rule_version IS NULL)
            AND (malaysia_dpia_rule_version IS NULL) = (malaysia_dpia_rule_source IS NULL)
        );

ALTER TABLE processing_activities
    ALTER COLUMN malaysia_dpia_total_data_subjects DROP DEFAULT,
    ALTER COLUMN malaysia_dpia_sensitive_data_subjects DROP DEFAULT,
    ALTER COLUMN malaysia_dpia_legal_or_significant_effects DROP DEFAULT,
    ALTER COLUMN malaysia_dpia_systematic_monitoring DROP DEFAULT,
    ALTER COLUMN malaysia_dpia_innovative_technology DROP DEFAULT,
    ALTER COLUMN malaysia_dpia_denial_or_restriction_of_rights DROP DEFAULT,
    ALTER COLUMN malaysia_dpia_location_or_behaviour_tracking DROP DEFAULT,
    ALTER COLUMN malaysia_dpia_children_or_vulnerable_data_subjects DROP DEFAULT,
    ALTER COLUMN malaysia_dpia_high_risk_automated_decision_making DROP DEFAULT,
    ALTER COLUMN malaysia_dpia_recommendation DROP DEFAULT,
    ALTER COLUMN malaysia_dpia_reasons DROP DEFAULT;

CREATE TYPE malaysia_pdpa_transfer_basis AS ENUM (
    'SUBSTANTIALLY_SIMILAR_LAW',
    'ADEQUATE_EQUIVALENT_PROTECTION',
    'DATA_SUBJECT_CONSENT',
    'DATA_SUBJECT_CONTRACT',
    'THIRD_PARTY_CONTRACT',
    'LEGAL_PROCEEDINGS',
    'ADVERSE_ACTION',
    'REASONABLE_PRECAUTIONS',
    'VITAL_INTERESTS'
);

CREATE TYPE malaysia_pdpa_transfer_approval_status AS ENUM (
    'PENDING',
    'APPROVED',
    'REJECTED'
);

ALTER TABLE processing_activity_transfer_impact_assessments
    ADD COLUMN malaysia_transfer_basis malaysia_pdpa_transfer_basis,
    ADD COLUMN malaysia_destination_country country_code,
    ADD COLUMN malaysia_recipient_third_party_id TEXT REFERENCES third_parties(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    ADD COLUMN malaysia_receiver_registration_number TEXT,
    ADD COLUMN malaysia_receiver_contact TEXT,
    ADD COLUMN malaysia_transfer_purpose TEXT,
    ADD COLUMN malaysia_personal_data_categories TEXT,
    ADD COLUMN malaysia_safeguards TEXT,
    ADD COLUMN malaysia_approval_status malaysia_pdpa_transfer_approval_status,
    ADD COLUMN malaysia_approved_by_profile_id TEXT REFERENCES iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    ADD COLUMN malaysia_approval_notes TEXT,
    ADD COLUMN malaysia_reviewed_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN malaysia_next_review_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN malaysia_review_evidence TEXT,
    ADD COLUMN malaysia_rule_version TEXT,
    ADD COLUMN malaysia_rule_source TEXT,
    ADD CONSTRAINT processing_activity_tia_malaysia_destination_foreign
        CHECK (
            malaysia_destination_country IS NULL
            OR malaysia_destination_country NOT IN ('MY', 'GLOBAL')
        ),
    ADD CONSTRAINT processing_activity_tia_malaysia_details_complete
        CHECK (
            malaysia_transfer_basis IS NULL
            OR (
                malaysia_destination_country IS NOT NULL
                AND malaysia_recipient_third_party_id IS NOT NULL
                AND malaysia_receiver_contact IS NOT NULL
                AND malaysia_transfer_purpose IS NOT NULL
                AND malaysia_personal_data_categories IS NOT NULL
                AND malaysia_safeguards IS NOT NULL
                AND malaysia_approval_status IS NOT NULL
                AND malaysia_rule_version IS NOT NULL
                AND malaysia_rule_source IS NOT NULL
            )
        ),
    ADD CONSTRAINT processing_activity_tia_malaysia_approval_complete
        CHECK (
            malaysia_approval_status IS NULL
            OR (
                malaysia_approval_status = 'PENDING'
                AND malaysia_approved_by_profile_id IS NULL
                AND malaysia_reviewed_at IS NULL
                AND malaysia_next_review_at IS NULL
            )
            OR (
                malaysia_approval_status = 'APPROVED'
                AND malaysia_approved_by_profile_id IS NOT NULL
                AND malaysia_reviewed_at IS NOT NULL
                AND malaysia_review_evidence IS NOT NULL
            )
            OR (
                malaysia_approval_status = 'REJECTED'
                AND malaysia_approved_by_profile_id IS NOT NULL
                AND malaysia_reviewed_at IS NOT NULL
                AND malaysia_approval_notes IS NOT NULL
                AND malaysia_next_review_at IS NULL
            )
        ),
    ADD CONSTRAINT processing_activity_tia_malaysia_tia_review_due
        CHECK (
            malaysia_approval_status IS DISTINCT FROM 'APPROVED'
            OR malaysia_transfer_basis NOT IN ('SUBSTANTIALLY_SIMILAR_LAW', 'ADEQUATE_EQUIVALENT_PROTECTION')
            OR (
                malaysia_next_review_at IS NOT NULL
                AND malaysia_next_review_at <= malaysia_reviewed_at + INTERVAL '3 years'
            )
        );
