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
