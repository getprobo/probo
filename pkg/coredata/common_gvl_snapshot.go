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
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type (
	CommonGVLSnapshot struct {
		ID                      gid.GID         `db:"id"`
		VendorListVersion       int             `db:"vendor_list_version"`
		GVLSpecificationVersion int             `db:"gvl_specification_version"`
		TCFPolicyVersion        int             `db:"tcf_policy_version"`
		LastUpdated             time.Time       `db:"last_updated"`
		FetchedAt               time.Time       `db:"fetched_at"`
		Payload                 json.RawMessage `db:"payload"`
	}
)

func (s *CommonGVLSnapshot) LoadByVendorListVersion(
	ctx context.Context,
	conn pg.Querier,
	version int,
) error {
	q := `
SELECT
    id,
    vendor_list_version,
    gvl_specification_version,
    tcf_policy_version,
    last_updated,
    fetched_at,
    payload
FROM
    common_gvl_snapshots
WHERE
    vendor_list_version = @vendor_list_version
LIMIT 1;
`

	args := pgx.StrictNamedArgs{"vendor_list_version": version}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query common gvl snapshot: %w", err)
	}

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CommonGVLSnapshot])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect common gvl snapshot: %w", err)
	}

	*s = row

	return nil
}

func (s *CommonGVLSnapshot) Insert(
	ctx context.Context,
	conn pg.Querier,
) error {
	q := `
INSERT INTO common_gvl_snapshots (
    id,
    vendor_list_version,
    gvl_specification_version,
    tcf_policy_version,
    last_updated,
    fetched_at,
    payload
) VALUES (
    @id,
    @vendor_list_version,
    @gvl_specification_version,
    @tcf_policy_version,
    @last_updated,
    @fetched_at,
    @payload
)
`

	args := pgx.StrictNamedArgs{
		"id":                        s.ID,
		"vendor_list_version":       s.VendorListVersion,
		"gvl_specification_version": s.GVLSpecificationVersion,
		"tcf_policy_version":        s.TCFPolicyVersion,
		"last_updated":              s.LastUpdated,
		"fetched_at":                s.FetchedAt,
		"payload":                   s.Payload,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot insert common gvl snapshot: %w", err)
	}

	return nil
}
