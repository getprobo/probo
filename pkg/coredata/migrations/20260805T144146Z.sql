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

-- Rename org risk_assessment* entities to risk_analysis*.
-- Third-party risk assessments (third_party_risk_assessments) are unchanged.

ALTER TABLE risk_assessments RENAME TO risk_analyses;
ALTER TABLE risk_assessment_scopes RENAME TO risk_analysis_scopes;
ALTER TABLE risk_assessment_nodes RENAME TO risk_analysis_nodes;
ALTER TABLE risk_assessment_boundaries RENAME TO risk_analysis_boundaries;
ALTER TABLE risk_assessment_processes RENAME TO risk_analysis_processes;
ALTER TABLE risk_assessment_threats RENAME TO risk_analysis_threats;
ALTER TABLE risk_assessment_scenarios RENAME TO risk_analysis_scenarios;
ALTER TABLE risk_assessment_scenario_threats RENAME TO risk_analysis_scenario_threats;
ALTER TABLE risk_assessment_scenario_risks RENAME TO risk_analysis_scenario_risks;

ALTER TABLE risk_analysis_scopes RENAME COLUMN risk_assessment_id TO risk_analysis_id;
ALTER TABLE risk_analysis_nodes RENAME COLUMN risk_assessment_scope_id TO risk_analysis_scope_id;
ALTER TABLE risk_analysis_boundaries RENAME COLUMN risk_assessment_scope_id TO risk_analysis_scope_id;
ALTER TABLE risk_analysis_processes RENAME COLUMN risk_assessment_scope_id TO risk_analysis_scope_id;
ALTER TABLE risk_analysis_threats RENAME COLUMN risk_assessment_scope_id TO risk_analysis_scope_id;
ALTER TABLE risk_analysis_scenarios RENAME COLUMN risk_assessment_scope_id TO risk_analysis_scope_id;
ALTER TABLE risk_analysis_scenario_threats RENAME COLUMN risk_assessment_scenario_id TO risk_analysis_scenario_id;
ALTER TABLE risk_analysis_scenario_threats RENAME COLUMN risk_assessment_threat_id TO risk_analysis_threat_id;
ALTER TABLE risk_analysis_scenario_risks RENAME COLUMN risk_assessment_scenario_id TO risk_analysis_scenario_id;

ALTER TYPE risk_assessment_node_type RENAME TO risk_analysis_node_type;

-- Rename constraints so Go PgError mapping stays aligned with table names.
-- The org risk_assessments PK is risk_assessments_pkey1 because the earlier
-- vendor risk_assessments table (now third_party_risk_assessments) still owns
-- the risk_assessments_pkey constraint name after its table rename.
ALTER TABLE risk_analyses RENAME CONSTRAINT risk_assessments_pkey1 TO risk_analyses_pkey;
ALTER TABLE risk_analysis_scopes RENAME CONSTRAINT risk_assessment_scopes_pkey TO risk_analysis_scopes_pkey;
ALTER TABLE risk_analysis_nodes RENAME CONSTRAINT risk_assessment_nodes_pkey TO risk_analysis_nodes_pkey;
ALTER TABLE risk_analysis_boundaries RENAME CONSTRAINT risk_assessment_boundaries_pkey TO risk_analysis_boundaries_pkey;
ALTER TABLE risk_analysis_processes RENAME CONSTRAINT risk_assessment_processes_pkey TO risk_analysis_processes_pkey;
ALTER TABLE risk_analysis_threats RENAME CONSTRAINT risk_assessment_threats_pkey TO risk_analysis_threats_pkey;
ALTER TABLE risk_analysis_scenarios RENAME CONSTRAINT risk_assessment_scenarios_pkey TO risk_analysis_scenarios_pkey;
ALTER TABLE risk_analysis_scenario_threats RENAME CONSTRAINT risk_assessment_scenario_threats_pkey TO risk_analysis_scenario_threats_pkey;
ALTER TABLE risk_analysis_scenario_risks RENAME CONSTRAINT risk_assessment_scenario_risks_pkey TO risk_analysis_scenario_risks_pkey;

