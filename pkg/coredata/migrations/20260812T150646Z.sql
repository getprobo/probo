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

-- third_parties.common_third_party_id has no index of any kind. Catalog
-- cleanup repoints these rows when two catalog entries are merged, matching
-- on the losing catalog id alone (WHERE common_third_party_id = <loser>),
-- deliberately without an organization or tenant predicate because a global
-- catalog row is referenced from every tenant. Duplicate detection likewise
-- counts them per catalog row across the whole table. Both are otherwise
-- full sequential scans of a large multi-tenant table with no other access
-- path available, so this is not a speculative index.
--
-- Partial: the column is NULL for every third party that was not imported
-- from the global catalog, which is the large majority of the table.
--
-- IF NOT EXISTS so this is a no-op where the index was already built by hand.
-- third_parties is large and CREATE INDEX takes a SHARE lock, which blocks
-- writes for the duration while still allowing reads, so an operator may
-- create it CONCURRENTLY ahead of the deploy. CONCURRENTLY cannot be used here: the migration runner wraps each
-- file in a transaction, and CONCURRENTLY is not allowed inside one.
CREATE INDEX IF NOT EXISTS third_parties_common_third_party_id_idx
    ON third_parties (common_third_party_id)
    WHERE common_third_party_id IS NOT NULL;

-- Redundant since it was created: common_third_party_domains already has
-- common_third_party_domains_party_domain_key on
-- (common_third_party_id, domain), whose leading column serves every lookup
-- this index could.
DROP INDEX IF EXISTS idx_common_third_party_domains_third_party;
