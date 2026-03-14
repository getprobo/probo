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
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	AccessEntry struct {
		ID                     gid.GID                   `db:"id"`
		AccessReviewCampaignID gid.GID                   `db:"access_review_campaign_id"`
		AccessSourceID         gid.GID                   `db:"access_source_id"`
		IdentityID             *gid.GID                  `db:"identity_id"`
		Email                  string                    `db:"email"`
		FullName               string                    `db:"full_name"`
		Role                   string                    `db:"role"`
		JobTitle               string                    `db:"job_title"`
		IsAdmin                bool                      `db:"is_admin"`
		MFAStatus              MFAStatus                 `db:"mfa_status"`
		AuthMethod             AccessEntryAuthMethod     `db:"auth_method"`
		LastLogin              *time.Time                `db:"last_login"`
		AccountCreatedAt       *time.Time                `db:"account_created_at"`
		ExternalID             string                    `db:"external_id"`
		AccountKey             string                    `db:"account_key"`
		IncrementalTag         AccessEntryIncrementalTag `db:"incremental_tag"`
		Flag                   AccessEntryFlag           `db:"flag"`
		FlagReason             *string                   `db:"flag_reason"`
		Decision               AccessEntryDecision       `db:"decision"`
		DecisionNote           *string                   `db:"decision_note"`
		DecidedBy              *gid.GID                  `db:"decided_by"`
		DecidedAt              *time.Time                `db:"decided_at"`
		CreatedAt              time.Time                 `db:"created_at"`
		UpdatedAt              time.Time                 `db:"updated_at"`
	}

	AccessEntries []*AccessEntry
)

func (e AccessEntry) CursorKey(orderBy AccessEntryOrderField) page.CursorKey {
	switch orderBy {
	case AccessEntryOrderFieldCreatedAt:
		return page.NewCursorKey(e.ID, e.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (e *AccessEntry) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	q := `
SELECT ar.organization_id
FROM access_entries ae
JOIN access_review_campaigns arc ON arc.id = ae.access_review_campaign_id
JOIN access_reviews ar ON ar.id = arc.access_review_id
WHERE ae.id = $1
LIMIT 1;
`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, e.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query access entry authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (e *AccessEntry) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	id gid.GID,
) error {
	q := `
SELECT
    id,
    access_review_campaign_id,
    access_source_id,
    identity_id,
    email,
    full_name,
    role,
    job_title,
    is_admin,
    mfa_status,
    auth_method,
    last_login,
    account_created_at,
    external_id,
    account_key,
    incremental_tag,
    flag,
    flag_reason,
    decision,
    decision_note,
    decided_by,
    decided_at,
    created_at,
    updated_at
FROM
    access_entries
WHERE
    %s
    AND id = @id
LIMIT 1;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": id}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query access_entries: %w", err)
	}

	entry, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AccessEntry])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect access entry: %w", err)
	}

	*e = entry

	return nil
}

func (e *AccessEntry) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO
    access_entries (
        id,
        tenant_id,
        access_review_campaign_id,
        access_source_id,
        identity_id,
        email,
        full_name,
        role,
        job_title,
        is_admin,
        mfa_status,
        auth_method,
        last_login,
        account_created_at,
        external_id,
        account_key,
        incremental_tag,
        flag,
        flag_reason,
        decision,
        decision_note,
        decided_by,
        decided_at,
        created_at,
        updated_at
    )
VALUES (
    @id,
    @tenant_id,
    @access_review_campaign_id,
    @access_source_id,
    @identity_id,
    @email,
    @full_name,
    @role,
    @job_title,
    @is_admin,
    @mfa_status,
    @auth_method,
    @last_login,
    @account_created_at,
    @external_id,
    @account_key,
    @incremental_tag,
    @flag,
    @flag_reason,
    @decision,
    @decision_note,
    @decided_by,
    @decided_at,
    @created_at,
    @updated_at
);
`

	args := pgx.StrictNamedArgs{
		"id":                        e.ID,
		"tenant_id":                 scope.GetTenantID(),
		"access_review_campaign_id": e.AccessReviewCampaignID,
		"access_source_id":          e.AccessSourceID,
		"identity_id":               e.IdentityID,
		"email":                     e.Email,
		"full_name":                 e.FullName,
		"role":                      e.Role,
		"job_title":                 e.JobTitle,
		"is_admin":                  e.IsAdmin,
		"mfa_status":                e.MFAStatus,
		"auth_method":               e.AuthMethod,
		"last_login":                e.LastLogin,
		"account_created_at":        e.AccountCreatedAt,
		"external_id":               e.ExternalID,
		"account_key":               e.AccountKey,
		"incremental_tag":           e.IncrementalTag,
		"flag":                      e.Flag,
		"flag_reason":               e.FlagReason,
		"decision":                  e.Decision,
		"decision_note":             e.DecisionNote,
		"decided_by":                e.DecidedBy,
		"decided_at":                e.DecidedAt,
		"created_at":                e.CreatedAt,
		"updated_at":                e.UpdatedAt,
	}
	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert access_entry: %w", err)
	}

	return nil
}

func (e *AccessEntry) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE access_entries
SET
    flag = @flag,
    flag_reason = @flag_reason,
    decision = @decision,
    decision_note = @decision_note,
    decided_by = @decided_by,
    decided_at = @decided_at,
    updated_at = @updated_at
WHERE
    %s
    AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":            e.ID,
		"flag":          e.Flag,
		"flag_reason":   e.FlagReason,
		"decision":      e.Decision,
		"decision_note": e.DecisionNote,
		"decided_by":    e.DecidedBy,
		"decided_at":    e.DecidedAt,
		"updated_at":    e.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update access_entry: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (entries *AccessEntries) LoadByCampaignID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	campaignID gid.GID,
	cursor *page.Cursor[AccessEntryOrderField],
) error {
	q := `
SELECT
    id,
    access_review_campaign_id,
    access_source_id,
    identity_id,
    email,
    full_name,
    role,
    job_title,
    is_admin,
    mfa_status,
    auth_method,
    last_login,
    account_created_at,
    external_id,
    account_key,
    incremental_tag,
    flag,
    flag_reason,
    decision,
    decision_note,
    decided_by,
    decided_at,
    created_at,
    updated_at
FROM
    access_entries
WHERE
    %s
    AND access_review_campaign_id = @campaign_id
    AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"campaign_id": campaignID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query access_entries: %w", err)
	}

	result, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[AccessEntry])
	if err != nil {
		return fmt.Errorf("cannot collect access_entries: %w", err)
	}

	*entries = result

	return nil
}

func (entries *AccessEntries) LoadByCampaignIDAndSourceID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	campaignID gid.GID,
	sourceID gid.GID,
	cursor *page.Cursor[AccessEntryOrderField],
) error {
	q := `
SELECT
    id,
    access_review_campaign_id,
    access_source_id,
    identity_id,
    email,
    full_name,
    role,
    job_title,
    is_admin,
    mfa_status,
    auth_method,
    last_login,
    account_created_at,
    external_id,
    account_key,
    incremental_tag,
    flag,
    flag_reason,
    decision,
    decision_note,
    decided_by,
    decided_at,
    created_at,
    updated_at
FROM
    access_entries
WHERE
    %s
    AND access_review_campaign_id = @campaign_id
    AND access_source_id = @source_id
    AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"campaign_id": campaignID, "source_id": sourceID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query access_entries: %w", err)
	}

	result, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[AccessEntry])
	if err != nil {
		return fmt.Errorf("cannot collect access_entries: %w", err)
	}

	*entries = result

	return nil
}

