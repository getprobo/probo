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
	"fmt"
	"time"

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
	if source, ok := message.Metadata[deliverySourceEventMetadata].(string); ok {
		executionID, err := gid.ParseGID(source)
		if err == nil &&
			executionID.EntityType() == coredata.AgentExecutionEntityType &&
			message.ChannelID != nil &&
			message.MessageTS != nil {
			if err := bindExecutionAnchor(ctx, tx, scope, executionID, message); err != nil {
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
	if err := destination.LoadByTarget(ctx, tx, scope, ProviderName, namespace, key); err != nil {
		return fmt.Errorf("cannot load delivered bot destination: %w", err)
	}

	now := time.Now()
	destination.VerifiedAt = &now
	destination.UpdatedAt = now
	if err := destination.MarkVerified(ctx, tx, scope); err != nil {
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
	tx pg.Tx,
	scope coredata.Scoper,
	executionID gid.GID,
	message *coredata.SlackbotMessage,
) error {
	now := time.Now()
	anchor := &coredata.AgentExecutionAnchor{
		ID:                     gid.New(scope.GetTenantID(), coredata.AgentExecutionAnchorEntityType),
		OrganizationID:         message.OrganizationID,
		AgentExecutionID:       executionID,
		Provider:               ProviderName,
		ExternalConversationID: *message.ChannelID,
		ExternalMessageID:      *message.MessageTS,
		CreatedAt:              now,
		UpdatedAt:              now,
	}
	if _, err := anchor.Upsert(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot bind delivered execution anchor: %w", err)
	}

	var execution coredata.AgentExecution
	if err := execution.LoadByID(ctx, tx, scope, executionID); err != nil {
		return fmt.Errorf("cannot load delivered execution: %w", err)
	}

	var coordinates ExecutionSourceCoordinates
	if len(execution.SourceCoordinates) > 0 {
		if err := json.Unmarshal(execution.SourceCoordinates, &coordinates); err != nil {
			return fmt.Errorf("cannot decode delivered execution coordinates: %w", err)
		}
	}
	coordinates.ChannelID = *message.ChannelID
	coordinates.ThreadTS = *message.MessageTS

	encoded, err := json.Marshal(coordinates)
	if err != nil {
		return fmt.Errorf("cannot encode delivered execution coordinates: %w", err)
	}
	if err := execution.UpdateSourceCoordinates(ctx, tx, scope, encoded, now); err != nil {
		return fmt.Errorf("cannot update delivered execution coordinates: %w", err)
	}

	return nil
}
