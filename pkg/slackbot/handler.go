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

package slackbot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"runtime/debug"
	"strings"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/llm"
)

const (
	agentFailureReply  = "Sorry, something went wrong while processing your message. Please try again."
	emptyReplyFallback = "I couldn't generate a response."

	postMessageMaxAttempts = 3
)

type (
	bindingGate interface {
		Lookup(ctx context.Context, teamID, slackUserID string) (*coredata.SlackIdentityBinding, error)
		BindURL(teamID, slackUserID string) (string, error)
	}

	slackMessenger interface {
		OpenIM(ctx context.Context, userID string) (string, error)
		PostMessage(ctx context.Context, channel, text, threadTS string) error
		PostMessageWithBlocks(ctx context.Context, channel, text, threadTS string, blocks []any) error
		SetStatus(ctx context.Context, channelID, threadTS, status string) error
		AddReaction(ctx context.Context, channel, name, timestamp string) error
	}

	Handler struct {
		signingSecret string
		baseAgent     *agent.Agent
		client        slackMessenger
		bindings      bindingGate
		pg            *pg.Client
		checkpointer  *PGCheckpointer
		logger        *log.Logger
		serviceCtx    context.Context
	}

	replyTarget struct {
		channel  string
		threadTS string
	}

	singleAgentRegistry struct {
		a *agent.Agent
	}
)

var mentionRE = regexp.MustCompile(`<@[A-Z0-9]+>`)

func NewHandler(
	signingSecret string,
	a *agent.Agent,
	c *Client,
	bindings *BindingService,
	pgClient *pg.Client,
	serviceCtx context.Context,
	l *log.Logger,
) *Handler {
	return &Handler{
		signingSecret: signingSecret,
		baseAgent:     a,
		client:        c,
		bindings:      bindings,
		pg:            pgClient,
		checkpointer:  NewCheckpointer(pgClient),
		logger:        l,
		serviceCtx:    serviceCtx,
	}
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
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"challenge": envelope.Challenge})

		return
	}

	w.WriteHeader(http.StatusOK)
	go h.processEvent(h.serviceCtx, envelope)
}

func (h *Handler) processEvent(ctx context.Context, envelope Envelope) {
	defer h.recoverPanic(ctx, "processEvent")

	if envelope.Type != EnvelopeTypeEventCallback || envelope.Event == nil {
		return
	}

	h.dispatch(ctx, envelope.EventID, envelope.TeamID, envelope.Event)
}

func (h *Handler) dispatch(ctx context.Context, eventID, teamID string, event *EventBody) {
	if event.BotID != "" {
		return
	}

	if event.Subtype == EventSubtypeMessageChanged {
		h.handleEditedMessage(ctx, eventID, teamID, event)

		return
	}

	switch event.Type {
	case EventTypeAppMention:
		h.handleInteraction(ctx, eventID, teamID, event)
	case EventTypeMessage:
		if event.ChannelType == ChannelTypeIM {
			h.handleInteraction(ctx, eventID, teamID, event)
		}
	case EventTypeReactionAdded:
		if event.Item == nil || event.Item.Type != ReactionItemTypeMessage {
			return
		}

		h.handleInteraction(
			ctx,
			eventID,
			teamID,
			&EventBody{
				Type: event.Type,
				User: event.User,
				Text: fmt.Sprintf(
					"@%s reacted with :%s: to a message in this thread",
					event.User,
					event.Reaction,
				),
				Channel:  event.Item.Channel,
				TS:       event.Item.TS,
				ThreadTS: event.Item.TS,
			},
		)
	}
}

func (h *Handler) handleEditedMessage(ctx context.Context, eventID, teamID string, event *EventBody) {
	if event.Message == nil || event.Message.BotID != "" {
		return
	}

	previousText := ""
	if event.PreviousMessage != nil {
		previousText = cleanText(event.PreviousMessage.Text)
	}

	newText := cleanText(event.Message.Text)
	if newText == "" {
		return
	}

	contextText := fmt.Sprintf(
		"[The user edited their message. Previous: %q. New: %q]",
		previousText,
		newText,
	)

	h.handleInteraction(
		ctx,
		eventID,
		teamID,
		&EventBody{
			Type:        event.Type,
			User:        event.Message.User,
			Text:        contextText,
			Channel:     event.Channel,
			ChannelType: event.ChannelType,
			TS:          event.Message.TS,
			ThreadTS:    event.Message.ThreadTS,
		},
	)
}

