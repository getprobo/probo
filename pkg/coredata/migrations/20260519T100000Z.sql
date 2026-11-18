-- Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
--
-- Permission is hereby granted, free of charge, to any person obtaining a copy
-- of this software and associated documentation files (the "Software"), to deal
-- in the Software without restriction, including without limitation the rights
-- to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
-- copies of the Software, and to permit persons to whom the Software is
-- furnished to do so, subject to the following conditions:
--
-- The above copyright notice and this permission notice shall be included in
-- all copies or substantial portions of the Software.
--
-- THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
-- IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
-- FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
-- AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
-- LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
-- OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
-- SOFTWARE.

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
