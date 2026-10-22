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

INSERT INTO task_activities (
    id,
    tenant_id,
    organization_id,
    task_id,
    actor_profile_id,
    activity_type,
    field,
    old_value,
    new_value,
    created_at
)
SELECT
    generate_gid(decode_base64_unpadded(tc.tenant_id), 133),
    tc.tenant_id,
    tc.organization_id,
    tc.task_id,
    tc.owner_profile_id,
    'COMMENTED',
    NULL,
    NULL,
    tc.id,
    tc.created_at
FROM
    task_comments tc
WHERE
    NOT EXISTS (
        SELECT
            1
        FROM
            task_activities ta
        WHERE
            ta.activity_type = 'COMMENTED'
            AND ta.new_value = tc.id
    );