func (h *Handler) handleInteraction(ctx context.Context, eventID, teamID string, event *EventBody) {
	text := cleanText(event.Text)
	if text == "" {
		return
	}

	if teamID == "" || event.User == "" {
		h.logger.WarnCtx(
			ctx,
			"slack event missing team or user",
			log.String("event_id", eventID),
			log.String("event_type", event.Type.String()),
		)

		return
	}

	target := replyTargetFor(*event)

	binding, err := h.bindings.Lookup(ctx, teamID, event.User)
	if err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
		h.logger.ErrorCtx(
			ctx,
			"cannot lookup slack identity binding",
			log.Error(err),
			log.String("event_id", eventID),
		)

		return
	}
	if binding == nil {
		claimed, err := claimEventID(ctx, h.pg, eventID)
		if err != nil {
			h.logger.ErrorCtx(
				ctx,
				"cannot claim slack event for bind cta",
				log.Error(err),
				log.String("event_id", eventID),
			)

			return
		}
		if !claimed {
			return
		}

		bindURL, err := h.bindings.BindURL(teamID, event.User)
		if err != nil {
			h.logger.ErrorCtx(
				ctx,
				"cannot build slack bind url",
				log.Error(err),
				log.String("event_id", eventID),
			)

			return
		}

		h.postBindRequired(ctx, eventID, target, event, bindURL)

		return
	}

	sessionID := sessionIDFor(
		teamID,
		event.Channel,
		event.ChannelType,
		event.ThreadTS,
		event.TS,
		event.User,
	)
	if sessionID == "" {
		h.logger.WarnCtx(
			ctx,
			"cannot build slackbot session id",
			log.String("team_id", teamID),
			log.String("channel", event.Channel),
			log.String("event_id", eventID),
		)

		return
	}

	agentID, err := upsertAgent(ctx, h.pg, sessionID, target.channel, target.threadTS, event.User)
	if err != nil {
		h.logger.ErrorCtx(
			ctx,
			"cannot upsert agent",
			log.Error(err),
			log.String("session_id", sessionID),
		)

		return
	}

	if err := insertInteraction(ctx, h.pg, agentID, eventID, event.Type.String(), *event); err != nil {
		h.logger.ErrorCtx(
			ctx,
			"cannot insert interaction",
			log.Error(err),
			log.String("agent_id", agentID),
		)

		return
	}

	claimed, err := claimAgent(ctx, h.pg, agentID)
	if err != nil {
		h.logger.ErrorCtx(
			ctx,
			"cannot claim agent",
			log.Error(err),
			log.String("agent_id", agentID),
		)

		return
	}
	if !claimed {
		return
	}

	go h.runActor(ctx, agentID)
}

func (h *Handler) runActor(ctx context.Context, agentID string) {
	defer h.recoverPanic(ctx, "runActor")
	defer func() {
		if err := markAvailable(ctx, h.pg, agentID); err != nil {
			h.logger.WarnCtx(
				ctx,
				"cannot mark agent available",
				log.Error(err),
				log.String("agent_id", agentID),
			)
		}
	}()

	for {
		interactions, err := listPending(ctx, h.pg, agentID)
		if err != nil {
			h.logger.ErrorCtx(
				ctx,
				"cannot list pending interactions",
				log.Error(err),
				log.String("agent_id", agentID),
			)

			break
		}

		if len(interactions) == 0 {
			break
		}

		if !h.processInteractionBatch(ctx, agentID, interactions) {
			break
		}
	}
}

