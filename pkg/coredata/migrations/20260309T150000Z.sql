ALTER TABLE iam_membership_profiles
    ALTER COLUMN kind TYPE TEXT USING kind::TEXT,
    ADD COLUMN nick_name TEXT,
    ADD COLUMN locale TEXT,
    ADD COLUMN timezone TEXT,
    ADD COLUMN profile_url TEXT;
