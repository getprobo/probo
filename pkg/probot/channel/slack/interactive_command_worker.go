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
	"encoding/hex"
	"errors"
	"fmt"
	"maps"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/bot"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/probot"
	"go.probo.inc/probo/pkg/probot/identitybinding"
)

type (
	InteractiveInstallationResolver interface {
		ClientByTeamID(ctx context.Context, teamID string) (*Client, *coredata.SlackbotInstallation, error)
	}

	InteractiveMessageResolver interface {
		GetInitialByChannelAndTS(
			ctx context.Context,
			organizationID gid.GID,
			channelID string,
			messageTS string,
		) (*bot.DeliveredMessage, error)
	}

	interactiveCommandHandler struct {
		pg            *pg.Client
		encryptionKey cipher.EncryptionKey
		installations InteractiveInstallationResolver
		bindings      identitybinding.Gate
		messages      InteractiveMessageResolver
		capabilities  *probot.CapabilityRegistry
		replies       interactiveReplyPoster
		logger        *log.Logger
		staleAfter    time.Duration
		retryBase     time.Duration
		retryMax      time.Duration
		now           func() time.Time
	}
)

const (
	defaultInteractiveCommandMaxConcurrency = 4
	defaultInteractiveCommandStaleAfter     = 5 * time.Minute
	defaultInteractiveCommandRetryBase      = time.Second
	defaultInteractiveCommandRetryMax       = 5 * time.Minute
)

var (
	_ worker.Handler[coredata.SlackbotInteractiveCommand] = (*interactiveCommandHandler)(nil)
	_ worker.StaleRecoverer                               = (*interactiveCommandHandler)(nil)
)

func NewInteractiveCommandWorker(
	pgClient *pg.Client,
	encryptionKey cipher.EncryptionKey,
	installations InteractiveInstallationResolver,
	bindings identitybinding.Gate,
	messages InteractiveMessageResolver,
	capabilities *probot.CapabilityRegistry,
	logger *log.Logger,
	opts ...worker.Option,
) *worker.Worker[coredata.SlackbotInteractiveCommand] {
	h := &interactiveCommandHandler{
		pg:            pgClient,
		encryptionKey: encryptionKey,
		installations: installations,
		bindings:      bindings,
		messages:      messages,
		capabilities:  capabilities,
		replies:       newResponseURLPoster(logger),
		logger:        logger,
		staleAfter:    defaultInteractiveCommandStaleAfter,
		retryBase:     defaultInteractiveCommandRetryBase,
		retryMax:      defaultInteractiveCommandRetryMax,
		now:           time.Now,
	}

	workerOpts := append(
		[]worker.Option{worker.WithMaxConcurrency(defaultInteractiveCommandMaxConcurrency)},
		opts...,
	)

	return worker.New(
		"slackbot-interactive-command-worker",
		h,
		logger,
		workerOpts...,
	)
}

func (h *interactiveCommandHandler) Claim(
	ctx context.Context,
) (coredata.SlackbotInteractiveCommand, error) {
	var command coredata.SlackbotInteractiveCommand

	err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			return command.ClaimNextForUpdateSkipLocked(ctx, tx, h.now())
		},
	)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return coredata.SlackbotInteractiveCommand{}, worker.ErrNoTask
		}

		return coredata.SlackbotInteractiveCommand{}, fmt.Errorf(
			"cannot claim Slack interactive command: %w",
			err,
		)
	}

	return command, nil
}

func (h *interactiveCommandHandler) Process(
	ctx context.Context,
	command coredata.SlackbotInteractiveCommand,
) error {
	logger := h.logger.With(log.String("command_id", command.ID.String()))
	processingErr := h.dispatch(ctx, &command)

	persistErr := h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			now := h.now()
			applyInteractiveCommandOutcome(
				&command,
				processingErr,
				now,
				h.retryDelay(command.AttemptCount),
			)

			return command.UpdateProcessingState(ctx, conn)
		},
	)
	if persistErr != nil {
		return fmt.Errorf("cannot persist Slack interactive command outcome: %w", persistErr)
	}

	if processingErr != nil {
		logger.ErrorCtx(
			ctx,
			"cannot process Slack interactive command",
			log.Int("attempt_count", command.AttemptCount),
			log.Bool("dead_lettered", command.DeadLetteredAt != nil),
			log.Error(processingErr),
		)

		return processingErr
	}

	fields := []log.Attr{log.Int("attempt_count", command.AttemptCount)}
	if command.OrganizationID != nil {
		fields = append(fields, log.String("organization_id", command.OrganizationID.String()))
	}

	logger.InfoCtx(ctx, "processed Slack interactive command", fields...)

	return nil
}

