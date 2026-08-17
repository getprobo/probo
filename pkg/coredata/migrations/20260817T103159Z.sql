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

DO $$
DECLARE
    cname text;
BEGIN
    SELECT c.conname
    INTO cname
    FROM pg_constraint c
    JOIN pg_class t ON t.oid = c.conrelid
    WHERE t.relname = 'bot_delivery_destinations'
      AND c.contype = 'u'
      AND pg_get_constraintdef(c.oid) = 'UNIQUE (provider, target_namespace, target_key)';

    IF cname IS NOT NULL THEN
        EXECUTE format('ALTER TABLE bot_delivery_destinations DROP CONSTRAINT %I', cname);
    END IF;
END $$;

ALTER TABLE bot_delivery_destinations
    ADD CONSTRAINT bot_delivery_destinations_org_target_key
    UNIQUE (tenant_id, organization_id, provider, target_namespace, target_key);
