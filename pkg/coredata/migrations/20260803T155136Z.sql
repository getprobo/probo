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

ALTER TABLE trust_centers RENAME TO compliance_portals;

ALTER TABLE cp_accesses RENAME COLUMN trust_center_id TO compliance_portal_id;
ALTER TABLE cp_references RENAME COLUMN trust_center_id TO compliance_portal_id;
ALTER TABLE cp_documents RENAME COLUMN trust_center_id TO compliance_portal_id;
ALTER TABLE cp_audits RENAME COLUMN trust_center_id TO compliance_portal_id;
ALTER TABLE cp_third_parties RENAME COLUMN trust_center_id TO compliance_portal_id;
ALTER TABLE cp_files RENAME COLUMN trust_center_id TO compliance_portal_id;
ALTER TABLE compliance_frameworks RENAME COLUMN trust_center_id TO compliance_portal_id;
ALTER TABLE compliance_custom_links RENAME COLUMN trust_center_id TO compliance_portal_id;
ALTER TABLE compliance_portal_commitment_groups RENAME COLUMN trust_center_id TO compliance_portal_id;
ALTER TABLE compliance_portal_commitments RENAME COLUMN trust_center_id TO compliance_portal_id;
ALTER TABLE rights_requests RENAME COLUMN trust_center_id TO compliance_portal_id;
ALTER TABLE cp_document_accesses RENAME COLUMN trust_center_access_id TO compliance_portal_access_id;
ALTER TABLE cp_document_accesses RENAME COLUMN trust_center_file_id TO compliance_portal_file_id;

ALTER TYPE trust_center_visibility RENAME TO compliance_portal_visibility;
ALTER TYPE trust_center_document_access_status RENAME TO compliance_portal_document_access_status;
ALTER TABLE cp_files RENAME COLUMN trust_center_visibility TO compliance_portal_visibility;

DO $$
DECLARE
    table_name TEXT;
    constraint_name TEXT;
    new_constraint_name TEXT;
BEGIN
    FOREACH table_name IN ARRAY ARRAY[
        'cp_accesses',
        'cp_references',
        'cp_documents',
        'cp_audits',
        'cp_third_parties',
        'cp_files',
        'cp_document_accesses',
        'compliance_portals',
        'compliance_frameworks',
        'compliance_custom_links',
        'compliance_portal_commitment_groups',
        'compliance_portal_commitments',
        'rights_requests'
    ]
    LOOP
        FOR constraint_name IN
            SELECT conname
            FROM pg_constraint
            WHERE conrelid = table_name::regclass
        LOOP
            new_constraint_name := replace(
                constraint_name,
                'trust_center_id',
                'compliance_portal_id'
            );
            new_constraint_name := replace(
                new_constraint_name,
                'trust_centers',
                'compliance_portals'
            );
            new_constraint_name := replace(
                new_constraint_name,
                'trust_center_visibility',
                'compliance_portal_visibility'
            );
            new_constraint_name := replace(
                new_constraint_name,
                'trust_center_access_id',
                'compliance_portal_access_id'
            );
            new_constraint_name := replace(
                new_constraint_name,
                'trust_center_file_id',
                'compliance_portal_file_id'
            );

            IF left(table_name, 3) = 'cp_' THEN
                new_constraint_name := replace(
                    new_constraint_name,
                    'cp_id',
                    'compliance_portal_id'
                );
                new_constraint_name := replace(
                    new_constraint_name,
                    'cp_visibility',
                    'compliance_portal_visibility'
                );
                new_constraint_name := replace(
                    new_constraint_name,
                    'cp_access_id',
                    'compliance_portal_access_id'
                );
                new_constraint_name := replace(
                    new_constraint_name,
                    'cp_file_id',
                    'compliance_portal_file_id'
                );
            END IF;

            IF new_constraint_name <> constraint_name THEN
                EXECUTE format(
                    'ALTER TABLE %I RENAME CONSTRAINT %I TO %I',
                    table_name,
                    constraint_name,
                    new_constraint_name
                );
            END IF;
        END LOOP;
    END LOOP;
END
$$;
