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

package asana

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	client "go.probo.inc/probo/pkg/task/sync/asana/client"
	"go.probo.inc/probo/pkg/task/sync/shared"
	"go.probo.inc/probo/pkg/webhook"
	webhooktypes "go.probo.inc/probo/pkg/webhook/types"
)

type asanaInbox struct {
	ConnectorID string          `json:"connector_id"`
	Body        json.RawMessage `json:"body"`
}

func (s *Service) ApplyAsanaWebhook(ctx context.Context, envelope []byte) error {
	var inbox asanaInbox
	if err := json.Unmarshal(envelope, &inbox); err != nil {
		return fmt.Errorf("cannot unmarshal Asana webhook envelope: %w", err)
	}

	connectorID, err := gid.ParseGID(inbox.ConnectorID)
	if err != nil {
		return fmt.Errorf("cannot parse Asana connector id: %w", err)
	}

	batch, err := client.ParseEvents(inbox.Body)
	if err != nil {
		return fmt.Errorf("cannot parse Asana webhook events: %w", err)
	}

	scope := coredata.NewScopeFromObjectID(connectorID)

	account, err := s.asanaClientByConnectorID(ctx, scope, connectorID)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil
		}

		return fmt.Errorf("cannot load Asana connector: %w", err)
	}

	deleted := map[string]struct{}{}
	removedFromProject := map[string]map[string]struct{}{}
	tasks := map[string]struct{}{}
	added := map[string]string{}

	for _, event := range batch.Events {
		taskGID := event.TaskGID()
		if taskGID == "" {
			continue
		}

		if event.DeletedTask() {
			deleted[taskGID] = struct{}{}

			continue
		}

		if projectGID, ok := event.ProjectRemoval(); ok {
			projects := removedFromProject[taskGID]
			if projects == nil {
				projects = map[string]struct{}{}
				removedFromProject[taskGID] = projects
			}

			projects[projectGID] = struct{}{}
		}

		if event.AddedTask(event.Parent.GID) {
			added[event.Resource.GID] = event.Parent.GID
		}

		tasks[taskGID] = struct{}{}
	}

	for taskGID := range deleted {
		if err := s.unlinkAsanaTask(ctx, connectorID, taskGID); err != nil {
			return fmt.Errorf("cannot unlink deleted Asana task: %w", err)
		}

		delete(tasks, taskGID)
		delete(added, taskGID)
	}

	for taskGID, projects := range removedFromProject {
		link, err := s.asanaLink(ctx, connectorID, taskGID)
		if err != nil {
			return fmt.Errorf("cannot load Asana task link: %w", err)
		}

		if link == nil {
			continue
		}

		projectGID, err := link.DestinationTeamID()
		if err != nil {
			return fmt.Errorf("cannot read Asana project: %w", err)
		}

		if _, ok := projects[projectGID]; !ok {
			continue
		}

		if err := s.unlinkAsanaTask(ctx, connectorID, taskGID); err != nil {
			return fmt.Errorf("cannot unlink Asana task removed from project: %w", err)
		}

		delete(tasks, taskGID)
		delete(added, taskGID)
	}

	for taskGID := range tasks {
		link, err := s.asanaLink(ctx, connectorID, taskGID)
		if err != nil {
			return fmt.Errorf("cannot load linked Asana task: %w", err)
		}

		if link == nil {
			projectGID := added[taskGID]
			if projectGID == "" || !s.shouldImportAsanaTask(account.settings, projectGID) {
				continue
			}

			if err := s.importAsanaTask(ctx, scope, account, projectGID, taskGID); err != nil {
				if errors.Is(err, shared.ErrTaskAlreadyLinked) {
					continue
				}

				return fmt.Errorf("cannot import Asana task: %w", err)
			}

			continue
		}

		if err := s.pullAsanaTask(ctx, scope, account, link); err != nil {
			return fmt.Errorf("cannot pull Asana task: %w", err)
		}
	}

	return nil
}

func (s *Service) shouldImportAsanaTask(settings coredata.AsanaConnectorSettings, projectGID string) bool {
	return settings.ImportNewTasks && settings.DefaultProjectGID != "" && settings.DefaultProjectGID == projectGID
}

