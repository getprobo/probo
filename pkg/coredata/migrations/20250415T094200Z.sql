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

ALTER TABLE risks ADD COLUMN inherent_likelihood_int INTEGER;
ALTER TABLE risks ADD COLUMN inherent_impact_int INTEGER;
ALTER TABLE risks ADD COLUMN residual_likelihood_int INTEGER;
ALTER TABLE risks ADD COLUMN residual_impact_int INTEGER;

UPDATE risks SET 
    inherent_likelihood_int = CASE
        WHEN inherent_likelihood <= 0.20 THEN 1
        WHEN inherent_likelihood <= 0.40 THEN 2
        WHEN inherent_likelihood <= 0.60 THEN 3
        WHEN inherent_likelihood <= 0.80 THEN 4
        ELSE 5
    END,
    inherent_impact_int = CASE
        WHEN inherent_impact <= 0.20 THEN 1
        WHEN inherent_impact <= 0.40 THEN 2
        WHEN inherent_impact <= 0.60 THEN 3
        WHEN inherent_impact <= 0.80 THEN 4
        ELSE 5
    END,
    residual_likelihood_int = CASE
        WHEN residual_likelihood <= 0.20 THEN 1
        WHEN residual_likelihood <= 0.40 THEN 2
        WHEN residual_likelihood <= 0.60 THEN 3
        WHEN residual_likelihood <= 0.80 THEN 4
        ELSE 5
    END,
    residual_impact_int = CASE
        WHEN residual_impact <= 0.20 THEN 1
        WHEN residual_impact <= 0.40 THEN 2
        WHEN residual_impact <= 0.60 THEN 3
        WHEN residual_impact <= 0.80 THEN 4
        ELSE 5
    END;

ALTER TABLE risks DROP COLUMN inherent_likelihood;
ALTER TABLE risks DROP COLUMN inherent_impact;
ALTER TABLE risks DROP COLUMN residual_likelihood;
ALTER TABLE risks DROP COLUMN residual_impact;

ALTER TABLE risks RENAME COLUMN inherent_likelihood_int TO inherent_likelihood;
ALTER TABLE risks RENAME COLUMN inherent_impact_int TO inherent_impact;
ALTER TABLE risks RENAME COLUMN residual_likelihood_int TO residual_likelihood;
ALTER TABLE risks RENAME COLUMN residual_impact_int TO residual_impact;
