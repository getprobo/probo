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

ALTER TABLE iam_membership_profiles
    ADD COLUMN preferred_language TEXT,
    ADD COLUMN given_name TEXT,
    ADD COLUMN family_name TEXT,
    ADD COLUMN formatted_name TEXT,
    ADD COLUMN middle_name TEXT,
    ADD COLUMN honorific_prefix TEXT,
    ADD COLUMN honorific_suffix TEXT,
    ADD COLUMN employee_number TEXT,
    ADD COLUMN department TEXT,
    ADD COLUMN cost_center TEXT,
    ADD COLUMN enterprise_organization TEXT,
    ADD COLUMN division TEXT,
    ADD COLUMN manager_value TEXT;
