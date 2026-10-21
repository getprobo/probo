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

package notification_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/notification"
)

// mentionFixture seeds two organizations (in the same tenant) with one
// membership (and matching profile) each, plus a deactivated profile, so
// tests can exercise every branch of InsertTaskMentions against a real
// database instead of relying on the GraphQL API being able to observe
// cross-tenant/cross-organization data (it cannot, by design).
type mentionFixture struct {
	scope                coredata.Scoper
	taskOrganizationID   gid.GID
	otherOrganizationID  gid.GID
	actorProfileID       gid.GID
	mentionedProfileID   gid.GID
	deactivatedProfileID gid.GID
	outsiderProfileID    gid.GID
}

func seedMentionFixture(t *testing.T, ctx context.Context, client *pg.Client) mentionFixture {
	t.Helper()

	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID, gid.Nil)

	taskOrganizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	otherOrganizationID := gid.New(tenantID, coredata.OrganizationEntityType)

	actorProfileID := gid.New(tenantID, coredata.MembershipProfileEntityType)
	mentionedProfileID := gid.New(tenantID, coredata.MembershipProfileEntityType)
	deactivatedProfileID := gid.New(tenantID, coredata.MembershipProfileEntityType)
	outsiderProfileID := gid.New(tenantID, coredata.MembershipProfileEntityType)

	now := time.Now().UTC()

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		organization := coredata.Organization{
			ID:        taskOrganizationID,
			TenantID:  tenantID,
			Name:      "Mention Test Org",
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := organization.Insert(ctx, tx); err != nil {
			return fmt.Errorf("cannot insert organization: %w", err)
		}

		otherOrganization := coredata.Organization{
			ID:        otherOrganizationID,
			TenantID:  tenantID,
			Name:      "Mention Test Other Org",
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := otherOrganization.Insert(ctx, tx); err != nil {
			return fmt.Errorf("cannot insert other organization: %w", err)
		}

		profiles := []struct {
			id             gid.GID
			organizationID gid.GID
			fullName       string
			state          coredata.ProfileState
			role           coredata.MembershipRole
		}{
			{actorProfileID, taskOrganizationID, "Actor", coredata.ProfileStateActive, coredata.MembershipRoleOwner},
			{mentionedProfileID, taskOrganizationID, "Mentioned", coredata.ProfileStateActive, coredata.MembershipRoleOwner},
			{deactivatedProfileID, taskOrganizationID, "Deactivated", coredata.ProfileStateDeactivated, coredata.MembershipRoleOwner},
			{outsiderProfileID, otherOrganizationID, "Outsider", coredata.ProfileStateActive, coredata.MembershipRoleOwner},
		}

		for _, p := range profiles {
			emailAddress, err := mail.ParseAddr(fmt.Sprintf("%s@example.com", p.id))
			if err != nil {
				return fmt.Errorf("cannot parse email address: %w", err)
			}

			identity := coredata.Identity{
				ID:                   gid.New(gid.NilTenant, coredata.IdentityEntityType),
				EmailAddress:         emailAddress,
				FullName:             p.fullName,
				EmailAddressVerified: true,
				CreatedAt:            now,
				UpdatedAt:            now,
			}
			if err := identity.Insert(ctx, tx); err != nil {
				return fmt.Errorf("cannot insert identity: %w", err)
			}

			membership := coredata.Membership{
				ID:             gid.New(tenantID, coredata.MembershipEntityType),
				IdentityID:     identity.ID,
				OrganizationID: p.organizationID,
				Role:           p.role,
				CreatedAt:      now,
				UpdatedAt:      now,
			}
			if err := membership.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert membership: %w", err)
			}

			profile := coredata.MembershipProfile{
				ID:             p.id,
				IdentityID:     identity.ID,
				OrganizationID: p.organizationID,
				Source:         coredata.ProfileSourceManual,
				State:          p.state,
				FullName:       p.fullName,
				CreatedAt:      now,
				UpdatedAt:      now,
			}
			if err := profile.Insert(ctx, tx); err != nil {
				return fmt.Errorf("cannot insert membership profile: %w", err)
			}
		}

		return nil
	}))

	return mentionFixture{
		scope:                scope,
		taskOrganizationID:   taskOrganizationID,
		otherOrganizationID:  otherOrganizationID,
		actorProfileID:       actorProfileID,
		mentionedProfileID:   mentionedProfileID,
		deactivatedProfileID: deactivatedProfileID,
		outsiderProfileID:    outsiderProfileID,
	}
}

