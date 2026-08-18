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
	"strings"

	"go.gearno.de/kit/log"
)

const threadTranscriptMaxMessages = 1000

type threadCollector interface {
	ListThreadReplies(ctx context.Context, channel string, threadTS string) ([]ThreadReply, error)
}

var (
	_ threadCollector = (*Client)(nil)
)

func formatThreadTranscript(replies []ThreadReply, botUserID, botID string) string {
	return formatKeptThreadTranscript(keptThreadReplies(replies, botUserID, botID), botUserID)
}

func keptThreadReplies(replies []ThreadReply, botUserID, botID string) []ThreadReply {
	filtered := make([]ThreadReply, 0, len(replies))
	for _, reply := range replies {
		if reply.Subtype == "message_deleted" {
			continue
		}

		text := cleanText(reply.Text)
		if text == "" {
			continue
		}

		if isForeignBotReply(reply, botUserID, botID) {
			continue
		}

		filtered = append(
			filtered,
			ThreadReply{
				User:  reply.User,
				BotID: reply.BotID,
				Text:  text,
				TS:    reply.TS,
			},
		)
	}

	if len(filtered) > threadTranscriptMaxMessages {
		kept := make([]ThreadReply, 0, threadTranscriptMaxMessages)
		kept = append(kept, filtered[0])
		kept = append(kept, filtered[len(filtered)-(threadTranscriptMaxMessages-1):]...)
		filtered = kept
	}

	return filtered
}

func formatKeptThreadTranscript(filtered []ThreadReply, botUserID string) string {
	if len(filtered) == 0 {
		return ""
	}

	var builder strings.Builder
	builder.WriteString("Thread:\n")

	for _, reply := range filtered {
		appendTranscriptLine(&builder, threadSpeaker(reply.User, reply.BotID, botUserID), reply.Text)
	}

	return strings.TrimSpace(builder.String())
}

func isForeignBotReply(reply ThreadReply, botUserID, botID string) bool {
	if reply.BotID == "" {
		return false
	}

	if reply.User != "" {
		return reply.User != botUserID
	}

	return botID == "" || reply.BotID != botID
}

func threadSpeaker(user, replyBotID, botUserID string) string {
	speaker := user
	if speaker == "" && replyBotID != "" {
		speaker = botUserID
	}

	if speaker == "" {
		speaker = "unknown"
	}

	return speaker
}

func appendTranscriptLine(builder *strings.Builder, speaker, text string) {
	builder.WriteString("<@")
	builder.WriteString(speaker)
	builder.WriteString(">: ")
	builder.WriteString(text)
	builder.WriteString("\n")
}

func threadReplyHasTS(replies []ThreadReply, ts string) bool {
	if ts == "" {
		return false
	}

	for _, reply := range replies {
		if reply.TS == ts {
			return true
		}
	}

	return false
}

func appendTriggeringEventIfMissing(
	transcript string,
	kept []ThreadReply,
	event EventBody,
	botUserID string,
) string {
	fallback := cleanText(event.Text)
	if transcript == "" {
		return fallback
	}

	if threadReplyHasTS(kept, event.TS) || fallback == "" {
		return transcript
	}

	return transcript + "\n<@" + threadSpeaker(event.User, event.BotID, botUserID) + ">: " + fallback
}

func (s *Service) collectThreadTranscript(
	ctx context.Context,
	slackClient *Client,
	event EventBody,
	botUserID string,
) string {
	fallback := cleanText(event.Text)
	if event.ChannelType == ChannelTypeIM {
		return fallback
	}

	target := replyTargetFor(event)
	if slackClient == nil || target.channel == "" || target.threadTS == "" {
		return fallback
	}

	replies, err := slackClient.ListThreadReplies(ctx, target.channel, target.threadTS)
	if err != nil && len(replies) == 0 {
		if isThreadCollectionFallbackError(err) {
			s.logger.WarnCtx(
				ctx,
				"cannot collect Slack thread replies; using mention text",
				log.String("error_code", slackAPIErrorCode(err)),
			)

			return fallback
		}

		s.logger.WarnCtx(
			ctx,
			"cannot collect Slack thread replies; using mention text",
			log.Error(err),
		)

		return fallback
	}

	if err != nil {
		s.logger.WarnCtx(
			ctx,
			"cannot collect remaining Slack thread replies; using collected messages",
			log.String("error_code", slackAPIErrorCode(err)),
		)
	}

	botID := resolveInstalledBotID(replies, botUserID)
	kept := keptThreadReplies(replies, botUserID, botID)
	transcript := formatKeptThreadTranscript(kept, botUserID)

	return appendTriggeringEventIfMissing(transcript, kept, event, botUserID)
}

func resolveInstalledBotID(replies []ThreadReply, botUserID string) string {
	if botUserID == "" {
		return ""
	}

	for _, reply := range replies {
		if reply.User == botUserID && reply.BotID != "" {
			return reply.BotID
		}
	}

	return ""
}

func isThreadCollectionFallbackError(err error) bool {
	switch slackAPIErrorCode(err) {
	case "missing_scope", "not_in_channel", "channel_not_found", "thread_not_found":
		return true
	}

	return false
}

func slackAPIErrorCode(err error) string {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Code
	}

	return ""
}
