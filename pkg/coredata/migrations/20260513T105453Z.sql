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

CREATE UNIQUE INDEX common_third_party_domains_domain_key
    ON common_third_party_domains (domain);

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
    description           TEXT NOT NULL DEFAULT '',
    max_age_seconds       INTEGER,
    confidence            REAL NOT NULL DEFAULT 0,
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
