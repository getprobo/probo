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

CREATE TYPE policy_status_new AS ENUM ('DRAFT', 'PUBLISHED');

ALTER TABLE policy_versions ADD COLUMN temp_status TEXT;

UPDATE policy_versions SET temp_status = 
  CASE WHEN status::TEXT = 'ACTIVE' THEN 'PUBLISHED' 
       ELSE status::TEXT 
  END;

ALTER TABLE policy_versions DROP COLUMN status;

ALTER TABLE policy_versions ADD COLUMN status policy_status_new;

UPDATE policy_versions SET status = temp_status::policy_status_new;

ALTER TABLE policy_versions DROP COLUMN temp_status;

DROP TYPE policy_status;
ALTER TYPE policy_status_new RENAME TO policy_status;

CREATE UNIQUE INDEX policy_one_draft_version_idx ON policy_versions (policy_id, status) 
WHERE status = 'DRAFT';
