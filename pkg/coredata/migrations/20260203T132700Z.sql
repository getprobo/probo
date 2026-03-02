-- 1. Add peoples column onto profiles
ALTER TABLE
    iam_membership_profiles
ADD
    COLUMN identity_id TEXT REFERENCES identities(id),
ADD
    COLUMN organization_id TEXT REFERENCES organizations(id),
ADD
    COLUMN additional_email_addresses CITEXT [],
ADD
    COLUMN kind PEOPLE_KIND,
ADD
    COLUMN contract_start_date DATE,
ADD
    COLUMN contract_end_date DATE,
ADD
    COLUMN position TEXT;

UPDATE
    iam_membership_profiles mp
SET
    identity_id = m.identity_id,
    organization_id = m.organization_id
FROM
    iam_memberships m
WHERE
    m.id = mp.membership_id;

ALTER TABLE
    iam_membership_profiles
ALTER COLUMN
    identity_id
SET
    NOT NULL,
ALTER COLUMN
    organization_id
SET
    NOT NULL;

CREATE UNIQUE INDEX idx_profiles_identity_id_organization_id ON iam_membership_profiles(identity_id, organization_id);

-- 2. Add profile references to all table referencing peoples
-- was owner_id, NOT NULL
ALTER TABLE
    assets
ADD
    COLUMN owner_profile_id TEXT REFERENCES iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE RESTRICT;

-- was owner_id, NOT NULL
ALTER TABLE
    continual_improvements
ADD
    COLUMN owner_profile_id TEXT REFERENCES iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE RESTRICT;

-- was owner_id, NOT NULL
ALTER TABLE
    data
ADD
    COLUMN owner_profile_id TEXT REFERENCES iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE RESTRICT;

-- was owner_id, NOT NULL
ALTER TABLE
    document_versions
ADD
    COLUMN owner_profile_id TEXT REFERENCES iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE RESTRICT;

-- was attendee_id, NOT NULL
ALTER TABLE
    meeting_attendees
ADD
    COLUMN attendee_profile_id TEXT REFERENCES iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE CASCADE;

-- was owner_id, NOT NULL
ALTER TABLE
    nonconformities
ADD
    COLUMN owner_profile_id TEXT REFERENCES iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE RESTRICT;

-- was owner_id, NOT NULL
ALTER TABLE
    obligations
ADD
    COLUMN owner_profile_id TEXT REFERENCES iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE RESTRICT;

-- was owner_id, can be null
ALTER TABLE
    documents
ADD
    COLUMN owner_profile_id TEXT REFERENCES iam_membership_profiles(id);

-- was signed_by, NOT NULL
ALTER TABLE
    document_version_signatures
ADD
    COLUMN signed_by_profile_id TEXT REFERENCES iam_membership_profiles(id);

-- was data_protection_officer_id, can be null
ALTER TABLE
    processing_activities
ADD
    COLUMN dpo_profile_id TEXT REFERENCES iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE RESTRICT;

-- was owner_id, can be null
ALTER TABLE
    risks
ADD
    COLUMN owner_profile_id TEXT REFERENCES iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE
SET
    NULL;

-- was owner_id, NOT NULL
ALTER TABLE
    states_of_applicability
ADD
    COLUMN owner_profile_id TEXT REFERENCES iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE RESTRICT;

-- was assigned_to, can be null
ALTER TABLE
    tasks
ADD
    COLUMN assigned_to_profile_id TEXT REFERENCES iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE
SET
    NULL;

-- was business_owner_id, can be null
ALTER TABLE
    vendors
ADD
    COLUMN business_owner_profile_id TEXT REFERENCES iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE
SET
    NULL;

-- was security_owner_id, can be null
ALTER TABLE
    vendors
ADD
    COLUMN security_owner_profile_id TEXT REFERENCES iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE
SET
    NULL;

-- 3. Create missing identities (one row per email to avoid "cannot affect row a second time")
INSERT INTO
    identities (
        id,
        created_at,
        updated_at,
        email_address,
        email_address_verified,
        full_name
    )
SELECT
    generate_gid('\x0000000000000000' :: bytea, 11),
    NOW(),
    NOW(),
    p.primary_email_address,
    FALSE,
    p.full_name
FROM
    (
        SELECT
            DISTINCT ON (primary_email_address) primary_email_address,
            full_name
        FROM
            peoples
        ORDER BY
            primary_email_address
    ) p ON CONFLICT (email_address) DO
UPDATE
SET
    full_name = EXCLUDED.full_name;

-- 4. Create missing memberships
WITH people_identities AS (
    SELECT
        i.id AS identity_id,
        p.tenant_id AS tenant_id,
        p.organization_id AS organization_id,
        p.contract_end_date AS contract_end_date
    FROM
        peoples p
        INNER JOIN identities i ON i.email_address = p.primary_email_address
)
INSERT INTO
    iam_memberships (
        id,
        tenant_id,
        identity_id,
        organization_id,
        role,
        source,
        state,
        created_at,
        updated_at
    )
