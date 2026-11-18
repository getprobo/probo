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

-- Rename vendor concept to third party everywhere in the schema.

-- Rename the vendor_category enum.

ALTER TYPE vendor_category RENAME TO third_party_category;

-- Rename the main vendor tables.

ALTER TABLE vendors RENAME TO third_parties;
ALTER TABLE vendor_contacts RENAME TO third_party_contacts;
ALTER TABLE vendor_services RENAME TO third_party_services;
ALTER TABLE vendor_compliance_reports RENAME TO third_party_compliance_reports;
ALTER TABLE vendor_business_associate_agreements RENAME TO third_party_business_associate_agreements;
ALTER TABLE vendor_data_privacy_agreements RENAME TO third_party_data_privacy_agreements;
ALTER TABLE vendor_risk_assessments RENAME TO third_party_risk_assessments;

-- Rename the junction tables.

ALTER TABLE asset_vendors RENAME TO asset_third_parties;
ALTER TABLE data_vendors RENAME TO data_third_parties;
ALTER TABLE processing_activity_vendors RENAME TO processing_activity_third_parties;

-- Rename vendor_id columns on child tables.

ALTER TABLE third_party_contacts RENAME COLUMN vendor_id TO third_party_id;
ALTER TABLE third_party_services RENAME COLUMN vendor_id TO third_party_id;
ALTER TABLE third_party_compliance_reports RENAME COLUMN vendor_id TO third_party_id;
ALTER TABLE third_party_business_associate_agreements RENAME COLUMN vendor_id TO third_party_id;
ALTER TABLE third_party_data_privacy_agreements RENAME COLUMN vendor_id TO third_party_id;
ALTER TABLE third_party_risk_assessments RENAME COLUMN vendor_id TO third_party_id;

-- Rename vendor_id columns on junction tables.

ALTER TABLE asset_third_parties RENAME COLUMN vendor_id TO third_party_id;
ALTER TABLE data_third_parties RENAME COLUMN vendor_id TO third_party_id;
ALTER TABLE processing_activity_third_parties RENAME COLUMN vendor_id TO third_party_id;

-- Rename generated_documents.vendors_document_id.

ALTER TABLE generated_documents RENAME COLUMN vendors_document_id TO third_parties_document_id;

-- Rename webhook event type enum values.

ALTER TYPE webhook_event_type RENAME VALUE 'vendor:created' TO 'third-party:created';
ALTER TYPE webhook_event_type RENAME VALUE 'vendor:updated' TO 'third-party:updated';
ALTER TYPE webhook_event_type RENAME VALUE 'vendor:deleted' TO 'third-party:deleted';