func (h *interactiveCommandHandler) dispatch(
	ctx context.Context,
	command *coredata.SlackbotInteractiveCommand,
) error {
	raw, err := cipher.Decrypt(command.EncryptedPayload, h.encryptionKey)
	if err != nil {
		return permanent(fmt.Errorf("cannot decrypt command payload: %w", err))
	}

	payload, err := DecodeInteractivePayload(raw)
	if err != nil {
		return permanent(err)
	}

	if len(payload.Actions) == 0 {
		return nil
	}

	action := payload.Actions[0]
	if action.Value == "" && action.SelectedOption.Value == "" {
		return nil
	}

	if action.ActionID == "" ||
		action.ActionTS == "" ||
		payload.Container.ChannelID == "" ||
		payload.Container.MessageTS == "" {
		return permanent(fmt.Errorf("slack interactive command is incomplete"))
	}

	if h.installations == nil || h.bindings == nil || h.messages == nil || h.capabilities == nil {
		return fmt.Errorf("slack interactive command dependencies are unavailable")
	}

	_, installation, err := h.installations.ClientByTeamID(ctx, payload.Team.ID)
	if err != nil {
		if errors.Is(err, ErrSlackbotNotInstalled) ||
			errors.Is(err, coredata.ErrResourceNotFound) {
			return permanent(fmt.Errorf("slack installation is unavailable: %w", err))
		}

		return fmt.Errorf("cannot reload Slack installation: %w", err)
	}

	command.OrganizationID = &installation.OrganizationID

	binding, err := h.bindings.Lookup(ctx, payload.ActorSubject())
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			h.replyInteractiveFailure(ctx, payload.ResponseURL, bindRequiredText)

			return permanent(fmt.Errorf("slack identity binding is missing or revoked: %w", err))
		}

		return fmt.Errorf("cannot recheck Slack identity binding: %w", err)
	}

	delivered, err := h.messages.GetInitialByChannelAndTS(
		ctx,
		installation.OrganizationID,
		payload.Container.ChannelID,
		payload.Container.MessageTS,
	)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return permanent(fmt.Errorf("delivered Slack message is unavailable: %w", err))
		}

		return fmt.Errorf("cannot resolve delivered Slack message: %w", err)
	}

	if delivered.Message.OrganizationID != installation.OrganizationID {
		return permanent(fmt.Errorf("slack workspace does not own delivered message"))
	}

	attributes := make(map[string]any, len(delivered.Message.Attributes))
	maps.Copy(attributes, delivered.Message.Attributes)

	probotAction := probot.Action{
		ID:               action.ActionID,
		Value:            action.Value,
		SelectedValue:    action.SelectedOption.Value,
		DeduplicationKey: hex.EncodeToString(command.RequestDigest),
		ResponseToken:    payload.ResponseURL,
		ActorIdentityID:  binding.IdentityID,
		Message: bot.Message{
			ID:             delivered.Message.ID,
			OrganizationID: delivered.Message.OrganizationID,
			Type:           delivered.Message.Type,
			Attributes:     attributes,
		},
	}
	probotAction, err = h.capabilities.NormalizeActionAlias(probotAction)
	if err != nil {
		h.replyInteractiveFailure(ctx, payload.ResponseURL, interactiveFailedText)

		return permanent(
			fmt.Errorf("cannot normalize Slack interactive action: %w", err),
		)
	}

	_, err = h.capabilities.HandleAction(ctx, probotAction)
	if err != nil {
		if errors.Is(err, probot.ErrCapabilityForbidden) ||
			errors.Is(err, probot.ErrCapabilityInvalidInput) ||
			errors.Is(err, probot.ErrCapabilityNotFound) {
			h.replyInteractiveFailure(
				ctx,
				payload.ResponseURL,
				interactiveFailureText(err),
			)

			return permanent(err)
		}

		return fmt.Errorf("cannot dispatch Slack capability action: %w", err)
	}

	return nil
}

func (h *interactiveCommandHandler) replyInteractiveFailure(
	ctx context.Context,
	responseURL string,
	text string,
) {
	if h.replies == nil || responseURL == "" {
		return
	}

	if err := h.replies.PostEphemeralReply(ctx, responseURL, text); err != nil && h.logger != nil {
		h.logger.ErrorCtx(ctx, "cannot post Slack interactive error", log.Error(err))
	}
}

func interactiveFailureText(err error) string {
	if errors.Is(err, probot.ErrCapabilityForbidden) {
		return interactiveForbiddenText
	}

	return interactiveFailedText
}

func (h *interactiveCommandHandler) RecoverStale(ctx context.Context) error {
	err := h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return coredata.ResetStaleSlackbotInteractiveCommands(
				ctx,
				conn,
				h.now(),
				h.staleAfter,
			)
		},
	)
	if err != nil {
		return fmt.Errorf("cannot recover stale Slack interactive commands: %w", err)
	}

	return nil
}

func (h *interactiveCommandHandler) retryDelay(attempt int) time.Duration {
	return probot.ExponentialRetryDelay(attempt, h.retryBase, h.retryMax)
}

func applyInteractiveCommandOutcome(
	command *coredata.SlackbotInteractiveCommand,
	processingErr error,
	now time.Time,
	retryDelay time.Duration,
) {
	command.ProcessingStartedAt = nil

	command.UpdatedAt = now
	if processingErr == nil {
		command.ProcessedAt = &now
		command.NextAttemptAt = nil
		command.LastError = nil
		command.DeadLetteredAt = nil

		return
	}

	command.LastError = new(processingErr.Error())
	if isPermanent(processingErr) ||
		command.AttemptCount >= command.MaxAttempts {
		command.AttemptCount = command.MaxAttempts
		command.NextAttemptAt = nil
		command.DeadLetteredAt = &now

		return
	}

	nextAttemptAt := now.Add(retryDelay)
	command.NextAttemptAt = &nextAttemptAt
}
