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

-- Convert flag from single TEXT to TEXT array
ALTER TABLE access_entries
    ADD COLUMN flags TEXT[] NOT NULL DEFAULT '{}';

-- Migrate existing data: copy non-NONE flag values into the array
UPDATE access_entries
SET flags = ARRAY[flag]
WHERE flag != 'NONE';

-- Convert flag_reason to flag_reasons array
ALTER TABLE access_entries
    ADD COLUMN flag_reasons TEXT[] NOT NULL DEFAULT '{}';

-- Migrate existing flag_reason
UPDATE access_entries
SET flag_reasons = ARRAY[flag_reason]
WHERE flag_reason IS NOT NULL AND flag_reason != '';

-- Drop old columns in a single statement
ALTER TABLE access_entries
    DROP COLUMN flag,
    DROP COLUMN flag_reason;
