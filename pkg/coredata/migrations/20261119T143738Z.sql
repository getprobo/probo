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

CREATE TABLE corrective_action_plans (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    finding_id TEXT NOT NULL REFERENCES findings(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT,
    owner_profile_id TEXT REFERENCES iam_membership_profiles(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    due_date DATE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE corrective_action_plans_internal_controls (
    corrective_action_plan_id TEXT NOT NULL REFERENCES corrective_action_plans(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    internal_control_id TEXT NOT NULL REFERENCES internal_controls(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    PRIMARY KEY (corrective_action_plan_id, internal_control_id)
);

-- Tasks now reach a finding through plan -> internal control -> task.
-- Each finding with direct task links gets one plan linked to the
-- internal controls of those tasks. Tasks without any internal control
-- cannot be reached that way, so they are listed in the plan description.
INSERT INTO corrective_action_plans (
    id,
    tenant_id,
    organization_id,
    finding_id,
    title,
    description,
    owner_profile_id,
    due_date,
    created_at,
    updated_at
)
SELECT
    generate_gid(decode_base64_unpadded(f.tenant_id), 148),
    f.tenant_id,
    f.organization_id,
    f.id,
    'Corrective action',
    unlinked.description,
    f.owner_id,
    f.due_date,
    now(),
    now()
FROM
    findings f
LEFT JOIN LATERAL (
    SELECT
        'Tasks previously linked to this finding without an internal control:'
            || E'\n'
            || string_agg('- ' || t.reference_id || ' ' || t.name, E'\n' ORDER BY t.reference_id) AS description
    FROM
        findings_tasks ft
    INNER JOIN
        tasks t ON t.id = ft.task_id
    WHERE
        ft.finding_id = f.id
        AND NOT EXISTS (
            SELECT
                1
            FROM
                internal_controls_tasks ict
            WHERE
                ict.task_id = t.id
        )
) unlinked ON TRUE
WHERE
    EXISTS (
        SELECT
            1
        FROM
            findings_tasks ft
        WHERE
            ft.finding_id = f.id
    );

INSERT INTO corrective_action_plans_internal_controls (
    corrective_action_plan_id,
    internal_control_id,
    tenant_id,
    organization_id,
    created_at
)
SELECT DISTINCT
    cap.id,
    ict.internal_control_id,
    cap.tenant_id,
    cap.organization_id,
    now()
FROM
    corrective_action_plans cap
INNER JOIN
    findings_tasks ft ON ft.finding_id = cap.finding_id
INNER JOIN
    internal_controls_tasks ict ON ict.task_id = ft.task_id
        AND ict.organization_id = cap.organization_id;

DROP TABLE findings_tasks;
