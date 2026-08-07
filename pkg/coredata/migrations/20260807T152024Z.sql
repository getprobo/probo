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

-- Stamp rows created by the visitor request flow so portal UI can distinguish
-- them from admin-granted access (console Merge inserts with no request).
ALTER TABLE cp_document_accesses
    ADD COLUMN requested_at TIMESTAMP WITH TIME ZONE;

-- Historical provenance (no requested_at yet):
-- - Visitor BulkInsert creates REQUESTED with created_at = updated_at.
-- - Later grant/reject/revoke only bumps updated_at, so created_at < updated_at.
-- - Console Merge inserts GRANTED/etc. with created_at = updated_at (same @now).
-- Stamp pending requests and any row whose lifecycle advanced after insert so
-- prior visitor requests that are no longer REQUESTED still flip the flag.
UPDATE cp_document_accesses
SET requested_at = created_at
WHERE status = 'REQUESTED'::compliance_portal_document_access_status
   OR created_at < updated_at;