ALTER TABLE risk_analysis_nodes RENAME CONSTRAINT risk_assessment_nodes_unique_name TO risk_analysis_nodes_unique_name;
ALTER TABLE risk_analysis_boundaries RENAME CONSTRAINT risk_assessment_boundaries_unique_name TO risk_analysis_boundaries_unique_name;
ALTER TABLE risk_analysis_processes RENAME CONSTRAINT risk_assessment_processes_unique_name TO risk_analysis_processes_unique_name;
ALTER TABLE risk_analysis_threats RENAME CONSTRAINT risk_assessment_threats_unique_name TO risk_analysis_threats_unique_name;

-- Rename foreign-key constraints to match the new table/column names.
ALTER TABLE risk_analysis_scopes RENAME CONSTRAINT risk_assessment_scopes_risk_assessment_id_fkey TO risk_analysis_scopes_risk_analysis_id_fkey;
ALTER TABLE risk_analysis_nodes RENAME CONSTRAINT risk_assessment_nodes_risk_assessment_scope_id_fkey TO risk_analysis_nodes_risk_analysis_scope_id_fkey;
ALTER TABLE risk_analysis_nodes RENAME CONSTRAINT risk_assessment_nodes_boundary_id_fkey TO risk_analysis_nodes_boundary_id_fkey;
ALTER TABLE risk_analysis_boundaries RENAME CONSTRAINT risk_assessment_boundaries_risk_assessment_scope_id_fkey TO risk_analysis_boundaries_risk_analysis_scope_id_fkey;
ALTER TABLE risk_analysis_boundaries RENAME CONSTRAINT risk_assessment_boundaries_parent_boundary_id_fkey TO risk_analysis_boundaries_parent_boundary_id_fkey;
ALTER TABLE risk_analysis_processes RENAME CONSTRAINT risk_assessment_processes_risk_assessment_scope_id_fkey TO risk_analysis_processes_risk_analysis_scope_id_fkey;
ALTER TABLE risk_analysis_processes RENAME CONSTRAINT risk_assessment_processes_source_node_id_fkey TO risk_analysis_processes_source_node_id_fkey;
ALTER TABLE risk_analysis_processes RENAME CONSTRAINT risk_assessment_processes_target_node_id_fkey TO risk_analysis_processes_target_node_id_fkey;
ALTER TABLE risk_analysis_threats RENAME CONSTRAINT risk_assessment_threats_risk_assessment_scope_id_fkey TO risk_analysis_threats_risk_analysis_scope_id_fkey;
ALTER TABLE risk_analysis_threats RENAME CONSTRAINT risk_assessment_threats_process_id_fkey TO risk_analysis_threats_process_id_fkey;
ALTER TABLE risk_analysis_scenarios RENAME CONSTRAINT risk_assessment_scenarios_risk_assessment_scope_id_fkey TO risk_analysis_scenarios_risk_analysis_scope_id_fkey;
ALTER TABLE risk_analysis_scenario_threats RENAME CONSTRAINT risk_assessment_scenario_threa_risk_assessment_scenario_id_fkey TO risk_analysis_scenario_threats_risk_analysis_scenario_id_fkey;
ALTER TABLE risk_analysis_scenario_threats RENAME CONSTRAINT risk_assessment_scenario_threats_risk_assessment_threat_id_fkey TO risk_analysis_scenario_threats_risk_analysis_threat_id_fkey;
ALTER TABLE risk_analysis_scenario_risks RENAME CONSTRAINT risk_assessment_scenario_risks_risk_assessment_scenario_id_fkey TO risk_analysis_scenario_risks_risk_analysis_scenario_id_fkey;
ALTER TABLE risk_analysis_scenario_risks RENAME CONSTRAINT risk_assessment_scenario_risks_risk_id_fkey TO risk_analysis_scenario_risks_risk_id_fkey;
ALTER TABLE risk_analyses RENAME CONSTRAINT risk_assessments_organization_id_fkey TO risk_analyses_organization_id_fkey;
