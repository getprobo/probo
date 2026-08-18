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

package console_v1

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/probot"
	slackchannel "go.probo.inc/probo/pkg/probot/channel/slack"
)

type (
	recordingDestinations struct {
		current      *coredata.BotDeliveryDestination
		setCalls     []string
		restoreCalls []restoreCall
		clearCalls   int
		onRestore    func(expectedExternalDestinationID string)
	}

	restoreCall struct {
		previous                      *coredata.BotDeliveryDestination
		expectedExternalDestinationID string
	}

	failingWelcomeQueue struct{}
)

func (d *recordingDestinations) GetDestination(
	_ context.Context,
	_ coredata.Scoper,
	_ gid.GID,
	_ probot.DeliveryTarget,
) (*coredata.BotDeliveryDestination, error) {
	if d.current == nil {
		return nil, slackchannel.ErrSlackbotChannelNotFound
	}

	return d.current, nil
}

func (d *recordingDestinations) SetDestination(
	_ context.Context,
	_ coredata.Scoper,
	_ gid.GID,
	_ probot.DeliveryTarget,
	externalDestinationID string,
) (*coredata.BotDeliveryDestination, error) {
	d.setCalls = append(d.setCalls, externalDestinationID)
	d.current = &coredata.BotDeliveryDestination{
		ExternalDestinationID: externalDestinationID,
		ExternalName:          "channel-" + externalDestinationID,
	}

	return d.current, nil
}

func (d *recordingDestinations) RestoreDestination(
	_ context.Context,
	_ coredata.Scoper,
	_ gid.GID,
	_ probot.DeliveryTarget,
	previous *coredata.BotDeliveryDestination,
	expectedExternalDestinationID string,
) (*coredata.BotDeliveryDestination, error) {
	d.restoreCalls = append(
		d.restoreCalls,
		restoreCall{
			previous:                      previous,
			expectedExternalDestinationID: expectedExternalDestinationID,
		},
	)

	if d.onRestore != nil {
		d.onRestore(expectedExternalDestinationID)
	}

	// CAS: only restore/clear when current still matches what this operation wrote.
	if d.current == nil || d.current.ExternalDestinationID != expectedExternalDestinationID {
		return nil, nil
	}

	if previous == nil {
		d.current = nil

		return nil, nil
	}

	d.current = &coredata.BotDeliveryDestination{
		ExternalDestinationID: previous.ExternalDestinationID,
		ExternalName:          previous.ExternalName,
		VerifiedAt:            previous.VerifiedAt,
	}

	return d.current, nil
}

func (d *recordingDestinations) ClearDestination(
	_ context.Context,
	_ coredata.Scoper,
	_ gid.GID,
	_ probot.DeliveryTarget,
) error {
	d.clearCalls++
	d.current = nil

	return nil
}

func (failingWelcomeQueue) QueueWelcome(
	_ context.Context,
	_ gid.GID,
	_ gid.GID,
) error {
	return errors.New("welcome queue unavailable")
}

func TestApplySlackbotNotificationChannel_RestoresPriorChannel(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	compliancePortalID := gid.New(tenantID, coredata.CompliancePortalEntityType)
	verifiedAt := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	destinations := &recordingDestinations{
		current: &coredata.BotDeliveryDestination{
			ExternalDestinationID: "C-old",
			ExternalName:          "old-channel",
			VerifiedAt:            &verifiedAt,
		},
	}

	_, err := applySlackbotNotificationChannel(
		t.Context(),
		destinations,
		failingWelcomeQueue{},
		coredata.NewScope(tenantID),
		organizationID,
		compliancePortalID,
		"C-new",
	)
	require.Error(t, err)
	assert.ErrorContains(t, err, "cannot queue Slackbot welcome")
	assert.Equal(t, []string{"C-new"}, destinations.setCalls)
	require.Len(t, destinations.restoreCalls, 1)
	assert.Equal(t, "C-new", destinations.restoreCalls[0].expectedExternalDestinationID)
	require.NotNil(t, destinations.restoreCalls[0].previous)
	assert.Equal(t, "C-old", destinations.restoreCalls[0].previous.ExternalDestinationID)
	require.NotNil(t, destinations.restoreCalls[0].previous.VerifiedAt)
	assert.True(t, destinations.restoreCalls[0].previous.VerifiedAt.Equal(verifiedAt))
	assert.Equal(t, 0, destinations.clearCalls)
	require.NotNil(t, destinations.current.VerifiedAt)
	assert.True(t, destinations.current.VerifiedAt.Equal(verifiedAt))
	assert.Equal(t, "C-old", destinations.current.ExternalDestinationID)
}

