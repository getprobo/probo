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

CREATE TYPE treatment_plan_event_type AS ENUM (
    'CREATED',
    'UPDATED',
    'DELETED',
    'MEASURE_LINKED',
    'MEASURE_UNLINKED'
);

CREATE TABLE treatment_plan_events (
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    risk_analysis_id TEXT NOT NULL REFERENCES risk_analyses(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    treatment_plan_id TEXT NOT NULL,
    event_type treatment_plan_event_type NOT NULL,
    risk_id TEXT NOT NULL REFERENCES risks(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    owner_profile_id TEXT NOT NULL REFERENCES iam_membership_profiles(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    treatment risk_treatment NOT NULL,
    inherent_likelihood INTEGER NOT NULL,
    inherent_impact INTEGER NOT NULL,
    residual_likelihood INTEGER NOT NULL,
    residual_impact INTEGER NOT NULL,
    measure_ids TEXT[] NOT NULL,
    category TEXT NOT NULL,
    treatment_plan_created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    treatment_plan_updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TYPE measure_event_type AS ENUM (
    'CREATED',
    'UPDATED',
    'DELETED'
);

CREATE TABLE measure_events (
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    measure_id TEXT NOT NULL,
    event_type measure_event_type NOT NULL,
    name TEXT NOT NULL,
    category TEXT NOT NULL,
    state TEXT NOT NULL,
    measure_created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);

-- Seed current treatment plans and measures as CREATED events, the
-- starting point for as-of reconstruction. History before this
-- migration cannot be recovered.

INSERT INTO treatment_plan_events (
    tenant_id,
    organization_id,
    risk_analysis_id,
    treatment_plan_id,
    event_type,
    risk_id,
    owner_profile_id,
    treatment,
    inherent_likelihood,
    inherent_impact,
    residual_likelihood,
    residual_impact,
    measure_ids,
    category,
    treatment_plan_created_at,
    treatment_plan_updated_at,
    created_at
)
SELECT
    tp.tenant_id,
    tp.organization_id,
    tp.risk_analysis_id,
    tp.id,
    'CREATED',
    tp.risk_id,
    tp.owner_profile_id,
    tp.treatment,
    tp.inherent_likelihood,
    tp.inherent_impact,
    tp.residual_likelihood,
    tp.residual_impact,
    COALESCE(
        (
            SELECT array_agg(tpm.measure_id ORDER BY tpm.measure_id)
            FROM treatment_plans_measures tpm
            WHERE tpm.treatment_plan_id = tp.id
        ),
        ARRAY[]::TEXT[]
    ),
    r.category,
    tp.created_at,
    tp.updated_at,
    tp.created_at
FROM treatment_plans tp
INNER JOIN risks r ON r.id = tp.risk_id;

INSERT INTO measure_events (
    tenant_id,
    organization_id,
    measure_id,
    event_type,
    name,
    category,
    state,
    measure_created_at,
    created_at
)
SELECT
    m.tenant_id,
    m.organization_id,
    m.id,
    'CREATED',
    m.name,
    m.category,
    m.state::text,
    m.created_at,
    m.created_at
FROM measures m;
