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

ALTER TABLE organizations ADD COLUMN tenant_id TEXT;
UPDATE organizations SET tenant_id = id;
ALTER TABLE organizations ALTER COLUMN tenant_id SET NOT NULL;

ALTER TABLE frameworks ADD COLUMN tenant_id TEXT;
UPDATE frameworks f SET tenant_id = f.organization_id;
ALTER TABLE frameworks ALTER COLUMN tenant_id SET NOT NULL;

ALTER TABLE controls ADD COLUMN tenant_id TEXT;
UPDATE controls c SET tenant_id = (SELECT organization_id FROM frameworks f WHERE f.id = c.framework_id);
ALTER TABLE controls ALTER COLUMN tenant_id SET NOT NULL;

ALTER TABLE tasks ADD COLUMN tenant_id TEXT;
UPDATE tasks t SET tenant_id = (SELECT c.tenant_id FROM controls_tasks ct JOIN controls c ON ct.control_id = c.id WHERE ct.task_id = t.id LIMIT 1);
ALTER TABLE tasks ALTER COLUMN tenant_id SET NOT NULL;

ALTER TABLE evidences ADD COLUMN tenant_id TEXT;
UPDATE evidences e SET tenant_id = (SELECT t.tenant_id FROM tasks t WHERE t.id = e.task_id);
ALTER TABLE evidences ALTER COLUMN tenant_id SET NOT NULL;

ALTER TABLE controls_tasks ADD COLUMN tenant_id TEXT;
UPDATE controls_tasks ct SET tenant_id = (SELECT f.tenant_id FROM frameworks f JOIN controls c ON c.framework_id = f.id WHERE c.id = ct.control_id);
ALTER TABLE controls_tasks ALTER COLUMN tenant_id SET NOT NULL;

ALTER TABLE control_state_transitions ADD COLUMN tenant_id TEXT;
UPDATE control_state_transitions cst SET tenant_id = (SELECT c.tenant_id FROM controls c WHERE c.id = cst.control_id);
ALTER TABLE control_state_transitions ALTER COLUMN tenant_id SET NOT NULL;

ALTER TABLE task_state_transitions ADD COLUMN tenant_id TEXT;
UPDATE task_state_transitions tst SET tenant_id = (SELECT t.tenant_id FROM tasks t WHERE t.id = tst.task_id);
ALTER TABLE task_state_transitions ALTER COLUMN tenant_id SET NOT NULL;

ALTER TABLE evidence_state_transitions ADD COLUMN tenant_id TEXT;
UPDATE evidence_state_transitions est SET tenant_id = (SELECT e.tenant_id FROM evidences e WHERE e.id = est.evidence_id);
ALTER TABLE evidence_state_transitions ALTER COLUMN tenant_id SET NOT NULL;

ALTER TABLE peoples ADD COLUMN tenant_id TEXT;
UPDATE peoples p SET tenant_id = p.organization_id;
ALTER TABLE peoples ALTER COLUMN tenant_id SET NOT NULL;

ALTER TABLE vendors ADD COLUMN tenant_id TEXT;
UPDATE vendors v SET tenant_id = v.organization_id;
ALTER TABLE vendors ALTER COLUMN tenant_id SET NOT NULL;

ALTER TABLE policies ADD COLUMN tenant_id TEXT;
UPDATE policies p SET tenant_id = p.organization_id;
ALTER TABLE policies ALTER COLUMN tenant_id SET NOT NULL;
