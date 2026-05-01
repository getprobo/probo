-- Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
--
-- Permission to use, copy, modify, and/or distribute this software for any
-- purpose with or without fee is hereby granted, provided that the above
-- copyright notice and this permission notice appear in all copies.
--
-- THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
-- REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
-- AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
-- INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
-- LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
-- OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
-- PERFORMANCE OF THIS SOFTWARE.

-- Link access_sources to cloud_accounts. ON DELETE RESTRICT defends
-- a live access source from a stray cloud-account delete; the
-- service layer maps the resulting 23503 to coredata.ErrResourceInUse
-- so the resolver renders gqlutils.Conflict instead of a 500.
ALTER TABLE access_sources
    ADD COLUMN cloud_account_id TEXT NULL REFERENCES cloud_accounts(id) ON DELETE RESTRICT;

-- The CHECK constraint defends the "at most one of {connector_id,
-- cloud_account_id, csv_data}" invariant at the DB level: a source
-- may carry zero targets (configured later via Update) but never
-- two simultaneous ones. NOT VALID acquires only
-- ShareUpdateExclusiveLock and takes effect for new writes
-- immediately; existing rows are validated in the next migration
-- once the column is in place across replicas.
ALTER TABLE access_sources
    ADD CONSTRAINT access_sources_target_at_most_one
    CHECK (num_nonnulls(connector_id, cloud_account_id, csv_data) <= 1)
    NOT VALID;
