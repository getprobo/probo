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

CREATE TYPE service_criticality AS ENUM ('LOW', 'MEDIUM', 'HIGH');
CREATE TYPE risk_tier AS ENUM ('CRITICAL', 'SIGNIFICANT', 'GENERAL');

ALTER TABLE vendors
    ADD COLUMN service_start_date TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    ADD COLUMN service_termination_date TIMESTAMP WITH TIME ZONE,
    ADD COLUMN service_criticality service_criticality NOT NULL DEFAULT 'LOW',
    ADD COLUMN risk_tier risk_tier NOT NULL DEFAULT 'GENERAL',
    ADD COLUMN status_page_url TEXT;

UPDATE vendors SET service_start_date = created_at;

ALTER TABLE vendors ALTER COLUMN service_start_date DROP DEFAULT;
ALTER TABLE vendors ALTER COLUMN service_criticality DROP DEFAULT;
ALTER TABLE vendors ALTER COLUMN risk_tier DROP DEFAULT;