func (s *Service) asanaLink(
	ctx context.Context,
	connectorID gid.GID,
	taskGID string,
) (*coredata.TaskExternalLink, error) {
	var found *coredata.TaskExternalLink

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			links := coredata.TaskExternalLinks{}
			if err := links.LoadByExternalID(
				ctx,
				conn,
				coredata.NewNoScope(),
				coredata.ConnectorProviderAsana,
				taskGID,
			); err != nil {
				return fmt.Errorf("cannot load Asana task link: %w", err)
			}

			for _, link := range links {
				if link.ConnectorID == connectorID {
					found = link

					return nil
				}
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return found, nil
}

func (s *Service) unlinkAsanaTask(ctx context.Context, connectorID gid.GID, taskGID string) error {
	link, err := s.asanaLink(ctx, connectorID, taskGID)
	if err != nil {
		return fmt.Errorf("cannot load Asana task link: %w", err)
	}

	if link == nil {
		return nil
	}

	scope := coredata.NewScopeFromObjectID(link.OrganizationID)

	err = s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			current := &coredata.TaskExternalLink{}
			if err := current.LoadByTaskID(ctx, tx, scope, link.TaskID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return nil
				}

				return fmt.Errorf("cannot load task external link: %w", err)
			}

			if current.ExternalID != taskGID || current.Provider != coredata.ConnectorProviderAsana {
				return nil
			}

			if err := current.Delete(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot delete task external link: %w", err)
			}

			if err := shared.DeleteCommentLinks(ctx, tx, scope, current.TaskID); err != nil {
				return fmt.Errorf("cannot delete Asana comment links: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return fmt.Errorf("cannot unlink Asana task: %w", err)
	}

	return nil
}

func (s *Service) pullAsanaTask(
	ctx context.Context,
	scope coredata.Scoper,
	account asanaAccount,
	link *coredata.TaskExternalLink,
) error {
	remote, err := account.client.GetTask(ctx, link.ExternalID)
	if err != nil {
		if client.IsNotFound(err) {
			return s.unlinkAsanaTask(ctx, link.ConnectorID, link.ExternalID)
		}

		return fmt.Errorf("cannot load Asana task: %w", err)
	}

	projectGID, err := link.DestinationTeamID()
	if err != nil {
		return fmt.Errorf("cannot read Asana project: %w", err)
	}

	mapping := projectMapping(account.settings, projectGID)
	sectionGID := sectionForProject(remote, projectGID)

	optionGID := ""
	if mapping != nil {
		optionGID = enumOption(remote, mapping.PriorityFieldGID)
	}

	hash := asanaSnapshotHash(remote, sectionGID, optionGID)
	changed := link.ContentHash == nil || *link.ContentHash != hash

	task := &coredata.Task{}

	err = s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := task.LoadByID(ctx, conn, scope, link.TaskID); err != nil {
				return fmt.Errorf("cannot load task: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return fmt.Errorf("cannot load task: %w", err)
	}

	content := task.Content

	if changed {
		content, err = s.contentFromAsanaHTML(ctx, scope, account, link.OrganizationID, remote.HTMLNotes)
		if err != nil {
			return fmt.Errorf("cannot convert Asana description: %w", err)
		}

		err = s.pg.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				current := &coredata.Task{}
				if err := current.LoadByID(ctx, tx, scope, link.TaskID); err != nil {
					return fmt.Errorf("cannot load task: %w", err)
				}

				if err := s.applyRemoteTask(ctx, tx, scope, current, remote, content, projectGID, mapping, false); err != nil {
					return fmt.Errorf("cannot apply Asana task: %w", err)
				}

				return nil
			},
		)
		if err != nil {
			return fmt.Errorf("cannot update task from Asana: %w", err)
		}
	}

	attachments, err := account.client.ListAttachments(ctx, remote.GID)
	if err != nil {
		return fmt.Errorf("cannot list Asana attachments: %w", err)
	}

	meta, err := readAsanaFiles(link.Metadata)
	if err != nil {
		return fmt.Errorf("cannot read Asana files: %w", err)
	}

	content, files, err := s.pullAsanaFiles(
		ctx,
		scope,
		account,
		link.OrganizationID,
		content,
		task.Content,
		meta.Files,
		attachments,
		remote.GID,
		"task",
	)
	if err != nil {
		return fmt.Errorf("cannot pull Asana files: %w", err)
	}

	meta.Files = files

	encoded, err := writeAsanaFiles(meta)
	if err != nil {
		return fmt.Errorf("cannot encode Asana files: %w", err)
	}

	err = s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			currentTask := &coredata.Task{}
			if err := currentTask.LoadByID(ctx, tx, scope, link.TaskID); err != nil {
				return fmt.Errorf("cannot load task: %w", err)
			}

			if currentTask.Content != content {
				previous := currentTask.Content
				currentTask.Content = content

				currentTask.UpdatedAt = time.Now()

				if err := currentTask.Update(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot store Asana task content: %w", err)
				}

				if err := shared.NotifyMentionedProfiles(
					ctx,
					tx,
					scope,
					currentTask.OrganizationID,
					currentTask.ID,
					currentTask.Name,
					previous,
					currentTask.Content,
					nil,
				); err != nil {
					return fmt.Errorf("cannot notify synced mentions: %w", err)
				}
			}

			current := &coredata.TaskExternalLink{}
			if err := current.LoadByTaskID(ctx, tx, scope, link.TaskID); err != nil {
				return fmt.Errorf("cannot load task external link: %w", err)
			}

			current.ContentHash = &hash
			current.Metadata = encoded

			current.UpdatedAt = time.Now()
			if !remote.ModifiedAt.IsZero() {
				current.RemoteUpdatedAt = &remote.ModifiedAt
			}

			if err := current.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot store Asana sync hash: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return fmt.Errorf("cannot store Asana task sync: %w", err)
	}

	link.Metadata = encoded

	if err := s.pullAsanaComments(ctx, scope, account, link, remote.GID); err != nil {
		return fmt.Errorf("cannot pull Asana comments: %w", err)
	}

	return nil
}

