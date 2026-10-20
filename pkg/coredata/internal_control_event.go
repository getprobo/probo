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
	InternalControlEvent struct {
		OrganizationID           gid.GID                              `db:"organization_id"`
		InternalControlID        gid.GID                              `db:"internal_control_id"`
		EventType                InternalControlEventType             `db:"event_type"`
		Name                     string                               `db:"name"`
		Category                 string                               `db:"category"`
		State                    InternalControlState                 `db:"state"`
		ImplementationStatus     *InternalControlImplementationStatus `db:"implementation_status"`
		InternalControlCreatedAt time.Time                            `db:"internal_control_created_at"`
		CreatedAt                time.Time                            `db:"created_at"`
	}

	InternalControlEvents []*InternalControlEvent
)

func NewInternalControlEvent(
	internalControl *InternalControl,
	eventType InternalControlEventType,
	now time.Time,
) *InternalControlEvent {
	status := internalControl.ImplementationStatus

	return &InternalControlEvent{
		OrganizationID:           internalControl.OrganizationID,
		InternalControlID:        internalControl.ID,
		EventType:                eventType,
		Name:                     internalControl.Name,
		Category:                 internalControl.Category,
		State:                    internalControl.State,
		ImplementationStatus:     &status,
		InternalControlCreatedAt: internalControl.CreatedAt,
		CreatedAt:                now,
	}
}

func internalControlEventImplementationStatus(event *InternalControlEvent) InternalControlImplementationStatus {
	if event.ImplementationStatus != nil && event.ImplementationStatus.IsValid() {
		return *event.ImplementationStatus
	}

	return ImplementationStatusForInternalControlState(event.State)
}

func (e *InternalControlEvent) InternalControl() *InternalControl {
	return &InternalControl{
		ID:                   e.InternalControlID,
		OrganizationID:       e.OrganizationID,
		Category:             e.Category,
		Name:                 e.Name,
		State:                e.State,
		ImplementationStatus: internalControlEventImplementationStatus(e),
		CreatedAt:            e.InternalControlCreatedAt,
		UpdatedAt:            e.CreatedAt,
	}
}

func (e *InternalControlEvent) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO
    internal_control_events (
        tenant_id,
        organization_id,
        internal_control_id,
        event_type,
        name,
        category,
        state,
        implementation_status,
        internal_control_created_at,
        created_at
    )
VALUES (
    @tenant_id,
    @organization_id,
    @internal_control_id,
    @event_type,
    @name,
    @category,
    @state,
    @implementation_status,
    @internal_control_created_at,
    @created_at
);
`

	args := pgx.StrictNamedArgs{
		"tenant_id":                   scope.GetTenantID(),
		"organization_id":             e.OrganizationID,
		"internal_control_id":         e.InternalControlID,
		"event_type":                  e.EventType,
		"name":                        e.Name,
		"category":                    e.Category,
		"state":                       e.State,
		"implementation_status":       e.ImplementationStatus,
		"internal_control_created_at": e.InternalControlCreatedAt,
		"created_at":                  e.CreatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert internal control event: %w", err)
	}

	return nil
}

func (es *InternalControlEvents) LoadLatestByInternalControlIDsAsOf(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	internalControlIDs []gid.GID,
	asOf time.Time,
	filter *InternalControlFilter,
) error {
	if len(internalControlIDs) == 0 {
		*es = nil
		return nil
	}

	q := `
SELECT
    organization_id,
    internal_control_id,
    event_type,
    name,
    category,
    state,
    implementation_status,
    internal_control_created_at,
    created_at
FROM (
    SELECT DISTINCT ON (internal_control_id)
        organization_id,
        internal_control_id,
        event_type,
        name,
        category,
        state,
        implementation_status,
        internal_control_created_at,
        created_at
    FROM
        internal_control_events
    WHERE
        %s
        AND internal_control_id = ANY(@internal_control_ids)
        AND created_at < @as_of
    ORDER BY
        internal_control_id,
        created_at DESC
) latest
WHERE
    event_type <> @deleted
    AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.EventSQLFragment())

	args := pgx.StrictNamedArgs{
		"internal_control_ids": internalControlIDs,
		"as_of":                asOf,
		"deleted":              InternalControlEventTypeDeleted,
	}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query internal control events: %w", err)
	}

	events, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[InternalControlEvent])
	if err != nil {
		return fmt.Errorf("cannot collect internal control events: %w", err)
	}

	*es = events

	return nil
}
