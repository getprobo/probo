// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	ThirdPartyThirdParty struct {
		ParentThirdPartyID gid.GID      `db:"parent_third_party_id"`
		ChildThirdPartyID  gid.GID      `db:"child_third_party_id"`
		TenantID           gid.TenantID `db:"tenant_id"`
		CreatedAt          time.Time    `db:"created_at"`
		Purpose            *string      `db:"purpose"`
	}

	ThirdPartyThirdParties []*ThirdPartyThirdParty
)

func (r *ThirdPartyThirdParty) Insert(ctx context.Context, conn pg.Tx, scope Scoper) error {
	q := `
INSERT INTO third_party_third_parties (
    parent_third_party_id,
    child_third_party_id,
    tenant_id,
    created_at,
    purpose
) VALUES (
    @parent_third_party_id,
    @child_third_party_id,
    @tenant_id,
    @created_at,
    @purpose
)
ON CONFLICT (parent_third_party_id, child_third_party_id) DO UPDATE SET
    purpose = COALESCE(EXCLUDED.purpose, third_party_third_parties.purpose)
`

	args := pgx.StrictNamedArgs{
		"parent_third_party_id": r.ParentThirdPartyID,
		"child_third_party_id":  r.ChildThirdPartyID,
		"tenant_id":             scope.GetTenantID(),
		"created_at":            r.CreatedAt,
		"purpose":               r.Purpose,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert third party third party: %w", err)
	}

	return nil
}

func (r *ThirdPartyThirdParty) Delete(ctx context.Context, conn pg.Tx, scope Scoper) error {
	q := `
DELETE FROM third_party_third_parties
WHERE %s
    AND parent_third_party_id = @parent_third_party_id
    AND child_third_party_id = @child_third_party_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"parent_third_party_id": r.ParentThirdPartyID,
		"child_third_party_id":  r.ChildThirdPartyID,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)

	return err
}

func (v *ThirdParties) CountByParentThirdPartyID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	parentThirdPartyID gid.GID,
) (int, error) {
	q := `
WITH children AS (
	SELECT
		tp.id,
		tp.tenant_id
	FROM
		third_parties tp
	INNER JOIN
		third_party_third_parties tpr ON tp.id = tpr.child_third_party_id
	WHERE
		tpr.parent_third_party_id = @parent_third_party_id
)
SELECT
	COUNT(id)
FROM
	children
WHERE %s
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"parent_third_party_id": parentThirdPartyID}
	maps.Copy(args, scope.SQLArguments())

	var count int

	err := conn.QueryRow(ctx, q, args).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("cannot count child third parties: %w", err)
	}

	return count, nil
}

func (v *ThirdParties) LoadByParentThirdPartyID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	parentThirdPartyID gid.GID,
	cursor *page.Cursor[ThirdPartyOrderField],
) error {
	q := `
WITH children AS (
	SELECT
		tp.id,
		tp.tenant_id,
		tp.organization_id,
		tp.common_third_party_id,
		tp.name,
		tp.description,
		tp.category,
		tp.headquarter_address,
		tp.legal_name,
		tp.website_url,
		tp.privacy_policy_url,
		tp.service_level_agreement_url,
		tp.data_processing_agreement_url,
		tp.business_associate_agreement_url,
		tp.subprocessors_list_url,
		tp.certifications,
		tp.countries,
		tp.business_owner_profile_id,
		tp.security_owner_profile_id,
		tp.status_page_url,
		tp.terms_of_service_url,
		tp.security_page_url,
		tp.trust_page_url,
		tp.show_on_trust_center,
		tp.first_level,
		tp.vetting_status,
		tp.vetting_website_url,
		tp.vetting_procedure,
		tp.vetting_processing_started_at,
		tp.vetting_error_message,
		tp.created_at,
		tp.updated_at
	FROM
		third_parties tp
	INNER JOIN
		third_party_third_parties tpr ON tp.id = tpr.child_third_party_id
	WHERE
		tpr.parent_third_party_id = @parent_third_party_id
)
SELECT
	id,
	organization_id,
	common_third_party_id,
	name,
	description,
	category,
	headquarter_address,
	legal_name,
	website_url,
	privacy_policy_url,
	service_level_agreement_url,
	data_processing_agreement_url,
	business_associate_agreement_url,
	subprocessors_list_url,
	certifications,
	countries,
	business_owner_profile_id,
	security_owner_profile_id,
	status_page_url,
	terms_of_service_url,
	security_page_url,
	trust_page_url,
	show_on_trust_center,
	first_level,
	vetting_status,
	vetting_website_url,
	vetting_procedure,
	vetting_processing_started_at,
	vetting_error_message,
	created_at,
	updated_at
FROM
	children
WHERE %s
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"parent_third_party_id": parentThirdPartyID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query child third parties: %w", err)
	}

	thirdParties, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ThirdParty])
	if err != nil {
		return fmt.Errorf("cannot collect child third parties: %w", err)
	}

	*v = thirdParties

	return nil
}
