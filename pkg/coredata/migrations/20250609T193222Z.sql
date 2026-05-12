-- Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
