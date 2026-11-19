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

package types

import (
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

func NewCorrectiveActionPlan(p *coredata.CorrectiveActionPlan) *CorrectiveActionPlan {
	return &CorrectiveActionPlan{
		ID:             p.ID,
		OrganizationID: p.OrganizationID,
		FindingID:      p.FindingID,
		Title:          p.Title,
		Description:    p.Description,
		OwnerID:        p.OwnerID,
		DueDate:        p.DueDate,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}

func NewListFindingCorrectiveActionPlansOutput(
	planPage *page.Page[*coredata.CorrectiveActionPlan, coredata.CorrectiveActionPlanOrderField],
) ListFindingCorrectiveActionPlansOutput {
	plans := make([]*CorrectiveActionPlan, 0, len(planPage.Data))
	for _, v := range planPage.Data {
		plans = append(plans, NewCorrectiveActionPlan(v))
	}

	var nextCursor *page.CursorKey

	if len(planPage.Data) > 0 {
		cursorKey := planPage.Data[len(planPage.Data)-1].CursorKey(planPage.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListFindingCorrectiveActionPlansOutput{
		NextCursor:            nextCursor,
		CorrectiveActionPlans: plans,
	}
}

func NewListCorrectiveActionPlanInternalControlsOutput(
	internalControlPage *page.Page[*coredata.InternalControl, coredata.InternalControlOrderField],
) ListCorrectiveActionPlanInternalControlsOutput {
	internalControls := make([]*InternalControl, 0, len(internalControlPage.Data))
	for _, v := range internalControlPage.Data {
		internalControls = append(internalControls, NewInternalControl(v))
	}

	var nextCursor *page.CursorKey

	if len(internalControlPage.Data) > 0 {
		cursorKey := internalControlPage.Data[len(internalControlPage.Data)-1].CursorKey(internalControlPage.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListCorrectiveActionPlanInternalControlsOutput{
		NextCursor:       nextCursor,
		InternalControls: internalControls,
	}
}

func NewListCorrectiveActionPlanTasksOutput(
	taskPage *page.Page[*coredata.Task, coredata.TaskOrderField],
	links map[gid.GID]*coredata.TaskExternalLink,
	internalControlIDs map[gid.GID][]gid.GID,
) ListCorrectiveActionPlanTasksOutput {
	tasks := make([]*Task, 0, len(taskPage.Data))
	for _, v := range taskPage.Data {
		tasks = append(tasks, NewTaskWithLink(v, links[v.ID], internalControlIDs[v.ID]))
	}

	var nextCursor *page.CursorKey

	if len(taskPage.Data) > 0 {
		cursorKey := taskPage.Data[len(taskPage.Data)-1].CursorKey(taskPage.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListCorrectiveActionPlanTasksOutput{
		NextCursor: nextCursor,
		Tasks:      tasks,
	}
}
