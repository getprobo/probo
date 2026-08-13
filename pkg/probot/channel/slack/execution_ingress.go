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

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/bot"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/llm"
)

const defaultAgentProfile = "probot"

func (h *Handler) enqueueExecutionInput(
	ctx context.Context,
	eventID string,
	installationTeamID string,
	actorTeamID string,
	organizationID gid.GID,
	identityID gid.GID,
	event EventBody,
	userText string,
) error {
	target := replyTargetFor(event)

	coordinates, err := json.Marshal(
		ExecutionSourceCoordinates{
			TeamID:         installationTeamID,
			ActorTeamID:    actorTeamID,
			ChannelID:      target.channel,
			ThreadTS:       target.threadTS,
			MessageTS:      event.TS,
			ExternalUserID: event.User,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot encode Slack execution coordinates: %w", err)
	}

	message, err := json.Marshal(
		llm.Message{
			Role:  llm.RoleUser,
			Parts: []llm.Part{llm.TextPart{Text: userText}},
		},
	)
	if err != nil {
		return fmt.Errorf("cannot encode Slack agent input: %w", err)
	}

	scope := coredata.NewScopeFromObjectID(organizationID)
	now := time.Now()

	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			exists, err := coredata.AgentInputExistsBySourceEventID(
				ctx,
				tx,
				scope,
				organizationID,
				ProviderName,
				eventID,
			)
			if err != nil {
				return fmt.Errorf("cannot check duplicate Slack agent input: %w", err)
			}

			if exists {
				return nil
			}

			execution, err := loadAnchoredExecution(
				ctx,
				tx,
				scope,
				organizationID,
				target,
			)
			if err != nil {
				return fmt.Errorf("cannot load anchored Slack execution: %w", err)
			}

			if execution == nil {
				sessionID := sessionIDFor(
					installationTeamID,
					event.Channel,
					event.ChannelType,
					event.ThreadTS,
					event.TS,
				)
				if sessionID == "" {
					return fmt.Errorf("cannot build Slack execution session ID")
				}

				trustedContext, err := loadThreadTrustedContext(
					ctx,
					tx,
					scope,
					organizationID,
					target,
				)
				if err != nil {
					return fmt.Errorf("cannot load Slack thread trusted context: %w", err)
				}

				source := ProviderName

				execution = &coredata.AgentExecution{
					ID:                gid.New(scope.GetTenantID(), coredata.AgentRunEntityType),
					OrganizationID:    organizationID,
					StartAgentName:    defaultAgentProfile,
					Source:            &source,
					SessionKey:        &sessionID,
					SourceCoordinates: coordinates,
					TrustedContext:    trustedContext,
					SessionMessages:   json.RawMessage("[]"),
					MaxAttempts:       coredata.AgentExecutionDefaultMaxAttempts,
					CreatedAt:         now,
					UpdatedAt:         now,
				}
				if _, err := execution.UpsertConversationalBySourceSession(
					ctx,
					tx,
					scope,
				); err != nil {
					return fmt.Errorf("cannot upsert direct Slack execution: %w", err)
				}
			} else if err := execution.UpdateSourceCoordinates(
				ctx,
				tx,
				scope,
				coordinates,
				now,
			); err != nil {
				return fmt.Errorf("cannot update anchored Slack execution coordinates: %w", err)
			}

			input := &coredata.AgentInput{
				ID:             gid.New(scope.GetTenantID(), coredata.AgentInputEntityType),
				OrganizationID: organizationID,
				AgentRunID:     execution.ID,
				Source:         ProviderName,
				SourceEventID:  &eventID,
				Purpose:        coredata.AgentInputPurposeUser,
				IdentityID:     new(identityID),
				Message:        message,
				MaxAttempts:    coredata.AgentInputDefaultMaxAttempts,
				CreatedAt:      now,
				UpdatedAt:      now,
			}
			if _, err := input.EnqueueIdempotently(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot enqueue Slack agent input: %w", err)
			}

			return nil
		},
	)
}

func loadThreadTrustedContext(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	organizationID gid.GID,
	target replyTarget,
) (json.RawMessage, error) {
	if target.channel == "" || target.threadTS == "" {
		return nil, nil
	}

	var subject coredata.BotThreadSubject

	err := subject.LoadByProviderCoordinates(
		ctx,
		tx,
		scope,
		organizationID,
		ProviderName,
		target.channel,
		target.threadTS,
	)
	if errors.Is(err, coredata.ErrResourceNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("cannot load Slack thread subject: %w", err)
	}

	var attributes map[string]any
	if len(subject.Attributes) > 0 {
		if err := json.Unmarshal(subject.Attributes, &attributes); err != nil {
			return nil, fmt.Errorf("cannot decode Slack thread subject attributes: %w", err)
		}
	}

	trusted, err := json.Marshal(
		bot.ConversationTrustedContext{
			Capability:       subject.Capability,
			MessageType:      subject.MessageType,
			Attributes:       attributes,
			SubjectNamespace: subject.SubjectNamespace,
			SubjectKey:       subject.SubjectKey,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot encode Slack thread trusted context: %w", err)
	}

	return trusted, nil
}

func loadAnchoredExecution(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	organizationID gid.GID,
	target replyTarget,
) (*coredata.AgentExecution, error) {
	if target.channel == "" || target.threadTS == "" {
		return nil, nil
	}

	var anchor coredata.AgentExecutionAnchor

	err := anchor.LoadByProviderCoordinates(
		ctx,
		tx,
		scope,
		organizationID,
		ProviderName,
		target.channel,
		target.threadTS,
	)
	if errors.Is(err, coredata.ErrResourceNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("cannot resolve Slack execution anchor: %w", err)
	}

	var execution coredata.AgentExecution
	if err := execution.LoadByID(ctx, tx, scope, anchor.AgentRunID); err != nil {
		return nil, fmt.Errorf("cannot load anchored Slack execution: %w", err)
	}

	if execution.ExecutionKind != coredata.AgentExecutionKindConversational {
		return nil, fmt.Errorf("slack execution anchor points to a non-conversational execution")
	}

	return &execution, nil
}
