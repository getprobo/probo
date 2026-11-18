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

-- Rename risk analysis scopes to diagrams.

ALTER TABLE risk_analysis_scopes RENAME TO risk_analysis_diagrams;

ALTER TABLE risk_analysis_nodes RENAME COLUMN risk_analysis_scope_id TO risk_analysis_diagram_id;
ALTER TABLE risk_analysis_boundaries RENAME COLUMN risk_analysis_scope_id TO risk_analysis_diagram_id;
ALTER TABLE risk_analysis_processes RENAME COLUMN risk_analysis_scope_id TO risk_analysis_diagram_id;
ALTER TABLE risk_analysis_threats RENAME COLUMN risk_analysis_scope_id TO risk_analysis_diagram_id;
ALTER TABLE risk_analysis_scenarios RENAME COLUMN risk_analysis_scope_id TO risk_analysis_diagram_id;

-- Rename constraints so Go PgError mapping stays aligned with table names.
ALTER TABLE risk_analysis_diagrams RENAME CONSTRAINT risk_analysis_scopes_pkey TO risk_analysis_diagrams_pkey;

-- Rename foreign-key constraints to match the new table/column names.
ALTER TABLE risk_analysis_nodes RENAME CONSTRAINT risk_analysis_nodes_risk_analysis_scope_id_fkey TO risk_analysis_nodes_risk_analysis_diagram_id_fkey;
ALTER TABLE risk_analysis_boundaries RENAME CONSTRAINT risk_analysis_boundaries_risk_analysis_scope_id_fkey TO risk_analysis_boundaries_risk_analysis_diagram_id_fkey;
ALTER TABLE risk_analysis_processes RENAME CONSTRAINT risk_analysis_processes_risk_analysis_scope_id_fkey TO risk_analysis_processes_risk_analysis_diagram_id_fkey;
ALTER TABLE risk_analysis_threats RENAME CONSTRAINT risk_analysis_threats_risk_analysis_scope_id_fkey TO risk_analysis_threats_risk_analysis_diagram_id_fkey;
ALTER TABLE risk_analysis_scenarios RENAME CONSTRAINT risk_analysis_scenarios_risk_analysis_scope_id_fkey TO risk_analysis_scenarios_risk_analysis_diagram_id_fkey;
ALTER TABLE risk_analysis_diagrams RENAME CONSTRAINT risk_analysis_scopes_risk_analysis_id_fkey TO risk_analysis_diagrams_risk_analysis_id_fkey;
