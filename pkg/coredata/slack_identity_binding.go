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
	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type SlackIdentityBinding struct {
	ID          gid.GID   `db:"id"`
	TeamID      string    `db:"team_id"`
	SlackUserID string    `db:"slack_user_id"`
	IdentityID  gid.GID   `db:"identity_id"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func (b *SlackIdentityBinding) LoadBySlackUser(
	ctx context.Context,
	conn pg.Querier,
	teamID, slackUserID string,
) error {
	q := `
SELECT
    id,
    team_id,
    slack_user_id,
    identity_id,
    created_at,
    updated_at
FROM
    slack_identity_bindings
WHERE
    team_id = @team_id
    AND slack_user_id = @slack_user_id
LIMIT 1;
`

	args := pgx.StrictNamedArgs{
		"team_id":       teamID,
		"slack_user_id": slackUserID,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query slack identity binding: %w", err)
	}

	binding, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[SlackIdentityBinding])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect slack identity binding: %w", err)
	}

	*b = binding

	return nil
}

func (b *SlackIdentityBinding) LoadByIdentityAndTeam(
	ctx context.Context,
	conn pg.Querier,
	identityID gid.GID,
	teamID string,
) error {
	q := `
SELECT
    id,
    team_id,
    slack_user_id,
    identity_id,
    created_at,
    updated_at
FROM
    slack_identity_bindings
WHERE
    identity_id = @identity_id
    AND team_id = @team_id
LIMIT 1;
`

	args := pgx.StrictNamedArgs{
		"identity_id": identityID,
		"team_id":     teamID,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query slack identity binding: %w", err)
	}

	binding, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[SlackIdentityBinding])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect slack identity binding: %w", err)
	}

	*b = binding

	return nil
}

func (b *SlackIdentityBinding) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	bindingID gid.GID,
) error {
	q := `
SELECT
    id,
    team_id,
    slack_user_id,
    identity_id,
    created_at,
    updated_at
FROM
    slack_identity_bindings
WHERE
    id = @id
LIMIT 1;
`

	args := pgx.StrictNamedArgs{"id": bindingID}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query slack identity binding: %w", err)
	}

	binding, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[SlackIdentityBinding])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect slack identity binding: %w", err)
	}

	*b = binding

	return nil
}

func (b *SlackIdentityBinding) Insert(
	ctx context.Context,
	conn pg.Querier,
) error {
	q := `
INSERT INTO slack_identity_bindings (
    id,
    team_id,
    slack_user_id,
    identity_id,
    created_at,
    updated_at
)
VALUES (
    @id,
    @team_id,
    @slack_user_id,
    @identity_id,
    @created_at,
    @updated_at
);
`

	args := pgx.StrictNamedArgs{
		"id":            b.ID,
		"team_id":       b.TeamID,
		"slack_user_id": b.SlackUserID,
		"identity_id":   b.IdentityID,
		"created_at":    b.CreatedAt,
		"updated_at":    b.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "slack_identity_bindings_team_id_slack_user_id_key",
				"slack_identity_bindings_identity_id_team_id_key":
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot insert slack identity binding: %w", err)
	}

	return nil
}

func (b *SlackIdentityBinding) Delete(
	ctx context.Context,
	conn pg.Querier,
) error {
	q := `
DELETE FROM slack_identity_bindings
WHERE
    id = @id;
`

	args := pgx.StrictNamedArgs{"id": b.ID}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete slack identity binding: %w", err)
	}

	return nil
}
