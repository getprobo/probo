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

ALTER TABLE controls DROP CONSTRAINT IF EXISTS controls_framework_id_fkey;
ALTER TABLE controls_mesures DROP CONSTRAINT IF EXISTS controls_mesures_control_id_fkey;
ALTER TABLE controls_policies DROP CONSTRAINT IF EXISTS controls_policies_control_id_fkey;

ALTER TABLE controls_mesures
    ADD CONSTRAINT controls_mesures_control_id_fkey
    FOREIGN KEY (control_id)
    REFERENCES controls(id)
    ON DELETE CASCADE;

ALTER TABLE controls_policies
    ADD CONSTRAINT controls_policies_control_id_fkey
    FOREIGN KEY (control_id)
    REFERENCES controls(id)
    ON DELETE CASCADE;

ALTER TABLE controls
    ADD CONSTRAINT controls_framework_id_fkey
    FOREIGN KEY (framework_id)
    REFERENCES frameworks(id)
    ON DELETE CASCADE;

DELETE FROM controls WHERE framework_id NOT IN (SELECT id FROM frameworks);
