// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"go.gearno.de/kit/httpclient"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/probot/identitybinding"
)

type (
	Service struct {
		installations *InstallationService
		bindings      identitybinding.Gate
		bindPrompts   *BindPromptService
		pg            *pg.Client
		logger        *log.Logger
		httpClient    *http.Client
	}

	replyTarget struct {
		channel  string
		threadTS string
	}

	InvalidEnvelopeError struct {
		err error
	}
)

var mentionRE = regexp.MustCompile(`<@[A-Z0-9]+>`)

func (e *InvalidEnvelopeError) Error() string {
	return e.err.Error()
}

func (e *InvalidEnvelopeError) Unwrap() error {
	return e.err
}

func invalidEnvelope(msg string) error {
	return &InvalidEnvelopeError{err: errors.New(msg)}
}

func validateEventCallback(envelope Envelope) error {
	if envelope.Type != EnvelopeTypeEventCallback {
		return invalidEnvelope("unsupported event envelope")
	}

	if strings.TrimSpace(envelope.EventID) == "" {
		return invalidEnvelope("missing event_id")
	}

	if envelope.Event == nil || envelope.Event.Type == "" {
		return invalidEnvelope("missing event")
	}

	return nil
}

func NewService(
	bindings identitybinding.Gate,
	installations *InstallationService,
	pgClient *pg.Client,
	l *log.Logger,
) *Service {
	return &Service{
		installations: installations,
		bindings:      bindings,
		pg:            pgClient,
		logger:        l,
		httpClient: httpclient.DefaultPooledClient(
			httpclient.WithLogger(l),
			httpclient.WithSSRFProtection(),
		),
	}
}

func (s *Service) SetBindPrompts(store *BindPromptService) {
	s.bindPrompts = store
}

func (s *Service) InteractiveActorBound(
	ctx context.Context,
	payload InteractivePayload,
) (bool, error) {
	if s == nil || s.bindings == nil {
		return true, nil
	}

	binding, err := s.bindings.Lookup(ctx, payload.ActorSubject())
	if err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
		return false, fmt.Errorf("cannot lookup Slack identity binding: %w", err)
	}

	return binding != nil, nil
}

func (s *Service) clientForTeam(
	ctx context.Context,
	teamID string,
) (*Client, gid.GID, string, error) {
	if s.installations == nil {
		return nil, gid.Nil, "", ErrSlackbotNotInstalled
	}

	client, installation, err := s.installations.ClientByTeamID(ctx, teamID)
	if err != nil {
		return nil, gid.Nil, "", err
	}

	if installation == nil {
		return nil, gid.Nil, "", ErrSlackbotNotInstalled
	}

	return client, installation.OrganizationID, installation.BotUserID, nil
}

func (s *Service) EnqueueEvent(ctx context.Context, envelope Envelope) error {
	if err := validateEventCallback(envelope); err != nil {
		return err
	}

	if s.pg == nil {
		return fmt.Errorf("cannot enqueue Slack event: inbox unavailable")
	}

	rawBody, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("cannot encode Slack event: %w", err)
	}

	event := coredata.NewSlackbotEvent(envelope.EventID, rawBody)

	err = s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			_, err := event.Insert(ctx, conn)
			if err != nil {
				return fmt.Errorf("cannot insert Slack event: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return fmt.Errorf("cannot persist Slack event: %w", err)
	}

	return nil
}

func (s *Service) ProcessEvent(ctx context.Context, envelope Envelope) error {
	if err := validateEventCallback(envelope); err != nil {
		return permanent(err)
	}

	return s.dispatch(
		ctx,
		envelope.EventID,
		envelope.InstallationTeamID(),
		envelope.Time(),
		envelope.Event,
	)
}

func (s *Service) dispatch(
	ctx context.Context,
	eventID string,
	teamID string,
	eventTime *time.Time,
	event *EventBody,
) error {
	if event.Type == EventTypeAppUninstalled ||
		event.Type == EventTypeTokensRevoked {
		if teamID == "" {
			return permanent(fmt.Errorf("slack uninstall event has no team ID"))
		}

		if s.installations == nil {
			return nil
		}

		if err := s.installations.DisableByTeamID(
			ctx,
			teamID,
			eventTime,
			event.RevokedBotUserIDs(),
		); err != nil {
			return fmt.Errorf("cannot disable revoked Slack installation: %w", err)
		}

		return nil
	}

	if event.BotID != "" {
		return nil
	}

	if event.Subtype == EventSubtypeMessageChanged {
		if event.ChannelType == ChannelTypeIM {
			return s.handleEditedMessage(ctx, eventID, teamID, event)
		}

		return nil
	}

	switch event.Type {
	case EventTypeAppMention, EventTypeMessage:
		if shouldHandleConversationEvent(event) {
			return s.handleInteraction(ctx, eventID, teamID, event)
		}
	}

	return nil
}

func shouldHandleConversationEvent(event *EventBody) bool {
	return event != nil &&
		(event.Type == EventTypeAppMention ||
			(event.Type == EventTypeMessage && event.ChannelType == ChannelTypeIM))
}

