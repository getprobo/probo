// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package coredata

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
)

type OIDCState struct {
	ID           string       `db:"id"`
	Provider     OIDCProvider `db:"provider"`
	Nonce        string       `db:"nonce"`
	CodeVerifier string       `db:"code_verifier"`
	ContinueURL  string       `db:"continue_url"`
	CreatedAt    time.Time    `db:"created_at"`
	ExpiresAt    time.Time    `db:"expires_at"`
}

func (s *OIDCState) Insert(ctx context.Context, conn pg.Conn) error {
	query := `
INSERT INTO iam_oidc_states (id, provider, nonce, code_verifier, continue_url, created_at, expires_at)
VALUES (@id, @provider, @nonce, @code_verifier, @continue_url, @created_at, @expires_at)
`

	args := pgx.StrictNamedArgs{
		"id":            s.ID,
		"provider":      s.Provider,
		"nonce":         s.Nonce,
		"code_verifier": s.CodeVerifier,
		"continue_url":  s.ContinueURL,
		"created_at":    s.CreatedAt,
		"expires_at":    s.ExpiresAt,
	}

	_, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot insert oidc_state: %w", err)
	}

	return nil
}

func (s *OIDCState) LoadByIDForUpdate(ctx context.Context, conn pg.Conn, id string) error {
	query := `
SELECT id, provider, nonce, code_verifier, continue_url, created_at, expires_at
FROM iam_oidc_states
WHERE id = @id
FOR UPDATE
`

	rows, err := conn.Query(ctx, query, pgx.StrictNamedArgs{"id": id})
	if err != nil {
		return fmt.Errorf("cannot query oidc_state: %w", err)
	}

	state, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[OIDCState])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect oidc_state: %w", err)
	}

	*s = state
	return nil
}

func (s *OIDCState) Delete(ctx context.Context, conn pg.Conn) error {
	query := `DELETE FROM iam_oidc_states WHERE id = @id`

	_, err := conn.Exec(ctx, query, pgx.StrictNamedArgs{"id": s.ID})
	if err != nil {
		return fmt.Errorf("cannot delete oidc_state: %w", err)
	}

	return nil
}

func (s *OIDCState) DeleteExpired(ctx context.Context, conn pg.Conn, now time.Time) (int64, error) {
	query := `DELETE FROM iam_oidc_states WHERE expires_at < @now`

	result, err := conn.Exec(ctx, query, pgx.StrictNamedArgs{"now": now})
	if err != nil {
		return 0, fmt.Errorf("cannot delete expired oidc_states: %w", err)
	}

	return result.RowsAffected(), nil
}
