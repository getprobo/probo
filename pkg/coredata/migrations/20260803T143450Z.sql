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

ALTER TABLE trust_center_accesses RENAME TO cp_accesses;
ALTER TABLE trust_center_audits RENAME TO cp_audits;
ALTER TABLE trust_center_document_accesses RENAME TO cp_document_accesses;
ALTER TABLE trust_center_documents RENAME TO cp_documents;
ALTER TABLE trust_center_files RENAME TO cp_files;
ALTER TABLE trust_center_references RENAME TO cp_references;
ALTER TABLE trust_center_third_parties RENAME TO cp_third_parties;

DO $$
DECLARE
    table_name TEXT;
    constraint_name TEXT;
    new_constraint_name TEXT;
BEGIN
    FOREACH table_name IN ARRAY ARRAY[
        'cp_accesses',
        'cp_audits',
        'cp_document_accesses',
        'cp_documents',
        'cp_files',
        'cp_references',
        'cp_third_parties'
    ]
    LOOP
        FOR constraint_name IN
            SELECT conname
            FROM pg_constraint
            WHERE conrelid = table_name::regclass
        LOOP
            new_constraint_name := CASE constraint_name
                WHEN 'trust_center_document_accesse_trust_center_access_id_docume_key'
                    THEN 'cp_document_accesses_cp_access_id_document_id_key'
                WHEN 'trust_center_document_accesse_trust_center_access_id_report_key'
                    THEN 'cp_document_accesses_cp_access_id_report_id_key'
                WHEN 'trust_center_document_accesses_trust_center_access_id_report_fi'
                    THEN 'cp_document_accesses_cp_access_id_report_file_id_key'
                ELSE replace(constraint_name, 'trust_center', 'cp')
            END;

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
