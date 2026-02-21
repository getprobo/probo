// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package coredata

import (
	"context"
	"fmt"
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type (
	MeetingAttendee struct {
		MeetingID      gid.GID   `db:"meeting_id"`
		OrganizationID gid.GID   `db:"organization_id"`
		AttendeeID     gid.GID   `db:"attendee_profile_id"`
		CreatedAt      time.Time `db:"created_at"`
	}

	MeetingAttendees []*MeetingAttendee
)

func (ma *MeetingAttendees) LoadByMeetingID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	meetingID gid.GID,
) error {
	q := `
SELECT
    meeting_id,
    attendee_profile_id,
    organization_id,
    created_at
FROM
    meeting_attendees
WHERE
    %s
    AND meeting_id = @meeting_id
ORDER BY
    created_at ASC
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{"meeting_id": meetingID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query meeting attendees: %w", err)
	}

	attendees, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[MeetingAttendee])
	if err != nil {
		return fmt.Errorf("cannot collect meeting attendees: %w", err)
	}

	*ma = attendees
	return nil
}

func (ma *MeetingAttendees) Merge(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	meetingID gid.GID,
	organizationID gid.GID,
	attendeeIDs []gid.GID,
) error {
	q := `
WITH attendee_ids AS (
	SELECT
		unnest(@attendee_ids::text[]) AS attendee_profile_id,
		@tenant_id AS tenant_id,
		@meeting_id AS meeting_id,
		@organization_id AS organization_id,
		@created_at::timestamptz AS created_at
)
MERGE INTO meeting_attendees AS tgt
USING attendee_ids AS src
ON tgt.tenant_id = src.tenant_id
	AND tgt.meeting_id = src.meeting_id
	AND tgt.attendee_profile_id = src.attendee_profile_id
WHEN NOT MATCHED THEN
	INSERT (tenant_id, meeting_id, attendee_profile_id, organization_id, created_at)
	VALUES (src.tenant_id, src.meeting_id, src.attendee_profile_id, src.organization_id, src.created_at)
WHEN NOT MATCHED BY SOURCE
	AND tgt.tenant_id = @tenant_id AND tgt.meeting_id = @meeting_id
	THEN DELETE
`

	args := pgx.StrictNamedArgs{
		"tenant_id":       scope.GetTenantID(),
		"meeting_id":      meetingID,
		"organization_id": organizationID,
		"created_at":      time.Now(),
		"attendee_ids":    attendeeIDs,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot merge meeting attendees: %w", err)
	}

	return nil
}
