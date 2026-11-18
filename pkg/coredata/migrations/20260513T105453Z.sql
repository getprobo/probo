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

-- common_third_party_domains: maps all known eTLD+1 domains operated by a
-- third party (e.g. Google owns googletagmanager.com, google-analytics.com,
-- doubleclick.net). Used for fast domain-based tracker attribution.
CREATE TABLE common_third_party_domains (
    id                    TEXT PRIMARY KEY,
    common_third_party_id TEXT NOT NULL REFERENCES common_third_parties(id) ON DELETE CASCADE,
    domain                CITEXT NOT NULL,
    created_at            TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at            TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE UNIQUE INDEX common_third_party_domains_party_domain_key
    ON common_third_party_domains (common_third_party_id, domain);

CREATE INDEX idx_common_third_party_domains_third_party
    ON common_third_party_domains (common_third_party_id);

-- common_tracker_patterns: global (not tenant/org scoped) knowledge base of
-- tracker patterns linked to third parties. Grows over time as we discover
-- new tracker-to-vendor mappings.
CREATE TABLE common_tracker_patterns (
    id                    TEXT PRIMARY KEY,
    common_third_party_id TEXT REFERENCES common_third_parties(id) ON DELETE SET NULL,
    tracker_type          tracker_type NOT NULL,
    pattern               TEXT NOT NULL,
    match_type            cookie_pattern_match_type NOT NULL,
    description           TEXT NOT NULL,
    max_age_seconds       INTEGER,
    confidence            REAL NOT NULL,
    created_at            TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at            TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE UNIQUE INDEX idx_common_tracker_patterns_unique
    ON common_tracker_patterns (tracker_type, pattern, COALESCE(max_age_seconds, -1));

-- Link tracker_patterns to the common knowledge base and to third_parties.
ALTER TABLE tracker_patterns
    ADD COLUMN common_tracker_pattern_id  TEXT REFERENCES common_tracker_patterns(id) ON DELETE SET NULL,
    ADD COLUMN third_party_id             TEXT REFERENCES third_parties(id) ON DELETE SET NULL;

-- Link third_parties to the common third party they were created from.
ALTER TABLE third_parties
    ADD COLUMN common_third_party_id TEXT REFERENCES common_third_parties(id) ON DELETE SET NULL;

-- Pre-extracted eTLD+1 for fast domain-based joins in the mapping worker.
ALTER TABLE detected_trackers
    ADD COLUMN initiator_domain CITEXT;

CREATE INDEX idx_detected_trackers_initiator_domain
    ON detected_trackers (initiator_domain)
    WHERE initiator_domain IS NOT NULL;
