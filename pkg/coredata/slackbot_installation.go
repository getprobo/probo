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
	"encoding"
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
)

type (
	SlackbotInstallationStatus string

	SlackbotInstallationCredentials struct {
		AccessToken  string
		RefreshToken *string
	}

	SlackbotInstallation struct {
		ID                    gid.GID                    `db:"id"`
		OrganizationID        gid.GID                    `db:"organization_id"`
		TeamID                string                     `db:"team_id"`
		BotUserID             string                     `db:"bot_user_id"`
		EncryptedAccessToken  []byte                     `db:"encrypted_access_token"`
		EncryptedRefreshToken []byte                     `db:"encrypted_refresh_token"`
		AccessTokenExpiresAt  *time.Time                 `db:"access_token_expires_at"`
		Scopes                []string                   `db:"scopes"`
		Status                SlackbotInstallationStatus `db:"status"`
		CreatedAt             time.Time                  `db:"created_at"`
		UpdatedAt             time.Time                  `db:"updated_at"`
	}
)

const (
	SlackbotInstallationStatusActive   SlackbotInstallationStatus = "ACTIVE"
	SlackbotInstallationStatusDisabled SlackbotInstallationStatus = "DISABLED"
)

var (
	_ fmt.Stringer             = SlackbotInstallationStatus("")
	_ encoding.TextMarshaler   = SlackbotInstallationStatus("")
	_ encoding.TextUnmarshaler = (*SlackbotInstallationStatus)(nil)
)

func SlackbotInstallationStatuses() []SlackbotInstallationStatus {
	return []SlackbotInstallationStatus{
		SlackbotInstallationStatusActive,
		SlackbotInstallationStatusDisabled,
	}
}

func (v SlackbotInstallationStatus) IsValid() bool {
	switch v {
	case SlackbotInstallationStatusActive, SlackbotInstallationStatusDisabled:
		return true
	}

	return false
}

func (v SlackbotInstallationStatus) String() string {
	return string(v)
}

func (v SlackbotInstallationStatus) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *SlackbotInstallationStatus) UnmarshalText(text []byte) error {
	value := SlackbotInstallationStatus(text)
	if !value.IsValid() {
		return fmt.Errorf("invalid SlackbotInstallationStatus value: %q", string(text))
	}

	*v = value

	return nil
}

func NewSlackbotInstallation(scope Scoper, organizationID gid.GID) *SlackbotInstallation {
	now := time.Now()

	return &SlackbotInstallation{
		ID:             gid.New(scope.GetTenantID(), SlackbotInstallationEntityType),
		OrganizationID: organizationID,
		Status:         SlackbotInstallationStatusActive,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func (s *SlackbotInstallation) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	team_id,
	bot_user_id,
	encrypted_access_token,
	encrypted_refresh_token,
	access_token_expires_at,
	scopes,
	status,
	created_at,
	updated_at
FROM slackbot_installations
WHERE %s
	AND organization_id = @organization_id
LIMIT 1
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	return s.load(ctx, conn, q, args)
}

func (s *SlackbotInstallation) LoadByOrganizationIDForUpdate(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
	organizationID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	team_id,
	bot_user_id,
	encrypted_access_token,
	encrypted_refresh_token,
	access_token_expires_at,
	scopes,
	status,
	created_at,
	updated_at
FROM slackbot_installations
WHERE %s
	AND organization_id = @organization_id
LIMIT 1
FOR UPDATE
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	return s.load(ctx, tx, q, args)
}

func (s *SlackbotInstallation) LoadByTeamID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	teamID string,
) error {
	q := `
SELECT
	id,
	organization_id,
	team_id,
	bot_user_id,
	encrypted_access_token,
	encrypted_refresh_token,
	access_token_expires_at,
	scopes,
	status,
	created_at,
	updated_at
FROM slackbot_installations
WHERE %s
	AND team_id = @team_id
LIMIT 1
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"team_id": teamID}
	maps.Copy(args, scope.SQLArguments())

	return s.load(ctx, conn, q, args)
}

func (s *SlackbotInstallation) Upsert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	encryptionKey cipher.EncryptionKey,
	credentials SlackbotInstallationCredentials,
) (bool, error) {
	encryptedAccessToken, encryptedRefreshToken, err := encryptSlackbotInstallationCredentials(
		encryptionKey,
		credentials,
	)
	if err != nil {
		return false, fmt.Errorf("cannot encrypt Slackbot installation credentials: %w", err)
	}

	q := `
INSERT INTO slackbot_installations (
	id,
	tenant_id,
	organization_id,
	team_id,
	bot_user_id,
	encrypted_access_token,
	encrypted_refresh_token,
	access_token_expires_at,
	scopes,
	status,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@team_id,
	@bot_user_id,
	@encrypted_access_token,
	@encrypted_refresh_token,
	@access_token_expires_at,
	@scopes,
	@status,
	@created_at,
	@updated_at
)
ON CONFLICT (organization_id) DO UPDATE SET
	team_id = EXCLUDED.team_id,
	bot_user_id = EXCLUDED.bot_user_id,
	encrypted_access_token = EXCLUDED.encrypted_access_token,
	encrypted_refresh_token = EXCLUDED.encrypted_refresh_token,
	access_token_expires_at = EXCLUDED.access_token_expires_at,
	scopes = EXCLUDED.scopes,
	status = EXCLUDED.status,
	updated_at = EXCLUDED.updated_at
RETURNING
	id,
	organization_id,
	team_id,
	bot_user_id,
	encrypted_access_token,
	encrypted_refresh_token,
	access_token_expires_at,
	scopes,
	status,
	created_at,
	updated_at
`

	originalID := s.ID
	args := pgx.StrictNamedArgs{
		"id":                      s.ID,
		"tenant_id":               scope.GetTenantID(),
		"organization_id":         s.OrganizationID,
		"team_id":                 s.TeamID,
		"bot_user_id":             s.BotUserID,
		"encrypted_access_token":  encryptedAccessToken,
		"encrypted_refresh_token": encryptedRefreshToken,
		"access_token_expires_at": s.AccessTokenExpiresAt,
		"scopes":                  s.Scopes,
		"status":                  s.Status,
		"created_at":              s.CreatedAt,
		"updated_at":              s.UpdatedAt,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "slackbot_installations_team_id_key" {
				return false, ErrResourceAlreadyExists
			}
		}

		return false, fmt.Errorf("cannot upsert Slackbot installation: %w", err)
	}
	defer rows.Close()

	installation, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[SlackbotInstallation])
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "slackbot_installations_team_id_key" {
				return false, ErrResourceAlreadyExists
			}
		}

		return false, fmt.Errorf("cannot collect Slackbot installation upsert result: %w", err)
	}

	*s = installation

	return originalID == s.ID, nil
}

func (s *SlackbotInstallation) UpdateCredentials(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	encryptionKey cipher.EncryptionKey,
	credentials SlackbotInstallationCredentials,
) error {
	encryptedAccessToken, encryptedRefreshToken, err := encryptSlackbotInstallationCredentials(
		encryptionKey,
		credentials,
	)
	if err != nil {
		return fmt.Errorf("cannot encrypt Slackbot installation credentials: %w", err)
	}

	q := `
UPDATE slackbot_installations
SET
	encrypted_access_token = @encrypted_access_token,
	encrypted_refresh_token = @encrypted_refresh_token,
	access_token_expires_at = @access_token_expires_at,
	scopes = @scopes,
	updated_at = @updated_at
WHERE %s
	AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                      s.ID,
		"encrypted_access_token":  encryptedAccessToken,
		"encrypted_refresh_token": encryptedRefreshToken,
		"access_token_expires_at": s.AccessTokenExpiresAt,
		"scopes":                  s.Scopes,
		"updated_at":              s.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update Slackbot installation credentials: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	s.EncryptedAccessToken = encryptedAccessToken
	s.EncryptedRefreshToken = encryptedRefreshToken

	return nil
}

