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

package slack

import (
	"context"
	"errors"
	"fmt"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/probot"
)

func (s *MessageService) GetInitialByChannelAndTS(
	ctx context.Context,
	organizationID gid.GID,
	channelID string,
	messageTS string,
) (*probot.DeliveredMessage, error) {
	if organizationID != gid.Nil {
		scope := coredata.NewScopeFromObjectID(organizationID)
		if s.modern != nil {
			message, err := s.modern.GetInitialByOrganizationIDChannelAndTS(
				ctx,
				scope,
				organizationID,
				channelID,
				messageTS,
			)
			if err == nil {
				return modernMessage(message), nil
			}
			if !errors.Is(err, coredata.ErrResourceNotFound) {
				return nil, fmt.Errorf("cannot load Slack message: %w", err)
			}
		}

		if s.legacy == nil {
			return nil, coredata.ErrResourceNotFound
		}
		message, err := s.legacy.GetInitialByOrganizationIDChannelAndTS(
			ctx,
			scope,
			organizationID,
			channelID,
			messageTS,
		)
		if err != nil {
			return nil, fmt.Errorf("cannot load legacy Slack message: %w", err)
		}

		return legacyMessage(message), nil
	}

	if s.legacy == nil {
		return nil, coredata.ErrResourceNotFound
	}
	if s.logger != nil {
		s.logger.WarnCtx(
			ctx,
			"falling back to unscoped legacy Slack message lookup",
			log.String("backend", BackendLegacySlack),
		)
	}

	message, err := s.legacy.GetInitialByChannelAndTS(
		ctx,
		coredata.NewNoScope(),
		channelID,
		messageTS,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load legacy Slack message: %w", err)
	}

	return legacyMessage(message), nil
}

func (s *MessageService) GetInitialMessage(
	ctx context.Context,
	organizationID gid.GID,
	anchor probot.MessageAnchor,
) (*probot.DeliveredMessage, error) {
	return s.GetInitialByChannelAndTS(
		ctx,
		organizationID,
		anchor.ConversationID,
		anchor.MessageID,
	)
}

func (s *MessageService) GetMessage(
	ctx context.Context,
	scope coredata.Scoper,
	messageID gid.GID,
) (*probot.DeliveredMessage, error) {
	switch messageID.EntityType() {
	case coredata.SlackbotMessageEntityType:
		message, err := s.modern.GetByID(ctx, scope, messageID)
		if err != nil {
			return nil, err
		}
		return modernMessage(message), nil
	case coredata.SlackMessageEntityType:
		message, err := s.legacy.GetByID(ctx, scope, messageID)
		if err != nil {
			return nil, err
		}
		return legacyMessage(message), nil
	default:
		return nil, fmt.Errorf("unsupported message entity type %d", messageID.EntityType())
	}
}

func (s *MessageService) UpdateMessage(
	ctx context.Context,
	messageID gid.GID,
	message probot.Message,
	intent probot.MessageIntent,
) error {
	scope := coredata.NewScopeFromObjectID(message.OrganizationID)
	delivered, err := s.GetMessage(ctx, scope, messageID)
	if err != nil {
		return fmt.Errorf("cannot load message for update: %w", err)
	}

	if delivered.Backend == BackendSlackbot {
		_, err = s.modern.QueueRevision(
			ctx,
			scope,
			messageID,
			RenderMessageIntent(intent),
			cloneNotificationData(message.Attributes),
		)
		if err != nil {
			return fmt.Errorf("cannot queue Slack message revision: %w", err)
		}
		s.logNotificationRoute(
			ctx,
			BackendSlackbot,
			message.OrganizationID,
			message.Type,
			"",
		)
		return nil
	}

	_, err = s.legacy.UpdateViaResponseURL(
		ctx,
		scope,
		messageID,
		"",
		RenderMessageIntent(intent),
		cloneNotificationData(message.Attributes),
	)
	if err != nil {
		return fmt.Errorf("cannot update legacy Slack message: %w", err)
	}

	s.logNotificationRoute(
		ctx,
		BackendLegacySlack,
		message.OrganizationID,
		message.Type,
		"",
	)

	return nil
}

func modernMessage(message *coredata.SlackbotMessage) *probot.DeliveredMessage {
	return &probot.DeliveredMessage{
		Message: probot.Message{
			ID:             message.ID,
			OrganizationID: message.OrganizationID,
			Type:           message.MessageType,
			Attributes:     cloneNotificationData(message.Metadata),
		},
		CreatedAt: message.CreatedAt,
		Backend:   BackendSlackbot,
	}
}

func legacyMessage(message *coredata.SlackMessage) *probot.DeliveredMessage {
	attributes := cloneNotificationData(message.Metadata)
	if message.RequesterEmail != nil {
		attributes["requester_email"] = message.RequesterEmail.String()
	}

	messageType := message.Type.String()
	if original, ok := attributes[deliveryMessageTypeMetadata].(string); ok && original != "" {
		messageType = original
	}

	return &probot.DeliveredMessage{
		Message: probot.Message{
			ID:             message.ID,
			OrganizationID: message.OrganizationID,
			Type:           messageType,
			Attributes:     attributes,
		},
		CreatedAt: message.CreatedAt,
		Backend:   BackendLegacySlack,
	}
}
