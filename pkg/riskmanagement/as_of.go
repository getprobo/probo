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
	"cmp"
	"context"
	"fmt"
	"maps"
	"slices"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	RiskAnalysisMatrixEntry struct {
		InherentLikelihood int
		InherentImpact     int
		NetLikelihood      int
		NetImpact          int
		ResidualLikelihood int
		ResidualImpact     int
	}

	RiskAnalysisMatrixMeasure struct {
		ID    gid.GID
		Name  *string
		State coredata.MeasureState
	}

	TreatmentPlansAsOfPage struct {
		Page         *page.Page[*coredata.TreatmentPlan, coredata.TreatmentPlanOrderField]
		TotalCount   int
		ProgressByID map[gid.GID]TreatmentProgress
		MeasuresByID map[gid.GID][]RiskAnalysisMatrixMeasure
	}
)

func (s *Service) loadMatrixCellsAsOf(
	ctx context.Context,
	scope coredata.Scoper,
	analysisID gid.GID,
	asOf time.Time,
) ([]*coredata.RiskAnalysisMatrixCell, error) {
	var cells []*coredata.RiskAnalysisMatrixCell

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			analysis := &coredata.RiskAnalysis{}
			if err := analysis.LoadByID(ctx, conn, scope, analysisID); err != nil {
				return fmt.Errorf("cannot load risk analysis: %w", err)
			}

			folded, states, err := loadTreatmentPlansAsOf(ctx, conn, scope, analysisID, asOf)
			if err != nil {
				return fmt.Errorf("cannot load treatment plans as of: %w", err)
			}

			entries := make([]RiskAnalysisMatrixEntry, 0, len(folded))
			for _, event := range folded {
				measureIDs, err := event.LinkedMeasureIDs()
				if err != nil {
					return fmt.Errorf("cannot parse treatment plan measure ids: %w", err)
				}

				entries = append(
					entries,
					matrixEntryFromPlan(
						event.TreatmentPlan(),
						progressFromStates(measureIDs, states),
					),
				)
			}

			cells = cellsFromMatrixEntries(entries)

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load risk analysis matrix cells as of: %w", err)
	}

	return cells, nil
}

func loadTreatmentPlansAsOf(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	analysisID gid.GID,
	asOf time.Time,
) ([]*coredata.TreatmentPlanEvent, map[gid.GID]coredata.MeasureState, error) {
	var events coredata.TreatmentPlanEvents
	if err := events.LoadLatestByRiskAnalysisIDAsOf(ctx, conn, scope, analysisID, asOf); err != nil {
		return nil, nil, fmt.Errorf("cannot load treatment plan events: %w", err)
	}

	measureIDs, err := uniqueEventMeasureIDs(events)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot collect treatment plan measure ids: %w", err)
	}

	measureEvents, err := loadMeasureEventsAsOf(ctx, conn, scope, measureIDs, asOf)
	if err != nil {
		return nil, nil, err
	}

	return events, measureStatesFromEvents(measureEvents), nil
}

func uniqueEventMeasureIDs(events []*coredata.TreatmentPlanEvent) ([]gid.GID, error) {
	seen := make(map[gid.GID]struct{})
	ids := make([]gid.GID, 0)

	for _, event := range events {
		measureIDs, err := event.LinkedMeasureIDs()
		if err != nil {
			return nil, fmt.Errorf("cannot parse treatment plan measure ids: %w", err)
		}

		for _, measureID := range measureIDs {
			if _, ok := seen[measureID]; ok {
				continue
			}

			seen[measureID] = struct{}{}
			ids = append(ids, measureID)
		}
	}

	return ids, nil
}

func loadMeasureEventsAsOf(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	measureIDs []gid.GID,
	asOf time.Time,
) (coredata.MeasureEvents, error) {
	var events coredata.MeasureEvents
	if err := events.LoadLatestByMeasureIDsAsOf(
		ctx,
		conn,
		scope,
		measureIDs,
		asOf,
		coredata.NewMeasureFilter(nil, nil, nil),
	); err != nil {
		return nil, fmt.Errorf("cannot load measure events: %w", err)
	}

	return events, nil
}

