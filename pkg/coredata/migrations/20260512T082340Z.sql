-- Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