func (s *Service) handleEditedMessage(
	ctx context.Context,
	eventID, teamID string,
	event *EventBody,
) error {
	if event.Message == nil || event.Message.BotID != "" {
		return nil
	}

	previousText := ""
	if event.PreviousMessage != nil {
		previousText = cleanText(event.PreviousMessage.Text)
	}

	newText := cleanText(event.Message.Text)
	if newText == "" {
		return nil
	}

	contextText := fmt.Sprintf(
		"[The user edited their message. Previous: %q. New: %q]",
		previousText,
		newText,
	)

	return s.handleInteraction(
		ctx,
		eventID,
		teamID,
		&EventBody{
			Type:        event.Type,
			User:        event.Message.User,
			UserTeam:    event.Message.UserTeam,
			Text:        contextText,
			Channel:     event.Channel,
			ChannelType: event.ChannelType,
			TS:          event.Message.TS,
			ThreadTS:    event.Message.ThreadTS,
		},
	)
}

func (s *Service) handleInteraction(
	ctx context.Context,
	eventID, teamID string,
	event *EventBody,
) error {
	text := cleanText(event.Text)
	if text == "" {
		return nil
	}

	if teamID == "" || event.User == "" {
		return permanent(fmt.Errorf("slack %s event is missing team or user", event.Type))
	}

	target := replyTargetFor(*event)
	actorTeamID := event.ActorTeamID(teamID)

	slackClient, organizationID, botUserID, err := s.clientForTeam(ctx, teamID)
	if err != nil {
		return fmt.Errorf("cannot resolve Slack installation: %w", err)
	}

	subject := IdentitySubject(actorTeamID, event.User)

	binding, err := s.bindings.Lookup(ctx, subject)
	if err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
		return fmt.Errorf("cannot lookup Slack identity binding: %w", err)
	}

	if binding == nil {
		return s.handleUnboundInteraction(
			ctx,
			eventID,
			slackClient,
			target,
		)
	}

	return s.handleBoundInteraction(
		ctx,
		eventID,
		teamID,
		actorTeamID,
		organizationID,
		botUserID,
		binding.IdentityID,
		slackClient,
		event,
	)
}

func (s *Service) handleUnboundInteraction(
	ctx context.Context,
	eventID string,
	slackClient *Client,
	target replyTarget,
) error {
	claimed, err := isEventIDClaimed(ctx, s.pg, eventID)
	if err != nil {
		return fmt.Errorf("cannot check Slack bind CTA delivery: %w", err)
	}

	if claimed {
		return nil
	}

	if err := s.postBindRequired(
		ctx,
		slackClient,
		target,
		eventID,
	); err != nil {
		return fmt.Errorf("cannot post Slack bind CTA: %w", err)
	}

	if _, err := claimEventID(ctx, s.pg, eventID); err != nil {
		return fmt.Errorf("cannot mark Slack bind CTA delivered: %w", err)
	}

	return nil
}

func (s *Service) handleBoundInteraction(
	ctx context.Context,
	eventID string,
	teamID string,
	actorTeamID string,
	organizationID gid.GID,
	botUserID string,
	identityID gid.GID,
	slackClient *Client,
	event *EventBody,
) error {
	userText := s.collectThreadTranscript(ctx, slackClient, *event, botUserID)
	if userText == "" {
		return nil
	}

	// The turn coordinates persisted below are what the run hook uses to clear
	// the indicator, so both sides derive it from the same reply target.
	if slackClient != nil {
		target := replyTargetFor(*event)

		setAssistantWorkingStatus(
			ctx,
			s.logger,
			slackClient,
			target.channel,
			assistantStatusThreadTS(target.threadTS, event.TS),
		)
	}

	if err := s.enqueueExecutionInput(
		ctx,
		eventID,
		teamID,
		actorTeamID,
		organizationID,
		identityID,
		*event,
		userText,
	); err != nil {
		return fmt.Errorf("cannot enqueue Slack execution input: %w", err)
	}

	return nil
}

// replyTargetFor derives the conversation destination from a Slack event.
// Used only when creating the agent row; runtime posts use the stored scope.
func replyTargetFor(payload EventBody) replyTarget {
	threadTS := payload.ThreadTS
	if threadTS == "" {
		threadTS = payload.TS
	}

	if payload.ChannelType == ChannelTypeIM {
		threadTS = ""
	}

	return replyTarget{
		channel:  payload.Channel,
		threadTS: threadTS,
	}
}

func (s *Service) postBindRequired(
	ctx context.Context,
	slackClient *Client,
	target replyTarget,
	eventID string,
) error {
	if _, err := slackClient.CreateMessage(
		ctx,
		target.channel,
		bindRequiredText,
		target.threadTS,
		stableSlackClientMsgID(eventID, "bind-prompt"),
	); err != nil {
		return fmt.Errorf("cannot post Slack bind prompt: %w", err)
	}

	return nil
}

func stableSlackClientMsgID(eventID string, purpose string) string {
	sum := sha256.Sum256([]byte(eventID + "\x00" + purpose))
	sum[6] = (sum[6] & 0x0f) | 0x50
	sum[8] = (sum[8] & 0x3f) | 0x80

	return fmt.Sprintf(
		"%x-%x-%x-%x-%x",
		sum[0:4],
		sum[4:6],
		sum[6:8],
		sum[8:10],
		sum[10:16],
	)
}

func cleanText(text string) string {
	return strings.TrimSpace(mentionRE.ReplaceAllString(text, ""))
}

func IdentitySubject(teamID, slackUserID string) identitybinding.Subject {
	return identitybinding.Subject{
		Provider:         ProviderName,
		ExternalTenantID: teamID,
		ExternalUserID:   slackUserID,
	}
}
