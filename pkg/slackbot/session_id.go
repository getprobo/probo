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

package slackbot

import "strings"

// sessionIDFor builds a workspace-scoped conversation key so agents cannot
// collide across teams, channels, or users that share a thread.
//
// Formats:
//   - IM:    {teamID}:{channel}
//   - other: {teamID}:{channel}:{threadTS|ts}:{slackUserID}
func sessionIDFor(teamID, channel string, channelType ChannelType, threadTS, ts, slackUserID string) string {
	if teamID == "" || channel == "" {
		return ""
	}

	if channelType == ChannelTypeIM {
		return teamID + ":" + channel
	}

	if slackUserID == "" {
		return ""
	}

	conversation := threadTS
	if conversation == "" {
		conversation = ts
	}
	if conversation == "" {
		return ""
	}

	return teamID + ":" + channel + ":" + conversation + ":" + slackUserID
}

func teamIDFromSessionID(sessionID string) string {
	if i := strings.Index(sessionID, ":"); i > 0 {
		return sessionID[:i]
	}

	return ""
}
