ALTER TABLE iam_membership_profiles
    ADD COLUMN preferred_language TEXT,
    ADD COLUMN given_name TEXT,
    ADD COLUMN family_name TEXT,
    ADD COLUMN formatted_name TEXT,
    ADD COLUMN middle_name TEXT,
    ADD COLUMN honorific_prefix TEXT,
    ADD COLUMN honorific_suffix TEXT,
    ADD COLUMN employee_number TEXT,
    ADD COLUMN department TEXT,
    ADD COLUMN cost_center TEXT,
    ADD COLUMN enterprise_organization TEXT,
    ADD COLUMN division TEXT,
    ADD COLUMN manager_value TEXT;