func measureStatesFromEvents(events coredata.MeasureEvents) map[gid.GID]coredata.MeasureState {
	states := make(map[gid.GID]coredata.MeasureState, len(events))
	for _, event := range events {
		states[event.MeasureID] = event.State
	}

	return states
}

func measureNamesFromEvents(events coredata.MeasureEvents) map[gid.GID]string {
	names := make(map[gid.GID]string, len(events))
	for _, event := range events {
		names[event.MeasureID] = event.Name
	}

	return names
}

func progressFromStates(
	measureIDs []gid.GID,
	states map[gid.GID]coredata.MeasureState,
) TreatmentProgress {
	progress := TreatmentProgress{}

	for _, measureID := range measureIDs {
		progress.Total++

		switch states[measureID] {
		case coredata.MeasureStateImplemented:
			progress.Done++
		case coredata.MeasureStateInProgress:
			progress.InProgress++
		case coredata.MeasureStateNotImplemented:
			progress.NotImplemented++
		}
	}

	return progress
}

func matrixEntryFromPlan(
	plan *coredata.TreatmentPlan,
	progress TreatmentProgress,
) RiskAnalysisMatrixEntry {
	netLikelihood, netImpact, _ := NetScores(plan, progress)

	return RiskAnalysisMatrixEntry{
		InherentLikelihood: plan.InherentLikelihood,
		InherentImpact:     plan.InherentImpact,
		NetLikelihood:      netLikelihood,
		NetImpact:          netImpact,
		ResidualLikelihood: plan.ResidualLikelihood,
		ResidualImpact:     plan.ResidualImpact,
	}
}

func measuresFromIDs(
	measureIDs []gid.GID,
	names map[gid.GID]string,
	states map[gid.GID]coredata.MeasureState,
) []RiskAnalysisMatrixMeasure {
	items := make([]RiskAnalysisMatrixMeasure, 0, len(measureIDs))
	for _, measureID := range measureIDs {
		item := RiskAnalysisMatrixMeasure{
			ID:    measureID,
			State: coredata.MeasureStateUnknown,
		}
		if name, ok := names[measureID]; ok {
			item.Name = &name
		}

		if state, ok := states[measureID]; ok {
			item.State = state
		}

		items = append(items, item)
	}

	slices.SortFunc(items, compareMatrixMeasures)

	return items
}

func compareMatrixMeasures(a, b RiskAnalysisMatrixMeasure) int {
	aName := ""
	if a.Name != nil {
		aName = *a.Name
	}

	bName := ""
	if b.Name != nil {
		bName = *b.Name
	}

	if cmp := cmp.Compare(aName, bName); cmp != 0 {
		return cmp
	}

	return cmp.Compare(a.ID.String(), b.ID.String())
}

func cellsFromMatrixEntries(entries []RiskAnalysisMatrixEntry) []*coredata.RiskAnalysisMatrixCell {
	type cellKey struct {
		scoreType  coredata.TreatmentPlanScoreType
		likelihood int
		impact     int
	}

	counts := make(map[cellKey]int, len(entries)*3)
	for _, entry := range entries {
		counts[cellKey{coredata.TreatmentPlanScoreTypeInherent, entry.InherentLikelihood, entry.InherentImpact}]++
		counts[cellKey{coredata.TreatmentPlanScoreTypeNet, entry.NetLikelihood, entry.NetImpact}]++
		counts[cellKey{coredata.TreatmentPlanScoreTypeResidual, entry.ResidualLikelihood, entry.ResidualImpact}]++
	}

	cells := make([]*coredata.RiskAnalysisMatrixCell, 0, len(counts))
	for key, count := range counts {
		cells = append(cells, &coredata.RiskAnalysisMatrixCell{
			Type:       key.scoreType,
			Likelihood: key.likelihood,
			Impact:     key.impact,
			Count:      count,
		})
	}

	slices.SortFunc(cells, func(a, b *coredata.RiskAnalysisMatrixCell) int {
		if cmp := cmp.Compare(a.Type.String(), b.Type.String()); cmp != 0 {
			return cmp
		}

		if cmp := cmp.Compare(a.Likelihood, b.Likelihood); cmp != 0 {
			return cmp
		}

		return cmp.Compare(a.Impact, b.Impact)
	})

	return cells
}

