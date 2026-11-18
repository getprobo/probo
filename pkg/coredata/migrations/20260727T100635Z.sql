-- Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

CREATE TABLE third_party_administrators (
    third_party_id            text                     NOT NULL,
    administrator_profile_id  text                     NOT NULL,
    tenant_id                 text                     NOT NULL,
    organization_id           text                     NOT NULL,
    created_at                timestamp with time zone NOT NULL,
    updated_at                timestamp with time zone NOT NULL,
    PRIMARY KEY (third_party_id, administrator_profile_id),
    FOREIGN KEY (third_party_id) REFERENCES third_parties(id) ON UPDATE CASCADE ON DELETE CASCADE,
    FOREIGN KEY (administrator_profile_id) REFERENCES iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE RESTRICT
);

-- Migrate existing business and security owners into administrators (dedupe).
INSERT INTO third_party_administrators (
    third_party_id,
    administrator_profile_id,
    tenant_id,
    organization_id,
    created_at,
    updated_at
)
SELECT
    id,
    business_owner_profile_id,
    tenant_id,
    organization_id,
    created_at,
    updated_at
FROM third_parties
WHERE business_owner_profile_id IS NOT NULL
ON CONFLICT (third_party_id, administrator_profile_id) DO NOTHING;

INSERT INTO third_party_administrators (
    third_party_id,
    administrator_profile_id,
    tenant_id,
    organization_id,
    created_at,
    updated_at
)
SELECT
    id,
    security_owner_profile_id,
    tenant_id,
    organization_id,
    created_at,
    updated_at
FROM third_parties
WHERE security_owner_profile_id IS NOT NULL
ON CONFLICT (third_party_id, administrator_profile_id) DO NOTHING;