func (h *Handler) processInteractionBatch(ctx context.Context, agentID string, interactions []interaction) bool {
	scope, err := loadAgentScope(ctx, h.pg, agentID)
	if err != nil {
		h.logger.ErrorCtx(
			ctx,
			"cannot load agent scope",
			log.Error(err),
			log.String("agent_id", agentID),
		)

		return false
	}

	owned, foreign := partitionInteractionsByUser(interactions, scope.SlackUserID)
	if len(foreign) > 0 {
		h.logger.WarnCtx(
			ctx,
			"dropping interactions from other slack users",
			log.String("agent_id", agentID),
			log.Int("foreign_count", len(foreign)),
		)
		h.markInteractionsProcessed(ctx, agentID, foreign)
	}
	if len(owned) == 0 {
		return true
	}

	lastInteraction := owned[len(owned)-1]

	target := replyTarget{
		channel:  scope.Channel,
		threadTS: scope.ThreadTS,
	}

	slackUserID := scope.SlackUserID
	if slackUserID == "" {
		slackUserID = lastInteraction.Payload.User
	}

	teamID := teamIDFromSessionID(scope.SessionID)
	binding, err := h.lookupBinding(ctx, teamID, slackUserID)
	if err != nil {
		h.logger.ErrorCtx(
			ctx,
			"cannot verify slack identity binding",
			log.Error(err),
			log.String("agent_id", agentID),
		)

		return false
	}
	if binding == nil {
		h.logger.WarnCtx(
			ctx,
			"refusing agent run without slack identity binding",
			log.String("agent_id", agentID),
		)

		return false
	}

	if lastInteraction.EventType != EventTypeReactionAdded.String() {
		if err := h.client.SetStatus(ctx, target.channel, target.threadTS, "is thinking..."); err != nil {
			h.logger.WarnCtx(
				ctx,
				"cannot set assistant status",
				log.Error(err),
				log.String("agent_id", agentID),
			)
		}
	}

	messages := buildMessages(owned)
	sess := newAgentSession(h.pg, agentID)
	a := h.baseAgent.Clone(agent.WithSession(sess, agentID))

	runCtx := agent.WithRunContext(
		ctx,
		&RunContext{
			Channel:    target.channel,
			ThreadTS:   target.threadTS,
			IdentityID: binding.IdentityID,
		},
	)
	result, err := a.Run(runCtx, messages, agent.WithCheckpointer(h.checkpointer, agentID))

	if lastInteraction.EventType != EventTypeReactionAdded.String() {
		if clearErr := h.client.SetStatus(ctx, target.channel, target.threadTS, ""); clearErr != nil {
			h.logger.WarnCtx(
				ctx,
				"cannot clear assistant status",
				log.Error(clearErr),
				log.String("agent_id", agentID),
			)
		}
	}

	if err != nil {
		h.logger.ErrorCtx(
			ctx,
			"agent run failed",
			log.Error(err),
			log.String("agent_id", agentID),
		)

		if postErr := h.postReply(ctx, target, agentFailureReply); postErr != nil {
			h.logger.ErrorCtx(
				ctx,
				"cannot post agent failure reply",
				log.Error(postErr),
				log.String("agent_id", agentID),
			)
		}

		h.markInteractionsProcessed(ctx, agentID, owned)

		return false
	}

	reply := extractReplyText(result)
	if reply == "" {
		h.logger.WarnCtx(
			ctx,
			"agent produced empty reply, using fallback",
			log.String("agent_id", agentID),
			log.Int("message_count", len(result.Messages)),
		)
		reply = emptyReplyFallback
	}

	if err := h.postReply(ctx, target, reply); err != nil {
		h.logger.ErrorCtx(
			ctx,
			"cannot post message",
			log.Error(err),
			log.String("agent_id", agentID),
		)

		return false
	}

	h.markInteractionsProcessed(ctx, agentID, owned)

	return true
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
	eventID string,
	target replyTarget,
	event *EventBody,
	bindURL string,
) {
	if event.ChannelType == ChannelTypeIM {
		if err := h.client.PostMessageWithBlocks(
			ctx,
			target.channel,
			bindRequiredDMText,
			"",
			bindRequiredBlocks(bindURL),
		); err != nil {
			h.logger.ErrorCtx(
				ctx,
				"cannot post slack bind dm",
				log.Error(err),
				log.String("event_id", eventID),
			)
		}

		return
	}

	if err := h.client.PostMessage(
		ctx,
		target.channel,
		bindRequiredPublicText,
		target.threadTS,
	); err != nil {
		h.logger.ErrorCtx(
			ctx,
			"cannot post slack bind public reply",
			log.Error(err),
			log.String("event_id", eventID),
		)
	}

	dmChannel, err := h.client.OpenIM(ctx, event.User)
	if err != nil {
		h.logger.ErrorCtx(
			ctx,
			"cannot open slack dm for bind link",
			log.Error(err),
			log.String("event_id", eventID),
		)

		return
	}

	if err := h.client.PostMessageWithBlocks(
		ctx,
		dmChannel,
		bindRequiredDMText,
		"",
		bindRequiredBlocks(bindURL),
	); err != nil {
		h.logger.ErrorCtx(
			ctx,
			"cannot post slack bind dm",
			log.Error(err),
			log.String("event_id", eventID),
		)
	}
}

func (h *Handler) postReply(ctx context.Context, target replyTarget, text string) error {
	var lastErr error

	for attempt := range postMessageMaxAttempts {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return fmt.Errorf("cannot post message: %w", ctx.Err())
			case <-time.After(time.Duration(attempt) * 500 * time.Millisecond):
			}
		}

		if err := h.client.PostMessage(ctx, target.channel, text, target.threadTS); err != nil {
			lastErr = err

			continue
		}

		return nil
	}

	return fmt.Errorf("cannot post message after %d attempts: %w", postMessageMaxAttempts, lastErr)
}

func (h *Handler) markInteractionsProcessed(ctx context.Context, agentID string, interactions []interaction) {
	ids := make([]string, len(interactions))
	for i, in := range interactions {
		ids[i] = in.InteractionID
	}
	if err := markProcessed(ctx, h.pg, ids); err != nil {
		h.logger.WarnCtx(
			ctx,
			"cannot mark interactions processed",
			log.Error(err),
			log.String("agent_id", agentID),
		)
	}
}

