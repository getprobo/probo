// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

package types

import (
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/page"
)

func NewTask(t *coredata.Task) *Task {
	return &Task{
		ID:           t.ID,
		Name:         t.Name,
		Description:  t.Description,
		State:        t.State,
		TimeEstimate: t.TimeEstimate,
		CreatedAt:    t.CreatedAt,
		UpdatedAt:    t.UpdatedAt,
		Deadline:     t.Deadline,
	}
}

func NewListTasksOutput(taskPage *page.Page[*coredata.Task, coredata.TaskOrderField]) ListTasksOutput {
	tasks := make([]*Task, 0, len(taskPage.Data))
	for _, v := range taskPage.Data {
		tasks = append(tasks, NewTask(v))
	}

	var nextCursor *page.CursorKey
	if len(taskPage.Data) > 0 {
		cursorKey := taskPage.Data[len(taskPage.Data)-1].CursorKey(taskPage.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListTasksOutput{
		NextCursor: nextCursor,
		Tasks:      tasks,
	}
}
