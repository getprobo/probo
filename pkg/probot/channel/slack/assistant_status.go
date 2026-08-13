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

	"go.gearno.de/kit/log"
)

const assistantWorkingStatus = "is working on your request..."

var assistantLoadingMessages = []string{
	"Looking that up…",
	"Checking your workspace…",
	"Drafting a reply…",
}

type assistantStatusSetter interface {
	SetStatus(
		ctx context.Context,
		channelID string,
		threadTS string,
		status string,
		loadingMessages []string,
	) error
}

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
