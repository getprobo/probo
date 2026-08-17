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
	"fmt"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/probot"
	slackchannel "go.probo.inc/probo/pkg/probot/channel/slack"
)

func applySlackbotNotificationChannel(
	ctx context.Context,
	destinations BotDeliveryDestinations,
	messages ComplianceMessages,
	scope coredata.Scoper,
	organizationID gid.GID,
	compliancePortalID gid.GID,
	channelID string,
) (*coredata.BotDeliveryDestination, error) {
	target := probot.DeliveryTarget{
		Namespace: "compliance_portal",
		Key:       compliancePortalID.String(),
	}

	previous, err := destinations.GetDestination(ctx, scope, organizationID, target)
	if errors.Is(err, slackchannel.ErrSlackbotChannelNotFound) {
		previous = nil
		err = nil
	}
	if err != nil {
		return nil, fmt.Errorf("cannot load Slackbot notification channel: %w", err)
	}

	destination, err := destinations.SetDestination(
		ctx,
		scope,
		organizationID,
		target,
		channelID,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot set Slackbot notification channel: %w", err)
	}

	if err := messages.QueueWelcome(ctx, organizationID, compliancePortalID); err != nil {
		if restoreErr := restoreSlackbotNotificationChannel(
			ctx,
			destinations,
			scope,
			organizationID,
			target,
			previous,
		); restoreErr != nil {
			return nil, fmt.Errorf(
				"cannot queue Slackbot welcome: %w: %w",
				err,
				restoreErr,
			)
		}

		return nil, fmt.Errorf("cannot queue Slackbot welcome: %w", err)
	}

	return destination, nil
}

func restoreSlackbotNotificationChannel(
	ctx context.Context,
	destinations BotDeliveryDestinations,
	scope coredata.Scoper,
	organizationID gid.GID,
	target probot.DeliveryTarget,
	previous *coredata.BotDeliveryDestination,
) error {
	if previous == nil {
		if err := destinations.ClearDestination(ctx, scope, organizationID, target); err != nil {
			return fmt.Errorf("cannot clear Slackbot notification channel: %w", err)
		}

		return nil
	}

	if _, err := destinations.SetDestination(
		ctx,
		scope,
		organizationID,
		target,
		previous.ExternalDestinationID,
	); err != nil {
		return fmt.Errorf("cannot restore Slackbot notification channel: %w", err)
	}

	return nil
}
