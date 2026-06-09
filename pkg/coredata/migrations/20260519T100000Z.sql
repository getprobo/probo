-- Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

ALTER TABLE third_parties ADD COLUMN first_level boolean NOT NULL DEFAULT true;
ALTER TABLE third_parties ALTER COLUMN first_level DROP DEFAULT;

CREATE TABLE third_party_third_parties (
    parent_third_party_id text NOT NULL REFERENCES third_parties(id) ON DELETE CASCADE,
    child_third_party_id  text NOT NULL REFERENCES third_parties(id) ON DELETE CASCADE,
    tenant_id             bytea NOT NULL,
    created_at            timestamptz NOT NULL,
    PRIMARY KEY (parent_third_party_id, child_third_party_id),
    CHECK (parent_third_party_id <> child_third_party_id)
);