func (s *Service) ListTreatmentPlansAsOf(
	ctx context.Context,
	scope coredata.Scoper,
	analysisID gid.GID,
	asOf time.Time,
	cursor *page.Cursor[coredata.TreatmentPlanOrderField],
	filter *coredata.TreatmentPlanFilter,
	includeMeasures bool,
) (*TreatmentPlansAsOfPage, error) {
	filter, err := prepareTreatmentPlanFilter(filter)
	if err != nil {
		return nil, fmt.Errorf("cannot list treatment plans as of: %w", err)
	}

	result := &TreatmentPlansAsOfPage{}

	err = s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			analysis := &coredata.RiskAnalysis{}
			if err := analysis.LoadByID(ctx, conn, scope, analysisID); err != nil {
				return fmt.Errorf("cannot load risk analysis: %w", err)
			}

			var plans coredata.TreatmentPlans
			if err := plans.LoadByRiskAnalysisIDAsOf(
				ctx,
				conn,
				scope,
				analysisID,
				asOf,
				cursor,
				filter,
			); err != nil {
				return fmt.Errorf("cannot load treatment plans as of: %w", err)
			}

			total, err := plans.CountByRiskAnalysisIDAsOf(
				ctx,
				conn,
				scope,
				analysisID,
				asOf,
				filter,
			)
			if err != nil {
				return fmt.Errorf("cannot count treatment plans as of: %w", err)
			}

			paged := page.NewPage(plans, cursor)

			progressByID, measuresByID, err := loadAsOfPlanExtras(
				ctx,
				conn,
				scope,
				analysisID,
				asOf,
				paged.Data,
				includeMeasures,
			)
			if err != nil {
				return err
			}

			result.Page = paged
			result.TotalCount = total
			result.ProgressByID = progressByID
			result.MeasuresByID = measuresByID

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot list treatment plans as of: %w", err)
	}

	return result, nil
}

func (s *Service) ListTreatmentPlansForMeasureIDAsOf(
	ctx context.Context,
	scope coredata.Scoper,
	measureID gid.GID,
	asOf time.Time,
	cursor *page.Cursor[coredata.TreatmentPlanOrderField],
	filter *coredata.TreatmentPlanFilter,
) (*TreatmentPlansAsOfPage, error) {
	filter, err := prepareTreatmentPlanFilter(filter)
	if err != nil {
		return nil, fmt.Errorf("cannot list treatment plans as of: %w", err)
	}

	result := &TreatmentPlansAsOfPage{}

	err = s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var plans coredata.TreatmentPlans
			if err := plans.LoadByMeasureIDAsOf(
				ctx,
				conn,
				scope,
				measureID,
				asOf,
				cursor,
				filter,
			); err != nil {
				return fmt.Errorf("cannot load treatment plans as of: %w", err)
			}

			total, err := plans.CountByMeasureIDAsOf(
				ctx,
				conn,
				scope,
				measureID,
				asOf,
				filter,
			)
			if err != nil {
				return fmt.Errorf("cannot count treatment plans as of: %w", err)
			}

			paged := page.NewPage(plans, cursor)
			progressByID := make(map[gid.GID]TreatmentProgress, len(paged.Data))
			byAnalysis := make(map[gid.GID][]*coredata.TreatmentPlan)

			for _, plan := range paged.Data {
				byAnalysis[plan.RiskAnalysisID] = append(byAnalysis[plan.RiskAnalysisID], plan)
			}

			for analysisID, group := range byAnalysis {
				progress, _, err := loadAsOfPlanExtras(
					ctx,
					conn,
					scope,
					analysisID,
					asOf,
					group,
					false,
				)
				if err != nil {
					return err
				}

				maps.Copy(progressByID, progress)
			}

			result.Page = paged
			result.TotalCount = total
			result.ProgressByID = progressByID

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot list measure treatment plans as of: %w", err)
	}

	return result, nil
}

