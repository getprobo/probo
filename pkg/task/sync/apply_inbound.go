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

package tasksync

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/task/sync/linear"
)

func (s *Service) ApplyInboundIssue(
	ctx context.Context,
	envelope *linear.WebhookEnvelope,
) error {
	if envelope.Type != "Issue" {
		return nil
	}

	if envelope.Action != "update" && envelope.Action != "create" {
		return nil
	}

	data, err := envelope.IssueData()
	if err != nil {
		return err
	}

	if data.ID == "" {
		return nil
	}

	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			links := coredata.TaskExternalLinks{}
			if err := links.LoadByExternalID(
				ctx,
				tx,
				coredata.NewNoScope(),
				coredata.ConnectorProviderLinear,
				data.ID,
			); err != nil {
				return fmt.Errorf("cannot load task external links: %w", err)
			}

			link, ok := PickLinkForLinearOrganization(links, envelope.OrganizationID)
			if !ok {
				return nil
			}

			if isAppActor(link.Metadata, envelope.Actor.ID) {
				return nil
			}

			scope := coredata.NewScope(link.OrganizationID.TenantID())
			if err := link.LoadByTaskIDForUpdate(ctx, tx, scope, link.TaskID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return nil
				}

				return fmt.Errorf("cannot lock task external link: %w", err)
			}

			if link.ExternalID != data.ID {
				return nil
			}

			if isAppActor(link.Metadata, envelope.Actor.ID) {
				return nil
			}

			remoteUpdatedAt := linearDateTime(data.UpdatedAt)
			if inboundEventIsStale(remoteUpdatedAt, link.RemoteUpdatedAt) {
				return nil
			}

			task := &coredata.Task{}
			if err := task.LoadByIDForUpdate(ctx, tx, scope, link.TaskID); err != nil {
				return fmt.Errorf("cannot load task %q: %w", link.TaskID, err)
			}

			mapped, err := mapInboundIssue(task, data)
			if err != nil {
				return err
			}

			hash := ContentHash(mapped.Name, mapped.Markdown, mapped.State, mapped.Priority, mapped.Deadline)
			if link.ContentHash != nil && *link.ContentHash == hash {
				return touchInboundLink(ctx, tx, scope, link, data, remoteUpdatedAt, hash)
			}

			oldState := task.State
			oldPriority := task.Priority

			task.Name = mapped.Name
			task.Content = mapped.Content
			task.State = mapped.State
			task.Priority = mapped.Priority
			task.Deadline = mapped.Deadline
			task.UpdatedAt = time.Now()

			if task.State != oldState || task.Priority != oldPriority {
				if err := task.NextRankForStatePriority(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot get next rank: %w", err)
				}
			}

			if err := task.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update task from Linear: %w", err)
			}

			return touchInboundLink(ctx, tx, scope, link, data, remoteUpdatedAt, hash)
		},
	)
}

type inboundMappedFields struct {
	Name     string
	Content  string
	Markdown string
	State    coredata.TaskState
	Priority coredata.TaskPriority
	Deadline *time.Time
}

func mapInboundIssue(task *coredata.Task, data *linear.IssueWebhookData) (inboundMappedFields, error) {
	markdown, err := ContentToMarkdown(task.Content)
	if err != nil {
		return inboundMappedFields{}, err
	}

	mapped := inboundMappedFields{
		Name:     task.Name,
		Content:  task.Content,
		Markdown: markdown,
		State:    task.State,
		Priority: task.Priority,
		Deadline: task.Deadline,
	}

	if data.Has("title") {
		mapped.Name = data.Title
	}

	if data.Has("description") {
		content, err := MarkdownToContent(data.Description)
		if err != nil {
			return inboundMappedFields{}, err
		}

		mapped.Content = content
		mapped.Markdown = data.Description
	}

	if data.Has("state") {
		mapped.State = LinearTypeToTaskState(data.State.Type)
	}

	if data.Has("priority") {
		mapped.Priority = LinearPriorityToTask(data.Priority)
	}

	if data.Has("dueDate") {
		mapped.Deadline = LinearDateToDeadline(data.DueDate)
	}

	return mapped, nil
}

func PickLinkForLinearOrganization(
	links coredata.TaskExternalLinks,
	linearOrganizationID string,
) (*coredata.TaskExternalLink, bool) {
	if linearOrganizationID == "" {
		return nil, false
	}

	var match *coredata.TaskExternalLink

	for _, link := range links {
		if link.DestinationLinearOrganizationID() != linearOrganizationID {
			continue
		}

		if match != nil {
			return nil, false
		}

		match = link
	}

	if match == nil {
		return nil, false
	}

	return match, true
}

func touchInboundLink(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	link *coredata.TaskExternalLink,
	data *linear.IssueWebhookData,
	remoteUpdatedAt *time.Time,
	hash string,
) error {
	link.ContentHash = &hash

	if remoteUpdatedAt != nil {
		link.RemoteUpdatedAt = remoteUpdatedAt
	}

	if data.Has("identifier") && data.Identifier != "" {
		link.ExternalIdentifier = data.Identifier
	}

	if data.Has("url") && data.URL != "" {
		link.ExternalURL = data.URL
	}

	link.UpdatedAt = time.Now()

	if err := link.Update(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot update task external link: %w", err)
	}

	return nil
}

func isAppActor(metadata json.RawMessage, actorID string) bool {
	if actorID == "" || len(metadata) == 0 {
		return false
	}

	var payload struct {
		AppActorID string `json:"app_actor_id"`
	}

	if err := json.Unmarshal(metadata, &payload); err != nil {
		return false
	}

	return payload.AppActorID != "" && payload.AppActorID == actorID
}

func inboundEventIsStale(remoteUpdatedAt, stored *time.Time) bool {
	if stored == nil {
		return false
	}

	if remoteUpdatedAt == nil {
		return true
	}

	return !remoteUpdatedAt.After(*stored)
}

func linearDateTime(value string) *time.Time {
	if value == "" {
		return nil
	}

	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return &t
	}

	if t, err := time.Parse(time.RFC3339Nano, value); err == nil {
		return &t
	}

	return nil
}
