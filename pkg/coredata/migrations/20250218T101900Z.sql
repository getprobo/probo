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
