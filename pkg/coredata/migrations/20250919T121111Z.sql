ALTER TABLE organizations
    ADD COLUMN description TEXT,
    ADD COLUMN website_url TEXT,
    ADD COLUMN email CITEXT,
    ADD COLUMN headquarter_address TEXT;