func (entries *AccessEntries) CountByCampaignID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	campaignID gid.GID,
) (int, error) {
	q := `
SELECT COUNT(id)
FROM access_entries
WHERE
    %s
    AND access_review_campaign_id = @campaign_id;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"campaign_id": campaignID}
	maps.Copy(args, scope.SQLArguments())

	var count int
	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count access_entries: %w", err)
	}

	return count, nil
}

func (entries *AccessEntries) CountByCampaignIDAndSourceID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	campaignID gid.GID,
	sourceID gid.GID,
) (int, error) {
	q := `
SELECT COUNT(id)
FROM access_entries
WHERE
    %s
    AND access_review_campaign_id = @campaign_id
    AND access_source_id = @source_id;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"campaign_id": campaignID, "source_id": sourceID}
	maps.Copy(args, scope.SQLArguments())

	var count int
	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count access_entries: %w", err)
	}

	return count, nil
}

func (entries *AccessEntries) CountPendingByCampaignID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	campaignID gid.GID,
) (int, error) {
	q := `
SELECT COUNT(id)
FROM access_entries
WHERE
    %s
    AND access_review_campaign_id = @campaign_id
    AND decision = 'PENDING';
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"campaign_id": campaignID}
	maps.Copy(args, scope.SQLArguments())

	var count int
	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count pending access_entries: %w", err)
	}

	return count, nil
}

func (entries *AccessEntries) BulkInsert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	for _, entry := range *entries {
		if err := entry.Insert(ctx, conn, scope); err != nil {
			return fmt.Errorf("cannot bulk insert access_entry: %w", err)
		}
	}

	return nil
}
