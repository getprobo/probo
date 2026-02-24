-- Follow-up for access_entry_decision enum migration.
-- PostgreSQL requires a transaction boundary before newly added enum values
-- can be safely used in DML.

UPDATE access_entries
SET decision = 'ESCALATE'
WHERE decision = 'MODIFY';