func (h *Handler) ResumePending(ctx context.Context) error {
	agents, err := listResumable(ctx, h.pg)
	if err != nil {
		return err
	}

	registry := &singleAgentRegistry{a: h.baseAgent}

	for _, a := range agents {
		go h.resumeActor(ctx, a, registry)
	}

	return nil
}

func (h *Handler) resumeActor(ctx context.Context, r resumableAgent, registry agent.AgentRegistry) {
	defer h.recoverPanic(ctx, "resumeActor")

	teamID := teamIDFromSessionID(r.SessionID)
	binding, err := h.lookupBinding(ctx, teamID, r.SlackUserID)
	if err != nil {
		h.logger.ErrorCtx(
			ctx,
			"cannot verify slack identity binding for resume",
			log.Error(err),
			log.String("agent_id", r.AgentID),
		)
		_ = markAvailable(ctx, h.pg, r.AgentID)

		return
	}
	if binding == nil {
		h.logger.WarnCtx(
			ctx,
			"refusing agent resume without slack identity binding",
			log.String("agent_id", r.AgentID),
		)
		_ = markAvailable(ctx, h.pg, r.AgentID)

		return
	}

	runCtx := agent.WithRunContext(
		ctx,
		&RunContext{
			Channel:    r.Channel,
			ThreadTS:   r.ThreadTS,
			IdentityID: binding.IdentityID,
		},
	)
	result, err := agent.Restore(runCtx, h.checkpointer, r.AgentID, registry)
	if err != nil {
		h.logger.ErrorCtx(
			ctx,
			"cannot restore agent",
			log.Error(err),
			log.String("agent_id", r.AgentID),
		)
		_ = markAvailable(ctx, h.pg, r.AgentID)

		return
	}

	if err := saveMessages(ctx, h.pg, r.AgentID, result.Messages); err != nil {
		h.logger.WarnCtx(
			ctx,
			"cannot save messages after restore",
			log.Error(err),
			log.String("agent_id", r.AgentID),
		)
	}

	reply := extractReplyText(result)
	if reply != "" {
		target := replyTarget{
			channel:  r.Channel,
			threadTS: r.ThreadTS,
		}
		if err := h.postReply(ctx, target, reply); err != nil {
			h.logger.ErrorCtx(
				ctx,
				"cannot post resumed message",
				log.Error(err),
				log.String("agent_id", r.AgentID),
			)
		}
	}

	h.runActor(ctx, r.AgentID)
}

func (h *Handler) recoverPanic(ctx context.Context, operation string) {
	r := recover()
	if r == nil {
		return
	}

	h.logger.ErrorCtx(
		ctx,
		"panic in slackbot goroutine",
		log.String("operation", operation),
		log.Any("panic", r),
		log.String("stack", string(debug.Stack())),
	)
}

func partitionInteractionsByUser(
	interactions []interaction,
	slackUserID string,
) (owned, foreign []interaction) {
	if slackUserID == "" {
		return interactions, nil
	}

	for _, in := range interactions {
		if in.Payload.User == "" || in.Payload.User == slackUserID {
			owned = append(owned, in)

			continue
		}

		foreign = append(foreign, in)
	}

	return owned, foreign
}

func buildMessages(interactions []interaction) []llm.Message {
	messages := make([]llm.Message, 0, len(interactions))
	for _, i := range interactions {
		text := cleanText(i.Payload.Text)
		if text == "" {
			continue
		}

		contextLine := "Slack context: channel=" + i.Payload.Channel + ", message_ts=" + i.Payload.TS
		if i.Payload.ThreadTS != "" {
			contextLine += ", thread_ts=" + i.Payload.ThreadTS
		}
		if i.Payload.User != "" {
			contextLine += ", user=" + i.Payload.User
		}

		messages = append(
			messages,
			llm.Message{
				Role:  llm.RoleUser,
				Parts: []llm.Part{llm.TextPart{Text: "[" + contextLine + "]\n\n" + text}},
			},
		)
	}

	return messages
}

func extractReplyText(result *agent.Result) string {
	for i := len(result.Messages) - 1; i >= 0; i-- {
		if result.Messages[i].Role == llm.RoleAssistant {
			return result.Messages[i].Text()
		}
	}

	return ""
}

func cleanText(text string) string {
	return strings.TrimSpace(mentionRE.ReplaceAllString(text, ""))
}

func (r *singleAgentRegistry) Agent(_ string) (*agent.Agent, error) {
	return r.a, nil
}

func (h *Handler) lookupBinding(
	ctx context.Context,
	teamID, slackUserID string,
) (*coredata.SlackIdentityBinding, error) {
	if teamID == "" || slackUserID == "" {
		return nil, nil
	}

	binding, err := h.bindings.Lookup(ctx, teamID, slackUserID)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return binding, nil
}
