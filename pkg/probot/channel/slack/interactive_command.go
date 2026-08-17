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
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/probot/identitybinding"
)

type (
	InteractivePayload struct {
		ResponseURL string `json:"response_url"`
		Team        struct {
			ID string `json:"id"`
		} `json:"team"`
		User struct {
			ID     string `json:"id"`
			TeamID string `json:"team_id"`
		} `json:"user"`
		Actions []struct {
			ActionID       string `json:"action_id"`
			ActionTS       string `json:"action_ts"`
			Value          string `json:"value"`
			SelectedOption struct {
				Value string `json:"value"`
			} `json:"selected_option"`
		} `json:"actions"`
		Container struct {
			MessageTS string `json:"message_ts"`
			ChannelID string `json:"channel_id"`
		} `json:"container"`
	}

	InteractiveResponse struct {
		Success bool   `json:"success"`
		Message string `json:"message,omitempty"`
	}

	InteractiveCommandInbox struct {
		pg            *pg.Client
		encryptionKey cipher.EncryptionKey
	}
)

func NewInteractiveCommandInbox(
	pgClient *pg.Client,
	encryptionKey cipher.EncryptionKey,
) *InteractiveCommandInbox {
	return &InteractiveCommandInbox{
		pg:            pgClient,
		encryptionKey: encryptionKey,
	}
}

func DecodeInteractivePayload(raw []byte) (InteractivePayload, error) {
	var payload InteractivePayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return InteractivePayload{}, fmt.Errorf("cannot decode Slack interactive payload: %w", err)
	}

	if payload.Team.ID == "" || payload.User.ID == "" {
		return InteractivePayload{}, fmt.Errorf("slack interactive payload is missing team or user")
	}

	return payload, nil
}

func (p InteractivePayload) ActorSubject() identitybinding.Subject {
	actorTeamID := p.User.TeamID
	if actorTeamID == "" {
		actorTeamID = p.Team.ID
	}

	return IdentitySubject(actorTeamID, p.User.ID)
}

func (i *InteractiveCommandInbox) Enqueue(
	ctx context.Context,
	verifiedPayload []byte,
) (bool, error) {
	encryptedPayload, err := cipher.Encrypt(verifiedPayload, i.encryptionKey)
	if err != nil {
		return false, fmt.Errorf("cannot encrypt Slack interactive payload: %w", err)
	}

	digest := sha256.Sum256(verifiedPayload)
	command := coredata.NewSlackbotInteractiveCommand(digest[:], encryptedPayload)

	var inserted bool

	err = i.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var err error

			inserted, err = command.Insert(ctx, conn)
			if err != nil {
				return fmt.Errorf("cannot insert Slack interactive command: %w", err)
			}

			if inserted {
				return nil
			}

			inserted, err = command.ResetDeadLetteredByRequestDigest(ctx, conn, time.Now())
			if err != nil {
				return fmt.Errorf("cannot reset Slack interactive command: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return false, fmt.Errorf("cannot persist Slack interactive command: %w", err)
	}

	return inserted, nil
}
