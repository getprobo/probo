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
	"strconv"
	"strings"
	"time"
)

type (
	Envelope struct {
		Type           EnvelopeType    `json:"type"`
		Challenge      string          `json:"challenge,omitempty"`
		TeamID         string          `json:"team_id,omitempty"`
		ContextTeamID  string          `json:"context_team_id,omitempty"`
		Authorizations []Authorization `json:"authorizations,omitempty"`
		EventID        string          `json:"event_id,omitempty"`
		EventTime      int64           `json:"event_time,omitempty"`
		Event          *EventBody      `json:"event,omitempty"`
	}

	Authorization struct {
		TeamID string `json:"team_id,omitempty"`
	}

	EventBody struct {
		Type            EventType      `json:"type"`
		Subtype         EventSubtype   `json:"subtype,omitempty"`
		User            string         `json:"user,omitempty"`
		UserTeam        string         `json:"user_team,omitempty"`
		SourceTeam      string         `json:"source_team,omitempty"`
		BotID           string         `json:"bot_id,omitempty"`
		Text            string         `json:"text,omitempty"`
		Channel         string         `json:"channel,omitempty"`
		ChannelType     ChannelType    `json:"channel_type,omitempty"`
		TS              string         `json:"ts,omitempty"`
		EventTS         string         `json:"event_ts,omitempty"`
		ThreadTS        string         `json:"thread_ts,omitempty"`
		Reaction        string         `json:"reaction,omitempty"`
		Item            *ReactionItem  `json:"item,omitempty"`
		Message         *EditedMsg     `json:"message,omitempty"`
		PreviousMessage *EditedMsg     `json:"previous_message,omitempty"`
		Tokens          *RevokedTokens `json:"tokens,omitempty"`
	}

	RevokedTokens struct {
		OAuth []string `json:"oauth,omitempty"`
		Bot   []string `json:"bot,omitempty"`
	}

	EditedMsg struct {
		Type     EditedMsgType `json:"type"`
		User     string        `json:"user,omitempty"`
		UserTeam string        `json:"user_team,omitempty"`
		BotID    string        `json:"bot_id,omitempty"`
		Text     string        `json:"text,omitempty"`
		TS       string        `json:"ts,omitempty"`
		ThreadTS string        `json:"thread_ts,omitempty"`
	}

	ReactionItem struct {
		Type    ReactionItemType `json:"type"`
		Channel string           `json:"channel,omitempty"`
		TS      string           `json:"ts,omitempty"`
	}
)

func (e Envelope) Time() *time.Time {
	if e.EventTime > 0 {
		return new(time.Unix(e.EventTime, 0).UTC())
	}

	if e.Event == nil {
		return nil
	}

	for _, raw := range []string{e.Event.EventTS, e.Event.TS} {
		if t, ok := parseSlackTimestamp(raw); ok {
			return new(t)
		}
	}

	return nil
}

func (e EventBody) RevokedBotUserIDs() []string {
	if e.Tokens == nil {
		return nil
	}

	return e.Tokens.Bot
}

func parseSlackTimestamp(raw string) (time.Time, bool) {
	if raw == "" {
		return time.Time{}, false
	}

	secondsPart, fracPart, _ := strings.Cut(raw, ".")
	seconds, err := strconv.ParseInt(secondsPart, 10, 64)
	if err != nil {
		return time.Time{}, false
	}

	var nsec int64

	if fracPart != "" {
		if len(fracPart) > 9 {
			fracPart = fracPart[:9]
		} else if len(fracPart) < 9 {
			fracPart += strings.Repeat("0", 9-len(fracPart))
		}

		nsec, err = strconv.ParseInt(fracPart, 10, 64)
		if err != nil {
			return time.Time{}, false
		}
	}

	return time.Unix(seconds, nsec).UTC(), true
}

func (e Envelope) InstallationTeamID() string {
	if e.ContextTeamID != "" {
		return e.ContextTeamID
	}

	for _, authorization := range e.Authorizations {
		if authorization.TeamID != "" {
			return authorization.TeamID
		}
	}

	return e.TeamID
}

func (e EventBody) ActorTeamID(installationTeamID string) string {
	if e.UserTeam != "" {
		return e.UserTeam
	}

	if e.SourceTeam != "" {
		return e.SourceTeam
	}

	return installationTeamID
}
