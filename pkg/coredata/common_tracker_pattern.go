// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type (
	CommonTrackerPattern struct {
		ID                    gid.GID                 `db:"id"`
		CommonThirdPartyID    *gid.GID                `db:"common_third_party_id"`
		TrackerType           TrackerType             `db:"tracker_type"`
		Pattern               string                  `db:"pattern"`
		MatchType             TrackerPatternMatchType `db:"match_type"`
		Description           string                  `db:"description"`
		MaxAgeSeconds         *int                    `db:"max_age_seconds"`
		Confidence            float32                 `db:"confidence"`
		EnrichmentRequestedAt *time.Time              `db:"enrichment_requested_at"`
		EnrichedAt            *time.Time              `db:"enriched_at"`
		CreatedAt             time.Time               `db:"created_at"`
		UpdatedAt             time.Time               `db:"updated_at"`
	}

	CommonTrackerPatterns []*CommonTrackerPattern
)

func (p *CommonTrackerPattern) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	id gid.GID,
) error {
	q := `
SELECT
    id,
    common_third_party_id,
    tracker_type,
    pattern,
    match_type,
    description,
    max_age_seconds,
    confidence,
    enrichment_requested_at,
    enriched_at,
    created_at,
    updated_at
FROM
    common_tracker_patterns
WHERE
    id = @id
LIMIT 1;
`

	args := pgx.StrictNamedArgs{"id": id}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query common tracker pattern: %w", err)
	}

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CommonTrackerPattern])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect common tracker pattern: %w", err)
	}

	*p = row

	return nil
}

func (p *CommonTrackerPattern) LoadByPattern(
	ctx context.Context,
	conn pg.Querier,
	trackerType TrackerType,
	pattern string,
	maxAgeSeconds *int,
) error {
	q := `
SELECT
    id,
    common_third_party_id,
    tracker_type,
    pattern,
    match_type,
    description,
    max_age_seconds,
    confidence,
    enrichment_requested_at,
    enriched_at,
    created_at,
    updated_at
FROM
    common_tracker_patterns
WHERE
    tracker_type = @tracker_type
    AND pattern = @pattern
    AND COALESCE(max_age_seconds, -1) = COALESCE(@max_age_seconds, -1)
LIMIT 1;
`

	args := pgx.StrictNamedArgs{
		"tracker_type":    trackerType,
		"pattern":         pattern,
		"max_age_seconds": maxAgeSeconds,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query common tracker pattern: %w", err)
	}

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CommonTrackerPattern])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect common tracker pattern: %w", err)
	}

	*p = row

	return nil
}

