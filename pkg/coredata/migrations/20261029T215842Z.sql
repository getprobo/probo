-- Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

-- Device search reads the owner name, and as-of internal-control search
-- reads the event name. Store the same simple vector the other lists use
-- so both can use a GIN index.

ALTER TABLE internal_control_events
    ADD COLUMN search_vector tsvector
    GENERATED ALWAYS AS (to_tsvector('simple', COALESCE(name, ''))) STORED;

CREATE INDEX internal_control_events_search_idx
    ON internal_control_events USING gin(search_vector);

ALTER TABLE iam_membership_profiles
    ADD COLUMN search_vector tsvector
    GENERATED ALWAYS AS (to_tsvector('simple', COALESCE(full_name, ''))) STORED;

CREATE INDEX iam_membership_profiles_search_idx
    ON iam_membership_profiles USING gin(search_vector);
