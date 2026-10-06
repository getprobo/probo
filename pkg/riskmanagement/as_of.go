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

	RiskAnalysisMatrixInternalControl struct {
		ID    gid.GID
		Name  *string
		State coredata.InternalControlState
	}

	TreatmentPlansAsOfPage struct {
		Page                 *page.Page[*coredata.TreatmentPlan, coredata.TreatmentPlanOrderField]
		TotalCount           int
		ProgressByID         map[gid.GID]TreatmentProgress
		InternalControlsByID map[gid.GID][]RiskAnalysisMatrixInternalControl
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
				internalControlIDs, err := event.LinkedInternalControlIDs()
				if err != nil {
					return fmt.Errorf("cannot parse treatment plan internal control ids: %w", err)
				}

				entries = append(
					entries,
					matrixEntryFromPlan(
						event.TreatmentPlan(),
						progressFromStates(internalControlIDs, states),
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
) ([]*coredata.TreatmentPlanEvent, map[gid.GID]coredata.InternalControlState, error) {
	var events coredata.TreatmentPlanEvents
	if err := events.LoadLatestByRiskAnalysisIDAsOf(ctx, conn, scope, analysisID, asOf); err != nil {
		return nil, nil, fmt.Errorf("cannot load treatment plan events: %w", err)
	}

	internalControlIDs, err := uniqueEventInternalControlIDs(events)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot collect treatment plan internal control ids: %w", err)
	}

	internalControlEvents, err := loadInternalControlEventsAsOf(ctx, conn, scope, internalControlIDs, asOf)
	if err != nil {
		return nil, nil, err
	}

	return events, internalControlStatesFromEvents(internalControlEvents), nil
}

func uniqueEventInternalControlIDs(events []*coredata.TreatmentPlanEvent) ([]gid.GID, error) {
	seen := make(map[gid.GID]struct{})
	ids := make([]gid.GID, 0)

	for _, event := range events {
		internalControlIDs, err := event.LinkedInternalControlIDs()
		if err != nil {
			return nil, fmt.Errorf("cannot parse treatment plan internal control ids: %w", err)
		}

		for _, internalControlID := range internalControlIDs {
			if _, ok := seen[internalControlID]; ok {
				continue
			}

			seen[internalControlID] = struct{}{}
			ids = append(ids, internalControlID)
		}
	}

	return ids, nil
}

func loadInternalControlEventsAsOf(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	internalControlIDs []gid.GID,
	asOf time.Time,
) (coredata.InternalControlEvents, error) {
	var events coredata.InternalControlEvents
	if err := events.LoadLatestByInternalControlIDsAsOf(
		ctx,
		conn,
		scope,
		internalControlIDs,
		asOf,
		coredata.NewInternalControlFilter(nil, nil, nil),
	); err != nil {
		return nil, fmt.Errorf("cannot load internal control events: %w", err)
	}

	return events, nil
}

func internalControlStatesFromEvents(events coredata.InternalControlEvents) map[gid.GID]coredata.InternalControlState {
	states := make(map[gid.GID]coredata.InternalControlState, len(events))
	for _, event := range events {
		states[event.InternalControlID] = event.State
	}

	return states
}

func internalControlNamesFromEvents(events coredata.InternalControlEvents) map[gid.GID]string {
	names := make(map[gid.GID]string, len(events))
	for _, event := range events {
		names[event.InternalControlID] = event.Name
	}

	return names
}

func progressFromStates(
	internalControlIDs []gid.GID,
	states map[gid.GID]coredata.InternalControlState,
) TreatmentProgress {
	progress := TreatmentProgress{}

	for _, internalControlID := range internalControlIDs {
		progress.Total++

		switch states[internalControlID] {
		case coredata.InternalControlStateImplemented:
			progress.Done++
		case coredata.InternalControlStateInProgress:
			progress.InProgress++
		case coredata.InternalControlStateNotImplemented:
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

func internalControlsFromIDs(
	internalControlIDs []gid.GID,
	names map[gid.GID]string,
	states map[gid.GID]coredata.InternalControlState,
) []RiskAnalysisMatrixInternalControl {
	items := make([]RiskAnalysisMatrixInternalControl, 0, len(internalControlIDs))
	for _, internalControlID := range internalControlIDs {
		item := RiskAnalysisMatrixInternalControl{
			ID:    internalControlID,
			State: coredata.InternalControlStateUnknown,
		}
		if name, ok := names[internalControlID]; ok {
			item.Name = &name
		}

		if state, ok := states[internalControlID]; ok {
			item.State = state
		}

		items = append(items, item)
	}

	slices.SortFunc(items, compareMatrixInternalControls)

	return items
}

func compareMatrixInternalControls(a, b RiskAnalysisMatrixInternalControl) int {
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
	includeInternalControls bool,
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

			progressByID, internalControlsByID, err := loadAsOfPlanExtras(
				ctx,
				conn,
				scope,
				analysisID,
				asOf,
				paged.Data,
				includeInternalControls,
			)
			if err != nil {
				return err
			}

			result.Page = paged
			result.TotalCount = total
			result.ProgressByID = progressByID
			result.InternalControlsByID = internalControlsByID

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot list treatment plans as of: %w", err)
	}

	return result, nil
}

func (s *Service) ListTreatmentPlansForInternalControlIDAsOf(
	ctx context.Context,
	scope coredata.Scoper,
	internalControlID gid.GID,
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
			if err := plans.LoadByInternalControlIDAsOf(
				ctx,
				conn,
				scope,
				internalControlID,
				asOf,
				cursor,
				filter,
			); err != nil {
				return fmt.Errorf("cannot load treatment plans as of: %w", err)
			}

			total, err := plans.CountByInternalControlIDAsOf(
				ctx,
				conn,
				scope,
				internalControlID,
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
		return nil, fmt.Errorf("cannot list internal control treatment plans as of: %w", err)
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
	includeInternalControls bool,
) (map[gid.GID]TreatmentProgress, map[gid.GID][]RiskAnalysisMatrixInternalControl, error) {
	progressByID := make(map[gid.GID]TreatmentProgress, len(plans))
	internalControlsByID := make(map[gid.GID][]RiskAnalysisMatrixInternalControl, len(plans))

	if len(plans) == 0 {
		return progressByID, internalControlsByID, nil
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

	internalControlIDs, err := uniqueEventInternalControlIDs(events)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot collect treatment plan internal control ids: %w", err)
	}

	internalControlEvents, err := loadInternalControlEventsAsOf(ctx, conn, scope, internalControlIDs, asOf)
	if err != nil {
		return nil, nil, err
	}

	states := internalControlStatesFromEvents(internalControlEvents)

	var names map[gid.GID]string
	if includeInternalControls {
		names = internalControlNamesFromEvents(internalControlEvents)
	}

	for _, event := range events {
		linked, err := event.LinkedInternalControlIDs()
		if err != nil {
			return nil, nil, fmt.Errorf("cannot parse treatment plan internal control ids: %w", err)
		}

		progressByID[event.TreatmentPlanID] = progressFromStates(linked, states)
		if includeInternalControls {
			internalControlsByID[event.TreatmentPlanID] = internalControlsFromIDs(linked, names, states)
		}
	}

	return progressByID, internalControlsByID, nil
}

func (s *Service) ListInternalControlsAsOf(
	ctx context.Context,
	scope coredata.Scoper,
	analysisID gid.GID,
	planID gid.GID,
	asOf time.Time,
	cursor *page.Cursor[coredata.InternalControlOrderField],
	filter *coredata.InternalControlFilter,
) (*page.Page[*coredata.InternalControl, coredata.InternalControlOrderField], int, error) {
	if filter == nil {
		filter = coredata.NewInternalControlFilter(nil, nil, nil)
	}

	var (
		paged *page.Page[*coredata.InternalControl, coredata.InternalControlOrderField]
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
				paged = page.NewPage([]*coredata.InternalControl{}, cursor)
				return nil
			}

			linked, err := events[0].LinkedInternalControlIDs()
			if err != nil {
				return fmt.Errorf("cannot parse treatment plan internal control ids: %w", err)
			}

			var internalControls coredata.InternalControls
			if err := internalControls.LoadByIDsAsOf(
				ctx,
				conn,
				scope,
				linked,
				asOf,
				cursor,
				filter,
			); err != nil {
				return fmt.Errorf("cannot load internal controls as of: %w", err)
			}

			count, err := internalControls.CountByIDsAsOf(
				ctx,
				conn,
				scope,
				linked,
				asOf,
				filter,
			)
			if err != nil {
				return fmt.Errorf("cannot count internal controls as of: %w", err)
			}

			total = count
			paged = page.NewPage(internalControls, cursor)

			return nil
		},
	)
	if err != nil {
		return nil, 0, fmt.Errorf("cannot list internal controls as of: %w", err)
	}

	return paged, total, nil
}
