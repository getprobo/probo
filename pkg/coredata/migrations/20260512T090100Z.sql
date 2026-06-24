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

CREATE TABLE common_third_parties (
    id                               TEXT PRIMARY KEY,
    name                             TEXT NOT NULL,
    description                      TEXT,
    category                         vendor_category NOT NULL,
    headquarter_address              TEXT,
    legal_name                       TEXT,
    website_url                      TEXT,
    privacy_policy_url               TEXT,
    service_level_agreement_url      TEXT,
    service_software_agreement_url   TEXT,
    data_processing_agreement_url    TEXT,
    business_associate_agreement_url TEXT,
    subprocessors_list_url           TEXT,
    certifications                   TEXT[],
    status_page_url                  TEXT,
    terms_of_service_url             TEXT,
    security_page_url                TEXT,
    trust_page_url                   TEXT,
    created_at                       TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at                       TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE UNIQUE INDEX common_third_parties_name_key
    ON common_third_parties (lower(name));
