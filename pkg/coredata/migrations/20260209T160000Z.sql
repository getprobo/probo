ALTER TABLE iam_scim_bridges ADD COLUMN excluded_user_names TEXT[] NOT NULL DEFAULT '{}';

ALTER TABLE iam_scim_bridges ALTER COLUMN excluded_user_names DROP DEFAULT;