func (p CommonTrackerPattern) Insert(
	ctx context.Context,
	conn pg.Tx,
) error {
	q := `
INSERT INTO common_tracker_patterns (
    id,
    common_third_party_id,
    tracker_type,
    pattern,
    match_type,
    description,
    max_age_seconds,
    confidence,
    enrichment_requested_at,
    enriched_at,
    created_at,
    updated_at
) VALUES (
    @id,
    @common_third_party_id,
    @tracker_type,
    @pattern,
    @match_type,
    @description,
    @max_age_seconds,
    @confidence,
    @enrichment_requested_at,
    @enriched_at,
    @created_at,
    @updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                      p.ID,
		"common_third_party_id":   p.CommonThirdPartyID,
		"tracker_type":            p.TrackerType,
		"pattern":                 p.Pattern,
		"match_type":              p.MatchType,
		"description":             p.Description,
		"max_age_seconds":         p.MaxAgeSeconds,
		"confidence":              p.Confidence,
		"enrichment_requested_at": p.EnrichmentRequestedAt,
		"enriched_at":             p.EnrichedAt,
		"created_at":              p.CreatedAt,
		"updated_at":              p.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert common tracker pattern: %w", err)
	}

	return nil
}

func (p *CommonTrackerPattern) Upsert(
	ctx context.Context,
	conn pg.Tx,
) (inserted bool, err error) {
	// On insert, a description-less row is immediately queued for the
	// enrichment worker (enrichment_requested_at = NOW()). On conflict the
	// enrichment columns are left untouched, and an empty incoming
	// description never overwrites an existing one — descriptions are owned
	// by the enrichment worker, so mapping-side upserts must not clobber a
	// researched description with an empty string.
	q := `
INSERT INTO common_tracker_patterns (
    id,
    common_third_party_id,
    tracker_type,
    pattern,
    match_type,
    description,
    max_age_seconds,
    confidence,
    enrichment_requested_at,
    enriched_at,
    created_at,
    updated_at
) VALUES (
    @id,
    @common_third_party_id,
    @tracker_type,
    @pattern,
    @match_type,
    @description,
    @max_age_seconds,
    @confidence,
    CASE WHEN @description = '' THEN NOW() ELSE NULL END,
    NULL,
    @created_at,
    @updated_at
)
ON CONFLICT (tracker_type, pattern, COALESCE(max_age_seconds, -1)) DO UPDATE
SET
    common_third_party_id = EXCLUDED.common_third_party_id,
    match_type            = EXCLUDED.match_type,
    description           = CASE
        WHEN EXCLUDED.description = '' THEN common_tracker_patterns.description
        ELSE EXCLUDED.description
    END,
    confidence            = EXCLUDED.confidence,
    updated_at            = EXCLUDED.updated_at
RETURNING
    id,
    common_third_party_id,
    tracker_type,
    pattern,
    match_type,
    description,
    max_age_seconds,
    confidence,
    enrichment_requested_at,
    enriched_at,
    created_at,
    updated_at
`

	originalID := p.ID

	args := pgx.StrictNamedArgs{
		"id":                    p.ID,
		"common_third_party_id": p.CommonThirdPartyID,
		"tracker_type":          p.TrackerType,
		"pattern":               p.Pattern,
		"match_type":            p.MatchType,
		"description":           p.Description,
		"max_age_seconds":       p.MaxAgeSeconds,
		"confidence":            p.Confidence,
		"created_at":            p.CreatedAt,
		"updated_at":            p.UpdatedAt,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return false, fmt.Errorf("cannot upsert common tracker pattern: %w", err)
	}
	defer rows.Close()

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CommonTrackerPattern])
	if err != nil {
		return false, fmt.Errorf("cannot collect upsert result: %w", err)
	}

	*p = row

	return originalID == p.ID, nil
}

func (p CommonTrackerPattern) Delete(
	ctx context.Context,
	conn pg.Tx,
	id gid.GID,
) error {
	q := `DELETE FROM common_tracker_patterns WHERE id = @id`

	args := pgx.StrictNamedArgs{"id": id}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete common tracker pattern: %w", err)
	}

	return nil
}

func (ps *CommonTrackerPatterns) FindMatchingPattern(
	ctx context.Context,
	conn pg.Querier,
	trackerType TrackerType,
	identifier string,
) (*CommonTrackerPattern, error) {
	q := `
SELECT
    id,
    common_third_party_id,
    tracker_type,
    pattern,
    match_type,
    description,
    max_age_seconds,
    confidence,
    enrichment_requested_at,
    enriched_at,
    created_at,
    updated_at
FROM
    common_tracker_patterns
WHERE
    tracker_type = @tracker_type
    AND (
        (match_type = @match_type_glob
         AND @identifier LIKE
             replace(replace(replace(replace(
                 pattern, E'\\', E'\\\\'), '%', E'\\%'), '_', E'\\_'), '*', '%')
             ESCAPE E'\\')
        OR (match_type = @match_type_exact AND pattern = @identifier)
    )
ORDER BY
    CASE WHEN match_type = @match_type_exact AND pattern = @identifier THEN 0
         ELSE 1
    END,
    length(replace(pattern, '*', '')) DESC
LIMIT 1;
`

	args := pgx.StrictNamedArgs{
		"tracker_type":     trackerType,
		"identifier":       identifier,
		"match_type_glob":  TrackerPatternMatchTypeGlob,
		"match_type_exact": TrackerPatternMatchTypeExact,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query common tracker patterns: %w", err)
	}

	pattern, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CommonTrackerPattern])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("cannot collect common tracker pattern: %w", err)
	}

	return &pattern, nil
}

type CommonTrackerPatternSearchResult struct {
	Pattern        string      `db:"pattern"`
	Description    string      `db:"description"`
	TrackerType    TrackerType `db:"tracker_type"`
	ThirdPartyName *string     `db:"third_party_name"`
	Confidence     float32     `db:"confidence"`
}

func (ps *CommonTrackerPatterns) FindByKeyword(
	ctx context.Context,
	conn pg.Querier,
	fragment string,
	limit int,
) ([]CommonTrackerPatternSearchResult, error) {
	if limit <= 0 || limit > 20 {
		limit = 10
	}

	q := `
SELECT
    ctp.pattern,
    ctp.description,
    ctp.tracker_type,
    ct.name AS third_party_name,
    ctp.confidence
FROM
    common_tracker_patterns ctp
LEFT JOIN common_third_parties ct ON ct.id = ctp.common_third_party_id
WHERE
    ctp.pattern ILIKE '%' || @fragment || '%'
    OR ctp.description ILIKE '%' || @fragment || '%'
ORDER BY
    ctp.confidence DESC
LIMIT @limit;
`

	args := pgx.StrictNamedArgs{
		"fragment": fragment,
		"limit":    limit,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot search common tracker patterns: %w", err)
	}

	results, err := pgx.CollectRows(rows, pgx.RowToStructByName[CommonTrackerPatternSearchResult])
	if err != nil {
		return nil, fmt.Errorf("cannot collect common tracker pattern search results: %w", err)
	}

	return results, nil
}

func (ps *CommonTrackerPatterns) LoadByCommonThirdPartyID(
	ctx context.Context,
	conn pg.Querier,
	commonThirdPartyID gid.GID,
) error {
	q := `
SELECT
    id,
    common_third_party_id,
    tracker_type,
    pattern,
    match_type,
    description,
    max_age_seconds,
    confidence,
    enrichment_requested_at,
    enriched_at,
    created_at,
    updated_at
FROM
    common_tracker_patterns
WHERE
    common_third_party_id = @common_third_party_id
ORDER BY pattern ASC;
`

	args := pgx.StrictNamedArgs{"common_third_party_id": commonThirdPartyID}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query common tracker patterns: %w", err)
	}

	patterns, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[CommonTrackerPattern])
	if err != nil {
		return fmt.Errorf("cannot collect common tracker patterns: %w", err)
	}

	*ps = patterns

	return nil
}

// LoadNextForEnrichmentForUpdateSkipLocked claims the next common tracker
// pattern queued for description enrichment, oldest request first. It
// mirrors the mapping worker's claim pattern: the row is locked FOR
// UPDATE SKIP LOCKED so concurrent enrichment workers never pick the same
// row.
func (p *CommonTrackerPattern) LoadNextForEnrichmentForUpdateSkipLocked(
	ctx context.Context,
	tx pg.Tx,
) error {
	q := `
SELECT
    id,
    common_third_party_id,
    tracker_type,
    pattern,
    match_type,
    description,
    max_age_seconds,
    confidence,
    enrichment_requested_at,
    enriched_at,
    created_at,
    updated_at
FROM
    common_tracker_patterns
WHERE
    enrichment_requested_at IS NOT NULL
ORDER BY
    enrichment_requested_at ASC
FOR UPDATE SKIP LOCKED
LIMIT 1;
`

	rows, err := tx.Query(ctx, q)
	if err != nil {
		return fmt.Errorf("cannot query common tracker pattern for enrichment: %w", err)
	}

	pattern, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CommonTrackerPattern])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect common tracker pattern for enrichment: %w", err)
	}

	*p = pattern

	return nil
}

// ClearEnrichmentRequestedAt removes the row from the enrichment queue. It
// bumps updated_at so the stale-recovery clock starts at claim time.
func (p *CommonTrackerPattern) ClearEnrichmentRequestedAt(
	ctx context.Context,
	tx pg.Tx,
) error {
	q := `
UPDATE common_tracker_patterns
SET
    enrichment_requested_at = NULL,
    updated_at = NOW()
WHERE id = @id
`

	args := pgx.StrictNamedArgs{"id": p.ID}

	_, err := tx.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot clear enrichment requested at: %w", err)
	}

	p.EnrichmentRequestedAt = nil

	return nil
}

// SetEnriched records the researched description and marks the row
// terminally enriched so it is never re-queued.
func (p *CommonTrackerPattern) SetEnriched(
	ctx context.Context,
	tx pg.Tx,
	description string,
) error {
	q := `
UPDATE common_tracker_patterns
SET
    description = @description,
    enriched_at = NOW(),
    enrichment_requested_at = NULL,
    updated_at = NOW()
WHERE id = @id
`

	args := pgx.StrictNamedArgs{
		"id":          p.ID,
		"description": description,
	}

	result, err := tx.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot mark common tracker pattern enriched: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	p.Description = description

	return nil
}

// ResetStaleEnrichments re-queues rows whose enrichment was claimed but
// never completed (no enriched_at, still description-less) and have been
// idle longer than staleAfter, so a crashed or timed-out enrichment is
// retried.
func ResetStaleEnrichments(
	ctx context.Context,
	conn pg.Querier,
	staleAfter time.Duration,
) error {
	q := `
UPDATE common_tracker_patterns
SET
    enrichment_requested_at = NOW(),
    updated_at = NOW()
WHERE
    enrichment_requested_at IS NULL
    AND enriched_at IS NULL
    AND description = ''
    AND updated_at < @stale_before
`

	args := pgx.StrictNamedArgs{"stale_before": time.Now().Add(-staleAfter)}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot reset stale common tracker pattern enrichments: %w", err)
	}

	return nil
}

func (ps *CommonTrackerPatterns) LoadByIDs(
	ctx context.Context,
	conn pg.Querier,
	ids []gid.GID,
) error {
	q := `
SELECT
    id,
    common_third_party_id,
    tracker_type,
    pattern,
    match_type,
    description,
    max_age_seconds,
    confidence,
    enrichment_requested_at,
    enriched_at,
    created_at,
    updated_at
FROM
    common_tracker_patterns
WHERE
    id = ANY(@ids)
`

	args := pgx.StrictNamedArgs{"ids": ids}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query common tracker patterns: %w", err)
	}

	patterns, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[CommonTrackerPattern])
	if err != nil {
		return fmt.Errorf("cannot collect common tracker patterns: %w", err)
	}

	*ps = patterns

	return nil
}
