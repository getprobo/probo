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

-- Snapshot catalog scores onto a treatment plan for every risk that already
-- sits on a scenario of a 5x5 analysis. Skip incomplete catalog rows: a plan
-- requires treatment, owner, and inherent scores. Residual falls back to
-- inherent when the catalog residual was never set. Entity type 129 is
-- TreatmentPlanEntityType. Idempotent on (risk_id, risk_analysis_id).

WITH candidates AS (
    SELECT DISTINCT
        r.tenant_id,
        r.organization_id,
        r.id AS risk_id,
        ra.id AS risk_analysis_id,
        r.treatment,
        r.owner_profile_id,
        r.inherent_likelihood,
        r.inherent_impact,
        COALESCE(r.residual_likelihood, r.inherent_likelihood) AS residual_likelihood,
        COALESCE(r.residual_impact, r.inherent_impact) AS residual_impact
    FROM risk_analysis_scenario_risks srr
    INNER JOIN risk_analysis_scenarios s
        ON s.id = srr.risk_analysis_scenario_id
    INNER JOIN risk_analysis_diagrams d
        ON d.id = s.risk_analysis_diagram_id
    INNER JOIN risk_analyses ra
        ON ra.id = d.risk_analysis_id
    INNER JOIN risks r
        ON r.id = srr.risk_id
    WHERE ra.matrix_rows = 5
        AND ra.matrix_cols = 5
        AND r.treatment IS NOT NULL
        AND r.owner_profile_id IS NOT NULL
        AND r.inherent_likelihood IS NOT NULL
        AND r.inherent_impact IS NOT NULL
),
inserted AS (
    INSERT INTO treatment_plans (
        id,
        tenant_id,
        organization_id,
        risk_id,
        risk_analysis_id,
        treatment,
        owner_profile_id,
        inherent_likelihood,
        inherent_impact,
        residual_likelihood,
        residual_impact,
        created_at,
        updated_at
    )
    SELECT
        generate_gid(decode_base64_unpadded(c.tenant_id), 129),
        c.tenant_id,
        c.organization_id,
        c.risk_id,
        c.risk_analysis_id,
        c.treatment,
        c.owner_profile_id,
        c.inherent_likelihood,
        c.inherent_impact,
        c.residual_likelihood,
        c.residual_impact,
        NOW(),
        NOW()
    FROM candidates c
    ON CONFLICT (risk_id, risk_analysis_id) DO NOTHING
)
INSERT INTO treatment_plans_measures (
    treatment_plan_id,
    measure_id,
    tenant_id,
    organization_id,
    created_at
)
SELECT
    tp.id,
    rm.measure_id,
    tp.tenant_id,
    tp.organization_id,
    NOW()
FROM candidates c
INNER JOIN treatment_plans tp
    ON tp.risk_id = c.risk_id
    AND tp.risk_analysis_id = c.risk_analysis_id
INNER JOIN risks_measures rm
    ON rm.risk_id = c.risk_id
ON CONFLICT (treatment_plan_id, measure_id) DO NOTHING;