func (s *Service) importAsanaTask(
	ctx context.Context,
	scope coredata.Scoper,
	account asanaAccount,
	projectGID string,
	taskGID string,
) error {
	existing, err := s.asanaLink(ctx, account.connector.ID, taskGID)
	if err != nil {
		return fmt.Errorf("cannot load Asana task link: %w", err)
	}

	if existing != nil {
		return shared.ErrTaskAlreadyLinked
	}

	remote, err := account.client.GetTask(ctx, taskGID)
	if err != nil {
		if client.IsNotFound(err) {
			return nil
		}

		return fmt.Errorf("cannot load Asana task: %w", err)
	}

	if !taskInProject(remote, projectGID) {
		return nil
	}

	mapping := projectMapping(account.settings, projectGID)

	content, err := s.contentFromAsanaHTML(
		ctx,
		scope,
		account,
		account.connector.OrganizationID,
		remote.HTMLNotes,
	)
	if err != nil {
		return fmt.Errorf("cannot convert Asana description: %w", err)
	}

	attachments, err := account.client.ListAttachments(ctx, remote.GID)
	if err != nil {
		return fmt.Errorf("cannot list Asana attachments: %w", err)
	}

	content, files, err := s.pullAsanaFiles(
		ctx,
		scope,
		account,
		account.connector.OrganizationID,
		content,
		"",
		map[string]string{},
		attachments,
		remote.GID,
		"task",
	)
	if err != nil {
		return fmt.Errorf("cannot pull Asana files: %w", err)
	}

	referenceID, err := uuid.NewV4()
	if err != nil {
		return fmt.Errorf("cannot generate reference id: %w", err)
	}

	state := coredata.TaskStateTodo
	if remote.Completed {
		state = coredata.TaskStateDone
	}

	state = inboundStateFromSection(sectionForProject(remote, projectGID), remote.Completed, state, mapping)

	priority := coredata.TaskPriorityMedium
	if mapping != nil {
		if mapped, ok := inboundPriorityFromOption(enumOption(remote, mapping.PriorityFieldGID), priority, mapping); ok {
			priority = mapped
		}
	}

	deadline, err := shared.DateToDeadline(firstNonEmpty(remote.DueAt, remote.DueOn))
	if err != nil {
		return fmt.Errorf("cannot parse Asana deadline: %w", err)
	}

	now := time.Now()
	task := &coredata.Task{
		ID:             gid.New(scope.GetTenantID(), coredata.TaskEntityType),
		OrganizationID: account.connector.OrganizationID,
		Name:           remote.Name,
		Content:        content,
		State:          state,
		Priority:       priority,
		Deadline:       deadline,
		ReferenceID:    "custom-task-" + referenceID.String(),
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	sectionGID := sectionForProject(remote, projectGID)

	optionGID := ""
	if mapping != nil {
		optionGID = enumOption(remote, mapping.PriorityFieldGID)
	}

	hash := asanaSnapshotHash(remote, sectionGID, optionGID)

	encoded, err := writeAsanaFiles(asanaFileMeta{Files: files})
	if err != nil {
		return fmt.Errorf("cannot encode Asana files: %w", err)
	}

	destination, err := json.Marshal(coredata.TaskExternalLinkDestination{TeamID: projectGID, ProjectID: &projectGID})
	if err != nil {
		return fmt.Errorf("cannot marshal Asana destination: %w", err)
	}

	taskURL := remote.Permalink
	if taskURL == "" {
		taskURL, err = asanaTaskURL(s.asanaAPIBaseURL, projectGID, remote.GID)
		if err != nil {
			return fmt.Errorf("cannot build Asana task URL: %w", err)
		}
	}

	link := &coredata.TaskExternalLink{
		OrganizationID:     task.OrganizationID,
		TaskID:             task.ID,
		ConnectorID:        account.connector.ID,
		Provider:           coredata.ConnectorProviderAsana,
		ExternalID:         remote.GID,
		ExternalIdentifier: remote.GID,
		ExternalURL:        taskURL,
		Destination:        destination,
		Origin:             coredata.TaskExternalLinkOriginAsana,
		ContentHash:        &hash,
		Metadata:           encoded,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if !remote.ModifiedAt.IsZero() {
		link.RemoteUpdatedAt = &remote.ModifiedAt
	}

	err = s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if _, err := s.applyAsanaAssignee(ctx, tx, scope, task, remote.AssigneeEmail); err != nil {
				return fmt.Errorf("cannot apply Asana assignee: %w", err)
			}

			if err := task.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert imported Asana task: %w", err)
			}

			if err := link.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert imported Asana task link: %w", err)
			}

			if err := shared.NotifyMentionedProfiles(
				ctx,
				tx,
				scope,
				task.OrganizationID,
				task.ID,
				task.Name,
				"",
				task.Content,
				nil,
			); err != nil {
				return fmt.Errorf("cannot notify imported Asana mentions: %w", err)
			}

			if s.recordImportedTask == nil {
				return nil
			}

			if err := s.recordImportedTask(ctx, tx, scope, task, nil, now); err != nil {
				return fmt.Errorf("cannot record imported Asana task: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return fmt.Errorf("cannot import Asana task: %w", err)
	}

	if err := s.pullAsanaComments(ctx, scope, account, link, remote.GID); err != nil {
		return fmt.Errorf("cannot pull Asana comments: %w", err)
	}

	return nil
}

func (s *Service) asanaCommentContent(
	ctx context.Context,
	scope coredata.Scoper,
	commentID gid.GID,
) (string, error) {
	comment := &coredata.TaskComment{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := comment.LoadByID(ctx, conn, scope, commentID); err != nil {
				return fmt.Errorf("cannot load task comment: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return "", fmt.Errorf("cannot load Asana comment: %w", err)
	}

	return comment.Content, nil
}

func (s *Service) pullAsanaComments(
	ctx context.Context,
	scope coredata.Scoper,
	account asanaAccount,
	link *coredata.TaskExternalLink,
	taskGID string,
) error {
	stories, err := account.client.ListStories(ctx, taskGID)
	if err != nil {
		return fmt.Errorf("cannot list Asana comments: %w", err)
	}

	known, err := shared.CommentLinksByExternalID(ctx, s.pg, scope, link.TaskID)
	if err != nil {
		return fmt.Errorf("cannot load Asana comment links: %w", err)
	}

	seen := map[string]struct{}{}

	for _, story := range stories {
		if story.GID == "" {
			continue
		}

		seen[story.GID] = struct{}{}
		hash := shared.CommentContentHash(story.HTMLText)
		existing := known[story.GID]
		unchanged := existing != nil && existing.ContentHash != nil && *existing.ContentHash == hash

		meta, err := readAsanaFiles(link.Metadata)
		if err != nil {
			return fmt.Errorf("cannot read Asana comment files: %w", err)
		}

		if meta.CommentFiles == nil {
			meta.CommentFiles = map[string]map[string]string{}
		}

		previousFiles := meta.CommentFiles[story.GID]

		previous := ""

		if existing != nil && len(previousFiles) > 0 {
			loaded, err := s.asanaCommentContent(ctx, scope, existing.TaskCommentID)
			if err != nil {
				return fmt.Errorf("cannot load Asana comment content: %w", err)
			}

			previous = loaded
		}

		origin, err := asanaOrigin(s.asanaAPIBaseURL)
		if err != nil {
			return fmt.Errorf("cannot resolve Asana origin: %w", err)
		}

		content, err := asanaHTMLToContent(asanaMentionAnchorsToURLs(story.HTMLText, origin))
		if err != nil {
			return fmt.Errorf("cannot convert Asana comment: %w", err)
		}

		attachments, err := account.client.ListStoryAttachments(ctx, story.GID)
		if err != nil {
			return fmt.Errorf("cannot list Asana comment attachments: %w", err)
		}

		content, files, err := s.pullAsanaFiles(
			ctx,
			scope,
			account,
			link.OrganizationID,
			content,
			previous,
			meta.CommentFiles[story.GID],
			attachments,
			story.GID,
			"story",
		)
		if err != nil {
			return fmt.Errorf("cannot pull Asana comment files: %w", err)
		}

		meta.CommentFiles[story.GID] = files

		encoded, err := writeAsanaFiles(meta)
		if err != nil {
			return fmt.Errorf("cannot encode Asana comment files: %w", err)
		}

		if unchanged && sameStringMap(previousFiles, files) {
			continue
		}

		content, err = s.attachAsanaMentions(ctx, scope, account, link.OrganizationID, content, origin)
		if err != nil {
			return fmt.Errorf("cannot attach Asana mentions: %w", err)
		}

		if err := s.saveAsanaComment(ctx, scope, link, story, content, hash, encoded); err != nil {
			return fmt.Errorf("cannot save Asana comment: %w", err)
		}

		link.Metadata = encoded
	}

	for externalID, mapping := range known {
		if _, ok := seen[externalID]; ok {
			continue
		}

		if err := s.pg.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				return shared.DeleteSyncedComment(ctx, tx, scope, mapping)
			},
		); err != nil {
			return fmt.Errorf("cannot delete removed Asana comment: %w", err)
		}
	}

	return nil
}

func (s *Service) saveAsanaComment(
	ctx context.Context,
	scope coredata.Scoper,
	link *coredata.TaskExternalLink,
	story client.Story,
	content string,
	hash string,
	metadata json.RawMessage,
) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			mappings := coredata.TaskCommentExternalLinks{}
			if err := mappings.LoadByExternalID(
				ctx,
				tx,
				scope,
				coredata.ConnectorProviderAsana,
				story.GID,
			); err != nil {
				return fmt.Errorf("cannot load Asana comment link: %w", err)
			}

			now := time.Now()

			var mapping *coredata.TaskCommentExternalLink

			for _, candidate := range mappings {
				if candidate.TaskID == link.TaskID {
					mapping = candidate

					break
				}
			}

			if mapping == nil {
				ownerID, err := shared.ResolveCommentOwner(ctx, tx, scope, link, story.AuthorEmail)
				if err != nil {
					return fmt.Errorf("cannot resolve Asana comment owner: %w", err)
				}

				comment := &coredata.TaskComment{
					ID:             gid.New(scope.GetTenantID(), coredata.TaskCommentEntityType),
					OrganizationID: link.OrganizationID,
					TaskID:         link.TaskID,
					OwnerID:        ownerID,
					Content:        content,
					CreatedAt:      now,
					UpdatedAt:      now,
				}
				if err := comment.Insert(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot insert Asana comment: %w", err)
				}

				mapping = &coredata.TaskCommentExternalLink{
					TaskCommentID:  comment.ID,
					OrganizationID: link.OrganizationID,
					TaskID:         link.TaskID,
					ConnectorID:    link.ConnectorID,
					Provider:       coredata.ConnectorProviderAsana,
					ExternalID:     story.GID,
					ContentHash:    &hash,
					CreatedAt:      now,
					UpdatedAt:      now,
				}
				if !story.CreatedAt.IsZero() {
					mapping.RemoteUpdatedAt = &story.CreatedAt
				}

				if err := mapping.Insert(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot insert Asana comment link: %w", err)
				}

				if err := webhook.InsertData(
					ctx,
					tx,
					scope,
					comment.OrganizationID,
					coredata.WebhookEventTypeTaskCommentCreated,
					webhooktypes.NewTaskComment(comment),
				); err != nil {
					return fmt.Errorf("cannot emit task comment created webhook: %w", err)
				}

				if err := shared.NotifyCommentMentions(
					ctx,
					tx,
					scope,
					comment.TaskID,
					"",
					comment.Content,
					comment.OwnerID,
				); err != nil {
					return fmt.Errorf("cannot notify synced comment mentions: %w", err)
				}
			} else if err := shared.UpdateImportedComment(ctx, tx, scope, mapping, content, hash, mapping.RemoteUpdatedAt, nil); err != nil {
				return fmt.Errorf("cannot update Asana comment: %w", err)
			}

			current := &coredata.TaskExternalLink{}
			if err := current.LoadByTaskID(ctx, tx, scope, link.TaskID); err != nil {
				return fmt.Errorf("cannot load task external link: %w", err)
			}

			current.Metadata = metadata

			current.UpdatedAt = now
			if err := current.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot store Asana comment files: %w", err)
			}

			return nil
		},
	)
}
