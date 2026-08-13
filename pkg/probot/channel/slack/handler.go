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
	"io"
	"net/http"
	"regexp"
	"strings"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/probot/identitybinding"
)

type (
	slackMessenger interface {
		CreateMessage(
			ctx context.Context,
			channel string,
			text string,
			threadTS string,
			clientMsgID string,
		) (*MessageRef, error)
	}

	installationClientResolver interface {
		ClientByTeamID(ctx context.Context, teamID string) (*Client, *coredata.SlackbotInstallation, error)
		DisableByTeamID(ctx context.Context, teamID string) error
	}

	Handler struct {
		signingSecret string
		client        slackMessenger
		installations installationClientResolver
		bindings      identitybinding.Gate
		bindPrompts   bindPromptStore
		pg            *pg.Client
		logger        *log.Logger
	}

	replyTarget struct {
		channel  string
		threadTS string
	}
)

var mentionRE = regexp.MustCompile(`<@[A-Z0-9]+>`)

func NewHandler(
	signingSecret string,
	bindings identitybinding.Gate,
	installations installationClientResolver,
	pgClient *pg.Client,
	l *log.Logger,
) *Handler {
	return &Handler{
		signingSecret: signingSecret,
		installations: installations,
		bindings:      bindings,
		pg:            pgClient,
		logger:        l,
	}
}

func (h *Handler) SetBindPrompts(store bindPromptStore) {
	h.bindPrompts = store
}

