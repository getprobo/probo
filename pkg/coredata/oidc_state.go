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
	"go.probo.inc/probo/pkg/gid"
)

type OIDCState struct {
	ID             string       `db:"id"`
	Provider       OIDCProvider `db:"provider"`
	Nonce          string       `db:"nonce"`
	CodeVerifier   string       `db:"code_verifier"`
	ContinueURL    string       `db:"continue_url"`
	OrganizationID *gid.GID     `db:"organization_id"`
	CreatedAt      time.Time    `db:"created_at"`
	ExpiresAt      time.Time    `db:"expires_at"`
}

func (s *OIDCState) Insert(ctx context.Context, conn pg.Tx) error {
	query := `
INSERT INTO iam_oidc_states (id, provider, nonce, code_verifier, continue_url, organization_id, created_at, expires_at)
VALUES (@id, @provider, @nonce, @code_verifier, @continue_url, @organization_id, @created_at, @expires_at)
`

	args := pgx.StrictNamedArgs{
		"id":              s.ID,
		"provider":        s.Provider,
		"nonce":           s.Nonce,
		"code_verifier":   s.CodeVerifier,
		"continue_url":    s.ContinueURL,
		"organization_id": s.OrganizationID,
		"created_at":      s.CreatedAt,
		"expires_at":      s.ExpiresAt,
	}

	_, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot insert oidc_state: %w", err)
	}

	return nil
}

func (s *OIDCState) LoadByIDForUpdate(ctx context.Context, conn pg.Tx, id string) error {
	query := `
SELECT id, provider, nonce, code_verifier, continue_url, organization_id, created_at, expires_at
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

func (s *OIDCState) Delete(ctx context.Context, conn pg.Tx) error {
	query := `DELETE FROM iam_oidc_states WHERE id = @id`

	_, err := conn.Exec(ctx, query, pgx.StrictNamedArgs{"id": s.ID})
	if err != nil {
		return fmt.Errorf("cannot delete oidc_state: %w", err)
	}

	return nil
}

func (s *OIDCState) DeleteExpired(ctx context.Context, conn pg.Tx, now time.Time) (int64, error) {
	query := `DELETE FROM iam_oidc_states WHERE expires_at < @now`

	result, err := conn.Exec(ctx, query, pgx.StrictNamedArgs{"now": now})
	if err != nil {
		return 0, fmt.Errorf("cannot delete expired oidc_states: %w", err)
	}

	return result.RowsAffected(), nil
}
