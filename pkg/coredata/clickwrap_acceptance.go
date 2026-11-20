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
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/policy"
	"go.probo.inc/probo/pkg/page"
)

type (
	ClickwrapAcceptance struct {
		ID                      gid.GID                   `db:"id"`
		OrganizationID          gid.GID                   `db:"organization_id"`
		SigningConfigID         gid.GID                   `db:"signing_config_id"`
		DocumentVersionID       gid.GID                   `db:"document_version_id"`
		SubjectID               gid.GID                   `db:"subject_id"`
		AcceptedEmail           string                    `db:"accepted_email"`
		AcceptedFullName        *string                   `db:"accepted_full_name"`
		AcceptedAccountID       *string                   `db:"accepted_account_id"`
		ConsentText             string                    `db:"consent_text"`
		ContentSha256           string                    `db:"content_sha256"`
		IPAddress               string                    `db:"ip_address"`
		UserAgent               string                    `db:"user_agent"`
		SdkVersion              *string                   `db:"sdk_version"`
		CountryCode             *string                   `db:"country_code"`
		Mode                    DocumentSigningConfigMode `db:"mode"`
		IdempotencyKey          *string                   `db:"idempotency_key"`
		RevokedAt               *time.Time                `db:"revoked_at"`
		RevocationReason        *string                   `db:"revocation_reason"`
		Seal                    *string                   `db:"seal"`
		SealVersion             int                       `db:"seal_version"`
		TSAToken                []byte                    `db:"tsa_token"`
		SealedAt                *time.Time                `db:"sealed_at"`
		SealAttemptCount        int                       `db:"seal_attempt_count"`
		SealMaxAttempts         int                       `db:"seal_max_attempts"`
		SealLastAttemptedAt     *time.Time                `db:"seal_last_attempted_at"`
		SealLastError           *string                   `db:"seal_last_error"`
		SealProcessingStartedAt *time.Time                `db:"seal_processing_started_at"`
		AcceptedAt              time.Time                 `db:"accepted_at"`
		CreatedAt               time.Time                 `db:"created_at"`
		UpdatedAt               time.Time                 `db:"updated_at"`
	}

	ClickwrapAcceptances []*ClickwrapAcceptance
)

var clickwrapAcceptanceColumns = `
	id, organization_id, signing_config_id, document_version_id, subject_id,
	accepted_email, accepted_full_name, accepted_account_id,
	consent_text, content_sha256, ip_address, user_agent, sdk_version, country_code,
	mode, idempotency_key,
	revoked_at, revocation_reason,
	seal, seal_version, tsa_token, sealed_at,
	seal_attempt_count, seal_max_attempts, seal_last_attempted_at, seal_last_error,
	seal_processing_started_at,
	accepted_at, created_at, updated_at
`

