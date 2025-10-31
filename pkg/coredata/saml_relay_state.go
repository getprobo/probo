// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"go.probo.inc/probo/pkg/gid"
	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
)

type SAMLRelayState struct {
	Token          string    `db:"token"`
	OrganizationID gid.GID   `db:"organization_id"`
	SAMLConfigID   gid.GID   `db:"saml_config_id"`
	RequestID      string    `db:"request_id"`
	CreatedAt      time.Time `db:"created_at"`
	ExpiresAt      time.Time `db:"expires_at"`
}

type ErrRelayStateNotFound struct {
	Token string
}

func (e ErrRelayStateNotFound) Error() string {
	return "relay state token not found or invalid"
}

type ErrRelayStateExpired struct {
	Token     string
	ExpiresAt time.Time
}

func (e ErrRelayStateExpired) Error() string {
	return fmt.Sprintf("relay state token expired at %v", e.ExpiresAt)
}

func GenerateSecureToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("cannot generate random token: %w", err)
	}

	token := base64.URLEncoding.EncodeToString(b)
	return token, nil
}

func (s *SAMLRelayState) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	query := `
INSERT INTO auth_saml_relay_states (token, tenant_id, organization_id, saml_config_id, request_id, created_at, expires_at)
VALUES (@token, @tenant_id, @organization_id, @saml_config_id, @request_id, @created_at, @expires_at)
`

	args := pgx.NamedArgs{
		"token":           s.Token,
		"tenant_id":       scope.GetTenantID(),
		"organization_id": s.OrganizationID,
		"saml_config_id":  s.SAMLConfigID,
		"request_id":      s.RequestID,
		"created_at":      s.CreatedAt,
		"expires_at":      s.ExpiresAt,
	}

	_, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot insert saml_relay_state: %w", err)
	}

	return nil
}

func (s *SAMLRelayState) Load(
	ctx context.Context,
	conn pg.Conn,
	token string,
) error {
	query := `
SELECT token, organization_id, saml_config_id, request_id, created_at, expires_at
FROM auth_saml_relay_states
WHERE token = @token
LIMIT 1
`

	rows, err := conn.Query(ctx, query, pgx.NamedArgs{"token": token})
	if err != nil {
		return fmt.Errorf("cannot query saml_relay_states: %w", err)
	}

	state, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[SAMLRelayState])
	if err == pgx.ErrNoRows {
		return ErrRelayStateNotFound{Token: token}
	}
	if err != nil {
		return fmt.Errorf("cannot collect saml_relay_state: %w", err)
	}

	*s = state
	return nil
}

func (s *SAMLRelayState) IsExpired(now time.Time) bool {
	return now.After(s.ExpiresAt) || now.Equal(s.ExpiresAt)
}

func (s *SAMLRelayState) Delete(
	ctx context.Context,
	conn pg.Conn,
) error {
	query := `
DELETE FROM auth_saml_relay_states
WHERE token = @token
`

	_, err := conn.Exec(ctx, query, pgx.NamedArgs{"token": s.Token})
	if err != nil {
		return fmt.Errorf("cannot delete saml_relay_state: %w", err)
	}

	return nil
}

func DeleteExpiredSAMLRelayStates(ctx context.Context, conn pg.Conn, now time.Time) (int64, error) {
	query := `
DELETE FROM auth_saml_relay_states
WHERE expires_at < @now
`

	result, err := conn.Exec(ctx, query, pgx.NamedArgs{"now": now})
	if err != nil {
		return 0, fmt.Errorf("cannot delete expired saml_relay_states: %w", err)
	}

	return result.RowsAffected(), nil
}
