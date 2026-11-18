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

ALTER TABLE
    assets DROP COLUMN owner_id;

ALTER TABLE
    continual_improvements DROP COLUMN owner_id;

ALTER TABLE
    data DROP COLUMN owner_id;

ALTER TABLE
    document_versions DROP COLUMN owner_id;

ALTER TABLE
    meeting_attendees DROP COLUMN attendee_id;

ALTER TABLE
    nonconformities DROP COLUMN owner_id;

ALTER TABLE
    obligations DROP COLUMN owner_id;

ALTER TABLE
    documents DROP COLUMN owner_id;

ALTER TABLE
    document_version_signatures DROP COLUMN signed_by;

ALTER TABLE
    processing_activities DROP COLUMN data_protection_officer_id;

ALTER TABLE
    risks DROP COLUMN owner_id;

ALTER TABLE
    states_of_applicability DROP COLUMN owner_id;

ALTER TABLE
    tasks DROP COLUMN assigned_to;

ALTER TABLE
    vendors DROP COLUMN business_owner_id,
    DROP COLUMN security_owner_id;

DROP TABLE peoples;