SELECT
    generate_gid(decode_base64_unpadded(pi.tenant_id), 39),
    pi.tenant_id,
    pi.identity_id,
    pi.organization_id,
    'EMPLOYEE' :: authz_role,
    'MANUAL',
    CASE
        WHEN pi.contract_end_date IS NOT NULL
        AND pi.contract_end_date <= NOW() THEN 'INACTIVE' :: membership_state
        ELSE 'ACTIVE' :: membership_state
    END,
    NOW(),
    NOW()
FROM
    people_identities pi ON CONFLICT DO NOTHING;

-- 5. Create missing profiles
WITH people_memberships AS (
    SELECT
        i.id AS identity_id,
        m.organization_id AS organization_id,
        m.id AS membership_id,
        p.tenant_id AS tenant_id,
        p.kind AS kind,
        p.full_name AS full_name,
        p.additional_email_addresses AS additional_email_addresses,
        p.position AS position,
        p.contract_start_date AS contract_start_date,
        p.contract_start_date AS contract_end_date
    FROM
        peoples p
        INNER JOIN identities i ON i.email_address = p.primary_email_address
        INNER JOIN iam_memberships m ON m.identity_id = i.id
    WHERE
        p.organization_id = m.organization_id
)
INSERT INTO
    iam_membership_profiles (
        id,
        tenant_id,
        identity_id,
        organization_id,
        membership_id,
        full_name,
        kind,
        additional_email_addresses,
        position,
        contract_start_date,
        contract_end_date,
        created_at,
        updated_at
    )
SELECT
    generate_gid(decode_base64_unpadded(pm.tenant_id), 51),
    pm.tenant_id,
    pm.identity_id,
    pm.organization_id,
    pm.membership_id,
    pm.full_name,
    pm.kind,
    pm.additional_email_addresses,
    pm.position,
    pm.contract_start_date,
    pm.contract_end_date,
    NOW(),
    NOW()
FROM
    people_memberships pm ON CONFLICT DO NOTHING;

-- 6. Update all profiles from peoples
WITH people_memberships AS (
    SELECT
        m.id AS membership_id,
        COALESCE(
            p.additional_email_addresses,
            '{}' :: CITEXT []
        ) AS additional_email_addresses,
        COALESCE(p.kind, 'EMPLOYEE' :: PEOPLE_KIND) AS kind,
        p.contract_start_date,
        p.contract_end_date,
        p.position
    FROM
        iam_memberships m
        INNER JOIN identities i ON i.id = m.identity_id
        LEFT JOIN peoples p ON p.primary_email_address = i.email_address
        AND p.organization_id = m.organization_id
)
UPDATE
    iam_membership_profiles mp
SET
    additional_email_addresses = pm.additional_email_addresses,
    kind = pm.kind,
    contract_start_date = pm.contract_start_date,
    contract_end_date = pm.contract_end_date,
    position = pm.position
FROM
    people_memberships pm
WHERE
    mp.membership_id = pm.membership_id;

ALTER TABLE
    iam_membership_profiles
ALTER COLUMN
    additional_email_addresses
SET
    NOT NULL,
ALTER COLUMN
    kind
SET
    NOT NULL;

