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

-- Measures become internal controls: a stable code, how the control works,
-- separate run / evidence / test cadences, derived due dates, and two people
-- (owner and reviewer) so duties stay separated.

ALTER TABLE measures
    ADD COLUMN code TEXT,
    ADD COLUMN control_type TEXT,
    ADD COLUMN nature TEXT,
    ADD COLUMN operating_frequency INTERVAL,
    ADD COLUMN evidence_cadence INTERVAL,
    ADD COLUMN testing_cadence INTERVAL,
    ADD COLUMN next_evidence_due TIMESTAMP WITH TIME ZONE,
    ADD COLUMN next_test_due TIMESTAMP WITH TIME ZONE,
    ADD COLUMN implementation_status TEXT,
    ADD COLUMN owner_profile_id TEXT,
    ADD COLUMN reviewer_profile_id TEXT;

UPDATE measures
SET implementation_status = CASE state::text
    WHEN 'IN_PROGRESS' THEN 'IN_PROGRESS'
    WHEN 'IMPLEMENTED' THEN 'IMPLEMENTED'
    WHEN 'NOT_IMPLEMENTED' THEN 'NOT_IMPLEMENTED'
    ELSE 'NOT_IMPLEMENTED'
END;

ALTER TABLE measures
    ALTER COLUMN implementation_status SET DEFAULT 'NOT_IMPLEMENTED';

ALTER TABLE measures
    ALTER COLUMN implementation_status SET NOT NULL;

ALTER TABLE measures
    ALTER COLUMN implementation_status DROP DEFAULT;

ALTER TABLE measures
    ADD CONSTRAINT measures_control_type_check
        CHECK (
            control_type IS NULL
            OR control_type IN ('PREVENTIVE', 'DETECTIVE', 'CORRECTIVE')
        ),
    ADD CONSTRAINT measures_nature_check
        CHECK (
            nature IS NULL
            OR nature IN ('MANUAL')
        ),
    ADD CONSTRAINT measures_implementation_status_check
        CHECK (
            implementation_status IN (
                'NOT_IMPLEMENTED',
                'IN_PROGRESS',
                'IMPLEMENTED',
                'OPERATING'
            )
        ),
    ADD CONSTRAINT measures_owner_reviewer_distinct_check
        CHECK (
            owner_profile_id IS NULL
            OR reviewer_profile_id IS NULL
            OR owner_profile_id <> reviewer_profile_id
        ),
    ADD CONSTRAINT measures_owner_profile_id_fkey
        FOREIGN KEY (owner_profile_id)
        REFERENCES iam_membership_profiles(id)
        ON UPDATE CASCADE
        ON DELETE RESTRICT,
    ADD CONSTRAINT measures_reviewer_profile_id_fkey
        FOREIGN KEY (reviewer_profile_id)
        REFERENCES iam_membership_profiles(id)
        ON UPDATE CASCADE
        ON DELETE RESTRICT,
    ADD CONSTRAINT measures_organization_id_code_key
        UNIQUE (organization_id, code);

-- Rebuild the existing search document so a control code is findable.
ALTER TABLE measures DROP COLUMN search_vector;

ALTER TABLE measures ADD COLUMN search_vector tsvector
GENERATED ALWAYS AS (
    to_tsvector(
        'simple',
        COALESCE(name, '') || ' ' || COALESCE(code, '')
    )
) STORED;

CREATE INDEX measures_search_idx ON measures USING gin(search_vector);
