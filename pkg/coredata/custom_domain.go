// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/policy"
	"go.probo.inc/probo/pkg/page"
)

type (
	// CustomDomain is a domain owned by an organization used to serve its
	// compliance portal. Its TLS certificate lifecycle is owned by the generic
	// certificates table, referenced through CertificateID.
	CustomDomain struct {
		ID             gid.GID   `db:"id"`
		OrganizationID gid.GID   `db:"organization_id"`
		Domain         string    `db:"domain"`
		Managed        bool      `db:"managed"`
		CertificateID  *gid.GID  `db:"certificate_id"`
		CreatedAt      time.Time `db:"created_at"`
		UpdatedAt      time.Time `db:"updated_at"`
	}

	CustomDomains []*CustomDomain
)

func NewCustomDomain(
	tenantID gid.TenantID,
	organizationID gid.GID,
	domain string,
	managed bool,
) *CustomDomain {
	now := time.Now()

	return &CustomDomain{
		ID:             gid.New(tenantID, CustomDomainEntityType),
		OrganizationID: organizationID,
		Domain:         domain,
		Managed:        managed,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// AuthorizationAttributes returns the authorization attributes for policy evaluation.
func (cd *CustomDomain) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id, managed FROM custom_domains WHERE id = ANY(@resource_ids::text[])`

	args := pgx.StrictNamedArgs{
		"resource_ids": resourceIDs,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query authorization attributes: %w", err)
	}

	defer rows.Close()

	attrsByID := make(policy.AttributesByID)

	for rows.Next() {
		var (
			id, organizationID gid.GID
			managed            bool
		)

		if err := rows.Scan(&id, &organizationID, &managed); err != nil {
			return nil, fmt.Errorf("cannot scan authorization attributes: %w", err)
		}

		attrsByID[id] = policy.Attributes{
			"organization_id": organizationID.String(),
			"managed":         strconv.FormatBool(managed),
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate authorization attributes: %w", err)
	}

	return attrsByID, nil
}

func (cd *CustomDomain) CursorKey(field CustomDomainOrderField) page.CursorKey {
	switch field {
	case CustomDomainOrderFieldCreatedAt:
		return page.NewCursorKey(cd.ID, cd.CreatedAt)
	case CustomDomainOrderFieldDomain:
		return page.NewCursorKey(cd.ID, cd.Domain)
	case CustomDomainOrderFieldUpdatedAt:
		return page.NewCursorKey(cd.ID, cd.UpdatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", field))
}

func (cd *CustomDomain) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	domainID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	domain,
	managed,
	certificate_id,
	created_at,
	updated_at
FROM
	custom_domains
WHERE
	%s
	AND id = @id
LIMIT 1
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{"id": domainID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query custom domain: %w", err)
	}

	customDomain, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CustomDomain])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect custom domain: %w", err)
	}

	*cd = customDomain

	return nil
}

func (cd *CustomDomain) LoadByDomain(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	domain string,
) error {
	q := `
SELECT
	id,
	organization_id,
	domain,
	managed,
	certificate_id,
	created_at,
	updated_at
FROM
	custom_domains
WHERE
	%s
	AND domain = @domain
LIMIT 1
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{"domain": domain}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query custom domain: %w", err)
	}

	customDomain, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CustomDomain])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect custom domain: %w", err)
	}

	*cd = customDomain

	return nil
}

func (domains *CustomDomains) LoadByIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	ids []gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	domain,
	managed,
	certificate_id,
	created_at,
	updated_at
FROM
	custom_domains
WHERE
	%s
	AND id = ANY(@ids)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{"ids": ids}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query custom domains: %w", err)
	}

	result, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[CustomDomain])
	if err != nil {
		return fmt.Errorf("cannot collect custom domains: %w", err)
	}

	*domains = result

	return nil
}

func (cd *CustomDomain) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO custom_domains (
	id,
	tenant_id,
	organization_id,
	domain,
	managed,
	certificate_id,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@domain,
	@managed,
	@certificate_id,
	@created_at,
	@updated_at
)
`

	args := pgx.NamedArgs{
		"id":              cd.ID,
		"tenant_id":       scope.GetTenantID(),
		"organization_id": cd.OrganizationID,
		"domain":          cd.Domain,
		"managed":         cd.Managed,
		"certificate_id":  cd.CertificateID,
		"created_at":      cd.CreatedAt,
		"updated_at":      cd.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "custom_domains_domain_key" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot insert custom domain: %w", err)
	}

	return nil
}

func (cd *CustomDomain) Update(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE
	custom_domains
SET
	domain = @domain,
	managed = @managed,
	certificate_id = @certificate_id,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{
		"id":             cd.ID,
		"domain":         cd.Domain,
		"managed":        cd.Managed,
		"certificate_id": cd.CertificateID,
		"updated_at":     time.Now(),
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update custom domain: %w", err)
	}

	return nil
}

func (cd *CustomDomain) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
DELETE FROM
	custom_domains
WHERE
	%s
	AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{"id": cd.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete custom domain: %w", err)
	}

	return nil
}

// LoadReferencedCertificateIDs returns certificate IDs currently linked from
// any custom domain. Used by the certificate cache warmer to drop orphaned
// cache rows without joining across entity tables.
func (domains *CustomDomains) LoadReferencedCertificateIDs(
	ctx context.Context,
	conn pg.Querier,
) ([]gid.GID, error) {
	q := `
SELECT DISTINCT
	certificate_id
FROM
	custom_domains
WHERE
	certificate_id IS NOT NULL
`

	rows, err := conn.Query(ctx, q, pgx.NamedArgs{})
	if err != nil {
		return nil, fmt.Errorf("cannot query referenced certificate ids: %w", err)
	}

	certificateIDs, err := pgx.CollectRows(rows, pgx.RowTo[gid.GID])
	if err != nil {
		return nil, fmt.Errorf("cannot collect referenced certificate ids: %w", err)
	}

	return certificateIDs, nil
}
