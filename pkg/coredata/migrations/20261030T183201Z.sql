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

-- IF NOT EXISTS so this is a no-op where the index was already built by hand.
-- CREATE INDEX blocks writes for the build; an operator can create them
-- CONCURRENTLY before deploy, dropping any INVALID leftover first since
-- IF NOT EXISTS would keep it. CONCURRENTLY cannot run in this file: the
-- migration runner wraps each file in a transaction.
CREATE INDEX IF NOT EXISTS idx_access_review_entries_campaign_id
    ON access_review_entries (access_review_campaign_id);

CREATE INDEX IF NOT EXISTS idx_access_review_entry_decision_history_entry_id
    ON access_review_entry_decision_history (access_review_entry_id);
