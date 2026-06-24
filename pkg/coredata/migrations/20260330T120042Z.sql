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

-- Add organization_id to access_entries and access_entry_decision_history
-- to avoid JOINs in AuthorizationAttributes lookups.

-- 1. access_entries
ALTER TABLE access_entries
    ADD COLUMN organization_id TEXT REFERENCES organizations(id);

UPDATE access_entries ae
SET organization_id = arc.organization_id
FROM access_review_campaigns arc
WHERE ae.access_review_campaign_id = arc.id;

ALTER TABLE access_entries
    ALTER COLUMN organization_id SET NOT NULL;

-- 2. access_entry_decision_history
ALTER TABLE access_entry_decision_history
    ADD COLUMN organization_id TEXT REFERENCES organizations(id);

UPDATE access_entry_decision_history h
SET organization_id = ae.organization_id
FROM access_entries ae
WHERE h.access_entry_id = ae.id;

ALTER TABLE access_entry_decision_history
    ALTER COLUMN organization_id SET NOT NULL;
