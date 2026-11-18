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

CREATE TYPE trust_center_visibility AS ENUM (
    'NONE',
    'PRIVATE',
    'PUBLIC'
);

ALTER TABLE documents ADD COLUMN trust_center_visibility trust_center_visibility NOT NULL DEFAULT 'NONE';

UPDATE documents SET trust_center_visibility = CASE
    WHEN show_on_trust_center = true THEN 'PRIVATE'::trust_center_visibility
    ELSE 'NONE'::trust_center_visibility
END;

ALTER TABLE documents ALTER COLUMN trust_center_visibility DROP DEFAULT;
ALTER TABLE documents DROP COLUMN show_on_trust_center;

ALTER TABLE audits ADD COLUMN trust_center_visibility trust_center_visibility NOT NULL DEFAULT 'NONE';

UPDATE audits SET trust_center_visibility = CASE
    WHEN show_on_trust_center = true THEN 'PRIVATE'::trust_center_visibility
    ELSE 'NONE'::trust_center_visibility
END;

ALTER TABLE audits ALTER COLUMN trust_center_visibility DROP DEFAULT;
ALTER TABLE audits DROP COLUMN show_on_trust_center;
