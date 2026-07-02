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

-- The register/document model has fully replaced the snapshot system. Delete the
-- snapshot-scoped data; the schema (snapshot_id / source_id columns, snapshots
-- table, controls_snapshots, snapshots_type enum, snapshot-scoped indexes) is
-- dropped in a follow-up migration.

-- processing_activity_third_parties stores snapshot_id without an FK to snapshots,
-- so the cascade delete below would not reach it. Clean it up explicitly first.
DELETE FROM processing_activity_third_parties WHERE snapshot_id IS NOT NULL;

-- Every other table with a snapshot_id has a FOREIGN KEY (snapshot_id) REFERENCES
-- snapshots(id) ON DELETE CASCADE, so deleting all snapshots removes every
-- snapshot-scoped row across data, third_parties, assets, risks, findings,
-- obligations, processing_activities, statements_of_applicability,
-- applicability_statements, the third_party_* sub-tables, the processing_activity
-- DPIA/TIA tables, and the controls_snapshots junction.
DELETE FROM snapshots;
