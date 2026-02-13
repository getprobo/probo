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

package probo

import (
	"context"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/validator"
	"go.probo.inc/probo/pkg/webhook"

	webhooktypes "go.probo.inc/probo/pkg/webhook/types"
)

type MeetingService struct {
	svc *TenantService
}

type (
	CreateMeetingRequest struct {
		OrganizationID gid.GID
		Name           string
		Date           time.Time
		AttendeeIDs    []gid.GID
		Minutes        *string
	}

	UpdateMeetingRequest struct {
		MeetingID   gid.GID
		Name        *string
		Date        *time.Time
		AttendeeIDs []gid.GID
		Minutes     **string
	}
)

const (
	MinutesMaxLength = 50_000
)

func (cmr *CreateMeetingRequest) Validate() error {
	v := validator.New()

	v.Check(cmr.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(cmr.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(cmr.Date, "date", validator.Required())
	v.CheckEach(cmr.AttendeeIDs, "attendee_ids", func(index int, item any) {
		v.Check(item, fmt.Sprintf("attendee_ids[%d]", index), validator.Required(), validator.GID(coredata.MembershipProfileEntityType))
	})
	v.Check(cmr.Minutes, "minutes", validator.SafeText(MinutesMaxLength))

	return v.Error()
}

func (umr *UpdateMeetingRequest) Validate() error {
	v := validator.New()

	v.Check(umr.MeetingID, "meeting_id", validator.Required(), validator.GID(coredata.MeetingEntityType))
	v.Check(umr.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.CheckEach(umr.AttendeeIDs, "attendee_ids", func(index int, item any) {
		v.Check(item, fmt.Sprintf("attendee_ids[%d]", index), validator.Required(), validator.GID(coredata.MembershipProfileEntityType))
	})
	v.Check(umr.Minutes, "minutes", validator.SafeText(MinutesMaxLength))

	return v.Error()
}

func (s MeetingService) ListForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.MeetingOrderField],
) (*page.Page[*coredata.Meeting, coredata.MeetingOrderField], error) {
	var meetings coredata.Meetings
	organization := &coredata.Organization{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := organization.LoadByID(ctx, conn, s.svc.scope, organizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			err := meetings.LoadByOrganizationID(
				ctx,
				conn,
				s.svc.scope,
				organization.ID,
				cursor,
			)
			if err != nil {
				return fmt.Errorf("cannot load meetings: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(meetings, cursor), nil
}

func (s MeetingService) CountForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			meetings := &coredata.Meetings{}
			count, err = meetings.CountByOrganizationID(ctx, conn, s.svc.scope, organizationID)
			if err != nil {
				return fmt.Errorf("cannot count meetings: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s MeetingService) Get(
	ctx context.Context,
	meetingID gid.GID,
) (*coredata.Meeting, error) {
	meeting := &coredata.Meeting{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return meeting.LoadByID(ctx, conn, s.svc.scope, meetingID)
		},
	)

	if err != nil {
		return nil, err
	}

	return meeting, nil
}

func (s MeetingService) Create(
	ctx context.Context,
	req CreateMeetingRequest,
) (*coredata.Meeting, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	now := time.Now()
	var meeting *coredata.Meeting
	organization := &coredata.Organization{}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := organization.LoadByID(ctx, conn, s.svc.scope, req.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			meeting = &coredata.Meeting{
				ID:             gid.New(organization.ID.TenantID(), coredata.MeetingEntityType),
				OrganizationID: organization.ID,
				Name:           req.Name,
				Date:           req.Date,
				Minutes:        req.Minutes,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			if err := meeting.Insert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert meeting: %w", err)
			}

			if len(req.AttendeeIDs) > 0 {
				var attendeeProfiles coredata.MembershipProfiles
				if err := attendeeProfiles.LoadByIDs(ctx, conn, s.svc.scope, req.AttendeeIDs); err != nil {
					return fmt.Errorf("cannot load attendee profiles: %w", err)
				}

				var attendees coredata.MeetingAttendees
				if err := attendees.Merge(ctx, conn, s.svc.scope, meeting.ID, organization.ID, req.AttendeeIDs); err != nil {
					return fmt.Errorf("cannot merge meeting attendees: %w", err)
				}
			}

			if err := webhook.InsertData(ctx, conn, s.svc.scope, organization.ID, coredata.WebhookEventTypeMeetingCreated, webhooktypes.NewMeeting(meeting)); err != nil {
				return fmt.Errorf("cannot insert webhook event: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return meeting, nil
}

func (s MeetingService) GetAttendees(
	ctx context.Context,
	meetingID gid.GID,
) (coredata.MembershipProfiles, error) {
	var attendees coredata.MembershipProfiles

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return attendees.LoadByMeetingID(ctx, conn, s.svc.scope, meetingID)
		},
	)

	if err != nil {
		return nil, err
	}

	return attendees, nil
}

func (s MeetingService) Update(
	ctx context.Context,
	req UpdateMeetingRequest,
) (*coredata.Meeting, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	meeting := &coredata.Meeting{}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := meeting.LoadByID(ctx, conn, s.svc.scope, req.MeetingID); err != nil {
				return fmt.Errorf("cannot load meeting: %w", err)
			}

			if req.Name != nil {
				meeting.Name = *req.Name
			}
			if req.Date != nil {
				meeting.Date = *req.Date
			}
			if req.Minutes != nil {
				meeting.Minutes = *req.Minutes
			}

			meeting.UpdatedAt = time.Now()

			if err := meeting.Update(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update meeting: %w", err)
			}

			if req.AttendeeIDs != nil {
				var attendeeProfiles coredata.MembershipProfiles
				if err := attendeeProfiles.LoadByIDs(ctx, conn, s.svc.scope, req.AttendeeIDs); err != nil {
					return fmt.Errorf("cannot load attendee profiles: %w", err)
				}

				var attendees coredata.MeetingAttendees
				if err := attendees.Merge(ctx, conn, s.svc.scope, meeting.ID, meeting.OrganizationID, req.AttendeeIDs); err != nil {
					return fmt.Errorf("cannot merge meeting attendees: %w", err)
				}
			}

			if err := webhook.InsertData(ctx, conn, s.svc.scope, meeting.OrganizationID, coredata.WebhookEventTypeMeetingUpdated, webhooktypes.NewMeeting(meeting)); err != nil {
				return fmt.Errorf("cannot insert webhook event: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return meeting, nil
}

func (s MeetingService) Delete(
	ctx context.Context,
	meetingID gid.GID,
) error {
	meeting := &coredata.Meeting{ID: meetingID}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := meeting.LoadByID(ctx, conn, s.svc.scope, meetingID); err != nil {
				return fmt.Errorf("cannot load meeting: %w", err)
			}

			if err := webhook.InsertData(ctx, conn, s.svc.scope, meeting.OrganizationID, coredata.WebhookEventTypeMeetingDeleted, webhooktypes.NewMeeting(meeting)); err != nil {
				return fmt.Errorf("cannot insert webhook event: %w", err)
			}

			if err := meeting.Delete(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot delete meeting: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return err
	}

	return nil
}
