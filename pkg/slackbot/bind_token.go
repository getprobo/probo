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

import (
	"fmt"
	"time"

	"go.probo.inc/probo/pkg/statelesstoken"
)

const (
	TokenTypeSlackIdentityBind = "slack_identity_bind"

	bindTokenExpiry = 30 * time.Minute
)

type (
	BindTokenData struct {
		TeamID      string `json:"team_id"`
		SlackUserID string `json:"slack_user_id"`
	}
)

func newBindToken(secret string, teamID, slackUserID string) (string, error) {
	return statelesstoken.NewToken(
		secret,
		TokenTypeSlackIdentityBind,
		bindTokenExpiry,
		BindTokenData{
			TeamID:      teamID,
			SlackUserID: slackUserID,
		},
	)
}

func validateBindToken(secret, tokenString string) (*statelesstoken.Payload[BindTokenData], error) {
	payload, err := statelesstoken.ValidateToken[BindTokenData](
		secret,
		TokenTypeSlackIdentityBind,
		tokenString,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot validate slack bind token: %w", err)
	}

	return payload, nil
}
