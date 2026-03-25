ALTER TABLE access_entries
    ADD COLUMN account_type TEXT NOT NULL DEFAULT 'USER';
