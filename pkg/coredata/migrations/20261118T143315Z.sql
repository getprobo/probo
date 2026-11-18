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

-- Finding.SearchText now includes title; rebuild body_vector to match.
UPDATE search_resources sr
SET body_vector = to_tsvector(
    'simple',
    translate(
        COALESCE(f.title, '') || ' ' ||
        COALESCE(f.reference_id, '') || ' ' ||
        COALESCE(f.description, '') || ' ' ||
        COALESCE(f.source, '') || ' ' ||
        COALESCE(f.root_cause, '') || ' ' ||
        COALESCE(f.corrective_action, '') || ' ' ||
        COALESCE(f.effectiveness_check, '') || ' ' ||
        sr.resource_id,
        '-+',
        '  '
    )
)
FROM findings f
WHERE f.id = sr.resource_id;

-- Device.SearchTitle skips whitespace-only candidates; align stored titles.
UPDATE search_resources sr
SET title = COALESCE(
    NULLIF(BTRIM(d.hostname), ''),
    NULLIF(BTRIM(d.serial_number), ''),
    NULLIF(BTRIM(d.hardware_uuid), ''),
    sr.resource_id
)
FROM devices d
WHERE d.id = sr.resource_id
  AND d.deleted_at IS NULL;
