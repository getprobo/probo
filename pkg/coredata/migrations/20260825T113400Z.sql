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

CREATE TABLE treatment_plans (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    risk_id TEXT NOT NULL REFERENCES risks(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    risk_analysis_id TEXT NOT NULL REFERENCES risk_analyses(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    treatment risk_treatment NOT NULL,
    owner_profile_id TEXT NOT NULL REFERENCES iam_membership_profiles(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    inherent_likelihood INTEGER NOT NULL,
    inherent_impact INTEGER NOT NULL,
    inherent_risk_score INTEGER GENERATED ALWAYS AS (inherent_impact * inherent_likelihood) STORED,
    residual_likelihood INTEGER NOT NULL,
    residual_impact INTEGER NOT NULL,
    residual_risk_score INTEGER GENERATED ALWAYS AS (residual_impact * residual_likelihood) STORED,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    CONSTRAINT treatment_plans_risk_analysis_unique UNIQUE (risk_id, risk_analysis_id)
);

CREATE TABLE treatment_plans_measures (
    treatment_plan_id TEXT NOT NULL REFERENCES treatment_plans(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    measure_id TEXT NOT NULL REFERENCES measures(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    PRIMARY KEY (treatment_plan_id, measure_id)
);

ALTER TABLE risks
    ALTER COLUMN treatment DROP NOT NULL,
    ALTER COLUMN inherent_likelihood DROP NOT NULL,
    ALTER COLUMN inherent_impact DROP NOT NULL,
    ALTER COLUMN residual_likelihood DROP NOT NULL,
    ALTER COLUMN residual_impact DROP NOT NULL;

-- Risks linked to a scenario must not be deleted. Deleting the scenario still
-- removes the junction row (ON DELETE CASCADE on risk_analysis_scenario_id).

ALTER TABLE risk_analysis_scenario_risks
    DROP CONSTRAINT risk_analysis_scenario_risks_risk_id_fkey;

ALTER TABLE risk_analysis_scenario_risks
    ADD CONSTRAINT risk_analysis_scenario_risks_risk_id_fkey
        FOREIGN KEY (risk_id)
        REFERENCES risks(id)
        ON UPDATE CASCADE
        ON DELETE RESTRICT;
