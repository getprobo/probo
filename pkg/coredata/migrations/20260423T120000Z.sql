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

-- Rename vendor tables to third_party tables.

ALTER TABLE vendors RENAME TO third_parties;
ALTER TABLE vendor_contacts RENAME TO third_party_contacts;
ALTER TABLE vendor_services RENAME TO third_party_services;
ALTER TABLE vendor_compliance_reports RENAME TO third_party_compliance_reports;
ALTER TABLE vendor_business_associate_agreements RENAME TO third_party_business_associate_agreements;
ALTER TABLE vendor_data_privacy_agreements RENAME TO third_party_data_privacy_agreements;
ALTER TABLE vendor_risk_assessments RENAME TO third_party_risk_assessments;

-- Rename junction tables.

ALTER TABLE asset_vendors RENAME TO asset_third_parties;
ALTER TABLE data_vendors RENAME TO data_third_parties;
ALTER TABLE processing_activity_vendors RENAME TO processing_activity_third_parties;

-- Rename vendor_id columns in child tables.

ALTER TABLE third_party_contacts RENAME COLUMN vendor_id TO third_party_id;
ALTER TABLE third_party_services RENAME COLUMN vendor_id TO third_party_id;
ALTER TABLE third_party_compliance_reports RENAME COLUMN vendor_id TO third_party_id;
ALTER TABLE third_party_business_associate_agreements RENAME COLUMN vendor_id TO third_party_id;
ALTER TABLE third_party_data_privacy_agreements RENAME COLUMN vendor_id TO third_party_id;
ALTER TABLE third_party_risk_assessments RENAME COLUMN vendor_id TO third_party_id;

-- Rename vendor_id columns in junction tables.

ALTER TABLE asset_third_parties RENAME COLUMN vendor_id TO third_party_id;
ALTER TABLE data_third_parties RENAME COLUMN vendor_id TO third_party_id;
ALTER TABLE processing_activity_third_parties RENAME COLUMN vendor_id TO third_party_id;
