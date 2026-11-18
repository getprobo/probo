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

-- Portal visitor lifecycle lives on cp_accesses, not membership profiles.
-- state was left nullable after the profile migration; restore it as the
-- ACTIVE/INACTIVE gate. authenticated_at is observational (first sign-in).

UPDATE cp_accesses
SET
    state = 'ACTIVE'
WHERE
    state IS NULL;

ALTER TABLE cp_accesses
ALTER COLUMN state SET NOT NULL;

ALTER TABLE cp_accesses
ADD COLUMN authenticated_at TIMESTAMP WITH TIME ZONE;

UPDATE cp_accesses
SET
    authenticated_at = created_at;
