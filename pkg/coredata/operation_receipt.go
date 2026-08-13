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
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type OperationReceipt struct {
	ID             gid.GID   `db:"id"`
	OrganizationID gid.GID   `db:"organization_id"`
	OperationKey   string    `db:"operation_key"`
	CreatedAt      time.Time `db:"created_at"`
}

func NewOperationReceipt(
	scope Scoper,
	organizationID gid.GID,
	operationKey string,
) *OperationReceipt {
	return &OperationReceipt{
		ID:             gid.New(scope.GetTenantID(), OperationReceiptEntityType),
		OrganizationID: organizationID,
		OperationKey:   operationKey,
		CreatedAt:      time.Now(),
	}
}

// Claim inserts a receipt in the caller's transaction. A false result means
// the operation has already committed and all transactional side effects must
// be skipped.
func (r *OperationReceipt) Claim(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) (bool, error) {
	if r.OperationKey == "" {
		return false, fmt.Errorf("operation key must not be empty")
	}

	if r.OrganizationID.TenantID() != scope.GetTenantID() {
		return false, fmt.Errorf("operation receipt organization is outside scope")
	}

	q := `
INSERT INTO operation_receipts (
	id, tenant_id, organization_id, operation_key, created_at
) VALUES (
	@id, @tenant_id, @organization_id, @operation_key, @created_at
)
ON CONFLICT (organization_id, operation_key) DO NOTHING
`

	result, err := tx.Exec(
		ctx,
		q,
		pgx.StrictNamedArgs{
			"id":              r.ID,
			"tenant_id":       scope.GetTenantID(),
			"organization_id": r.OrganizationID,
			"operation_key":   r.OperationKey,
			"created_at":      r.CreatedAt,
		},
	)
	if err != nil {
		return false, fmt.Errorf("cannot claim operation receipt: %w", err)
	}

	return result.RowsAffected() > 0, nil
}

func DeleteOperationReceiptsBefore(
	ctx context.Context,
	conn pg.Querier,
	before time.Time,
	limit int,
) (int64, error) {
	q := `
WITH doomed AS (
	SELECT id
	FROM operation_receipts
	WHERE created_at < @before
	ORDER BY created_at ASC, id ASC
	LIMIT @limit
)
DELETE FROM operation_receipts
WHERE id IN (SELECT id FROM doomed)
`

	result, err := conn.Exec(
		ctx,
		q,
		pgx.StrictNamedArgs{
			"before": before,
			"limit":  limit,
		},
	)
	if err != nil {
		return 0, fmt.Errorf("cannot delete operation receipts: %w", err)
	}

	return result.RowsAffected(), nil
}

func (r *OperationReceipt) AuthorizationAttributes(
	_ context.Context,
	_ pg.Querier,
) (map[string]string, error) {
	return map[string]string{"organization_id": r.OrganizationID.String()}, nil
}
