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

-- A host-only cookie and a Domain-scoped cookie can share a name.
-- COALESCE treats a missing domain as one host-only/unknown bucket so
-- NULL cookie_domain does not bypass the unique index.
--
-- IF NOT EXISTS / IF EXISTS so an operator can build the new index
-- CONCURRENTLY and drop the old one CONCURRENTLY ahead of deploy:
--
--   CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS
--       idx_detected_trackers_unique_identifier_domain_per_banner
--       ON detected_trackers (
--           cookie_banner_id,
--           tracker_type,
--           identifier,
--           COALESCE(cookie_domain, '')
--       );
--   DROP INDEX CONCURRENTLY IF EXISTS
--       idx_detected_trackers_unique_identifier_per_banner;
--
-- CONCURRENTLY cannot be used here: the migration runner wraps each
-- file in a transaction, and CONCURRENTLY is not allowed inside one.
CREATE UNIQUE INDEX IF NOT EXISTS idx_detected_trackers_unique_identifier_domain_per_banner
    ON detected_trackers (
        cookie_banner_id,
        tracker_type,
        identifier,
        COALESCE(cookie_domain, '')
    );

DROP INDEX IF EXISTS idx_detected_trackers_unique_identifier_per_banner;
