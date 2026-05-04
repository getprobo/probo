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

ALTER TABLE cookies ADD COLUMN last_detected_at TIMESTAMPTZ NOT NULL DEFAULT now();
UPDATE cookies SET last_detected_at = created_at;
ALTER TABLE cookies ALTER COLUMN last_detected_at DROP DEFAULT;

ALTER TABLE cookie_patterns ADD COLUMN last_matched_at TIMESTAMPTZ;
UPDATE cookie_patterns SET last_matched_at = sub.max_detected
FROM (
    SELECT cookie_pattern_id, MAX(last_detected_at) AS max_detected
    FROM cookies
    GROUP BY cookie_pattern_id
) sub
WHERE cookie_patterns.id = sub.cookie_pattern_id;