func (h *Handler) clientForTeam(
	ctx context.Context,
	teamID string,
) (slackMessenger, gid.GID, string, error) {
	if h.installations == nil {
		if h.client == nil {
			return nil, gid.Nil, "", ErrSlackbotNotInstalled
		}

		return h.client, gid.Nil, "", nil
	}

	client, installation, err := h.installations.ClientByTeamID(ctx, teamID)
	if err != nil {
		return nil, gid.Nil, "", err
	}

	if installation == nil {
		return nil, gid.Nil, "", ErrSlackbotNotInstalled
	}

	return client, installation.OrganizationID, installation.BotUserID, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.WarnCtx(r.Context(), "cannot read request body", log.Error(err))
		http.Error(w, "cannot read body", http.StatusBadRequest)

		return
	}

	timestamp := r.Header.Get("X-Slack-Request-Timestamp")
	signature := r.Header.Get("X-Slack-Signature")

	if err := VerifySignature(h.signingSecret, timestamp, signature, body); err != nil {
		h.logger.WarnCtx(r.Context(), "invalid slack signature", log.Error(err))
		http.Error(w, "invalid signature", http.StatusUnauthorized)

		return
	}

	var envelope Envelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		h.logger.WarnCtx(r.Context(), "cannot decode slack event", log.Error(err))
		http.Error(w, "invalid json", http.StatusBadRequest)

		return
	}

	if envelope.Type == EnvelopeTypeURLVerification {
		if envelope.Challenge == "" {
			http.Error(w, "missing challenge", http.StatusBadRequest)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"challenge": envelope.Challenge})

		return
	}

	if envelope.Type != EnvelopeTypeEventCallback {
		http.Error(w, "unsupported event envelope", http.StatusBadRequest)

		return
	}

	if strings.TrimSpace(envelope.EventID) == "" {
		http.Error(w, "missing event_id", http.StatusBadRequest)

		return
	}

	if envelope.Event == nil || envelope.Event.Type == "" {
		http.Error(w, "missing event", http.StatusBadRequest)

		return
	}

	if h.pg == nil {
		h.logger.ErrorCtx(r.Context(), "Slackbot event inbox is unavailable")
		http.Error(w, "event inbox unavailable", http.StatusInternalServerError)

		return
	}

	event := coredata.NewSlackbotEvent(envelope.EventID, json.RawMessage(body))

	err = h.pg.WithConn(
		r.Context(),
		func(ctx context.Context, conn pg.Querier) error {
			_, err := event.Insert(ctx, conn)
			if err != nil {
				return fmt.Errorf("cannot insert Slack event: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		h.logger.ErrorCtx(
			r.Context(),
			"cannot persist Slackbot event before acknowledgement",
			log.String("event_id", envelope.EventID),
			log.Error(err),
		)
		http.Error(w, "cannot persist event", http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) ProcessEvent(ctx context.Context, envelope Envelope) error {
	if envelope.Type != EnvelopeTypeEventCallback {
		return &permanentEventError{
			err: fmt.Errorf("unexpected Slack envelope type %q", envelope.Type),
		}
	}

	if envelope.EventID == "" {
		return &permanentEventError{
			err: fmt.Errorf("slack event has no event ID"),
		}
	}

	if envelope.Event == nil || envelope.Event.Type == "" {
		return &permanentEventError{
			err: fmt.Errorf("slack event has no event body"),
		}
	}

	return h.dispatch(
		ctx,
		envelope.EventID,
		envelope.InstallationTeamID(),
		envelope.Event,
	)
}

func (h *Handler) dispatch(ctx context.Context, eventID, teamID string, event *EventBody) error {
	if event.Type == EventTypeAppUninstalled ||
		event.Type == EventTypeTokensRevoked {
		if teamID == "" {
			return &permanentEventError{
				err: fmt.Errorf("slack uninstall event has no team ID"),
			}
		}

		if h.installations == nil {
			return nil
		}

		if err := h.installations.DisableByTeamID(ctx, teamID); err != nil {
			return fmt.Errorf("cannot disable revoked Slack installation: %w", err)
		}

		return nil
	}

	if event.BotID != "" {
		return nil
	}

	if event.Subtype == EventSubtypeMessageChanged {
		if event.ChannelType == ChannelTypeIM {
			return h.handleEditedMessage(ctx, eventID, teamID, event)
		}

		return nil
	}

	switch event.Type {
	case EventTypeAppMention, EventTypeMessage:
		if shouldHandleConversationEvent(event) {
			return h.handleInteraction(ctx, eventID, teamID, event)
		}
	}

	return nil
}

func shouldHandleConversationEvent(event *EventBody) bool {
	return event != nil &&
		(event.Type == EventTypeAppMention ||
			(event.Type == EventTypeMessage && event.ChannelType == ChannelTypeIM))
}

func (h *Handler) handleEditedMessage(
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

	return h.handleInteraction(
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

func (h *Handler) handleInteraction(
	ctx context.Context,
	eventID, teamID string,
	event *EventBody,
) error {
	text := cleanText(event.Text)
	if text == "" {
		return nil
	}

	if teamID == "" || event.User == "" {
		return &permanentEventError{
			err: fmt.Errorf("slack %s event is missing team or user", event.Type),
		}
	}

	target := replyTargetFor(*event)
	actorTeamID := event.ActorTeamID(teamID)

	slackClient, organizationID, botUserID, err := h.clientForTeam(ctx, teamID)
	if err != nil {
		return fmt.Errorf("cannot resolve Slack installation: %w", err)
	}

	subject := IdentitySubject(actorTeamID, event.User)

	binding, err := h.bindings.Lookup(ctx, subject)
	if err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
		return fmt.Errorf("cannot lookup Slack identity binding: %w", err)
	}

	if binding == nil {
		return h.handleUnboundInteraction(
			ctx,
			eventID,
			slackClient,
			target,
		)
	}

	return h.handleBoundInteraction(
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

func (h *Handler) handleUnboundInteraction(
	ctx context.Context,
	eventID string,
	slackClient slackMessenger,
	target replyTarget,
) error {
	claimed, err := isEventIDClaimed(ctx, h.pg, eventID)
	if err != nil {
		return fmt.Errorf("cannot check Slack bind CTA delivery: %w", err)
	}

	if claimed {
		return nil
	}

	if err := h.postBindRequired(
		ctx,
		slackClient,
		target,
		eventID,
	); err != nil {
		return fmt.Errorf("cannot post Slack bind CTA: %w", err)
	}

	if _, err := claimEventID(ctx, h.pg, eventID); err != nil {
		return fmt.Errorf("cannot mark Slack bind CTA delivered: %w", err)
	}

	return nil
}

func (h *Handler) handleBoundInteraction(
	ctx context.Context,
	eventID string,
	teamID string,
	actorTeamID string,
	organizationID gid.GID,
	botUserID string,
	identityID gid.GID,
	slackClient slackMessenger,
	event *EventBody,
) error {
	userText := h.collectThreadTranscript(ctx, slackClient, *event, botUserID)
	if userText == "" {
		return nil
	}

	h.sendAssistantWorkingStatus(ctx, slackClient, *event)

	if err := h.enqueueExecutionInput(
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

func (h *Handler) sendAssistantWorkingStatus(
	ctx context.Context,
	slackClient slackMessenger,
	event EventBody,
) {
	setter, ok := slackClient.(assistantStatusSetter)
	if !ok {
		return
	}

	setAssistantWorkingStatus(
		ctx,
		h.logger,
		setter,
		event.Channel,
		assistantStatusThreadTS(event.ThreadTS, event.TS),
	)
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

func (h *Handler) postBindRequired(
	ctx context.Context,
	slackClient slackMessenger,
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
