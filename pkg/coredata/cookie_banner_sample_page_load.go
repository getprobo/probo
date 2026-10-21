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
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type SamplePageLoad struct {
	BrowserFamily BrowserFamily `db:"browser_family"`
	Presentation  string        `db:"presentation"`
	ConsentAction string        `db:"consent_action"`
	Count         int           `db:"count"`
}

func RecordSamplePageLoad(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
	sampleID gid.GID,
	family BrowserFamily,
	presentation string,
	consentAction string,
	count int64,
) error {
	if count == 0 {
		return nil
	}

	now := time.Now()

	q := `
WITH open_sample AS (
	SELECT id
	FROM cookie_banner_samples
	WHERE id = @sample_id
	  AND frozen_at IS NULL
	FOR UPDATE
)
INSERT INTO cookie_banner_sample_page_loads (
	sample_id,
	browser_family,
	presentation,
	consent_action,
	tenant_id,
	count,
	created_at,
	updated_at
)
SELECT
	@sample_id,
	@browser_family,
	@presentation,
	@consent_action,
	@tenant_id,
	@count,
	@now,
	@now
FROM open_sample
ON CONFLICT (sample_id, browser_family, presentation, consent_action) DO UPDATE
SET
	count = cookie_banner_sample_page_loads.count + EXCLUDED.count,
	updated_at = EXCLUDED.updated_at
WHERE EXISTS (
	SELECT 1
	FROM cookie_banner_samples
	WHERE id = cookie_banner_sample_page_loads.sample_id
	  AND frozen_at IS NULL
)
`

	args := pgx.StrictNamedArgs{
		"sample_id":      sampleID,
		"browser_family": family,
		"presentation":   presentation,
		"consent_action": consentAction,
		"tenant_id":      scope.GetTenantID(),
		"count":          count,
		"now":            now,
	}

	if _, err := tx.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot record sample page load: %w", err)
	}

	return nil
}

func LoadSamplePageLoads(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	sampleID gid.GID,
) ([]SamplePageLoad, error) {
	q := `
SELECT
	browser_family,
	presentation,
	consent_action,
	count
FROM cookie_banner_sample_page_loads
WHERE
	%s
	AND sample_id = @sample_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"sample_id": sampleID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query sample page loads: %w", err)
	}

	loads, err := pgx.CollectRows(rows, pgx.RowToStructByName[SamplePageLoad])
	if err != nil {
		return nil, fmt.Errorf("cannot collect sample page loads: %w", err)
	}

	return loads, nil
}

type SamplePageLoadBySample struct {
	SampleID      gid.GID       `db:"sample_id"`
	BrowserFamily BrowserFamily `db:"browser_family"`
	Presentation  string        `db:"presentation"`
	ConsentAction string        `db:"consent_action"`
	Count         int           `db:"count"`
}

func LoadSamplePageLoadsBySampleIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	sampleIDs []gid.GID,
) ([]SamplePageLoadBySample, error) {
	if len(sampleIDs) == 0 {
		return nil, nil
	}

	q := `
SELECT
	sample_id,
	browser_family,
	presentation,
	consent_action,
	count
FROM cookie_banner_sample_page_loads
WHERE
	%s
	AND sample_id = ANY(@sample_ids)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"sample_ids": sampleIDs}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query sample page loads: %w", err)
	}

	loads, err := pgx.CollectRows(rows, pgx.RowToStructByName[SamplePageLoadBySample])
	if err != nil {
		return nil, fmt.Errorf("cannot collect sample page loads: %w", err)
	}

	return loads, nil
}
