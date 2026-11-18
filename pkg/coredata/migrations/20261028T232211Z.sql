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

-- List, count, and export read one organization in created_at, id order.
-- The single-column created_at index walks every tenant to serve that
-- page, and organization_id alone cannot provide the order.
--
-- IF NOT EXISTS so this is a no-op where the index was already built by hand.
-- audit_log_entries is written on every audited mutation, and CREATE INDEX
-- takes a SHARE lock that blocks those writes for the whole build. An
-- operator can create it CONCURRENTLY before deploy. CONCURRENTLY cannot
-- run in this file: the migration runner wraps each file in a transaction.
CREATE INDEX IF NOT EXISTS idx_audit_log_entries_organization_id_created_at
    ON audit_log_entries (organization_id, tenant_id, created_at DESC, id DESC);

DROP INDEX IF EXISTS idx_audit_log_entries_organization_id;
DROP INDEX IF EXISTS idx_audit_log_entries_created_at;