func (a *ClickwrapAcceptance) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM clickwrap_acceptances WHERE id = ANY(@resource_ids::text[])`

	args := pgx.StrictNamedArgs{"resource_ids": resourceIDs}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query clickwrap acceptance authorization attributes: %w", err)
	}
	defer rows.Close()

	attrsByID := make(policy.AttributesByID)

	for rows.Next() {
		var id, organizationID gid.GID
		if err := rows.Scan(&id, &organizationID); err != nil {
			return nil, fmt.Errorf("cannot scan clickwrap acceptance authorization attributes: %w", err)
		}

		attrsByID[id] = policy.Attributes{
			"organization_id": organizationID.String(),
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate clickwrap acceptance authorization attributes: %w", err)
	}

	return attrsByID, nil
}

func (a *ClickwrapAcceptance) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO clickwrap_acceptances (
	id, tenant_id, organization_id,
	signing_config_id, document_version_id, subject_id,
	accepted_email, accepted_full_name, accepted_account_id,
	consent_text, content_sha256, ip_address, user_agent, sdk_version, country_code,
	mode, idempotency_key,
	seal_version, seal_attempt_count, seal_max_attempts,
	accepted_at, created_at, updated_at
) VALUES (
	@id, @tenant_id, @organization_id,
	@signing_config_id, @document_version_id, @subject_id,
	@accepted_email, @accepted_full_name, @accepted_account_id,
	@consent_text, @content_sha256, @ip_address, @user_agent, @sdk_version, @country_code,
	@mode, @idempotency_key,
	@seal_version, @seal_attempt_count, @seal_max_attempts,
	@accepted_at, @created_at, @updated_at
)
ON CONFLICT (organization_id, signing_config_id, idempotency_key) DO NOTHING
RETURNING ` + clickwrapAcceptanceColumns

	args := pgx.StrictNamedArgs{
		"id":                  a.ID,
		"tenant_id":           scope.GetTenantID(),
		"organization_id":     a.OrganizationID,
		"signing_config_id":   a.SigningConfigID,
		"document_version_id": a.DocumentVersionID,
		"subject_id":          a.SubjectID,
		"accepted_email":      a.AcceptedEmail,
		"accepted_full_name":  a.AcceptedFullName,
		"accepted_account_id": a.AcceptedAccountID,
		"consent_text":        a.ConsentText,
		"content_sha256":      a.ContentSha256,
		"ip_address":          a.IPAddress,
		"user_agent":          a.UserAgent,
		"sdk_version":         a.SdkVersion,
		"country_code":        a.CountryCode,
		"mode":                a.Mode,
		"idempotency_key":     a.IdempotencyKey,
		"seal_version":        a.SealVersion,
		"seal_attempt_count":  a.SealAttemptCount,
		"seal_max_attempts":   a.SealMaxAttempts,
		"accepted_at":         a.AcceptedAt,
		"created_at":          a.CreatedAt,
		"updated_at":          a.UpdatedAt,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert clickwrap acceptance: %w", err)
	}
	defer rows.Close()

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ClickwrapAcceptance])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect clickwrap acceptance insert: %w", err)
	}

	*a = row

	return nil
}

func (a *ClickwrapAcceptance) Update(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE clickwrap_acceptances SET
	revoked_at               = @revoked_at,
	revocation_reason        = @revocation_reason,
	seal                     = @seal,
	seal_version             = @seal_version,
	tsa_token                = @tsa_token,
	sealed_at                = @sealed_at,
	seal_attempt_count       = @seal_attempt_count,
	seal_last_attempted_at   = @seal_last_attempted_at,
	seal_last_error          = @seal_last_error,
	seal_processing_started_at = @seal_processing_started_at,
	updated_at               = @updated_at
WHERE %s AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                         a.ID,
		"revoked_at":                 a.RevokedAt,
		"revocation_reason":          a.RevocationReason,
		"seal":                       a.Seal,
		"seal_version":               a.SealVersion,
		"tsa_token":                  a.TSAToken,
		"sealed_at":                  a.SealedAt,
		"seal_attempt_count":         a.SealAttemptCount,
		"seal_last_attempted_at":     a.SealLastAttemptedAt,
		"seal_last_error":            a.SealLastError,
		"seal_processing_started_at": a.SealProcessingStartedAt,
		"updated_at":                 a.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update clickwrap acceptance: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (a *ClickwrapAcceptance) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	id gid.GID,
) error {
	q := `
SELECT ` + clickwrapAcceptanceColumns + `
FROM clickwrap_acceptances
WHERE %s AND id = @id
LIMIT 1
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": id}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query clickwrap acceptance: %w", err)
	}

	acc, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ClickwrapAcceptance])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect clickwrap acceptance: %w", err)
	}

	*a = acc

	return nil
}

func (a *ClickwrapAcceptance) LoadByIdempotencyKey(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	signingConfigID gid.GID,
	idempotencyKey string,
) error {
	q := `
SELECT ` + clickwrapAcceptanceColumns + `
FROM clickwrap_acceptances
WHERE %s
	AND organization_id   = @organization_id
	AND signing_config_id = @signing_config_id
	AND idempotency_key   = @idempotency_key
LIMIT 1
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"organization_id":   organizationID,
		"signing_config_id": signingConfigID,
		"idempotency_key":   idempotencyKey,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query clickwrap acceptance by idempotency key: %w", err)
	}

	acc, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ClickwrapAcceptance])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect clickwrap acceptance: %w", err)
	}

	*a = acc

	return nil
}

