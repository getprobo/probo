-- Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