func TestApplySlackbotNotificationChannel_ClearsWhenNoPreviousChannel(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	compliancePortalID := gid.New(tenantID, coredata.CompliancePortalEntityType)
	destinations := &recordingDestinations{}

	_, err := applySlackbotNotificationChannel(
		t.Context(),
		destinations,
		failingWelcomeQueue{},
		coredata.NewScope(tenantID),
		organizationID,
		compliancePortalID,
		"C-new",
	)
	require.Error(t, err)
	assert.ErrorContains(t, err, "cannot queue Slackbot welcome")
	assert.Equal(t, []string{"C-new"}, destinations.setCalls)
	require.Len(t, destinations.restoreCalls, 1)
	assert.Nil(t, destinations.restoreCalls[0].previous)
	assert.Equal(t, "C-new", destinations.restoreCalls[0].expectedExternalDestinationID)
	assert.Equal(t, 0, destinations.clearCalls)
	assert.Nil(t, destinations.current)
}

func TestApplySlackbotNotificationChannel_SkipsRestoreWhenDestinationChanged(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	compliancePortalID := gid.New(tenantID, coredata.CompliancePortalEntityType)
	destinations := &recordingDestinations{
		current: &coredata.BotDeliveryDestination{
			ExternalDestinationID: "C-old",
			ExternalName:          "old-channel",
		},
	}

	messages := welcomeQueueFunc(func(
		_ context.Context,
		_ gid.GID,
		_ gid.GID,
	) error {
		destinations.current = &coredata.BotDeliveryDestination{
			ExternalDestinationID: "C-other",
			ExternalName:          "other-channel",
		}

		return errors.New("welcome queue unavailable")
	})

	_, err := applySlackbotNotificationChannel(
		t.Context(),
		destinations,
		messages,
		coredata.NewScope(tenantID),
		organizationID,
		compliancePortalID,
		"C-new",
	)
	require.Error(t, err)
	assert.ErrorContains(t, err, "cannot queue Slackbot welcome")
	assert.Equal(t, []string{"C-new"}, destinations.setCalls)
	require.Len(t, destinations.restoreCalls, 1)
	assert.Equal(t, "C-new", destinations.restoreCalls[0].expectedExternalDestinationID)
	assert.Equal(t, 0, destinations.clearCalls)
	assert.Equal(t, "C-other", destinations.current.ExternalDestinationID)
}

func TestApplySlackbotNotificationChannel_SkipsClearWhenDestinationChangedBeforeCAS(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	compliancePortalID := gid.New(tenantID, coredata.CompliancePortalEntityType)
	destinations := &recordingDestinations{}
	destinations.onRestore = func(expectedExternalDestinationID string) {
		// Simulate a concurrent SetDestination after the failed welcome and
		// before the conditional clear — CAS must leave C-other alone.
		_ = expectedExternalDestinationID
		destinations.current = &coredata.BotDeliveryDestination{
			ExternalDestinationID: "C-other",
			ExternalName:          "other-channel",
		}
	}

	_, err := applySlackbotNotificationChannel(
		t.Context(),
		destinations,
		failingWelcomeQueue{},
		coredata.NewScope(tenantID),
		organizationID,
		compliancePortalID,
		"C-new",
	)
	require.Error(t, err)
	assert.ErrorContains(t, err, "cannot queue Slackbot welcome")
	require.Len(t, destinations.restoreCalls, 1)
	assert.Nil(t, destinations.restoreCalls[0].previous)
	assert.Equal(t, "C-new", destinations.restoreCalls[0].expectedExternalDestinationID)
	require.NotNil(t, destinations.current)
	assert.Equal(t, "C-other", destinations.current.ExternalDestinationID)
}

func TestApplySlackbotNotificationChannel_SkipsRestoreWhenDestinationChangedBeforeCAS(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	compliancePortalID := gid.New(tenantID, coredata.CompliancePortalEntityType)
	verifiedAt := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	destinations := &recordingDestinations{
		current: &coredata.BotDeliveryDestination{
			ExternalDestinationID: "C-old",
			ExternalName:          "old-channel",
			VerifiedAt:            &verifiedAt,
		},
	}
	destinations.onRestore = func(expectedExternalDestinationID string) {
		_ = expectedExternalDestinationID
		destinations.current = &coredata.BotDeliveryDestination{
			ExternalDestinationID: "C-other",
			ExternalName:          "other-channel",
		}
	}

	_, err := applySlackbotNotificationChannel(
		t.Context(),
		destinations,
		failingWelcomeQueue{},
		coredata.NewScope(tenantID),
		organizationID,
		compliancePortalID,
		"C-new",
	)
	require.Error(t, err)
	assert.ErrorContains(t, err, "cannot queue Slackbot welcome")
	require.Len(t, destinations.restoreCalls, 1)
	require.NotNil(t, destinations.restoreCalls[0].previous)
	assert.Equal(t, "C-old", destinations.restoreCalls[0].previous.ExternalDestinationID)
	assert.Equal(t, "C-new", destinations.restoreCalls[0].expectedExternalDestinationID)
	require.NotNil(t, destinations.current)
	assert.Equal(t, "C-other", destinations.current.ExternalDestinationID)
}

type welcomeQueueFunc func(context.Context, gid.GID, gid.GID) error

func (f welcomeQueueFunc) QueueWelcome(
	ctx context.Context,
	organizationID gid.GID,
	compliancePortalID gid.GID,
) error {
	return f(ctx, organizationID, compliancePortalID)
}
