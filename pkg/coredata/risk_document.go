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
	"fmt"
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type (
	RiskDocument struct {
		RiskID         gid.GID      `db:"risk_id"`
		DocumentID     gid.GID      `db:"document_id"`
		OrganizationID gid.GID      `db:"organization_id"`
		TenantID       gid.TenantID `db:"tenant_id"`
		CreatedAt      time.Time    `db:"created_at"`
	}

	RiskDocuments []*RiskDocument
)

func (rp RiskDocument) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO
    risks_documents (
        risk_id,
        document_id,
        organization_id,
        tenant_id,
        created_at
    )
VALUES (
    @risk_id,
    @document_id,
    @organization_id,
    @tenant_id,
    @created_at
);
`

	args := pgx.StrictNamedArgs{
		"risk_id":         rp.RiskID,
		"document_id":     rp.DocumentID,
		"organization_id": rp.OrganizationID,
		"tenant_id":       scope.GetTenantID(),
		"created_at":      rp.CreatedAt,
	}
	_, err := conn.Exec(ctx, q, args)

	return err
}

func (rp RiskDocument) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	riskID gid.GID,
	documentID gid.GID,
) error {
	q := `
DELETE
FROM
    risks_documents
WHERE
    %s
    AND risk_id = @risk_id
    AND document_id = @document_id;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"risk_id":     riskID,
		"document_id": documentID,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)

	return err
}

func (rp RiskDocument) DeleteByDocumentIDs(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	documentIDs []gid.GID,
) error {
	q := `
DELETE
FROM
    risks_documents
WHERE
    %s
    AND document_id = ANY(@document_ids);
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"document_ids": documentIDs,
	}
	maps.Copy(args, scope.SQLArguments())

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot delete risk document mappings by document ids: %w", err)
	}

	return nil
}
