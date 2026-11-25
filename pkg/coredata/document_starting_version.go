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
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type DocumentStartingVersion struct {
	DocumentID     gid.GID   `db:"document_id"`
	OrganizationID gid.GID   `db:"organization_id"`
	Major          int       `db:"major"`
	Minor          int       `db:"minor"`
	CreatedAt      time.Time `db:"created_at"`
}

// LoadByDocumentID loads the version number the document's first published
// version takes. It returns ErrResourceNotFound when none was chosen.
func (v *DocumentStartingVersion) LoadByDocumentID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	documentID gid.GID,
) error {
	q := `
SELECT
	document_id,
	organization_id,
	major,
	minor,
	created_at
FROM document_starting_versions
WHERE
	%s
	AND document_id = @document_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"document_id": documentID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query document starting version: %w", err)
	}

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[DocumentStartingVersion])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect document starting version: %w", err)
	}

	*v = *result

	return nil
}

func (v *DocumentStartingVersion) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO document_starting_versions (
	document_id,
	tenant_id,
	organization_id,
	major,
	minor,
	created_at
) VALUES (
	@document_id,
	@tenant_id,
	@organization_id,
	@major,
	@minor,
	@created_at
);
`

	args := pgx.StrictNamedArgs{
		"document_id":     v.DocumentID,
		"tenant_id":       scope.GetTenantID(),
		"organization_id": v.OrganizationID,
		"major":           v.Major,
		"minor":           v.Minor,
		"created_at":      v.CreatedAt,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot insert document starting version: %w", err)
	}

	return nil
}
