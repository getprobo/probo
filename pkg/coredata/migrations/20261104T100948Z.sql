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

-- Per-visitor NDA exception. nda_bypassed exempts the visitor from signing.
-- recorded_nda_* stores a copy signed in another system. The pair constraint
-- keeps a recorded file and its signature date together.

ALTER TABLE cp_accesses
    ADD COLUMN nda_bypassed BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE cp_accesses
    ALTER COLUMN nda_bypassed DROP DEFAULT;

ALTER TABLE cp_accesses
    ADD COLUMN recorded_nda_file_id TEXT,
    ADD COLUMN recorded_nda_signed_at TIMESTAMP WITH TIME ZONE;

ALTER TABLE cp_accesses
    ADD CONSTRAINT cp_accesses_recorded_nda_file_id_fkey
    FOREIGN KEY (recorded_nda_file_id)
    REFERENCES files(id)
    ON UPDATE CASCADE
    ON DELETE RESTRICT;

ALTER TABLE cp_accesses
    ADD CONSTRAINT cp_accesses_recorded_nda_pair_check
    CHECK (
        (
            recorded_nda_file_id IS NULL
            AND recorded_nda_signed_at IS NULL
        )
        OR (
            recorded_nda_file_id IS NOT NULL
            AND recorded_nda_signed_at IS NOT NULL
        )
    );
