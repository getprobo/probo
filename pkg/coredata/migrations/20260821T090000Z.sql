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

CREATE TYPE common_third_party_review AS ENUM (
    'UNREVIEWED',
    'VALIDATED',
    'REJECTED'
);

ALTER TABLE common_third_parties
    ADD COLUMN review common_third_party_review NOT NULL DEFAULT 'UNREVIEWED',
    ADD COLUMN rejected_verdict common_tracker_pattern_attribution,
    ADD COLUMN reviewed_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN reviewed_by TEXT;

-- The default existed only to backfill the rows already in the table. Drop it
-- so an insert that omits the state fails instead of silently claiming nobody
-- has reviewed the row.
ALTER TABLE common_third_parties
    ALTER COLUMN review DROP DEFAULT;

-- A rejection must say which terminal verdict its patterns earn, and only a
-- rejection may carry one: the mapping pipeline reads this column directly
-- rather than mapping a reason onto a verdict.
--
-- Written as NOT NULL plus a membership test rather than
-- "rejected_verdict IN (...)". A CHECK passes when it evaluates to NULL, and
-- "NULL IN (...)" is NULL rather than false, so the obvious spelling accepts
-- exactly the row it is meant to forbid: REJECTED with no verdict.
ALTER TABLE common_third_parties
    ADD CONSTRAINT common_third_parties_rejected_verdict_check CHECK (
        CASE review
            WHEN 'REJECTED' THEN
                rejected_verdict IS NOT NULL
                AND rejected_verdict IN ('FIRST_PARTY', 'NOT_ATTRIBUTABLE')
            ELSE rejected_verdict IS NULL
        END
    );

-- The mapping pipeline and the catalog lookups filter on this, and the
-- reviewed set is small next to the unreviewed one.
CREATE INDEX common_third_parties_review_idx
    ON common_third_parties (review);
