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
	"time"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/timespan"
)

type (
	Task struct {
		ID             gid.GID               `json:"id"`
		OrganizationID gid.GID               `json:"organizationId"`
		MeasureID      *gid.GID              `json:"measureId"`
		Name           string                `json:"name"`
		Content        string                `json:"content"`
		State          coredata.TaskState    `json:"state"`
		Priority       coredata.TaskPriority `json:"priority"`
		Rank           int                   `json:"rank"`
		TimeEstimate   *timespan.TimeSpan    `json:"timeEstimate"`
		AssignedToID   *gid.GID              `json:"assignedToId"`
		Deadline       *time.Time            `json:"deadline"`
		Recurrence     *timespan.TimeSpan    `json:"recurrence"`
		CreatedAt      time.Time             `json:"createdAt"`
		UpdatedAt      time.Time             `json:"updatedAt"`
	}

	TaskComment struct {
		ID             gid.GID   `json:"id"`
		OrganizationID gid.GID   `json:"organizationId"`
		TaskID         gid.GID   `json:"taskId"`
		OwnerID        *gid.GID  `json:"ownerId"`
		Content        string    `json:"content"`
		CreatedAt      time.Time `json:"createdAt"`
		UpdatedAt      time.Time `json:"updatedAt"`
	}
)

func NewTask(t *coredata.Task) *Task {
	return &Task{
		ID:             t.ID,
		OrganizationID: t.OrganizationID,
		MeasureID:      t.MeasureID,
		Name:           t.Name,
		Content:        t.Content,
		State:          t.State,
		Priority:       t.Priority,
		Rank:           t.Rank,
		TimeEstimate:   t.TimeEstimate,
		AssignedToID:   t.AssignedToID,
		Deadline:       t.Deadline,
		Recurrence:     t.Recurrence,
		CreatedAt:      t.CreatedAt,
		UpdatedAt:      t.UpdatedAt,
	}
}

func NewTaskComment(c *coredata.TaskComment) *TaskComment {
	return &TaskComment{
		ID:             c.ID,
		OrganizationID: c.OrganizationID,
		TaskID:         c.TaskID,
		OwnerID:        c.OwnerID,
		Content:        c.Content,
		CreatedAt:      c.CreatedAt,
		UpdatedAt:      c.UpdatedAt,
	}
}
