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
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/bot"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/probot"
	legacyslack "go.probo.inc/probo/pkg/slack"
)

const (
	deliveryTargetNamespaceMetadata  = "delivery_target_namespace"
	deliveryTargetKeyMetadata        = "delivery_target_key"
	deliveryMessageTypeMetadata      = "probot_message_type"
	deliverySourceEventMetadata      = coredata.SlackbotMessageMetadataSourceEventID
	deliveryAgentExecutionMetadata   = "agent_execution_id"
	deliverySubjectNamespaceMetadata = "bot_subject_namespace"
	deliverySubjectKeyMetadata       = "bot_subject_key"
	deliveryCapabilityMetadata       = "bot_capability"
	BackendSlackbot                  = "slackbot"
	BackendLegacySlack               = "legacy-slack"
)

var (
	ErrNoDeliveryDestination = errors.New("no bot delivery destination configured")
)

type (
	ModernMessageQueue interface {
		Queue(ctx context.Context, scope coredata.Scoper, req QueueNotificationRequest) (*coredata.SlackbotMessage, error)
		QueueRevision(ctx context.Context, scope coredata.Scoper, messageID gid.GID, body, metadata map[string]any) (*coredata.SlackbotMessage, error)
		GetByID(ctx context.Context, scope coredata.Scoper, messageID gid.GID) (*coredata.SlackbotMessage, error)
		GetBySourceEventID(ctx context.Context, scope coredata.Scoper, sourceEventID string) (*coredata.SlackbotMessage, error)
		GetInitialByOrganizationIDChannelAndTS(ctx context.Context, scope coredata.Scoper, organizationID gid.GID, channelID, messageTS string) (*coredata.SlackbotMessage, error)
	}

	LegacyMessageQueue interface {
		Queue(ctx context.Context, scope coredata.Scoper, req legacyslack.QueueRequest) (*coredata.SlackMessage, error)
		GetByID(ctx context.Context, scope coredata.Scoper, messageID gid.GID) (*coredata.SlackMessage, error)
		GetBySourceEventID(ctx context.Context, scope coredata.Scoper, sourceEventID string) (*coredata.SlackMessage, error)
		GetInitialByOrganizationIDChannelAndTS(ctx context.Context, scope coredata.Scoper, organizationID gid.GID, channelID, messageTS string) (*coredata.SlackMessage, error)
		GetInitialByChannelAndTS(ctx context.Context, scope coredata.Scoper, channelID, messageTS string) (*coredata.SlackMessage, error)
		UpdateViaResponseURL(ctx context.Context, scope coredata.Scoper, messageID gid.GID, responseURL string, body, metadata map[string]any) (*coredata.SlackMessage, error)
	}

	MessageService struct {
		pg            *pg.Client
		installations *InstallationService
		modern        ModernMessageQueue
		legacy        LegacyMessageQueue
		logger        *log.Logger
	}
)

var (
	_ probot.MessageDelivery     = (*MessageService)(nil)
	_ DeliverySuccessHook        = (*MessageService)(nil)
	_ InteractiveMessageResolver = (*MessageService)(nil)
)

func NewMessageService(
	pgClient *pg.Client,
	installations *InstallationService,
	modern ModernMessageQueue,
	legacy LegacyMessageQueue,
	logger *log.Logger,
) *MessageService {
	return &MessageService{
		pg:            pgClient,
		installations: installations,
		modern:        modern,
		legacy:        legacy,
		logger:        logger,
	}
}

func (s *MessageService) DeliverOutbound(
	ctx context.Context,
	delivery probot.OutboundDelivery,
) error {
	if delivery.Purpose == coredata.BotMessagePurposeUpdate {
		return s.updateOutbound(ctx, delivery)
	}

	return s.queue(ctx, delivery, true)
}

func (s *MessageService) updateOutbound(
	ctx context.Context,
	delivery probot.OutboundDelivery,
) error {
	scope := coredata.NewScopeFromObjectID(delivery.OrganizationID)

	var subject coredata.BotThreadSubject

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return subject.LoadBySubject(
				ctx,
				conn,
				scope,
				delivery.OrganizationID,
				delivery.SubjectNamespace,
				delivery.SubjectKey,
			)
		},
	)
	if errors.Is(err, coredata.ErrResourceNotFound) {
		return fmt.Errorf("cannot update bot message before the initial post is delivered")
	}

	if err != nil {
		return fmt.Errorf("cannot load bot thread subject: %w", err)
	}

	delivered, err := s.GetInitialByChannelAndTS(
		ctx,
		delivery.OrganizationID,
		subject.ExternalConversationID,
		subject.ExternalMessageID,
	)
	if err != nil {
		return fmt.Errorf("cannot load delivered bot message for update: %w", err)
	}

	message := delivery.Result.Message

	attributes := cloneNotificationData(message.Attributes)
	if delivery.SourceEventID != "" {
		attributes[deliverySourceEventMetadata] = delivery.SourceEventID
	}

	if delivery.AgentExecutionID != gid.Nil {
		attributes[deliveryAgentExecutionMetadata] = delivery.AgentExecutionID.String()
	}

	message.Attributes = attributes

	return s.UpdateMessage(
		ctx,
		delivered.Message.ID,
		message,
		delivery.Result.Intent,
	)
}

