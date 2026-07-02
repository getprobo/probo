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

CREATE TABLE risk_assessments (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE risk_assessment_scopes (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL,
    risk_assessment_id TEXT NOT NULL REFERENCES risk_assessments(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TYPE risk_assessment_node_type AS ENUM ('ENTITY', 'BOUNDARY', 'ASSET', 'DATA');

CREATE TABLE risk_assessment_nodes (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL,
    risk_assessment_scope_id TEXT NOT NULL REFERENCES risk_assessment_scopes(id) ON DELETE CASCADE,
    node_type risk_assessment_node_type NOT NULL,
    name TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    CONSTRAINT risk_assessment_nodes_unique_name UNIQUE (risk_assessment_scope_id, name)
);

CREATE TABLE risk_assessment_processes (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL,
    risk_assessment_scope_id TEXT NOT NULL REFERENCES risk_assessment_scopes(id) ON DELETE CASCADE,
    source_node_id TEXT NOT NULL REFERENCES risk_assessment_nodes(id) ON DELETE CASCADE,
    target_node_id TEXT NOT NULL REFERENCES risk_assessment_nodes(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    CONSTRAINT risk_assessment_processes_unique_name UNIQUE (risk_assessment_scope_id, name)
);

CREATE TABLE risk_assessment_threats (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL,
    risk_assessment_scope_id TEXT NOT NULL REFERENCES risk_assessment_scopes(id) ON DELETE CASCADE,
    process_id TEXT NOT NULL REFERENCES risk_assessment_processes(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    category TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    CONSTRAINT risk_assessment_threats_unique_name UNIQUE (risk_assessment_scope_id, name)
);

CREATE TABLE risk_assessment_scenarios (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL,
    risk_assessment_scope_id TEXT NOT NULL REFERENCES risk_assessment_scopes(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE risk_assessment_scenario_threats (
    tenant_id TEXT NOT NULL,
    risk_assessment_scenario_id TEXT NOT NULL REFERENCES risk_assessment_scenarios(id) ON DELETE CASCADE,
    risk_assessment_threat_id TEXT NOT NULL REFERENCES risk_assessment_threats(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (risk_assessment_scenario_id, risk_assessment_threat_id)
);

CREATE TABLE risk_assessment_scenario_risks (
    tenant_id TEXT NOT NULL,
    risk_assessment_scenario_id TEXT NOT NULL REFERENCES risk_assessment_scenarios(id) ON DELETE CASCADE,
    risk_id TEXT NOT NULL REFERENCES risks(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (risk_assessment_scenario_id, risk_id)
);
