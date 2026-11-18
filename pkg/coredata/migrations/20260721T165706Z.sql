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

-- entity_name is the short name shown in the hero ("Compliance at {{name}}.")
-- and TopBar brand. Prefill from the organization so existing portals keep a
-- sensible default; customized full titles in the old title column are dropped.
ALTER TABLE trust_centers
    ADD COLUMN entity_name TEXT NOT NULL DEFAULT '';

UPDATE trust_centers tc
SET entity_name = o.name
FROM organizations o
WHERE tc.organization_id = o.id;

ALTER TABLE trust_centers
    ALTER COLUMN entity_name DROP DEFAULT;

ALTER TABLE trust_centers
    DROP COLUMN title;