func loadAsOfPlanExtras(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	analysisID gid.GID,
	asOf time.Time,
	plans []*coredata.TreatmentPlan,
	includeMeasures bool,
) (map[gid.GID]TreatmentProgress, map[gid.GID][]RiskAnalysisMatrixMeasure, error) {
	progressByID := make(map[gid.GID]TreatmentProgress, len(plans))
	measuresByID := make(map[gid.GID][]RiskAnalysisMatrixMeasure, len(plans))

	if len(plans) == 0 {
		return progressByID, measuresByID, nil
	}

	planIDs := make([]gid.GID, 0, len(plans))
	for _, plan := range plans {
		planIDs = append(planIDs, plan.ID)
	}

	var events coredata.TreatmentPlanEvents
	if err := events.LoadLatestByTreatmentPlanIDsAsOf(
		ctx,
		conn,
		scope,
		analysisID,
		planIDs,
		asOf,
	); err != nil {
		return nil, nil, fmt.Errorf("cannot load treatment plan events: %w", err)
	}

	measureIDs, err := uniqueEventMeasureIDs(events)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot collect treatment plan measure ids: %w", err)
	}

	measureEvents, err := loadMeasureEventsAsOf(ctx, conn, scope, measureIDs, asOf)
	if err != nil {
		return nil, nil, err
	}

	states := measureStatesFromEvents(measureEvents)

	var names map[gid.GID]string
	if includeMeasures {
		names = measureNamesFromEvents(measureEvents)
	}

	for _, event := range events {
		linked, err := event.LinkedMeasureIDs()
		if err != nil {
			return nil, nil, fmt.Errorf("cannot parse treatment plan measure ids: %w", err)
		}

		progressByID[event.TreatmentPlanID] = progressFromStates(linked, states)
		if includeMeasures {
			measuresByID[event.TreatmentPlanID] = measuresFromIDs(linked, names, states)
		}
	}

	return progressByID, measuresByID, nil
}

func (s *Service) ListMeasuresAsOf(
	ctx context.Context,
	scope coredata.Scoper,
	analysisID gid.GID,
	planID gid.GID,
	asOf time.Time,
	cursor *page.Cursor[coredata.MeasureOrderField],
	filter *coredata.MeasureFilter,
) (*page.Page[*coredata.Measure, coredata.MeasureOrderField], int, error) {
	if filter == nil {
		filter = coredata.NewMeasureFilter(nil, nil, nil)
	}

	var (
		paged *page.Page[*coredata.Measure, coredata.MeasureOrderField]
		total int
	)

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			analysis := &coredata.RiskAnalysis{}
			if err := analysis.LoadByID(ctx, conn, scope, analysisID); err != nil {
				return fmt.Errorf("cannot load risk analysis: %w", err)
			}

			var events coredata.TreatmentPlanEvents
			if err := events.LoadLatestByTreatmentPlanIDsAsOf(
				ctx,
				conn,
				scope,
				analysisID,
				[]gid.GID{planID},
				asOf,
			); err != nil {
				return fmt.Errorf("cannot load treatment plan events: %w", err)
			}

			if len(events) == 0 {
				paged = page.NewPage([]*coredata.Measure{}, cursor)
				return nil
			}

			linked, err := events[0].LinkedMeasureIDs()
			if err != nil {
				return fmt.Errorf("cannot parse treatment plan measure ids: %w", err)
			}

			var measures coredata.Measures
			if err := measures.LoadByIDsAsOf(
				ctx,
				conn,
				scope,
				linked,
				asOf,
				cursor,
				filter,
			); err != nil {
				return fmt.Errorf("cannot load measures as of: %w", err)
			}

			count, err := measures.CountByIDsAsOf(
				ctx,
				conn,
				scope,
				linked,
				asOf,
				filter,
			)
			if err != nil {
				return fmt.Errorf("cannot count measures as of: %w", err)
			}

			total = count
			paged = page.NewPage(measures, cursor)

			return nil
		},
	)
	if err != nil {
		return nil, 0, fmt.Errorf("cannot list measures as of: %w", err)
	}

	return paged, total, nil
}
