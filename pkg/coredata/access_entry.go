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
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	AccessEntry struct {
		ID                     gid.GID                   `db:"id"`
		OrganizationID         gid.GID                   `db:"organization_id"`
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
		AccountType            AccessEntryAccountType    `db:"account_type"`
		LastLogin              *time.Time                `db:"last_login"`
		AccountCreatedAt       *time.Time                `db:"account_created_at"`
		ExternalID             string                    `db:"external_id"`
		AccountKey             string                    `db:"account_key"`
		IncrementalTag         AccessEntryIncrementalTag `db:"incremental_tag"`
		Flags                  []AccessEntryFlag         `db:"flags"`
		FlagReasons            []string                  `db:"flag_reasons"`
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

func (e *AccessEntry) AuthorizationAttributes(ctx context.Context, conn pg.Querier) (map[string]string, error) {
	q := `SELECT organization_id FROM access_entries WHERE id = $1 LIMIT 1;`

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
	conn pg.Querier,
	scope Scoper,
	id gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
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
    account_type,
    last_login,
    account_created_at,
    external_id,
    account_key,
    incremental_tag,
    flags,
    flag_reasons,
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
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO
    access_entries (
        id,
        tenant_id,
        organization_id,
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
        account_type,
        last_login,
        account_created_at,
        external_id,
        account_key,
        incremental_tag,
        flags,
        flag_reasons,
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
    @organization_id,
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
    @account_type,
    @last_login,
    @account_created_at,
    @external_id,
    @account_key,
    @incremental_tag,
    @flags,
    @flag_reasons,
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
		"organization_id":           e.OrganizationID,
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
		"account_type":              e.AccountType,
		"last_login":                e.LastLogin,
		"account_created_at":        e.AccountCreatedAt,
		"external_id":               e.ExternalID,
		"account_key":               e.AccountKey,
		"incremental_tag":           e.IncrementalTag,
		"flags":                     e.Flags,
		"flag_reasons":              e.FlagReasons,
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
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE access_entries
SET
    flags = @flags,
    flag_reasons = @flag_reasons,
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
		"flags":         e.Flags,
		"flag_reasons":  e.FlagReasons,
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
	conn pg.Querier,
	scope Scoper,
	campaignID gid.GID,
	cursor *page.Cursor[AccessEntryOrderField],
	filter *AccessEntryFilter,
) error {
	q := `
SELECT
    id,
    organization_id,
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
    account_type,
    last_login,
    account_created_at,
    external_id,
    account_key,
    incremental_tag,
    flags,
    flag_reasons,
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
    AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"campaign_id": campaignID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())
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
	conn pg.Querier,
	scope Scoper,
	campaignID gid.GID,
	sourceID gid.GID,
	cursor *page.Cursor[AccessEntryOrderField],
	filter *AccessEntryFilter,
) error {
	q := `
SELECT
    id,
    organization_id,
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
    account_type,
    last_login,
    account_created_at,
    external_id,
    account_key,
    incremental_tag,
    flags,
    flag_reasons,
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
    AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"campaign_id": campaignID, "source_id": sourceID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())
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
	conn pg.Querier,
	scope Scoper,
	campaignID gid.GID,
	filter *AccessEntryFilter,
) (int, error) {
	q := `
SELECT COUNT(id)
FROM access_entries
WHERE
    %s
    AND access_review_campaign_id = @campaign_id
    AND %s;
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.StrictNamedArgs{"campaign_id": campaignID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	var count int
	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count access_entries: %w", err)
	}

	return count, nil
}

func (entries *AccessEntries) CountByCampaignIDAndSourceID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	campaignID gid.GID,
	sourceID gid.GID,
	filter *AccessEntryFilter,
) (int, error) {
	q := `
SELECT COUNT(id)
FROM access_entries
WHERE
    %s
    AND access_review_campaign_id = @campaign_id
    AND access_source_id = @source_id
    AND %s;
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.StrictNamedArgs{"campaign_id": campaignID, "source_id": sourceID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	var count int
	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count access_entries: %w", err)
	}

	return count, nil
}

func (entries *AccessEntries) CountPendingByCampaignID(
	ctx context.Context,
	conn pg.Querier,
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

func (e *AccessEntry) LoadOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	entryID gid.GID,
) (gid.GID, error) {
	q := `SELECT organization_id FROM access_entries WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, entryID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return gid.GID{}, ErrResourceNotFound
		}
		return gid.GID{}, fmt.Errorf("cannot load organization id for access entry: %w", err)
	}

	return organizationID, nil
}

func (e *AccessEntry) UpdateFlags(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE access_entries
SET
    flags = @flags,
    flag_reasons = @flag_reasons,
    updated_at = @updated_at
WHERE
    %s
    AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":           e.ID,
		"flags":        e.Flags,
		"flag_reasons": e.FlagReasons,
		"updated_at":   e.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update access entry flags: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (e *AccessEntry) Upsert(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) error {
	q := `
INSERT INTO access_entries (
    id,
    tenant_id,
    organization_id,
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
    account_type,
    last_login,
    account_created_at,
    external_id,
    account_key,
    incremental_tag,
    flags,
    flag_reasons,
    decision,
    decision_note,
    decided_by,
    decided_at,
    created_at,
    updated_at
) VALUES (
    @id,
    @tenant_id,
    @organization_id,
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
    @account_type,
    @last_login,
    @account_created_at,
    @external_id,
    @account_key,
    @incremental_tag,
    @flags,
    @flag_reasons,
    @decision,
    @decision_note,
    @decided_by,
    @decided_at,
    @created_at,
    @updated_at
)
ON CONFLICT (access_review_campaign_id, access_source_id, account_key) DO UPDATE SET
    email = EXCLUDED.email,
    full_name = EXCLUDED.full_name,
    role = EXCLUDED.role,
    job_title = EXCLUDED.job_title,
    is_admin = EXCLUDED.is_admin,
    mfa_status = EXCLUDED.mfa_status,
    auth_method = EXCLUDED.auth_method,
    account_type = EXCLUDED.account_type,
    last_login = EXCLUDED.last_login,
    account_created_at = EXCLUDED.account_created_at,
    external_id = EXCLUDED.external_id,
    incremental_tag = EXCLUDED.incremental_tag,
    updated_at = EXCLUDED.updated_at
`
	args := pgx.StrictNamedArgs{
		"id":                        e.ID,
		"tenant_id":                 scope.GetTenantID(),
		"organization_id":           e.OrganizationID,
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
		"account_type":              e.AccountType,
		"last_login":                e.LastLogin,
		"account_created_at":        e.AccountCreatedAt,
		"external_id":               e.ExternalID,
		"account_key":               e.AccountKey,
		"incremental_tag":           e.IncrementalTag,
		"flags":                     e.Flags,
		"flag_reasons":              e.FlagReasons,
		"decision":                  e.Decision,
		"decision_note":             e.DecisionNote,
		"decided_by":                e.DecidedBy,
		"decided_at":                e.DecidedAt,
		"created_at":                e.CreatedAt,
		"updated_at":                e.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot upsert access entry: %w", err)
	}

	return nil
}

// BaselineAccountEntry holds minimal data from a previous campaign's entries
// for incremental diffing.
type BaselineAccountEntry struct {
	AccountKey string
	Email      string
	FullName   string
}

func (entries *AccessEntries) LoadBaselineBySourceID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	campaignID gid.GID,
	sourceID gid.GID,
) ([]BaselineAccountEntry, error) {
	q := fmt.Sprintf(`
SELECT account_key, email, full_name
FROM access_entries
WHERE %s
  AND access_review_campaign_id = @campaign_id
  AND access_source_id = @source_id
`, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"campaign_id": campaignID,
		"source_id":   sourceID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot load baseline entries: %w", err)
	}
	defer rows.Close()

	var result []BaselineAccountEntry
	for rows.Next() {
		var entry BaselineAccountEntry
		if err := rows.Scan(&entry.AccountKey, &entry.Email, &entry.FullName); err != nil {
			return nil, fmt.Errorf("cannot scan baseline entry: %w", err)
		}
		result = append(result, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate baseline entries: %w", err)
	}

	return result, nil
}

// LoadMembershipAccountsByOrganizationID loads IAM membership accounts for the
// given organization.
type MembershipAccount struct {
	ID        gid.GID
	Email     string
	FullName  string
	State     string
	Role      string
	CreatedAt time.Time
}

func LoadMembershipAccountsByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
) ([]MembershipAccount, error) {
	q := `
SELECT
    m.id,
    i.email_address,
    i.full_name,
    m.state,
    m.role,
    m.created_at
FROM
    iam_memberships m
JOIN
    identities i ON i.id = m.identity_id
WHERE
    m.%s
    AND m.organization_id = @organization_id
    AND m.state = 'ACTIVE'
ORDER BY
    i.email_address ASC
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"organization_id": organizationID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query membership accounts: %w", err)
	}
	defer rows.Close()

	var result []MembershipAccount
	for rows.Next() {
		var account MembershipAccount
		if err := rows.Scan(&account.ID, &account.Email, &account.FullName, &account.State, &account.Role, &account.CreatedAt); err != nil {
			return nil, fmt.Errorf("cannot scan membership account: %w", err)
		}
		result = append(result, account)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate membership accounts: %w", err)
	}

	return result, nil
}