func (s *MessageService) DeliverVerification(
	ctx context.Context,
	organizationID gid.GID,
	target probot.DeliveryTarget,
	message bot.Message,
	intent bot.MessageIntent,
) error {
	return s.queue(
		ctx,
		probot.OutboundDelivery{
			OrganizationID: organizationID,
			Purpose:        coredata.BotMessagePurposePost,
			Result: probot.OutboundMessage{
				Message:        message,
				Intent:         intent,
				DeliveryTarget: target,
			},
		},
		false,
	)
}

func (s *MessageService) queue(
	ctx context.Context,
	delivery probot.OutboundDelivery,
	requireVerified bool,
) error {
	result := delivery.Result
	if result.DeliveryTarget.Namespace == "" || result.DeliveryTarget.Key == "" {
		return fmt.Errorf("delivery target is incomplete")
	}

	scope := coredata.NewScopeFromObjectID(delivery.OrganizationID)

	destination, err := s.GetDestination(
		ctx,
		scope,
		delivery.OrganizationID,
		result.DeliveryTarget,
	)
	if err != nil &&
		(!errors.Is(err, ErrSlackbotChannelNotFound) || s.legacy == nil || !requireVerified) {
		return fmt.Errorf("cannot load Slack delivery destination: %w", err)
	}

	if err == nil && requireVerified && destination.VerifiedAt == nil && s.legacy == nil {
		return ErrNoDeliveryDestination
	}

	metadata := cloneNotificationData(result.Message.Attributes)
	metadata[deliveryMessageTypeMetadata] = result.Message.Type
	metadata[deliveryTargetNamespaceMetadata] = result.DeliveryTarget.Namespace

	metadata[deliveryTargetKeyMetadata] = result.DeliveryTarget.Key
	if delivery.SubjectNamespace != "" {
		metadata[deliverySubjectNamespaceMetadata] = delivery.SubjectNamespace
	}

	if delivery.SubjectKey != "" {
		metadata[deliverySubjectKeyMetadata] = delivery.SubjectKey
	}

	if delivery.Capability != "" {
		metadata[deliveryCapabilityMetadata] = delivery.Capability
	}

	var sourceEventID *string

	if delivery.SourceEventID != "" {
		source := delivery.SourceEventID
		sourceEventID = &source
		metadata[deliverySourceEventMetadata] = source
	}

	if delivery.AgentExecutionID != gid.Nil {
		metadata[deliveryAgentExecutionMetadata] = delivery.AgentExecutionID.String()
	}

	if err == nil && (!requireVerified || destination.VerifiedAt != nil) && s.modern != nil {
		return s.queueModern(ctx, scope, delivery, destination, metadata, sourceEventID)
	}

	return s.queueLegacy(ctx, scope, delivery, metadata, sourceEventID)
}

func (s *MessageService) queueModern(
	ctx context.Context,
	scope coredata.Scoper,
	delivery probot.OutboundDelivery,
	destination *coredata.BotDeliveryDestination,
	metadata map[string]any,
	sourceEventID *string,
) error {
	result := delivery.Result

	_, err := s.modern.Queue(
		ctx,
		scope,
		QueueNotificationRequest{
			ID:             result.Message.ID,
			OrganizationID: delivery.OrganizationID,
			ChannelID:      destination.ExternalDestinationID,
			MessageType:    result.Message.Type,
			Body:           RenderMessageIntent(result.Intent),
			Metadata:       metadata,
			SourceEventID:  sourceEventID,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot queue Slack message: %w", err)
	}

	s.logNotificationRoute(
		ctx,
		BackendSlackbot,
		delivery.OrganizationID,
		delivery.MessageType,
		delivery.Capability,
	)

	return nil
}

func (s *MessageService) queueLegacy(
	ctx context.Context,
	scope coredata.Scoper,
	delivery probot.OutboundDelivery,
	metadata map[string]any,
	sourceEventID *string,
) error {
	if s.legacy == nil {
		return ErrNoDeliveryDestination
	}

	result := delivery.Result

	_, err := s.legacy.Queue(
		ctx,
		scope,
		legacyslack.QueueRequest{
			ID:             result.Message.ID,
			OrganizationID: delivery.OrganizationID,
			MessageType:    coredata.SlackMessageTypeWelcome,
			Body:           RenderMessageIntent(result.Intent),
			Metadata:       metadata,
			SourceEventID:  sourceEventID,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot queue legacy Slack message: %w", err)
	}

	s.logNotificationRoute(
		ctx,
		BackendLegacySlack,
		delivery.OrganizationID,
		delivery.MessageType,
		delivery.Capability,
	)

	return nil
}

func (s *MessageService) GetDestination(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	target probot.DeliveryTarget,
) (*coredata.BotDeliveryDestination, error) {
	var destination coredata.BotDeliveryDestination

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return destination.LoadByTarget(
				ctx,
				conn,
				scope,
				organizationID,
				ProviderName,
				target.Namespace,
				target.Key,
			)
		},
	)
	if errors.Is(err, coredata.ErrResourceNotFound) {
		return nil, ErrSlackbotChannelNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("cannot load Slack delivery destination: %w", err)
	}

	return &destination, nil
}

