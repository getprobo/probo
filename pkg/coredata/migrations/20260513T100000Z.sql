-- Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
