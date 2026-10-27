// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

type (
	InsertRequest struct {
		OrganizationID     gid.GID
		RecipientProfileID gid.GID
		Type               coredata.NotificationType
		SubjectID          gid.GID
		Title              string
	}
)

type (
	notificationAttributes struct {
		Title string `json:"title"`
	}
)

func Insert(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	req InsertRequest,
) error {
	if !req.Type.IsValid() {
		return fmt.Errorf("cannot insert notification: invalid type %q", req.Type)
	}

	attributes, err := json.Marshal(notificationAttributes{Title: req.Title})
	if err != nil {
		return fmt.Errorf("cannot encode notification attributes: %w", err)
	}

	now := time.Now()
	notification := &coredata.Notification{
		ID:                 gid.New(scope.GetTenantID(), coredata.NotificationEntityType),
		OrganizationID:     req.OrganizationID,
		RecipientProfileID: req.RecipientProfileID,
		Type:               req.Type,
		SubjectID:          req.SubjectID,
		Attributes:         attributes,
		Status:             coredata.NotificationStatusPending,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if err := notification.Insert(ctx, tx, scope); err != nil {
		return err
	}

	return nil
}

func InsertTaskAssignment(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	organizationID gid.GID,
	taskID gid.GID,
	taskName string,
	assigneeID gid.GID,
	actorProfileID *gid.GID,
) error {
	if actorProfileID != nil && *actorProfileID == assigneeID {
		return nil
	}

	if err := Insert(
		ctx,
		tx,
		scope,
		InsertRequest{
			OrganizationID:     organizationID,
			RecipientProfileID: assigneeID,
			Type:               coredata.NotificationTypeTaskAssigned,
			SubjectID:          taskID,
			Title:              taskName,
		},
	); err != nil {
		return fmt.Errorf("cannot insert task assignment notification: %w", err)
	}

	return nil
}

// InsertTaskMentions notifies each mentioned profile that belongs to the
// organization and is not deactivated. The actor is never notified.
func InsertTaskMentions(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	organizationID gid.GID,
	taskID gid.GID,
	taskName string,
	mentionedIDs []gid.GID,
	actorProfileID *gid.GID,
) error {
	recipientIDs := make([]gid.GID, 0, len(mentionedIDs))
	for _, id := range mentionedIDs {
		if actorProfileID != nil && *actorProfileID == id {
			continue
		}

		recipientIDs = append(recipientIDs, id)
	}

	if len(recipientIDs) == 0 {
		return nil
	}

	var profiles coredata.MembershipProfiles
	if err := profiles.LoadExistingByIDs(ctx, tx, scope, recipientIDs); err != nil {
		return fmt.Errorf("cannot load mentioned profiles: %w", err)
	}

	for _, profile := range profiles {
		if profile.OrganizationID != organizationID || profile.State == coredata.ProfileStateDeactivated {
			continue
		}

		if err := Insert(
			ctx,
			tx,
			scope,
			InsertRequest{
				OrganizationID:     organizationID,
				RecipientProfileID: profile.ID,
				Type:               coredata.NotificationTypeTaskMentioned,
				SubjectID:          taskID,
				Title:              taskName,
			},
		); err != nil {
			return fmt.Errorf("cannot insert task mention notification: %w", err)
		}
	}

	return nil
}
