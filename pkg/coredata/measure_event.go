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

type (
	MeasureEvent struct {
		OrganizationID   gid.GID          `db:"organization_id"`
		MeasureID        gid.GID          `db:"measure_id"`
		EventType        MeasureEventType `db:"event_type"`
		Name             string           `db:"name"`
		Category         string           `db:"category"`
		State            MeasureState     `db:"state"`
		MeasureCreatedAt time.Time        `db:"measure_created_at"`
		CreatedAt        time.Time        `db:"created_at"`
	}

	MeasureEvents []*MeasureEvent
)

func NewMeasureEvent(
	measure *Measure,
	eventType MeasureEventType,
	now time.Time,
) *MeasureEvent {
	return &MeasureEvent{
		OrganizationID:   measure.OrganizationID,
		MeasureID:        measure.ID,
		EventType:        eventType,
		Name:             measure.Name,
		Category:         measure.Category,
		State:            measure.State,
		MeasureCreatedAt: measure.CreatedAt,
		CreatedAt:        now,
	}
}

func (e *MeasureEvent) Measure() *Measure {
	return &Measure{
		ID:             e.MeasureID,
		OrganizationID: e.OrganizationID,
		Category:       e.Category,
		Name:           e.Name,
		State:          e.State,
		CreatedAt:      e.MeasureCreatedAt,
		UpdatedAt:      e.CreatedAt,
	}
}

func (e *MeasureEvent) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO
    measure_events (
        tenant_id,
        organization_id,
        measure_id,
        event_type,
        name,
        category,
        state,
        measure_created_at,
        created_at
    )
VALUES (
    @tenant_id,
    @organization_id,
    @measure_id,
    @event_type,
    @name,
    @category,
    @state,
    @measure_created_at,
    @created_at
);
`

	args := pgx.StrictNamedArgs{
		"tenant_id":          scope.GetTenantID(),
		"organization_id":    e.OrganizationID,
		"measure_id":         e.MeasureID,
		"event_type":         e.EventType,
		"name":               e.Name,
		"category":           e.Category,
		"state":              e.State,
		"measure_created_at": e.MeasureCreatedAt,
		"created_at":         e.CreatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert measure event: %w", err)
	}

	return nil
}

func (es *MeasureEvents) LoadLatestByMeasureIDsAsOf(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	measureIDs []gid.GID,
	asOf time.Time,
	filter *MeasureFilter,
) error {
	if len(measureIDs) == 0 {
		*es = nil
		return nil
	}

	q := `
SELECT
    organization_id,
    measure_id,
    event_type,
    name,
    category,
    state,
    measure_created_at,
    created_at
FROM (
    SELECT DISTINCT ON (measure_id)
        organization_id,
        measure_id,
        event_type,
        name,
        category,
        state,
        measure_created_at,
        created_at
    FROM
        measure_events
    WHERE
        %s
        AND measure_id = ANY(@measure_ids)
        AND created_at < @as_of
    ORDER BY
        measure_id,
        created_at DESC
) latest
WHERE
    event_type <> @deleted
    AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.EventSQLFragment())

	args := pgx.StrictNamedArgs{
		"measure_ids": measureIDs,
		"as_of":       asOf,
		"deleted":     MeasureEventTypeDeleted,
	}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query measure events: %w", err)
	}

	events, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[MeasureEvent])
	if err != nil {
		return fmt.Errorf("cannot collect measure events: %w", err)
	}

	*es = events

	return nil
}
