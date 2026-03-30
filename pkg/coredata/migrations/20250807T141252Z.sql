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

ALTER TABLE frameworks DROP CONSTRAINT IF EXISTS frameworks_organization_id_fkey;
ALTER TABLE frameworks ADD CONSTRAINT frameworks_organization_id_fkey
   FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE;

ALTER TABLE peoples DROP CONSTRAINT IF EXISTS peoples_organization_id_fkey;
ALTER TABLE peoples ADD CONSTRAINT peoples_organization_id_fkey
   FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE;

ALTER TABLE vendors DROP CONSTRAINT IF EXISTS vendors_organization_id_fkey;
ALTER TABLE vendors ADD CONSTRAINT vendors_organization_id_fkey
   FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE;

ALTER TABLE documents DROP CONSTRAINT IF EXISTS policies_organization_id_fkey;
ALTER TABLE documents DROP CONSTRAINT IF EXISTS documents_organization_id_fkey;
ALTER TABLE documents ADD CONSTRAINT documents_organization_id_fkey
   FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE;

ALTER TABLE risks DROP CONSTRAINT IF EXISTS risks_organization_id_fkey;
ALTER TABLE risks ADD CONSTRAINT risks_organization_id_fkey
   FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE;

ALTER TABLE connectors DROP CONSTRAINT IF EXISTS connectors_organization_id_fkey;
ALTER TABLE connectors ADD CONSTRAINT connectors_organization_id_fkey
   FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE;

ALTER TABLE data DROP CONSTRAINT IF EXISTS data_organization_id_fkey;
ALTER TABLE data ADD CONSTRAINT data_organization_id_fkey
   FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE;

ALTER TABLE users_organizations ADD CONSTRAINT users_organizations_organization_id_fkey
   FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE;

ALTER TABLE controls DROP CONSTRAINT IF EXISTS controls_framework_id_fkey;
ALTER TABLE controls ADD CONSTRAINT controls_framework_id_fkey
   FOREIGN KEY (framework_id) REFERENCES frameworks(id) ON DELETE CASCADE;

ALTER TABLE document_versions DROP CONSTRAINT IF EXISTS document_versions_document_id_fkey;
ALTER TABLE document_versions ADD CONSTRAINT document_versions_document_id_fkey
    FOREIGN KEY (document_id) REFERENCES documents(id) ON DELETE CASCADE;

ALTER TABLE document_versions DROP CONSTRAINT IF EXISTS document_versions_published_by_fkey;
ALTER TABLE document_versions ADD CONSTRAINT document_versions_published_by_fkey
    FOREIGN KEY (published_by) REFERENCES peoples(id)
    ON DELETE SET NULL
    ON UPDATE CASCADE;
