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

-- Fix tracker_patterns rows created with entity type 88 (removed
-- CookiePatternEntityType) instead of 89 (TrackerPatternEntityType),
-- and detected_trackers rows migrated from the cookies table with entity
-- type 85 (removed CookieEntityType) instead of 90 (DetectedTrackerEntityType).
-- Rebuild each GID by replacing the entity type bytes and re-encoding.

WITH mapping AS (
    SELECT id AS old_id,
           generate_gid(extract_tenant_id(parse_gid(id)), 89) AS new_id
    FROM tracker_patterns
    WHERE extract_entity_type(parse_gid(id)) = 88
),
update_fk AS (
    UPDATE detected_trackers dt
    SET tracker_pattern_id = m.new_id
    FROM mapping m
    WHERE dt.tracker_pattern_id = m.old_id
)
UPDATE tracker_patterns tp
SET id = m.new_id
FROM mapping m
WHERE tp.id = m.old_id;

UPDATE detected_trackers
SET id = generate_gid(extract_tenant_id(parse_gid(id)), 90)
WHERE extract_entity_type(parse_gid(id)) = 85;
