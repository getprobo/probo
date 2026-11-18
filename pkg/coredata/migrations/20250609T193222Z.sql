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

CREATE TYPE data_classification_type AS ENUM (
    'PUBLIC',
    'INTERNAL',
    'CONFIDENTIAL',
    'SECRET'
);

ALTER TABLE data
ADD COLUMN data_classification data_classification_type;

UPDATE data
SET data_classification = CASE
    WHEN data_sensitivity = 'NONE' THEN 'PUBLIC'::data_classification_type
    WHEN data_sensitivity = 'LOW' THEN 'INTERNAL'::data_classification_type
    WHEN data_sensitivity = 'MEDIUM' THEN 'INTERNAL'::data_classification_type
    WHEN data_sensitivity = 'HIGH' THEN 'CONFIDENTIAL'::data_classification_type
    WHEN data_sensitivity = 'CRITICAL' THEN 'SECRET'::data_classification_type
END;

ALTER TABLE data
ALTER COLUMN data_classification SET NOT NULL;

ALTER TABLE data
DROP COLUMN data_sensitivity;
