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

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/agent"
)

const assistantWorkingStatus = "is working on your request..."

var assistantLoadingMessages = []string{
	"Looking that up…",
	"Checking your workspace…",
	"Drafting a reply…",
}

type (
	assistantStatusSetter interface {
		SetStatus(
			ctx context.Context,
			channelID string,
			threadTS string,
			status string,
			loadingMessages []string,
		) error
	}

	// assistantStatusResolver defers Slack client resolution until the status
	// actually has to be cleared.
	assistantStatusResolver func(ctx context.Context) (assistantStatusSetter, error)

	// assistantStatusHook clears the thread status indicator when a turn ends.
	// send_message only queues durable delivery, so a successful tool call is
	// not proof that Slack has posted yet. Clearing on run end avoids leaving
	// the indicator spinning through delivery retries or failures until Slack's
	// own two-minute timeout. Suspended runs keep the indicator because the turn
	// resumes later.
	assistantStatusHook struct {
		agent.NoOpHooks

		logger    *log.Logger
		resolve   assistantStatusResolver
		channelID string
		threadTS  string
	}
)

var (
	_ agent.RunHooks = (*assistantStatusHook)(nil)
)

func assistantStatusThreadTS(threadTS, messageTS string) string {
	if threadTS != "" {
		return threadTS
	}

	return messageTS
}

func setAssistantWorkingStatus(
	ctx context.Context,
	logger *log.Logger,
	setter assistantStatusSetter,
	channelID, threadTS string,
) {
	if setter == nil || channelID == "" || threadTS == "" {
		return
	}

	err := setter.SetStatus(
		ctx,
		channelID,
		threadTS,
		assistantWorkingStatus,
		assistantLoadingMessages,
	)
	if err != nil {
		logAssistantStatusError(ctx, logger, err)
	}
}

func newAssistantStatusHook(
	logger *log.Logger,
	resolve assistantStatusResolver,
	channelID, threadTS string,
) *assistantStatusHook {
	return &assistantStatusHook{
		logger:    logger,
		resolve:   resolve,
		channelID: channelID,
		threadTS:  threadTS,
	}
}

func (h *assistantStatusHook) OnRunEnd(
	ctx context.Context,
	_ *agent.Agent,
	_ *agent.Result,
	err error,
) {
	// A suspended run resumes later, so the turn is still in progress.
	if _, ok := errors.AsType[*agent.SuspendedError](err); ok {
		return
	}

	h.clear(context.WithoutCancel(ctx))
}

func (h *assistantStatusHook) clear(ctx context.Context) {
	if h.resolve == nil || h.channelID == "" || h.threadTS == "" {
		return
	}

	setter, err := h.resolve(ctx)
	if err != nil {
		logAssistantStatusError(ctx, h.logger, err)

		return
	}

	if setter == nil {
		return
	}

	if err := setter.SetStatus(ctx, h.channelID, h.threadTS, "", nil); err != nil {
		logAssistantStatusError(ctx, h.logger, err)
	}
}

func logAssistantStatusError(ctx context.Context, logger *log.Logger, err error) {
	if logger == nil {
		return
	}

	if code := slackAPIErrorCode(err); code != "" {
		logger.WarnCtx(
			ctx,
			"cannot set Slack assistant status",
			log.String("error_code", code),
		)

		return
	}

	logger.WarnCtx(ctx, "cannot set Slack assistant status", log.Error(err))
}
