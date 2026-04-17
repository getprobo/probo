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
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/uri"
)

type (
	OAuth2Consent struct {
		ID                  gid.GID                   `db:"id"`
		IdentityID          gid.GID                   `db:"identity_id"`
		SessionID           gid.GID                   `db:"session_id"`
		ClientID            gid.GID                   `db:"client_id"`
		Scopes              OAuth2Scopes              `db:"scopes"`
		RedirectURI         *uri.URI                  `db:"redirect_uri"`
		CodeChallenge       string                    `db:"code_challenge"`
		CodeChallengeMethod OAuth2CodeChallengeMethod `db:"code_challenge_method"`
		Nonce               string                    `db:"nonce"`
		State               string                    `db:"state"`
		DeviceCodeID        *gid.GID                  `db:"device_code_id"`
		Approved            bool                      `db:"approved"`
		CreatedAt           time.Time                 `db:"created_at"`
		UpdatedAt           time.Time                 `db:"updated_at"`
	}

	OAuth2Consents []*OAuth2Consent
)

func (c *OAuth2Consent) CursorKey(orderBy OAuth2ConsentOrderField) page.CursorKey {
	switch orderBy {
	case OAuth2ConsentOrderFieldCreatedAt:
		return page.NewCursorKey(c.ID, c.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (c *OAuth2Consent) AuthorizationAttributes(ctx context.Context, conn pg.Querier) (map[string]string, error) {
	q := `
SELECT
	identity_id,
	session_id
FROM
	iam_oauth2_consents
WHERE
	id = $1
LIMIT 1;
`

	var identityID, sessionID gid.GID
	if err := conn.QueryRow(ctx, q, c.ID).Scan(&identityID, &sessionID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}

		return nil, fmt.Errorf("cannot query oauth2_consent authorization attributes: %w", err)
	}

	return map[string]string{
		"identity_id": identityID.String(),
		"session_id":  sessionID.String(),
	}, nil
}

func (c *OAuth2Consent) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	id gid.GID,
) error {
	q := `
SELECT
	id,
	identity_id,
	session_id,
	client_id,
	scopes,
	redirect_uri,
	code_challenge,
	code_challenge_method,
	nonce,
	state,
	device_code_id,
	approved,
	created_at,
	updated_at
FROM
	iam_oauth2_consents
WHERE
	id = @id
LIMIT 1;
`

	rows, err := conn.Query(ctx, q, pgx.StrictNamedArgs{"id": id})
	if err != nil {
		return fmt.Errorf("cannot query oauth2_consent: %w", err)
	}

	consent, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[OAuth2Consent])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect oauth2_consent: %w", err)
	}

	*c = consent
	return nil
}

func (c *OAuth2Consent) LoadByIDForSession(
	ctx context.Context,
	conn pg.Querier,
	id gid.GID,
	identityID gid.GID,
	sessionID gid.GID,
) error {
	q := `
SELECT
	id,
	identity_id,
	session_id,
	client_id,
	scopes,
	redirect_uri,
	code_challenge,
	code_challenge_method,
	nonce,
	state,
	device_code_id,
	approved,
	created_at,
	updated_at
FROM
	iam_oauth2_consents
WHERE
	id = @id
	AND identity_id = @identity_id
	AND session_id = @session_id
LIMIT 1;
`

	rows, err := conn.Query(
		ctx,
		q,
		pgx.StrictNamedArgs{
			"id":          id,
			"identity_id": identityID,
			"session_id":  sessionID,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot query oauth2_consent: %w", err)
	}

	consent, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[OAuth2Consent])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect oauth2_consent: %w", err)
	}

	*c = consent
	return nil
}

func (c *OAuth2Consent) LoadByIDForSessionForUpdate(
	ctx context.Context,
	conn pg.Querier,
	id gid.GID,
	identityID gid.GID,
	sessionID gid.GID,
) error {
	q := `
SELECT
	id,
	identity_id,
	session_id,
	client_id,
	scopes,
	redirect_uri,
	code_challenge,
	code_challenge_method,
	nonce,
	state,
	device_code_id,
	approved,
	created_at,
	updated_at
FROM
	iam_oauth2_consents
WHERE
	id = @id
	AND identity_id = @identity_id
	AND session_id = @session_id
LIMIT 1
FOR UPDATE;
`

	rows, err := conn.Query(
		ctx,
		q,
		pgx.StrictNamedArgs{
			"id":          id,
			"identity_id": identityID,
			"session_id":  sessionID,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot query oauth2_consent: %w", err)
	}

	consent, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[OAuth2Consent])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect oauth2_consent: %w", err)
	}

	*c = consent
	return nil
}

