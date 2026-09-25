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
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func (s *MessageService) OnSlackbotDeliverySuccess(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	message *coredata.SlackbotMessage,
) error {
	if raw, ok := message.Metadata[deliveryAgentExecutionMetadata].(string); ok {
		executionID, err := gid.ParseGID(raw)
		if err == nil &&
			executionID.EntityType() == coredata.AgentExecutionEntityType &&
			message.ChannelID != nil &&
			message.MessageTS != nil {
			if err := bindExecutionAnchor(
				ctx,
				tx,
				scope,
				message.OrganizationID,
				executionID,
				*message.ChannelID,
				*message.MessageTS,
			); err != nil {
				return fmt.Errorf("cannot bind Slack execution anchor: %w", err)
			}
		}
	}

	if err := bindThreadSubject(ctx, tx, scope, message); err != nil {
		return fmt.Errorf("cannot bind Slack thread subject: %w", err)
	}

	namespace, namespaceOK := message.Metadata[deliveryTargetNamespaceMetadata].(string)

	key, keyOK := message.Metadata[deliveryTargetKeyMetadata].(string)
	if !namespaceOK || !keyOK {
		return nil
	}

	var destination coredata.BotDeliveryDestination
	if err := destination.LoadByTarget(
		ctx,
		tx,
		scope,
		message.OrganizationID,
		ProviderName,
		namespace,
		key,
	); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			if s.logger != nil {
				s.logger.InfoCtx(
					ctx,
					"skipping verification of missing bot destination",
					log.String("organization_id", message.OrganizationID.String()),
					log.String("target_namespace", namespace),
				)
			}

			return nil
		}

		return fmt.Errorf("cannot load delivered bot destination: %w", err)
	}

	now := time.Now()
	destination.VerifiedAt = &now

	destination.UpdatedAt = now
	if err := destination.MarkVerified(ctx, tx, scope); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			if s.logger != nil {
				s.logger.InfoCtx(
					ctx,
					"skipping verification of deleted bot destination",
					log.String("organization_id", message.OrganizationID.String()),
					log.String("target_namespace", namespace),
				)
			}

			return nil
		}

		return fmt.Errorf("cannot verify delivered bot destination: %w", err)
	}

	return nil
}

func bindThreadSubject(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	message *coredata.SlackbotMessage,
) error {
	if message.ChannelID == nil || message.MessageTS == nil {
		return nil
	}

	namespace, namespaceOK := message.Metadata[deliverySubjectNamespaceMetadata].(string)
	key, keyOK := message.Metadata[deliverySubjectKeyMetadata].(string)

	capability, capabilityOK := message.Metadata[deliveryCapabilityMetadata].(string)
	if !namespaceOK || namespace == "" || !keyOK || key == "" || !capabilityOK || capability == "" {
		return nil
	}

	messageType := message.MessageType
	if original, ok := message.Metadata[deliveryMessageTypeMetadata].(string); ok && original != "" {
		messageType = original
	}

	attributes := cloneNotificationData(message.Metadata)
	delete(attributes, deliveryTargetNamespaceMetadata)
	delete(attributes, deliveryTargetKeyMetadata)
	delete(attributes, deliveryMessageTypeMetadata)
	delete(attributes, deliverySourceEventMetadata)
	delete(attributes, deliveryAgentExecutionMetadata)
	delete(attributes, deliverySubjectNamespaceMetadata)
	delete(attributes, deliverySubjectKeyMetadata)
	delete(attributes, deliveryCapabilityMetadata)

	encoded, err := json.Marshal(attributes)
	if err != nil {
		return fmt.Errorf("cannot encode bot thread subject attributes: %w", err)
	}

	now := time.Now()

	subject := &coredata.BotThreadSubject{
		ID:                     gid.New(scope.GetTenantID(), coredata.BotThreadSubjectEntityType),
		OrganizationID:         message.OrganizationID,
		Provider:               ProviderName,
		ExternalConversationID: *message.ChannelID,
		ExternalMessageID:      *message.MessageTS,
		Capability:             capability,
		MessageType:            messageType,
		Attributes:             encoded,
		SubjectNamespace:       namespace,
		SubjectKey:             key,
		CreatedAt:              now,
		UpdatedAt:              now,
	}
	if _, err := subject.Upsert(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot bind delivered bot thread subject: %w", err)
	}

	return nil
}

func bindExecutionAnchor(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	organizationID gid.GID,
	executionID gid.GID,
	channelID string,
	messageTS string,
) error {
	now := time.Now()

	anchor := &coredata.AgentExecutionAnchor{
		ID:                     gid.New(scope.GetTenantID(), coredata.AgentExecutionAnchorEntityType),
		OrganizationID:         organizationID,
		AgentExecutionID:       executionID,
		Provider:               ProviderName,
		ExternalConversationID: channelID,
		ExternalMessageID:      messageTS,
		CreatedAt:              now,
		UpdatedAt:              now,
	}
	if _, err := anchor.Upsert(ctx, conn, scope); err != nil {
		return fmt.Errorf("cannot bind delivered execution anchor: %w", err)
	}

	var execution coredata.AgentExecution
	if err := execution.LoadByID(ctx, conn, scope, executionID); err != nil {
		return fmt.Errorf("cannot load delivered execution: %w", err)
	}

	var coordinates ExecutionSourceCoordinates
	if len(execution.SourceCoordinates) > 0 {
		if err := json.Unmarshal(execution.SourceCoordinates, &coordinates); err != nil {
			return fmt.Errorf("cannot decode delivered execution coordinates: %w", err)
		}
	}

	coordinates.ChannelID = channelID
	coordinates.ThreadTS = messageTS

	encoded, err := json.Marshal(coordinates)
	if err != nil {
		return fmt.Errorf("cannot encode delivered execution coordinates: %w", err)
	}

	if err := execution.UpdateSourceCoordinates(ctx, conn, scope, encoded, now); err != nil {
		return fmt.Errorf("cannot update delivered execution coordinates: %w", err)
	}

	return nil
}
