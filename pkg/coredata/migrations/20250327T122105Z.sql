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

-- Add organization_id column to mitigations table
ALTER TABLE mitigations ADD COLUMN organization_id TEXT;

-- Update mitigations to set organization_id based on framework's organization_id
UPDATE mitigations m
SET organization_id = f.organization_id
FROM frameworks f
WHERE m.framework_id = f.id;

-- Make organization_id NOT NULL after update
ALTER TABLE mitigations ALTER COLUMN organization_id SET NOT NULL;

-- Add foreign key constraint
ALTER TABLE mitigations ADD CONSTRAINT fk_mitigations_organization_id
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE;

-- Add index for performance
CREATE INDEX idx_mitigations_organization_id ON mitigations(organization_id); 
