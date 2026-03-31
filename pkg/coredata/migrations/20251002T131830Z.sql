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
