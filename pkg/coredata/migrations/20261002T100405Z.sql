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

-- Visitor grants belong on the portal catalog link, not the org-wide
-- document or audit PDF. After this, unlinking a catalog row CASCADE-deletes
-- its requests.

ALTER TABLE cp_document_accesses
    ADD COLUMN compliance_portal_document_id TEXT,
    ADD COLUMN compliance_portal_audit_id TEXT;

UPDATE cp_document_accesses da
SET compliance_portal_document_id = cpd.id
FROM cp_accesses acc, cp_documents cpd
WHERE da.compliance_portal_access_id = acc.id
    AND cpd.compliance_portal_id = acc.compliance_portal_id
    AND cpd.document_id = da.document_id
    AND da.document_id IS NOT NULL;

UPDATE cp_document_accesses da
SET compliance_portal_audit_id = cpa.id
FROM cp_accesses acc, audits, cp_audits cpa
WHERE da.compliance_portal_access_id = acc.id
    AND audits.report_file_id = da.report_file_id
    AND cpa.compliance_portal_id = acc.compliance_portal_id
    AND cpa.audit_id = audits.id
    AND da.report_file_id IS NOT NULL;

DELETE FROM cp_document_accesses
WHERE (document_id IS NOT NULL AND compliance_portal_document_id IS NULL)
   OR (report_file_id IS NOT NULL AND compliance_portal_audit_id IS NULL);

ALTER TABLE cp_document_accesses
    DROP CONSTRAINT cp_document_accesses_check;

ALTER TABLE cp_document_accesses
    DROP COLUMN document_id,
    DROP COLUMN report_file_id;

ALTER TABLE cp_document_accesses
    ADD CONSTRAINT cp_document_accesses_cp_document_id_fkey
        FOREIGN KEY (compliance_portal_document_id)
        REFERENCES cp_documents (id)
        ON DELETE CASCADE;

ALTER TABLE cp_document_accesses
    ADD CONSTRAINT cp_document_accesses_cp_audit_id_fkey
        FOREIGN KEY (compliance_portal_audit_id)
        REFERENCES cp_audits (id)
        ON DELETE CASCADE;

ALTER TABLE cp_document_accesses
    ADD CONSTRAINT cp_document_accesses_check CHECK (
        (compliance_portal_document_id IS NOT NULL)::int
        + (compliance_portal_audit_id IS NOT NULL)::int
        + (compliance_portal_file_id IS NOT NULL)::int
        = 1
    );

ALTER TABLE cp_document_accesses
    ADD CONSTRAINT cp_document_accesses_access_id_cp_document_id_key
        UNIQUE (compliance_portal_access_id, compliance_portal_document_id);

ALTER TABLE cp_document_accesses
    ADD CONSTRAINT cp_document_accesses_access_id_cp_audit_id_key
        UNIQUE (compliance_portal_access_id, compliance_portal_audit_id);