func (s *MessageService) SetDestination(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	target probot.DeliveryTarget,
	externalDestinationID string,
) (*coredata.BotDeliveryDestination, error) {
	client, _, err := s.installations.ClientByOrganizationID(ctx, scope, organizationID)
	if err != nil {
		return nil, fmt.Errorf("cannot load Slack client: %w", err)
	}

	conversation, err := findMemberConversation(ctx, client, externalDestinationID)
	if err != nil {
		return nil, fmt.Errorf("cannot find Slack conversation: %w", err)
	}

	destination := coredata.NewBotDeliveryDestination(
		scope,
		organizationID,
		ProviderName,
		target.Namespace,
		target.Key,
	)
	destination.ExternalDestinationID = conversation.ID
	destination.ExternalName = conversation.Name

	err = s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			_, err := destination.Upsert(ctx, tx, scope)
			if err != nil {
				return fmt.Errorf("cannot upsert Slack delivery destination: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot save Slack delivery destination: %w", err)
	}

	return destination, nil
}

// RestoreDestination writes a previously stored destination, including its
// verification timestamp, only when the current external destination ID still
// equals expectedExternalDestinationID (the value this operation wrote). When
// previous is nil, clears under the same compare-and-swap condition. A CAS miss
// (concurrent replace or clear) is a no-op success.
func (s *MessageService) RestoreDestination(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	target probot.DeliveryTarget,
	previous *coredata.BotDeliveryDestination,
	expectedExternalDestinationID string,
) (*coredata.BotDeliveryDestination, error) {
	if expectedExternalDestinationID == "" {
		return nil, fmt.Errorf("cannot restore Slack delivery destination without expected external destination id")
	}

	if previous == nil {
		if err := s.clearDestinationIf(
			ctx,
			scope,
			organizationID,
			target,
			expectedExternalDestinationID,
		); err != nil {
			return nil, err
		}

		return nil, nil
	}

	destination := coredata.NewBotDeliveryDestination(
		scope,
		organizationID,
		ProviderName,
		target.Namespace,
		target.Key,
	)
	destination.ExternalDestinationID = previous.ExternalDestinationID
	destination.ExternalName = previous.ExternalName
	destination.VerifiedAt = previous.VerifiedAt
	destination.UpdatedAt = time.Now()

	updated := false

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			err := destination.UpdateIfExternalDestinationID(
				ctx,
				tx,
				scope,
				expectedExternalDestinationID,
			)
			if errors.Is(err, coredata.ErrResourceNotFound) {
				return nil
			}

			if err != nil {
				return fmt.Errorf("cannot update Slack delivery destination: %w", err)
			}

			updated = true

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot restore Slack delivery destination: %w", err)
	}

	if !updated {
		return nil, nil
	}

	return destination, nil
}

func (s *MessageService) ClearDestination(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	target probot.DeliveryTarget,
) error {
	destination, err := s.GetDestination(ctx, scope, organizationID, target)
	if errors.Is(err, ErrSlackbotChannelNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("cannot load Slack delivery destination: %w", err)
	}

	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := destination.Delete(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot delete Slack delivery destination: %w", err)
			}

			return nil
		},
	)
}

// clearDestinationIf deletes the destination only when its external ID still
// matches expectedExternalDestinationID.
func (s *MessageService) clearDestinationIf(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	target probot.DeliveryTarget,
	expectedExternalDestinationID string,
) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := coredata.DeleteIfExternalDestinationID(
				ctx,
				tx,
				scope,
				organizationID,
				ProviderName,
				target.Namespace,
				target.Key,
				expectedExternalDestinationID,
			); err != nil {
				return fmt.Errorf("cannot delete Slack delivery destination: %w", err)
			}

			return nil
		},
	)
}

func (s *MessageService) logNotificationRoute(
	ctx context.Context,
	backend string,
	organizationID gid.GID,
	messageType string,
	capability string,
) {
	if s.logger == nil {
		return
	}

	fields := []log.Attr{
		log.String("backend", backend),
		log.String("organization_id", organizationID.String()),
		log.String("message_type", messageType),
	}
	if capability != "" {
		fields = append(fields, log.String("capability", capability))
	}

	s.logger.InfoCtx(ctx, "routed Slack notification", fields...)
}
