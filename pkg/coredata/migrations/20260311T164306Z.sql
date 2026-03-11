ALTER TABLE
    assets DROP COLUMN owner_id;

ALTER TABLE
    continual_improvements DROP COLUMN owner_id;

ALTER TABLE
    data DROP COLUMN owner_id;

ALTER TABLE
    document_versions DROP COLUMN owner_id;

ALTER TABLE
    meeting_attendees DROP COLUMN attendee_id;

ALTER TABLE
    nonconformities DROP COLUMN owner_id;

ALTER TABLE
    obligations DROP COLUMN owner_id;

ALTER TABLE
    documents DROP COLUMN owner_id;

ALTER TABLE
    document_version_signatures DROP COLUMN signed_by;

ALTER TABLE
    processing_activities DROP COLUMN data_protection_officer_id;

ALTER TABLE
    risks DROP COLUMN owner_id;

ALTER TABLE
    states_of_applicability DROP COLUMN owner_id;

ALTER TABLE
    tasks DROP COLUMN assigned_to;

ALTER TABLE
    vendors DROP COLUMN business_owner_id,
    DROP COLUMN security_owner_id;

DROP TABLE peoples;
