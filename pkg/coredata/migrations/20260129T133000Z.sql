ALTER TABLE
    trust_centers
ADD
    COLUMN logo_file_id TEXT REFERENCES files(id) ON UPDATE CASCADE ON DELETE
SET
    NULL,
ADD
    COLUMN dark_logo_file_id TEXT REFERENCES files(id) ON UPDATE CASCADE ON DELETE
SET
    NULL;

UPDATE
    trust_centers t
SET
    logo_file_id = o.logo_file_id
FROM
    organizations o
WHERE
    t.organization_id = o.id
    AND t.logo_file_id IS NULL;
