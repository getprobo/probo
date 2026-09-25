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

package riskmanagement

import (
	"context"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func insertTreatmentPlanEvent(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	tp *coredata.TreatmentPlan,
	eventType coredata.TreatmentPlanEventType,
	measureIDs []gid.GID,
	now time.Time,
) error {
	event := coredata.NewTreatmentPlanEvent(
		tp,
		eventType,
		measureIDs,
		now,
	)

	if err := event.Insert(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot insert treatment plan event: %w", err)
	}

	return nil
}

func loadTreatmentPlanMeasureIDs(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	planID gid.GID,
) ([]gid.GID, error) {
	var mappings coredata.TreatmentPlanMeasures
	if err := mappings.LoadByTreatmentPlanIDs(ctx, conn, scope, []gid.GID{planID}); err != nil {
		return nil, fmt.Errorf("cannot load treatment plan measures: %w", err)
	}

	ids := make([]gid.GID, 0, len(mappings))
	for _, mapping := range mappings {
		ids = append(ids, mapping.MeasureID)
	}

	return ids, nil
}
