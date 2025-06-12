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
	"fmt"
	"maps"
	"time"

	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/page"
	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
)

type (
	Organization struct {
		ID                    gid.GID      `db:"id"`
		TenantID              gid.TenantID `db:"tenant_id"`
		Name                  string       `db:"name"`
		LogoObjectKey         string       `db:"logo_object_key"`
		FoundingYear          *int         `db:"founding_year"`
		CompanyType           *string      `db:"company_type"`
		PreMarketFit          *bool        `db:"premarket_fit"`
		UsesCloudProviders    *bool        `db:"uses_cloud_providers"`
		AIFocused             *bool        `db:"ai_focused"`
		UsesAIGeneratedCode   *bool        `db:"uses_ai_generated_code"`
		VCBacked              *bool        `db:"vc_backed"`
		HasRaisedMoney        *bool        `db:"has_raised_money"`
		HasEnterpriseAccounts *bool        `db:"has_enterprise_accounts"`
		CreatedAt             time.Time    `db:"created_at"`
		UpdatedAt             time.Time    `db:"updated_at"`
	}

	Organizations []*Organization
)

func (o Organization) CursorKey(orderBy OrganizationOrderField) page.CursorKey {
	switch orderBy {
	case OrganizationOrderFieldCreatedAt:
		return page.NewCursorKey(o.ID, o.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (o *Organization) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
) error {
	q := `
SELECT
    tenant_id,
    id,
    name,
    logo_object_key,
	founding_year,
	company_type,
	premarket_fit,
	uses_cloud_providers,
	ai_focused,
	uses_ai_generated_code,
	has_raised_money,
	has_enterprise_accounts,
	vc_backed,
    created_at,
    updated_at
FROM
    organizations
WHERE
    %s
    AND id = @organization_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query organizations: %w", err)
	}

	organization, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Organization])
	if err != nil {
		return fmt.Errorf("cannot collect organization: %w", err)
	}

	*o = organization

	return nil
}

func (o *Organization) Insert(
	ctx context.Context,
	conn pg.Conn,
) error {
	q := `
INSERT INTO organizations (
    tenant_id,
    id,
    name,
    logo_object_key,
	founding_year,
	company_type,
	premarket_fit,
	uses_cloud_providers,
	ai_focused,
	uses_ai_generated_code,
	vc_backed,
	has_raised_money,
	has_enterprise_accounts,
    created_at,
    updated_at
) VALUES (@tenant_id, @id, @name, @logo_object_key, @founding_year, @company_type, @premarket_fit, @uses_cloud_providers, @ai_focused, @uses_ai_generated_code, @vc_backed, @has_raised_money, @has_enterprise_accounts, @created_at, @updated_at)
`

	args := pgx.StrictNamedArgs{
		"tenant_id":               o.TenantID,
		"id":                      o.ID,
		"name":                    o.Name,
		"logo_object_key":         o.LogoObjectKey,
		"founding_year":           o.FoundingYear,
		"company_type":            o.CompanyType,
		"premarket_fit":           o.PreMarketFit,
		"uses_cloud_providers":    o.UsesCloudProviders,
		"ai_focused":              o.AIFocused,
		"uses_ai_generated_code":  o.UsesAIGeneratedCode,
		"vc_backed":               o.VCBacked,
		"has_raised_money":        o.HasRaisedMoney,
		"has_enterprise_accounts": o.HasEnterpriseAccounts,
		"created_at":              o.CreatedAt,
		"updated_at":              o.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return err
	}

	return nil
}

func (o *Organization) Update(
	ctx context.Context,
	scope Scoper,
	conn pg.Conn,
) error {
	q := `
UPDATE organizations
SET
    name = @name,
    logo_object_key = @logo_object_key,
    founding_year = @founding_year,
    company_type = @company_type,
    premarket_fit = @premarket_fit,
    uses_cloud_providers = @uses_cloud_providers,
    ai_focused = @ai_focused,
    uses_ai_generated_code = @uses_ai_generated_code,
    vc_backed = @vc_backed,
    has_raised_money = @has_raised_money,
    has_enterprise_accounts = @has_enterprise_accounts,
    updated_at = @updated_at
WHERE
    %s
    AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                      o.ID,
		"name":                    o.Name,
		"logo_object_key":         o.LogoObjectKey,
		"founding_year":           o.FoundingYear,
		"company_type":            o.CompanyType,
		"premarket_fit":           o.PreMarketFit,
		"uses_cloud_providers":    o.UsesCloudProviders,
		"ai_focused":              o.AIFocused,
		"uses_ai_generated_code":  o.UsesAIGeneratedCode,
		"vc_backed":               o.VCBacked,
		"has_raised_money":        o.HasRaisedMoney,
		"has_enterprise_accounts": o.HasEnterpriseAccounts,
		"updated_at":              o.UpdatedAt,
	}

	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update organization: %w", err)
	}

	return nil
}