func notificationCount(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	scope coredata.Scoper,
	organizationID gid.GID,
	recipientProfileID gid.GID,
) int {
	t.Helper()

	var count int

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		var notifications coredata.Notifications

		n, err := notifications.CountByRecipientProfileID(
			ctx,
			conn,
			scope,
			organizationID,
			recipientProfileID,
			coredata.NewNotificationFilter(nil, nil),
		)
		if err != nil {
			return err
		}

		count = n

		return nil
	}))

	return count
}

func TestInsertTaskMentions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := test.PGClient(t)

	t.Run("notifies a mentioned member once", func(t *testing.T) {
		t.Parallel()

		fixture := seedMentionFixture(t, ctx, client)
		taskID := gid.New(fixture.taskOrganizationID.TenantID(), coredata.TaskEntityType)

		require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
			return notification.InsertTaskMentions(
				ctx,
				tx,
				fixture.scope,
				fixture.taskOrganizationID,
				taskID,
				"Mention task",
				[]gid.GID{fixture.mentionedProfileID},
				&fixture.actorProfileID,
			)
		}))

		assert.Equal(
			t,
			1,
			notificationCount(t, ctx, client, fixture.scope, fixture.taskOrganizationID, fixture.mentionedProfileID),
		)
	})

	t.Run("does not notify the actor who mentioned themselves", func(t *testing.T) {
		t.Parallel()

		fixture := seedMentionFixture(t, ctx, client)
		taskID := gid.New(fixture.taskOrganizationID.TenantID(), coredata.TaskEntityType)

		require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
			return notification.InsertTaskMentions(
				ctx,
				tx,
				fixture.scope,
				fixture.taskOrganizationID,
				taskID,
				"Self mention task",
				[]gid.GID{fixture.actorProfileID},
				&fixture.actorProfileID,
			)
		}))

		assert.Equal(
			t,
			0,
			notificationCount(t, ctx, client, fixture.scope, fixture.taskOrganizationID, fixture.actorProfileID),
		)
	})

	t.Run("does not notify a deactivated profile", func(t *testing.T) {
		t.Parallel()

		fixture := seedMentionFixture(t, ctx, client)
		taskID := gid.New(fixture.taskOrganizationID.TenantID(), coredata.TaskEntityType)

		require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
			return notification.InsertTaskMentions(
				ctx,
				tx,
				fixture.scope,
				fixture.taskOrganizationID,
				taskID,
				"Deactivated mention task",
				[]gid.GID{fixture.deactivatedProfileID},
				&fixture.actorProfileID,
			)
		}))

		assert.Equal(
			t,
			0,
			notificationCount(t, ctx, client, fixture.scope, fixture.taskOrganizationID, fixture.deactivatedProfileID),
		)
	})

	// This is the case the e2e black-box suite cannot exercise: a leaked
	// cross-organization notification would carry organization_id = the
	// task's organization and recipient_profile_id = a profile that only
	// exists in a different organization. No GraphQL caller can ever list
	// notifications for an organization they are not a member of, so the
	// outsider could never observe the leak (or its absence) through the
	// API either way. Asserting directly against the notifications table
	// is the only way to prove InsertTaskMentions actually drops the
	// cross-organization recipient instead of merely happening to be
	// unobservable.
	t.Run("does not notify a profile from another organization", func(t *testing.T) {
		t.Parallel()

		fixture := seedMentionFixture(t, ctx, client)
		taskID := gid.New(fixture.taskOrganizationID.TenantID(), coredata.TaskEntityType)

		require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
			return notification.InsertTaskMentions(
				ctx,
				tx,
				fixture.scope,
				fixture.taskOrganizationID,
				taskID,
				"Cross org mention task",
				[]gid.GID{fixture.outsiderProfileID},
				&fixture.actorProfileID,
			)
		}))

		assert.Equal(
			t,
			0,
			notificationCount(t, ctx, client, fixture.scope, fixture.taskOrganizationID, fixture.outsiderProfileID),
		)
	})
}
