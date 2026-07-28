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

ALTER TABLE device_postures
    ADD COLUMN correlation_id TEXT;

WITH report_ids AS (
    SELECT
        device_id,
        created_at,
        generate_gid(parse_tenant_id(MIN(tenant_id)), 110) AS correlation_id
    FROM
        device_postures
    GROUP BY
        device_id,
        created_at
)
UPDATE device_postures AS dp
SET
    correlation_id = report_ids.correlation_id
FROM
    report_ids
WHERE
    dp.device_id = report_ids.device_id
    AND dp.created_at = report_ids.created_at;

ALTER TABLE device_postures
    ALTER COLUMN correlation_id SET NOT NULL;

CREATE INDEX device_postures_device_id_correlation_id_created_at_idx
    ON device_postures (device_id, correlation_id, created_at DESC);
