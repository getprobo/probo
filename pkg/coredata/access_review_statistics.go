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

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type AccessReviewStatistics struct {
	TotalCount           int
	DecisionCounts       map[AccessReviewEntryDecision]int
	FlagCounts           map[AccessReviewEntryFlag]int
	IncrementalTagCounts map[AccessReviewEntryIncrementalTag]int
}

func (s *AccessReviewStatistics) LoadByCampaignID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	campaignID gid.GID,
) error {
	args := pgx.StrictNamedArgs{"campaign_id": campaignID}
	maps.Copy(args, scope.SQLArguments())

	s.DecisionCounts = make(map[AccessReviewEntryDecision]int)
	s.FlagCounts = make(map[AccessReviewEntryFlag]int)
	s.IncrementalTagCounts = make(map[AccessReviewEntryIncrementalTag]int)
	s.TotalCount = 0

	q := `
SELECT decision, COUNT(*) as count
FROM access_review_entries
WHERE
    %s
    AND access_review_campaign_id = @campaign_id
GROUP BY decision;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query access entry decision counts: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			decision AccessReviewEntryDecision
			count    int
		)

		if err := rows.Scan(&decision, &count); err != nil {
			return fmt.Errorf("cannot scan decision count: %w", err)
		}

		s.DecisionCounts[decision] = count
		s.TotalCount += count
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("cannot iterate decision counts: %w", err)
	}

	q = `
SELECT f, COUNT(*) as count
FROM access_review_entries, unnest(flags) AS f
WHERE
    %s
    AND access_review_campaign_id = @campaign_id
GROUP BY f;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	rows, err = conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query access entry flag counts: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			flag  AccessReviewEntryFlag
			count int
		)

		if err := rows.Scan(&flag, &count); err != nil {
			return fmt.Errorf("cannot scan flag count: %w", err)
		}

		s.FlagCounts[flag] = count
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("cannot iterate flag counts: %w", err)
	}

	q = `
SELECT incremental_tag, COUNT(*) as count
FROM access_review_entries
WHERE
    %s
    AND access_review_campaign_id = @campaign_id
GROUP BY incremental_tag;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	rows, err = conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query access entry incremental tag counts: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			tag   AccessReviewEntryIncrementalTag
			count int
		)

		if err := rows.Scan(&tag, &count); err != nil {
			return fmt.Errorf("cannot scan incremental tag count: %w", err)
		}

		s.IncrementalTagCounts[tag] = count
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("cannot iterate incremental tag counts: %w", err)
	}

	return nil
}

func (s *AccessReviewStatistics) LoadByCampaignIDAndSourceID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	campaignID gid.GID,
	sourceID gid.GID,
) error {
	args := pgx.StrictNamedArgs{
		"campaign_id": campaignID,
		"source_id":   sourceID,
	}
	maps.Copy(args, scope.SQLArguments())

	s.DecisionCounts = make(map[AccessReviewEntryDecision]int)
	s.FlagCounts = make(map[AccessReviewEntryFlag]int)
	s.IncrementalTagCounts = make(map[AccessReviewEntryIncrementalTag]int)
	s.TotalCount = 0

	q := `
SELECT decision, COUNT(*) as count
FROM access_review_entries
WHERE
    %s
    AND access_review_campaign_id = @campaign_id
    AND access_review_campaign_source_id = @source_id
GROUP BY decision;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query access entry decision counts: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			decision AccessReviewEntryDecision
			count    int
		)

		if err := rows.Scan(&decision, &count); err != nil {
			return fmt.Errorf("cannot scan decision count: %w", err)
		}

		s.DecisionCounts[decision] = count
		s.TotalCount += count
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("cannot iterate decision counts: %w", err)
	}

	q = `
SELECT f, COUNT(*) as count
FROM access_review_entries, unnest(flags) AS f
WHERE
    %s
    AND access_review_campaign_id = @campaign_id
    AND access_review_campaign_source_id = @source_id
GROUP BY f;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	rows, err = conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query access entry flag counts: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			flag  AccessReviewEntryFlag
			count int
		)

		if err := rows.Scan(&flag, &count); err != nil {
			return fmt.Errorf("cannot scan flag count: %w", err)
		}

		s.FlagCounts[flag] = count
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("cannot iterate flag counts: %w", err)
	}

	q = `
SELECT incremental_tag, COUNT(*) as count
FROM access_review_entries
WHERE
    %s
    AND access_review_campaign_id = @campaign_id
    AND access_review_campaign_source_id = @source_id
GROUP BY incremental_tag;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	rows, err = conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query access entry incremental tag counts: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			tag   AccessReviewEntryIncrementalTag
			count int
		)

		if err := rows.Scan(&tag, &count); err != nil {
			return fmt.Errorf("cannot scan incremental tag count: %w", err)
		}

		s.IncrementalTagCounts[tag] = count
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("cannot iterate incremental tag counts: %w", err)
	}

	return nil
}
