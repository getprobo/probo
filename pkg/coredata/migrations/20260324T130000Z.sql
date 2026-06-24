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

-- Add description to campaigns and decision audit trail

-- 1. Campaign description
ALTER TABLE access_review_campaigns ADD COLUMN description TEXT NOT NULL DEFAULT '';
ALTER TABLE access_review_campaigns ALTER COLUMN description DROP DEFAULT;

-- 2. Decision audit trail: immutable log of every decision recorded
CREATE TABLE access_entry_decision_history (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    access_entry_id TEXT NOT NULL REFERENCES access_entries(id) ON DELETE CASCADE,
    decision access_entry_decision NOT NULL,
    decision_note TEXT,
    decided_by TEXT,
    decided_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);
