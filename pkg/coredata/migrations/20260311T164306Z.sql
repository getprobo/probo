-- Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