func (s *SlackbotInstallation) UpdateStatus(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE slackbot_installations
SET
	status = @status,
	updated_at = @updated_at
WHERE %s
	AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":         s.ID,
		"status":     s.Status,
		"updated_at": s.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update Slackbot installation status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (s *SlackbotInstallation) DecryptCredentials(
	encryptionKey cipher.EncryptionKey,
) (SlackbotInstallationCredentials, error) {
	accessToken, err := cipher.Decrypt(s.EncryptedAccessToken, encryptionKey)
	if err != nil {
		return SlackbotInstallationCredentials{}, fmt.Errorf("cannot decrypt Slackbot access token: %w", err)
	}

	credentials := SlackbotInstallationCredentials{AccessToken: string(accessToken)}

	if len(s.EncryptedRefreshToken) > 0 {
		refreshToken, err := cipher.Decrypt(s.EncryptedRefreshToken, encryptionKey)
		if err != nil {
			return SlackbotInstallationCredentials{}, fmt.Errorf("cannot decrypt Slackbot refresh token: %w", err)
		}

		credentials.RefreshToken = new(string(refreshToken))
	}

	return credentials, nil
}

func (s *SlackbotInstallation) Delete(ctx context.Context, conn pg.Tx, scope Scoper) error {
	q := `DELETE FROM slackbot_installations WHERE %s AND id = @id`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": s.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete Slackbot installation: %w", err)
	}

	return nil
}

func (s *SlackbotInstallation) load(
	ctx context.Context,
	conn pg.Querier,
	q string,
	args pgx.StrictNamedArgs,
) error {
	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query Slackbot installation: %w", err)
	}

	installation, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[SlackbotInstallation])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect Slackbot installation: %w", err)
	}

	*s = installation

	return nil
}

func encryptSlackbotInstallationCredentials(
	encryptionKey cipher.EncryptionKey,
	credentials SlackbotInstallationCredentials,
) ([]byte, []byte, error) {
	encryptedAccessToken, err := cipher.Encrypt([]byte(credentials.AccessToken), encryptionKey)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot encrypt access token: %w", err)
	}

	var encryptedRefreshToken []byte
	if credentials.RefreshToken != nil {
		encryptedRefreshToken, err = cipher.Encrypt([]byte(*credentials.RefreshToken), encryptionKey)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot encrypt refresh token: %w", err)
		}
	}

	return encryptedAccessToken, encryptedRefreshToken, nil
}
