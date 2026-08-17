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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/probot"
	slackchannel "go.probo.inc/probo/pkg/probot/channel/slack"
)

type (
	recordingDestinations struct {
		previous   *coredata.BotDeliveryDestination
		setCalls   []string
		clearCalls int
	}

	failingWelcomeQueue struct{}
)

func (d *recordingDestinations) GetDestination(
	_ context.Context,
	_ coredata.Scoper,
	_ gid.GID,
	_ probot.DeliveryTarget,
) (*coredata.BotDeliveryDestination, error) {
	if d.previous == nil {
		return nil, slackchannel.ErrSlackbotChannelNotFound
	}

	return d.previous, nil
}

func (d *recordingDestinations) SetDestination(
	_ context.Context,
	_ coredata.Scoper,
	_ gid.GID,
	_ probot.DeliveryTarget,
	externalDestinationID string,
) (*coredata.BotDeliveryDestination, error) {
	d.setCalls = append(d.setCalls, externalDestinationID)

	return &coredata.BotDeliveryDestination{
		ExternalDestinationID: externalDestinationID,
		ExternalName:          "channel-" + externalDestinationID,
	}, nil
}

func (d *recordingDestinations) ClearDestination(
	_ context.Context,
	_ coredata.Scoper,
	_ gid.GID,
	_ probot.DeliveryTarget,
) error {
	d.clearCalls++

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
	destinations := &recordingDestinations{
		previous: &coredata.BotDeliveryDestination{
			ExternalDestinationID: "C-old",
			ExternalName:          "old-channel",
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
	assert.Equal(t, []string{"C-new", "C-old"}, destinations.setCalls)
	assert.Equal(t, 0, destinations.clearCalls)
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
	assert.Equal(t, 1, destinations.clearCalls)
}
