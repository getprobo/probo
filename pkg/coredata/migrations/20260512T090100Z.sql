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
