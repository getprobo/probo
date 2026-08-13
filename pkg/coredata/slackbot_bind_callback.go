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

package coredata

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
)

// SlackbotBindCallback stores an encrypted Slack response_url for bind
// confirmation. Rows are keyed by Slack team/user, not tenant.
type SlackbotBindCallback struct {
	TeamID               string    `db:"team_id"`
	UserID               string    `db:"user_id"`
	EncryptedResponseURL []byte    `db:"encrypted_response_url"`
	ExpiresAt            time.Time `db:"expires_at"`
	CreatedAt            time.Time `db:"created_at"`
}

func (c *SlackbotBindCallback) Upsert(ctx context.Context, conn pg.Querier) error {
	q := `
INSERT INTO slackbot_bind_callbacks (
	team_id,
	user_id,
	encrypted_response_url,
	expires_at,
	created_at
) VALUES (
	@team_id,
	@user_id,
	@encrypted_response_url,
	@expires_at,
	@created_at
)
ON CONFLICT (team_id, user_id) DO UPDATE
SET
	encrypted_response_url = EXCLUDED.encrypted_response_url,
	expires_at = EXCLUDED.expires_at,
	created_at = EXCLUDED.created_at
`

	_, err := conn.Exec(
		ctx,
		q,
		pgx.StrictNamedArgs{
			"team_id":                c.TeamID,
			"user_id":                c.UserID,
			"encrypted_response_url": c.EncryptedResponseURL,
			"expires_at":             c.ExpiresAt,
			"created_at":             c.CreatedAt,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot upsert Slackbot bind callback: %w", err)
	}

	return nil
}

func (c *SlackbotBindCallback) LoadByTeamAndUser(
	ctx context.Context,
	conn pg.Querier,
	teamID string,
	userID string,
) error {
	q := `
SELECT
	team_id,
	user_id,
	encrypted_response_url,
	expires_at,
	created_at
FROM slackbot_bind_callbacks
WHERE team_id = @team_id
	AND user_id = @user_id
`

	rows, err := conn.Query(
		ctx,
		q,
		pgx.StrictNamedArgs{
			"team_id": teamID,
			"user_id": userID,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot query Slackbot bind callback: %w", err)
	}

	callback, err := pgx.CollectExactlyOneRow(
		rows,
		pgx.RowToStructByName[SlackbotBindCallback],
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect Slackbot bind callback: %w", err)
	}

	*c = callback

	return nil
}

func DeleteSlackbotBindCallback(
	ctx context.Context,
	conn pg.Querier,
	teamID string,
	userID string,
) error {
	_, err := conn.Exec(
		ctx,
		`
DELETE FROM slackbot_bind_callbacks
WHERE team_id = @team_id
	AND user_id = @user_id
`,
		pgx.StrictNamedArgs{
			"team_id": teamID,
			"user_id": userID,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot delete Slackbot bind callback: %w", err)
	}

	return nil
}

func DeleteSlackbotBindCallbacksByTeamID(
	ctx context.Context,
	conn pg.Querier,
	teamID string,
) error {
	_, err := conn.Exec(
		ctx,
		`
DELETE FROM slackbot_bind_callbacks
WHERE team_id = @team_id
`,
		pgx.StrictNamedArgs{"team_id": teamID},
	)
	if err != nil {
		return fmt.Errorf("cannot delete Slackbot bind callbacks: %w", err)
	}

	return nil
}
