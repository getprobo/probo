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
	"go.probo.inc/probo/pkg/crypto/hash"
	"go.probo.inc/probo/pkg/crypto/rand"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/policy"
	"go.probo.inc/probo/pkg/page"
)

type (
	AgreementSite struct {
		ID              gid.GID           `db:"id"`
		OrganizationID  gid.GID           `db:"organization_id"`
		Name            string            `db:"name"`
		Mode            AgreementSiteMode `db:"mode"`
		HashedSecretKey []byte            `db:"hashed_secret_key"`
		AllowedOrigins  []string          `db:"allowed_origins"`
		CreatedAt       time.Time         `db:"created_at"`
		UpdatedAt       time.Time         `db:"updated_at"`
	}

	AgreementSites []*AgreementSite
)

const secretKeyPrefix = "sk_live_"
const testSecretKeyPrefix = "sk_test_"

// GenerateSecretKey creates a new random secret key for the site, stores its
// SHA-256 hash, and returns the plaintext. The plaintext is returned only once.
func (s *AgreementSite) GenerateSecretKey() (string, error) {
	raw, err := rand.HexString(32)
	if err != nil {
		return "", fmt.Errorf("cannot generate secret key: %w", err)
	}

	prefix := secretKeyPrefix
	if s.Mode == AgreementSiteModeTest {
		prefix = testSecretKeyPrefix
	}

	key := prefix + raw
	s.HashedSecretKey = []byte(hash.SHA256Hex([]byte(key)))

	return key, nil
}

func (s *AgreementSite) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM agreement_sites WHERE id = ANY(@resource_ids::text[])`

	args := pgx.StrictNamedArgs{
		"resource_ids": resourceIDs,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query agreement site authorization attributes: %w", err)
	}
	defer rows.Close()

	attrsByID := make(policy.AttributesByID)

	for rows.Next() {
		var id, organizationID gid.GID
		if err := rows.Scan(&id, &organizationID); err != nil {
			return nil, fmt.Errorf("cannot scan agreement site authorization attributes: %w", err)
		}

		attrsByID[id] = policy.Attributes{
			"organization_id": organizationID.String(),
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate agreement site authorization attributes: %w", err)
	}

	return attrsByID, nil
}

func (s *AgreementSite) CursorKey(orderBy AgreementSiteOrderField) page.CursorKey {
	switch orderBy {
	case AgreementSiteOrderFieldCreatedAt:
		return page.NewCursorKey(s.ID, s.CreatedAt)
	case AgreementSiteOrderFieldName:
		return page.NewCursorKey(s.ID, s.Name)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (s *AgreementSite) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO agreement_sites (
	id, tenant_id, organization_id, name, mode, hashed_secret_key, allowed_origins,
	created_at, updated_at
) VALUES (
	@id, @tenant_id, @organization_id, @name, @mode, @hashed_secret_key, @allowed_origins,
	@created_at, @updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                s.ID,
		"tenant_id":         scope.GetTenantID(),
		"organization_id":   s.OrganizationID,
		"name":              s.Name,
		"mode":              s.Mode,
		"hashed_secret_key": s.HashedSecretKey,
		"allowed_origins":   s.AllowedOrigins,
		"created_at":        s.CreatedAt,
		"updated_at":        s.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert agreement site: %w", err)
	}

	return nil
}

func (s *AgreementSite) Update(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE agreement_sites SET
	name = @name,
	allowed_origins = @allowed_origins,
	updated_at = @updated_at
WHERE %s AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":              s.ID,
		"name":            s.Name,
		"allowed_origins": s.AllowedOrigins,
		"updated_at":      s.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update agreement site: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (s *AgreementSite) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	id gid.GID,
) error {
	q := `
SELECT
	id, organization_id, name, mode, hashed_secret_key, allowed_origins,
	created_at, updated_at
FROM agreement_sites
WHERE %s AND id = @id
LIMIT 1
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": id}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query agreement site: %w", err)
	}

	site, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AgreementSite])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect agreement site: %w", err)
	}

	*s = site

	return nil
}

// LoadByHashedSecretKey loads a site by the SHA-256 hash of its secret key.
// Used to authenticate data-plane API requests.
func (s *AgreementSite) LoadByHashedSecretKey(
	ctx context.Context,
	conn pg.Querier,
	hashedKey []byte,
) error {
	q := `
SELECT
	id, organization_id, name, mode, hashed_secret_key, allowed_origins,
	created_at, updated_at
FROM agreement_sites
WHERE hashed_secret_key = @hashed_secret_key
LIMIT 1
`

	args := pgx.StrictNamedArgs{"hashed_secret_key": hashedKey}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query agreement site by key: %w", err)
	}

	site, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AgreementSite])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect agreement site: %w", err)
	}

	*s = site

	return nil
}

func (ss *AgreementSites) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[AgreementSiteOrderField],
) error {
	q := `
SELECT
	id, organization_id, name, mode, hashed_secret_key, allowed_origins,
	created_at, updated_at
FROM agreement_sites
WHERE
	%s
	AND organization_id = @organization_id
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query agreement sites: %w", err)
	}

	sites, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[AgreementSite])
	if err != nil {
		return fmt.Errorf("cannot collect agreement sites: %w", err)
	}

	*ss = sites

	return nil
}
