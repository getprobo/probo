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
	DocumentSigningConfigState string

	DocumentSigningConfigMode string

	DocumentSigningConfigKind string

	DocumentSigningReacceptancePolicy string

	DocumentSigningConfig struct {
		ID                 gid.GID                           `db:"id"`
		OrganizationID     gid.GID                           `db:"organization_id"`
		DocumentID         gid.GID                           `db:"document_id"`
		PublishableKey     string                            `db:"publishable_key"`
		State              DocumentSigningConfigState        `db:"state"`
		Mode               DocumentSigningConfigMode         `db:"mode"`
		Kind               DocumentSigningConfigKind         `db:"kind"`
		ReacceptancePolicy DocumentSigningReacceptancePolicy `db:"reacceptance_policy"`
		AllowedOrigins     []string                          `db:"allowed_origins"`
		CreatedAt          time.Time                         `db:"created_at"`
		UpdatedAt          time.Time                         `db:"updated_at"`
	}

	DocumentSigningConfigs []*DocumentSigningConfig
)

const (
	DocumentSigningConfigStateDraft    DocumentSigningConfigState = "DRAFT"
	DocumentSigningConfigStateActive   DocumentSigningConfigState = "ACTIVE"
	DocumentSigningConfigStateArchived DocumentSigningConfigState = "ARCHIVED"

	DocumentSigningConfigModeLive DocumentSigningConfigMode = "LIVE"
	DocumentSigningConfigModeTest DocumentSigningConfigMode = "TEST"

	DocumentSigningConfigKindTermsOfService      DocumentSigningConfigKind = "TERMS_OF_SERVICE"
	DocumentSigningConfigKindAcceptableUsePolicy DocumentSigningConfigKind = "ACCEPTABLE_USE_POLICY"
	DocumentSigningConfigKindOther               DocumentSigningConfigKind = "OTHER"

	DocumentSigningReacceptancePolicyOnMajor DocumentSigningReacceptancePolicy = "ON_MAJOR"
	DocumentSigningReacceptancePolicyOnAny   DocumentSigningReacceptancePolicy = "ON_ANY"
	DocumentSigningReacceptancePolicyNever   DocumentSigningReacceptancePolicy = "NEVER"
)

func (s DocumentSigningConfigState) String() string        { return string(s) }
func (m DocumentSigningConfigMode) String() string         { return string(m) }
func (k DocumentSigningConfigKind) String() string         { return string(k) }
func (p DocumentSigningReacceptancePolicy) String() string { return string(p) }

func (c *DocumentSigningConfig) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM document_signing_configs WHERE id = ANY(@resource_ids::text[])`

	args := pgx.StrictNamedArgs{
		"resource_ids": resourceIDs,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query document signing config authorization attributes: %w", err)
	}
	defer rows.Close()

	attrsByID := make(policy.AttributesByID)

	for rows.Next() {
		var id, organizationID gid.GID
		if err := rows.Scan(&id, &organizationID); err != nil {
			return nil, fmt.Errorf("cannot scan document signing config authorization attributes: %w", err)
		}

		attrsByID[id] = policy.Attributes{
			"organization_id": organizationID.String(),
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate document signing config authorization attributes: %w", err)
	}

	return attrsByID, nil
}

func (c *DocumentSigningConfig) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO document_signing_configs (
	id, tenant_id, organization_id, document_id,
	publishable_key,
	state, mode, kind, reacceptance_policy,
	allowed_origins,
	created_at, updated_at
) VALUES (
	@id, @tenant_id, @organization_id, @document_id,
	@publishable_key,
	@state, @mode, @kind, @reacceptance_policy,
	@allowed_origins,
	@created_at, @updated_at
)
`
	args := pgx.StrictNamedArgs{
		"id":                  c.ID,
		"tenant_id":           scope.GetTenantID(),
		"organization_id":     c.OrganizationID,
		"document_id":         c.DocumentID,
		"publishable_key":     c.PublishableKey,
		"state":               c.State,
		"mode":                c.Mode,
		"kind":                c.Kind,
		"reacceptance_policy": c.ReacceptancePolicy,
		"allowed_origins":     c.AllowedOrigins,
		"created_at":          c.CreatedAt,
		"updated_at":          c.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert document signing config: %w", err)
	}

	return nil
}

