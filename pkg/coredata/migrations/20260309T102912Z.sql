ALTER TABLE iam_membership_profiles ADD COLUMN search_vector tsvector
GENERATED ALWAYS AS (
    to_tsvector('simple',
        COALESCE(full_name, '') || ' ' ||
        COALESCE(position, '')
    )
) STORED;

CREATE INDEX users_search_idx ON iam_membership_profiles USING gin(search_vector);

ALTER TABLE identities ADD COLUMN search_vector tsvector
GENERATED ALWAYS AS (
    to_tsvector('simple',
        COALESCE(full_name, '') || ' ' ||
        COALESCE(email_address, '')
    )
) STORED;

CREATE INDEX identities_search_idx ON identities USING gin(search_vector);