func (a *ClickwrapAcceptance) LoadLatestBySubjectAndConfig(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	signingConfigID gid.GID,
	subjectID gid.GID,
) error {
	q := `
SELECT ` + clickwrapAcceptanceColumns + `
FROM clickwrap_acceptances
WHERE %s
	AND organization_id   = @organization_id
	AND signing_config_id = @signing_config_id
	AND subject_id        = @subject_id
	AND revoked_at IS NULL
ORDER BY accepted_at DESC
LIMIT 1
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"organization_id":   organizationID,
		"signing_config_id": signingConfigID,
		"subject_id":        subjectID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query latest clickwrap acceptance: %w", err)
	}

	acc, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ClickwrapAcceptance])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect latest clickwrap acceptance: %w", err)
	}

	*a = acc

	return nil
}

type ClickwrapAcceptanceOrderField string

const ClickwrapAcceptanceOrderFieldAcceptedAt ClickwrapAcceptanceOrderField = "accepted_at"

func (f ClickwrapAcceptanceOrderField) String() string { return string(f) }
func (f ClickwrapAcceptanceOrderField) Column() string { return string(f) }

func (a ClickwrapAcceptance) CursorKey(orderBy ClickwrapAcceptanceOrderField) page.CursorKey {
	switch orderBy {
	case ClickwrapAcceptanceOrderFieldAcceptedAt:
		return page.NewCursorKey(a.ID, a.AcceptedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (as *ClickwrapAcceptances) LoadBySigningConfigID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	signingConfigID gid.GID,
	cursor *page.Cursor[ClickwrapAcceptanceOrderField],
) error {
	q := `
SELECT ` + clickwrapAcceptanceColumns + `
FROM clickwrap_acceptances
WHERE %s
	AND signing_config_id = @signing_config_id
	AND %s
ORDER BY accepted_at DESC, id DESC
LIMIT @limit
`
	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"signing_config_id": signingConfigID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query clickwrap acceptances: %w", err)
	}

	acceptances, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ClickwrapAcceptance])
	if err != nil {
		return fmt.Errorf("cannot collect clickwrap acceptances: %w", err)
	}

	*as = acceptances

	return nil
}

func (as *ClickwrapAcceptances) LoadBySubjectID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	subjectID gid.GID,
	cursor *page.Cursor[ClickwrapAcceptanceOrderField],
) error {
	q := `
SELECT ` + clickwrapAcceptanceColumns + `
FROM clickwrap_acceptances
WHERE %s
	AND subject_id = @subject_id
	AND %s
ORDER BY accepted_at DESC, id DESC
LIMIT @limit
`
	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"subject_id": subjectID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query clickwrap acceptances: %w", err)
	}

	acceptances, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ClickwrapAcceptance])
	if err != nil {
		return fmt.Errorf("cannot collect clickwrap acceptances: %w", err)
	}

	*as = acceptances

	return nil
}

func (a *ClickwrapAcceptance) LoadNextPendingSealForUpdate(
	ctx context.Context,
	conn pg.Tx,
) error {
	q := `
SELECT ` + clickwrapAcceptanceColumns + `
FROM clickwrap_acceptances
WHERE seal IS NULL
	AND revoked_at IS NULL
	AND seal_attempt_count < seal_max_attempts
ORDER BY created_at ASC
LIMIT 1
FOR UPDATE SKIP LOCKED
`

	rows, err := conn.Query(ctx, q)
	if err != nil {
		return fmt.Errorf("cannot query pending seal clickwrap acceptances: %w", err)
	}

	acc, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ClickwrapAcceptance])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect pending seal clickwrap acceptance: %w", err)
	}

	*a = acc

	return nil
}
