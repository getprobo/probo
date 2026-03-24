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
	"fmt"
	"maps"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type AccessEntryStatistics struct {
	TotalCount           int
	DecisionCounts       map[AccessEntryDecision]int
	FlagCounts           map[AccessEntryFlag]int
	IncrementalTagCounts map[AccessEntryIncrementalTag]int
}

func (s *AccessEntryStatistics) LoadByCampaignID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	campaignID gid.GID,
) error {
	args := pgx.StrictNamedArgs{"campaign_id": campaignID}
	maps.Copy(args, scope.SQLArguments())

	s.DecisionCounts = make(map[AccessEntryDecision]int)
	s.FlagCounts = make(map[AccessEntryFlag]int)
	s.IncrementalTagCounts = make(map[AccessEntryIncrementalTag]int)
	s.TotalCount = 0

	q := `
SELECT decision, COUNT(*) as count
FROM access_entries
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
		var decision AccessEntryDecision
		var count int
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
SELECT flag, COUNT(*) as count
FROM access_entries
WHERE
    %s
    AND access_review_campaign_id = @campaign_id
GROUP BY flag;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	rows, err = conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query access entry flag counts: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var flag AccessEntryFlag
		var count int
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
FROM access_entries
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
		var tag AccessEntryIncrementalTag
		var count int
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

func (s *AccessEntryStatistics) LoadByCampaignIDAndSourceID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	campaignID gid.GID,
	sourceID gid.GID,
) error {
	args := pgx.StrictNamedArgs{
		"campaign_id": campaignID,
		"source_id":   sourceID,
	}
	maps.Copy(args, scope.SQLArguments())

	s.DecisionCounts = make(map[AccessEntryDecision]int)
	s.FlagCounts = make(map[AccessEntryFlag]int)
	s.IncrementalTagCounts = make(map[AccessEntryIncrementalTag]int)
	s.TotalCount = 0

	q := `
SELECT decision, COUNT(*) as count
FROM access_entries
WHERE
    %s
    AND access_review_campaign_id = @campaign_id
    AND access_source_id = @source_id
GROUP BY decision;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query access entry decision counts: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var decision AccessEntryDecision
		var count int
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
SELECT flag, COUNT(*) as count
FROM access_entries
WHERE
    %s
    AND access_review_campaign_id = @campaign_id
    AND access_source_id = @source_id
GROUP BY flag;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	rows, err = conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query access entry flag counts: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var flag AccessEntryFlag
		var count int
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
FROM access_entries
WHERE
    %s
    AND access_review_campaign_id = @campaign_id
    AND access_source_id = @source_id
GROUP BY incremental_tag;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	rows, err = conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query access entry incremental tag counts: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tag AccessEntryIncrementalTag
		var count int
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
