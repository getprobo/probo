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

ALTER TABLE data ADD COLUMN snapshot_id TEXT;
ALTER TABLE data ADD COLUMN source_id TEXT;

ALTER TABLE data ADD CONSTRAINT data_snapshot_id_fkey
    FOREIGN KEY (snapshot_id)
    REFERENCES snapshots(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE;

ALTER TABLE data ADD CONSTRAINT data_source_id_snapshot_id_key
    UNIQUE (source_id, snapshot_id);

ALTER TABLE vendors ADD COLUMN snapshot_id TEXT;
ALTER TABLE vendors ADD COLUMN source_id TEXT;

ALTER TABLE vendors ADD CONSTRAINT vendors_snapshot_id_fkey
    FOREIGN KEY (snapshot_id)
    REFERENCES snapshots(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE;

ALTER TABLE vendors ADD CONSTRAINT vendors_source_id_snapshot_id_key
    UNIQUE (source_id, snapshot_id);

ALTER TABLE data_vendors ADD COLUMN snapshot_id TEXT;

ALTER TABLE data_vendors ADD CONSTRAINT data_vendors_snapshot_id_fkey
    FOREIGN KEY (snapshot_id)
    REFERENCES snapshots(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE;
