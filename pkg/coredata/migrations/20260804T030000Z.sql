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

CREATE TYPE malaysia_pdpa_breach_status AS ENUM (
    'OPEN',
    'ASSESSING',
    'CONTAINED',
    'CLOSED'
);

CREATE TYPE malaysia_pdpa_breach_notification_decision AS ENUM (
    'PENDING',
    'NOT_REQUIRED',
    'COMMISSIONER_ONLY',
    'COMMISSIONER_AND_DATA_SUBJECTS'
);

CREATE TABLE malaysia_pdpa_breach_incidents (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT,
    occurred_at TIMESTAMP WITH TIME ZONE,
    discovered_at TIMESTAMP WITH TIME ZONE NOT NULL,
    awareness_at TIMESTAMP WITH TIME ZONE NOT NULL,
    affected_data_subjects BIGINT NOT NULL,
    affected_data_records BIGINT NOT NULL,
    personal_data_types TEXT NOT NULL,
    affected_system TEXT,
    likely_consequences TEXT,
    containment_actions TEXT,
    potential_physical_harm BOOLEAN NOT NULL,
    potential_financial_loss BOOLEAN NOT NULL,
    potential_credit_or_property_damage BOOLEAN NOT NULL,
    potential_illegal_use BOOLEAN NOT NULL,
    sensitive_personal_data BOOLEAN NOT NULL,
    potential_identity_fraud BOOLEAN NOT NULL,
    significant_harm BOOLEAN NOT NULL,
    significant_scale BOOLEAN NOT NULL,
    notification_recommendation malaysia_pdpa_breach_notification_decision NOT NULL,
    notification_reasons TEXT[] NOT NULL,
    notification_decision malaysia_pdpa_breach_notification_decision NOT NULL,
    decision_rationale TEXT,
    decision_evidence TEXT,
    assessed_by_profile_id TEXT NOT NULL REFERENCES iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    assessed_at TIMESTAMP WITH TIME ZONE NOT NULL,
    rule_version TEXT NOT NULL,
    rule_source TEXT NOT NULL,
    commissioner_notified_at TIMESTAMP WITH TIME ZONE,
    commissioner_notification_reference TEXT,
    commissioner_confirmation_received_at TIMESTAMP WITH TIME ZONE,
    commissioner_confirmation_reference TEXT,
    delayed_notification_reason TEXT,
    delayed_notification_evidence TEXT,
    data_subjects_notified_at TIMESTAMP WITH TIME ZONE,
    data_subjects_notification_evidence TEXT,
    status malaysia_pdpa_breach_status NOT NULL,
    created_by_profile_id TEXT NOT NULL REFERENCES iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    CONSTRAINT malaysia_pdpa_breach_subject_count_nonnegative
        CHECK (affected_data_subjects >= 0),
    CONSTRAINT malaysia_pdpa_breach_record_count_nonnegative
        CHECK (affected_data_records >= 0),
    CONSTRAINT malaysia_pdpa_breach_occurred_before_discovery
        CHECK (occurred_at IS NULL OR occurred_at <= discovered_at),
    CONSTRAINT malaysia_pdpa_breach_discovery_before_awareness
        CHECK (discovered_at <= awareness_at),
    CONSTRAINT malaysia_pdpa_breach_recommendation_not_pending
        CHECK (notification_recommendation <> 'PENDING'),
    CONSTRAINT malaysia_pdpa_breach_decision_has_rationale
        CHECK (notification_decision = 'PENDING' OR decision_rationale IS NOT NULL),
    CONSTRAINT malaysia_pdpa_breach_commissioner_after_awareness
        CHECK (commissioner_notified_at IS NULL OR commissioner_notified_at >= awareness_at),
    CONSTRAINT malaysia_pdpa_breach_commissioner_reference_has_notification
        CHECK (
            (commissioner_notified_at IS NULL AND commissioner_notification_reference IS NULL)
            OR (
                commissioner_notified_at IS NOT NULL
                AND commissioner_notification_reference IS NOT NULL
            )
        ),
    CONSTRAINT malaysia_pdpa_breach_confirmation_after_notification
        CHECK (
            commissioner_confirmation_received_at IS NULL
            OR (
                commissioner_notified_at IS NOT NULL
                AND commissioner_confirmation_received_at >= commissioner_notified_at
            )
        ),
    CONSTRAINT malaysia_pdpa_breach_confirmation_reference_has_confirmation
        CHECK (
            (
                commissioner_confirmation_received_at IS NULL
                AND commissioner_confirmation_reference IS NULL
            )
            OR (
                commissioner_confirmation_received_at IS NOT NULL
                AND commissioner_confirmation_reference IS NOT NULL
            )
        ),
    CONSTRAINT malaysia_pdpa_breach_delay_evidence_has_reason
        CHECK (delayed_notification_evidence IS NULL OR delayed_notification_reason IS NOT NULL),
    CONSTRAINT malaysia_pdpa_breach_data_subject_notice_after_commissioner
        CHECK (
            data_subjects_notified_at IS NULL
            OR (
                commissioner_notified_at IS NOT NULL
                AND data_subjects_notified_at >= commissioner_notified_at
            )
        ),
    CONSTRAINT malaysia_pdpa_breach_data_subject_evidence_has_notification
        CHECK (
            (data_subjects_notified_at IS NULL AND data_subjects_notification_evidence IS NULL)
            OR (
                data_subjects_notified_at IS NOT NULL
                AND data_subjects_notification_evidence IS NOT NULL
            )
        )
);

CREATE TABLE malaysia_pdpa_breach_status_history (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    incident_id TEXT NOT NULL REFERENCES malaysia_pdpa_breach_incidents(id) ON DELETE CASCADE,
    from_status malaysia_pdpa_breach_status,
    to_status malaysia_pdpa_breach_status NOT NULL,
    changed_by_profile_id TEXT NOT NULL REFERENCES iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    CONSTRAINT malaysia_pdpa_breach_status_changed
        CHECK (from_status IS NULL OR from_status <> to_status)
);