func (c *DocumentSigningConfig) Update(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE document_signing_configs SET
	state               = @state,
	mode                = @mode,
	kind                = @kind,
	reacceptance_policy = @reacceptance_policy,
	allowed_origins     = @allowed_origins,
	updated_at          = @updated_at
WHERE %s AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                  c.ID,
		"state":               c.State,
		"mode":                c.Mode,
		"kind":                c.Kind,
		"reacceptance_policy": c.ReacceptancePolicy,
		"allowed_origins":     c.AllowedOrigins,
		"updated_at":          c.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update document signing config: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (c *DocumentSigningConfig) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	id gid.GID,
) error {
	q := `
SELECT
	id, organization_id, document_id,
	publishable_key,
	state, mode, kind, reacceptance_policy,
	allowed_origins,
	created_at, updated_at
FROM document_signing_configs
WHERE %s AND id = @id
LIMIT 1
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": id}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query document signing config: %w", err)
	}

	cfg, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[DocumentSigningConfig])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect document signing config: %w", err)
	}

	*c = cfg

	return nil
}

func (c *DocumentSigningConfig) LoadByPublishableKey(
	ctx context.Context,
	conn pg.Querier,
	publishableKey string,
) error {
	q := `
SELECT
	id, organization_id, document_id,
	publishable_key,
	state, mode, kind, reacceptance_policy,
	allowed_origins,
	created_at, updated_at
FROM document_signing_configs
WHERE publishable_key = @publishable_key
LIMIT 1
`
	args := pgx.StrictNamedArgs{"publishable_key": publishableKey}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query document signing config by key: %w", err)
	}

	cfg, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[DocumentSigningConfig])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect document signing config: %w", err)
	}

	*c = cfg

	return nil
}

func (c *DocumentSigningConfig) LoadByDocumentID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	documentID gid.GID,
) error {
	q := `
SELECT
	id, organization_id, document_id,
	publishable_key,
	state, mode, kind, reacceptance_policy,
	allowed_origins,
	created_at, updated_at
FROM document_signing_configs
WHERE %s AND document_id = @document_id
LIMIT 1
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"document_id": documentID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query document signing config: %w", err)
	}

	cfg, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[DocumentSigningConfig])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect document signing config: %w", err)
	}

	*c = cfg

	return nil
}

type DocumentSigningConfigOrderField string

const DocumentSigningConfigOrderFieldCreatedAt DocumentSigningConfigOrderField = "created_at"

func (f DocumentSigningConfigOrderField) String() string { return string(f) }
func (f DocumentSigningConfigOrderField) Column() string { return string(f) }

func (c DocumentSigningConfig) CursorKey(orderBy DocumentSigningConfigOrderField) page.CursorKey {
	switch orderBy {
	case DocumentSigningConfigOrderFieldCreatedAt:
		return page.NewCursorKey(c.ID, c.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (cs *DocumentSigningConfigs) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[DocumentSigningConfigOrderField],
) error {
	q := `
SELECT
	id, organization_id, document_id,
	publishable_key,
	state, mode, kind, reacceptance_policy,
	allowed_origins,
	created_at, updated_at
FROM document_signing_configs
WHERE %s
	AND organization_id = @organization_id
	AND %s
ORDER BY created_at ASC, id ASC
LIMIT @limit
`
	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query document signing configs: %w", err)
	}

	configs, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[DocumentSigningConfig])
	if err != nil {
		return fmt.Errorf("cannot collect document signing configs: %w", err)
	}

	*cs = configs

	return nil
}