func (c *OAuth2Consent) LoadMatchingConsent(
	ctx context.Context,
	conn pg.Querier,
	identityID gid.GID,
	clientID gid.GID,
	scopes OAuth2Scopes,
) error {
	q := `
SELECT
	id,
	identity_id,
	session_id,
	client_id,
	scopes,
	redirect_uri,
	code_challenge,
	code_challenge_method,
	nonce,
	state,
	device_code_id,
	approved,
	created_at,
	updated_at
FROM
	iam_oauth2_consents
WHERE
	identity_id = @identity_id
	AND client_id = @client_id
	AND approved = TRUE
	AND scopes @> @scopes
	AND scopes <@ @scopes
LIMIT 1;
`

	rows, err := conn.Query(
		ctx,
		q,
		pgx.StrictNamedArgs{
			"identity_id": identityID,
			"client_id":   clientID,
			"scopes":      scopes,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot query oauth2_consent: %w", err)
	}

	consent, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[OAuth2Consent])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect oauth2_consent: %w", err)
	}

	*c = consent
	return nil
}

func (c *OAuth2Consent) Insert(ctx context.Context, conn pg.Tx) error {
	q := `
INSERT INTO iam_oauth2_consents (
	id,
	identity_id,
	session_id,
	client_id,
	scopes,
	redirect_uri,
	code_challenge,
	code_challenge_method,
	nonce,
	state,
	device_code_id,
	approved,
	created_at,
	updated_at
) VALUES (
	@id,
	@identity_id,
	@session_id,
	@client_id,
	@scopes,
	@redirect_uri,
	@code_challenge,
	@code_challenge_method,
	@nonce,
	@state,
	@device_code_id,
	@approved,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                    c.ID,
		"identity_id":           c.IdentityID,
		"session_id":            c.SessionID,
		"client_id":             c.ClientID,
		"scopes":                c.Scopes,
		"redirect_uri":          c.RedirectURI,
		"code_challenge":        c.CodeChallenge,
		"code_challenge_method": c.CodeChallengeMethod,
		"nonce":                 c.Nonce,
		"state":                 c.State,
		"device_code_id":        c.DeviceCodeID,
		"approved":              c.Approved,
		"created_at":            c.CreatedAt,
		"updated_at":            c.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" {
			return ErrResourceAlreadyExists
		}

		return fmt.Errorf("cannot insert oauth2_consent: %w", err)
	}

	return nil
}

func (c *OAuth2Consent) Update(ctx context.Context, conn pg.Tx) error {
	q := `
UPDATE iam_oauth2_consents
SET
	scopes = @scopes,
	approved = @approved,
	updated_at = @updated_at
WHERE
	id = @id
`

	_, err := conn.Exec(
		ctx,
		q,
		pgx.StrictNamedArgs{
			"id":         c.ID,
			"scopes":     c.Scopes,
			"approved":   c.Approved,
			"updated_at": c.UpdatedAt,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot update oauth2_consent: %w", err)
	}

	return nil
}

func (c *OAuth2Consent) Delete(ctx context.Context, conn pg.Tx) error {
	q := `
DELETE FROM iam_oauth2_consents
WHERE
	id = @id
`

	_, err := conn.Exec(ctx, q, pgx.StrictNamedArgs{"id": c.ID})
	if err != nil {
		return fmt.Errorf("cannot delete oauth2_consent: %w", err)
	}

	return nil
}

func (c *OAuth2Consents) LoadByIdentityID(
	ctx context.Context,
	conn pg.Querier,
	identityID gid.GID,
	cursor *page.Cursor[OAuth2ConsentOrderField],
) error {
	q := `
SELECT
	id,
	identity_id,
	session_id,
	client_id,
	scopes,
	redirect_uri,
	code_challenge,
	code_challenge_method,
	nonce,
	state,
	device_code_id,
	approved,
	created_at,
	updated_at
FROM
	iam_oauth2_consents
WHERE
	identity_id = @identity_id
	AND approved = TRUE
	AND %s
`

	q = fmt.Sprintf(
		q,
		cursor.SQLFragment(),
	)

	args := pgx.StrictNamedArgs{"identity_id": identityID}
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query oauth2_consents: %w", err)
	}

	consents, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[OAuth2Consent])
	if err != nil {
		return fmt.Errorf("cannot collect oauth2_consents: %w", err)
	}

	*c = consents
	return nil
}

func (c *OAuth2Consents) CountByIdentityID(
	ctx context.Context,
	conn pg.Querier,
	identityID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	iam_oauth2_consents
WHERE
	identity_id = @identity_id
	AND approved = TRUE;
`

	var count int
	err := conn.QueryRow(
		ctx,
		q,
		pgx.StrictNamedArgs{"identity_id": identityID},
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("cannot count oauth2_consents: %w", err)
	}

	return count, nil
}
