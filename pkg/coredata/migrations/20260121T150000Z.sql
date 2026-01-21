ALTER TABLE states_of_applicability_controls RENAME TO applicability_statements;

ALTER TABLE applicability_statements 
    RENAME CONSTRAINT states_of_applicability_controls_organization_id_fkey 
    TO applicability_statements_organization_id_fkey;

ALTER TABLE applicability_statements 
    RENAME CONSTRAINT states_of_applicability_controls_snapshot_id_fkey 
    TO applicability_statements_snapshot_id_fkey;

ALTER INDEX states_of_applicability_controls_pkey RENAME TO applicability_statements_pkey;
