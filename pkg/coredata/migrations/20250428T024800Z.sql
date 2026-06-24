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

CREATE TYPE policy_version_signature_state AS ENUM ('REQUESTED', 'SIGNED');

CREATE TABLE policy_version_signatures (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    policy_version_id TEXT NOT NULL REFERENCES policy_versions(id) ON DELETE CASCADE,
    state policy_version_signature_state NOT NULL,
    signed_by TEXT NOT NULL REFERENCES peoples(id),
    signed_at TIMESTAMP WITH TIME ZONE,
    requested_at TIMESTAMP WITH TIME ZONE NOT NULL,
    requested_by TEXT NOT NULL REFERENCES peoples(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE (policy_version_id, signed_by)
);
