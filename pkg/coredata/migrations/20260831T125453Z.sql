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

-- Catalog risks are identity only. Treatment, scores, and owner live on
-- treatment_plans (see 20260825T113400Z / 20260825T113401Z). Drop the
-- columns that that migration made nullable, the generated scores that
-- depended on them, and owner_profile_id (already nullable; copied onto
-- treatment plans in the backfill).

ALTER TABLE risks
    DROP COLUMN inherent_risk_score,
    DROP COLUMN residual_risk_score,
    DROP COLUMN treatment,
    DROP COLUMN inherent_likelihood,
    DROP COLUMN inherent_impact,
    DROP COLUMN residual_likelihood,
    DROP COLUMN residual_impact,
    DROP COLUMN owner_profile_id;
