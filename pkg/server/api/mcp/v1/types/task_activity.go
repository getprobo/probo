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
	"go.probo.inc/probo/pkg/page"
)

func NewTaskActivity(e *coredata.TaskActivity) *TaskActivity {
	return &TaskActivity{
		ID:             e.ID,
		OrganizationID: e.OrganizationID,
		TaskID:         e.TaskID,
		ActorID:        e.ActorID,
		ActivityType:   e.ActivityType,
		Field:          e.Field,
		OldValue:       e.OldValue,
		NewValue:       e.NewValue,
		CreatedAt:      e.CreatedAt,
	}
}

func NewListTaskActivitiesOutput(
	activityPage *page.Page[*coredata.TaskActivity, coredata.TaskActivityOrderField],
) ListTaskActivitiesOutput {
	activities := make([]*TaskActivity, 0, len(activityPage.Data))
	for _, v := range activityPage.Data {
		activities = append(activities, NewTaskActivity(v))
	}

	var nextCursor *page.CursorKey

	if activityPage.Info.HasNext && len(activityPage.Data) > 0 {
		cursorKey := activityPage.Data[len(activityPage.Data)-1].CursorKey(activityPage.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListTaskActivitiesOutput{
		NextCursor:     nextCursor,
		TaskActivities: activities,
	}
}
