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