-- 7. Fill new references to profiles (e.g. previous references to peoples)
WITH people_profiles AS (
    SELECT
        mp.id AS membership_profile_id,
        p.id AS people_id
    FROM
        iam_membership_profiles mp
        INNER JOIN iam_memberships m ON m.id = mp.membership_id
        INNER JOIN identities i ON i.id = m.identity_id
        LEFT JOIN peoples p ON p.primary_email_address = i.email_address
        AND m.organization_id = p.organization_id
),
updated_assets AS (
    UPDATE
        assets a
    SET
        owner_profile_id = pp.membership_profile_id
    FROM
        people_profiles pp
    WHERE
        a.owner_id = pp.people_id
),
updated_ci AS (
    UPDATE
        continual_improvements ci
    SET
        owner_profile_id = pp.membership_profile_id
    FROM
        people_profiles pp
    WHERE
        ci.owner_id = pp.people_id
),
updated_data AS (
    UPDATE
        data d
    SET
        owner_profile_id = pp.membership_profile_id
    FROM
        people_profiles pp
    WHERE
        d.owner_id = pp.people_id
),
updated_dv AS (
    UPDATE
        document_versions dv
    SET
        owner_profile_id = pp.membership_profile_id
    FROM
        people_profiles pp
    WHERE
        dv.owner_id = pp.people_id
),
updated_ma AS (
    UPDATE
        meeting_attendees ma
    SET
        attendee_profile_id = pp.membership_profile_id
    FROM
        people_profiles pp
    WHERE
        ma.attendee_id = pp.people_id
),
updated_nc AS (
    UPDATE
        nonconformities nc
    SET
        owner_profile_id = pp.membership_profile_id
    FROM
        people_profiles pp
    WHERE
        nc.owner_id = pp.people_id
),
updated_obligations AS (
    UPDATE
        obligations o
    SET
        owner_profile_id = pp.membership_profile_id
    FROM
        people_profiles pp
    WHERE
        o.owner_id = pp.people_id
),
updated_documents AS (
    UPDATE
        documents d
    SET
        owner_profile_id = pp.membership_profile_id
    FROM
        people_profiles pp
    WHERE
        d.owner_id = pp.people_id
),
updated_dvs AS (
    UPDATE
        document_version_signatures dvs
    SET
        signed_by_profile_id = pp.membership_profile_id
    FROM
        people_profiles pp
    WHERE
        dvs.signed_by = pp.people_id
),
updated_pa AS (
    UPDATE
        processing_activities pa
    SET
        dpo_profile_id = pp.membership_profile_id
    FROM
        people_profiles pp
    WHERE
        pa.data_protection_officer_id = pp.people_id
),
updated_risks AS (
    UPDATE
        risks r
    SET
        owner_profile_id = pp.membership_profile_id
    FROM
        people_profiles pp
    WHERE
        r.owner_id = pp.people_id
),
updated_soa AS (
    UPDATE
        states_of_applicability soa
    SET
        owner_profile_id = pp.membership_profile_id
    FROM
        people_profiles pp
    WHERE
        soa.owner_id = pp.people_id
),
updated_tasks AS (
    UPDATE
        tasks t
    SET
        assigned_to_profile_id = pp.membership_profile_id
    FROM
        people_profiles pp
    WHERE
        t.assigned_to = pp.people_id
),
updated_vendors AS (
    UPDATE
        vendors v
    SET
        business_owner_profile_id = pp.membership_profile_id
    FROM
        people_profiles pp
    WHERE
        v.business_owner_id = pp.people_id
)
UPDATE
    vendors v
SET
    security_owner_profile_id = pp.membership_profile_id
FROM
    people_profiles pp
WHERE
    v.security_owner_id = pp.people_id;

-- 8. Now that references are filled, add the NOT NULL constraints to those who need it
ALTER TABLE
    assets
ALTER COLUMN
    owner_id DROP NOT NULL,
ALTER COLUMN
    owner_profile_id
SET
    NOT NULL;

ALTER TABLE
    continual_improvements
ALTER COLUMN
    owner_id DROP NOT NULL,
ALTER COLUMN
    owner_profile_id
SET
    NOT NULL;

ALTER TABLE
    data
ALTER COLUMN
    owner_id DROP NOT NULL,
ALTER COLUMN
    owner_profile_id
SET
    NOT NULL;

ALTER TABLE
    document_versions
ALTER COLUMN
    owner_id DROP NOT NULL,
ALTER COLUMN
    owner_profile_id
SET
    NOT NULL;

ALTER TABLE
    meeting_attendees DROP CONSTRAINT meeting_attendees_pkey,
ALTER COLUMN
    attendee_id DROP NOT NULL,
ALTER COLUMN
    attendee_profile_id
SET
    NOT NULL;

WITH duplicates AS (
    SELECT
        meeting_id,
        attendee_profile_id,
        ROW_NUMBER() OVER (PARTITION BY meeting_id, attendee_profile_id) AS rn
    FROM
        meeting_attendees
)
DELETE FROM
    meeting_attendees ma USING duplicates d
WHERE
    ma.meeting_id = d.meeting_id
    AND ma.attendee_profile_id = d.attendee_profile_id
    AND d.rn > 1;

ALTER TABLE
    meeting_attendees
ADD
    PRIMARY KEY (meeting_id, attendee_profile_id);

ALTER TABLE
    nonconformities
ALTER COLUMN
    owner_id DROP NOT NULL,
ALTER COLUMN
    owner_profile_id
SET
    NOT NULL;

ALTER TABLE
    obligations
ALTER COLUMN
    owner_id DROP NOT NULL,
ALTER COLUMN
    owner_profile_id
SET
    NOT NULL;

ALTER TABLE
    document_version_signatures
ALTER COLUMN
    signed_by DROP NOT NULL,
ALTER COLUMN
    signed_by_profile_id
SET
    NOT NULL;

ALTER TABLE
    states_of_applicability
ALTER COLUMN
    owner_id DROP NOT NULL,
ALTER COLUMN
    owner_profile_id
SET
    NOT NULL;

-- 9. Expire any potential invitation that is now obsolete since we created memberships
ALTER TABLE
    iam_invitations
ALTER COLUMN
    email TYPE CITEXT;

WITH people_identities AS (
    SELECT
        m.organization_id AS organization_id,
        i.email_address AS email_address
    FROM
        iam_memberships m
        JOIN identities i ON i.id = m.identity_id
        JOIN peoples p ON p.primary_email_address = i.email_address
        AND p.organization_id = m.organization_id
)
UPDATE
    iam_invitations i
SET
    expires_at = NOW()
FROM
    people_identities pi
WHERE
    pi.organization_id = i.organization_id
    AND i.email = pi.email_address;