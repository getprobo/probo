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

package linear

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	linear "go.probo.inc/probo/pkg/task/sync/linear/client"
	"go.probo.inc/probo/pkg/task/sync/shared"
	"go.probo.inc/probo/pkg/webhook"
	webhooktypes "go.probo.inc/probo/pkg/webhook/types"
)

func (s *Service) copyTaskCommentsToLinear(
	ctx context.Context,
	scope coredata.Scoper,
	client *linear.Client,
	link *coredata.TaskExternalLink,
) ([]string, error) {
	comments, err := s.listTaskComments(ctx, scope, link.TaskID)
	if err != nil {
		return nil, fmt.Errorf("cannot list task comments: %w", err)
	}

	mapped, err := shared.CommentLinksByCommentID(ctx, s.pg, scope, link.TaskID)
	if err != nil {
		return nil, fmt.Errorf("cannot load comment links: %w", err)
	}

	var createdIDs []string

	for _, comment := range comments {
		if _, ok := mapped[comment.ID]; ok {
			continue
		}

		canonical, err := ContentToMarkdown(comment.Content)
		if err != nil {
			return createdIDs, fmt.Errorf("cannot convert task comment %q: %w", comment.ID, err)
		}

		if strings.TrimSpace(canonical) == "" {
			continue
		}

		markdown, err := s.linearCommentMarkdown(
			ctx,
			scope,
			client,
			comment.OrganizationID,
			comment.Content,
		)
		if err != nil {
			return createdIDs, fmt.Errorf("cannot convert task comment %q mentions: %w", comment.ID, err)
		}

		description, files, changed, err := s.embedLinearDescription(
			ctx,
			client,
			scope,
			link.ConnectorID,
			link.ExternalID,
			comment.OrganizationID,
			link.TaskID,
			markdown,
			comment.Content,
		)
		if err != nil {
			return createdIDs, fmt.Errorf("cannot embed Linear files for comment %q: %w", comment.ID, err)
		}

		if changed {
			link.Metadata, err = mergeAttachments(link.Metadata, files)
			if err != nil {
				return createdIDs, fmt.Errorf("cannot merge Linear attachments for comment %q: %w", comment.ID, err)
			}
		}

		author, err := s.commentAuthor(ctx, scope, comment.OwnerID)
		if err != nil {
			return createdIDs, fmt.Errorf("cannot load Linear comment author for %q: %w", comment.ID, err)
		}

		created, err := client.CreateComment(ctx, link.ExternalID, description, author)
		if err != nil {
			return createdIDs, fmt.Errorf("cannot create Linear comment for %q: %w", comment.ID, err)
		}

		createdIDs = append(createdIDs, created.ID)

		if err := s.saveCommentLink(ctx, scope, link, comment.ID, created, shared.CommentContentHash(canonical)); err != nil {
			return createdIDs, fmt.Errorf("cannot save Linear comment link for %q: %w", comment.ID, err)
		}
	}

	return createdIDs, nil
}

func (s *Service) syncLinkedComments(
	ctx context.Context,
	scope coredata.Scoper,
	client *linear.Client,
	link *coredata.TaskExternalLink,
) error {
	remote, err := client.ListIssueComments(ctx, link.ExternalID)
	if err != nil {
		return fmt.Errorf("cannot list Linear comments: %w", err)
	}

	mappedExternal, err := shared.CommentLinksByExternalID(ctx, s.pg, scope, link.TaskID)
	if err != nil {
		return fmt.Errorf("cannot load comment links: %w", err)
	}

	var importedIDs []gid.GID

	for _, comment := range remote {
		if _, ok := mappedExternal[comment.ID]; ok {
			continue
		}

		importedID, err := s.importLinearComment(ctx, scope, client, link, &comment)
		if err != nil {
			s.rollbackLinkedComments(ctx, scope, client, importedIDs, nil)

			return fmt.Errorf("cannot import Linear comment: %w", err)
		}

		if importedID != gid.Nil {
			importedIDs = append(importedIDs, importedID)
		}
	}

	createdIDs, err := s.copyTaskCommentsToLinear(ctx, scope, client, link)
	if err != nil {
		s.rollbackLinkedComments(ctx, scope, client, importedIDs, createdIDs)

		return fmt.Errorf("cannot copy task comments: %w", err)
	}

	return nil
}

func (s *Service) importLinearComment(
	ctx context.Context,
	scope coredata.Scoper,
	client *linear.Client,
	link *coredata.TaskExternalLink,
	remote *linear.Comment,
) (gid.GID, error) {
	if strings.TrimSpace(remote.Body) == "" {
		return gid.Nil, nil
	}

	markdown, files, metas, err := s.importLinearMarkdown(
		ctx,
		client,
		scope,
		link.ConnectorID,
		link.OrganizationID,
		remote.Body,
		attachmentsFromMetadata(link.Metadata),
	)
	if err != nil {
		return gid.Nil, fmt.Errorf("cannot import Linear comment %q files: %w", remote.ID, err)
	}

	if next, changed, err := mergedLinearAttachments(link.Metadata, files); err != nil {
		return gid.Nil, fmt.Errorf("cannot merge Linear attachments: %w", err)
	} else if changed {
		if err := s.saveLinearAttachments(
			ctx,
			scope,
			link.TaskID,
			link.ConnectorID,
			link.ExternalID,
			files,
		); err != nil {
			return gid.Nil, fmt.Errorf("cannot save Linear attachments: %w", err)
		}

		link.Metadata = next
	}

	content, canonical, err := s.contentFromLinearMarkdown(
		ctx,
		client,
		scope,
		link.OrganizationID,
		markdown,
		metas,
	)
	if err != nil {
		return gid.Nil, fmt.Errorf("cannot convert Linear comment %q: %w", remote.ID, err)
	}

	if strings.TrimSpace(canonical) == "" {
		return gid.Nil, nil
	}

	ownerID, err := s.commentOwner(ctx, scope, link, remote.UserEmail)
	if err != nil {
		return gid.Nil, fmt.Errorf("cannot resolve comment owner: %w", err)
	}

	now := time.Now()
	taskComment := &coredata.TaskComment{
		ID:             gid.New(scope.GetTenantID(), coredata.TaskCommentEntityType),
		OrganizationID: link.OrganizationID,
		TaskID:         link.TaskID,
		OwnerID:        ownerID,
		Content:        content,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	hash := shared.CommentContentHash(canonical)
	commentLink := newCommentLink(link, taskComment.ID, remote, hash, now)

	err = s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := taskComment.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert task comment: %w", err)
			}

			if err := commentLink.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert task comment external link: %w", err)
			}

			if err := insertSyncedCommentActivity(ctx, tx, scope, taskComment); err != nil {
				return fmt.Errorf("cannot record imported comment activity: %w", err)
			}

			actorIdentityID, err := linearActorIdentityID(ctx, tx, scope, link.OrganizationID, remote.UserEmail)
			if err != nil {
				return fmt.Errorf("cannot match Linear actor: %w", err)
			}

			if err := webhook.InsertData(
				ctx,
				tx,
				coredata.NewScope(link.OrganizationID.TenantID(), actorIdentityID),
				taskComment.OrganizationID,
				coredata.WebhookEventTypeTaskCommentCreated,
				webhooktypes.NewTaskComment(taskComment),
			); err != nil {
				return fmt.Errorf("cannot emit task comment created webhook: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceAlreadyExists) {
			return gid.Nil, nil
		}

		return gid.Nil, fmt.Errorf("cannot import Linear comment: %w", err)
	}

	return taskComment.ID, nil
}

func (s *Service) rollbackLinkedComments(
	ctx context.Context,
	scope coredata.Scoper,
	client *linear.Client,
	importedIDs []gid.GID,
	createdLinearIDs []string,
) {
	removedLinearIDs := make([]string, 0, len(createdLinearIDs))

	for _, linearID := range createdLinearIDs {
		if linearID == "" {
			continue
		}

		err := client.DeleteComment(ctx, linearID)
		if err != nil && !errors.Is(err, linear.ErrCommentNotFound) {
			if s.logger != nil {
				s.logger.WarnCtx(
					ctx,
					"cannot delete Linear comment after failed link",
					log.Error(err),
				)
			}

			continue
		}

		removedLinearIDs = append(removedLinearIDs, linearID)
	}

	if len(importedIDs) == 0 && len(removedLinearIDs) == 0 {
		return
	}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			for _, linearID := range removedLinearIDs {
				var links coredata.TaskCommentExternalLinks

				if err := links.LoadByExternalID(
					ctx,
					tx,
					scope,
					coredata.ConnectorProviderLinearSync,
					linearID,
				); err != nil {
					return fmt.Errorf("cannot load task comment external link: %w", err)
				}

				for _, link := range links {
					if err := link.Delete(ctx, tx, scope); err != nil {
						return fmt.Errorf("cannot delete task comment external link: %w", err)
					}
				}
			}

			for _, commentID := range importedIDs {
				comment := coredata.TaskComment{ID: commentID}
				if err := comment.Delete(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot delete imported task comment: %w", err)
				}

				if _, err := coredata.DeleteCommentedActivity(ctx, tx, scope, commentID); err != nil {
					return fmt.Errorf("cannot delete imported comment activity: %w", err)
				}
			}

			return nil
		},
	)
	if err != nil && s.logger != nil {
		s.logger.WarnCtx(
			ctx,
			"cannot delete synced comments after failed link",
			log.Error(err),
		)
	}
}

func (s *Service) ApplyInboundComment(
	ctx context.Context,
	envelope *linear.WebhookEnvelope,
) error {
	if envelope.Type != "Comment" {
		return nil
	}

	if envelope.Action == "remove" {
		return s.unlinkRemovedComment(ctx, envelope)
	}

	if envelope.Action != "create" && envelope.Action != "update" {
		return nil
	}

	data, err := envelope.CommentData()
	if err != nil {
		return fmt.Errorf("cannot read Linear comment: %w", err)
	}

	if data.ID == "" || data.IssueID == "" {
		return nil
	}

	prepared, err := s.prepareInboundComment(ctx, envelope, data)
	if err != nil {
		return fmt.Errorf("cannot prepare Linear comment: %w", err)
	}

	var unused *linearImport

	err = s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			link, scope, ok, err := lockedLinearLink(ctx, tx, data.IssueID, envelope.OrganizationID)
			if err != nil {
				return fmt.Errorf("cannot lock Linear task link: %w", err)
			}

			if !ok || isAppActor(link.Metadata, envelope.Actor.ID) {
				unused = prepared

				return nil
			}

			if err := s.applyInboundCommentTx(ctx, tx, scope, link, envelope, data, prepared); err != nil {
				return fmt.Errorf("cannot apply Linear comment: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		if cleanupErr := s.discardInboundImport(ctx, prepared); cleanupErr != nil {
			return fmt.Errorf("cannot discard Linear comment files: %w", errors.Join(err, cleanupErr))
		}

		if errors.Is(err, coredata.ErrResourceAlreadyExists) {
			return nil
		}

		return fmt.Errorf("cannot apply Linear comment: %w", err)
	}

	if unused == nil {
		s.retainInboundImport(prepared)
	}

	if err := s.discardInboundImport(ctx, unused); err != nil {
		return fmt.Errorf("cannot discard Linear comment files: %w", err)
	}

	return nil
}

func (s *Service) unlinkRemovedComment(
	ctx context.Context,
	envelope *linear.WebhookEnvelope,
) error {
	data, err := envelope.CommentData()
	if err != nil {
		return fmt.Errorf("cannot read Linear comment: %w", err)
	}

	if data.ID == "" || data.IssueID == "" {
		return nil
	}

	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			link, scope, ok, err := lockedLinearLink(ctx, tx, data.IssueID, envelope.OrganizationID)
			if err != nil {
				return fmt.Errorf("cannot lock Linear task link: %w", err)
			}

			if !ok {
				return nil
			}

			mappings := coredata.TaskCommentExternalLinks{}
			if err := mappings.LoadByExternalID(
				ctx,
				tx,
				scope,
				coredata.ConnectorProviderLinearSync,
				data.ID,
			); err != nil {
				return fmt.Errorf("cannot load task comment external link: %w", err)
			}

			actorIdentityID, err := linearActorIdentityID(ctx, tx, scope, link.OrganizationID, envelope.Actor.Email)
			if err != nil {
				return fmt.Errorf("cannot match Linear actor: %w", err)
			}

			actorScope := coredata.NewScope(link.OrganizationID.TenantID(), actorIdentityID)

			for _, mapping := range mappings {
				if mapping.TaskID != link.TaskID {
					continue
				}

				if err := shared.DeleteSyncedComment(ctx, tx, actorScope, mapping); err != nil {
					return fmt.Errorf("cannot delete Linear comment: %w", err)
				}
			}

			return nil
		},
	)
}

func (s *Service) prepareInboundComment(
	ctx context.Context,
	envelope *linear.WebhookEnvelope,
	data *linear.CommentWebhookData,
) (*linearImport, error) {
	if data == nil || !data.Has("body") || strings.TrimSpace(data.Body) == "" {
		return nil, nil
	}

	link, scope, ok, err := s.loadLinearLink(ctx, data.IssueID, envelope.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("cannot load Linear task link: %w", err)
	}

	if !ok {
		return nil, nil
	}

	if isAppActor(link.Metadata, envelope.Actor.ID) {
		return nil, nil
	}

	return s.prepareLinearImport(
		ctx,
		scope,
		link.ConnectorID,
		link.OrganizationID,
		data.Body,
		attachmentsFromMetadata(link.Metadata),
		false,
	)
}

func (s *Service) applyInboundCommentTx(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	link *coredata.TaskExternalLink,
	envelope *linear.WebhookEnvelope,
	data *linear.CommentWebhookData,
	prepared *linearImport,
) error {
	mappings := coredata.TaskCommentExternalLinks{}
	if err := mappings.LoadByExternalID(
		ctx,
		tx,
		scope,
		coredata.ConnectorProviderLinearSync,
		data.ID,
	); err != nil {
		return fmt.Errorf("cannot load task comment external link: %w", err)
	}

	var existing *coredata.TaskCommentExternalLink

	for _, mapping := range mappings {
		if mapping.TaskID == link.TaskID {
			existing = mapping
			break
		}
	}

	if !data.Has("body") && existing != nil {
		return nil
	}

	files := attachmentsFromMetadata(link.Metadata)

	var (
		body  string
		metas map[string]importedFileMeta
		err   error
	)

	var (
		content   string
		canonical string
	)

	if prepared != nil && prepared.mentionsApplied {
		content = prepared.mentionContent
		canonical = prepared.mentionCanonical
		files = mergeAttachmentMaps(files, prepared.files)
	} else if prepared != nil {
		body = prepared.markdown
		files = mergeAttachmentMaps(files, prepared.files)
		metas = prepared.metas

		content, err = MarkdownToContent(body)
		if err != nil {
			return fmt.Errorf("cannot parse Linear comment: %w", err)
		}

		content, canonical, err = applyImportedFiles(content, metas)
		if err != nil {
			return fmt.Errorf("cannot apply imported Linear comment files: %w", err)
		}
	} else {
		body, files, metas, err = s.importLinearMarkdown(
			ctx,
			nil,
			scope,
			link.ConnectorID,
			link.OrganizationID,
			data.Body,
			files,
		)
		if err != nil {
			return fmt.Errorf("cannot import Linear comment files: %w", err)
		}

		var client *linear.Client

		if textMayContainLinearProfile(body) {
			client, err = s.linearClientForConnectorID(ctx, scope, link.ConnectorID)
			if err != nil {
				return fmt.Errorf("cannot resolve Linear mentions: %w", err)
			}
		}

		content, canonical, err = s.contentFromLinearMarkdown(
			ctx,
			client,
			scope,
			link.OrganizationID,
			body,
			metas,
		)
		if err != nil {
			return fmt.Errorf("cannot convert Linear comment: %w", err)
		}
	}

	if strings.TrimSpace(canonical) == "" {
		if err := s.discardInboundImport(ctx, prepared); err != nil {
			return fmt.Errorf("cannot discard empty Linear comment: %w", err)
		}

		return nil
	}

	authorEmail := envelope.Actor.Email
	if data.User != nil && strings.TrimSpace(data.User.Email) != "" {
		authorEmail = data.User.Email
	}

	actorIdentityID, err := linearActorIdentityID(ctx, tx, scope, link.OrganizationID, authorEmail)
	if err != nil {
		return fmt.Errorf("cannot match Linear actor: %w", err)
	}

	actorScope := coredata.NewScope(link.OrganizationID.TenantID(), actorIdentityID)

	hash := shared.CommentContentHash(canonical)
	remoteUpdatedAt := linearDateTime(data.UpdatedAt)

	if existing != nil && inboundEventIsStale(remoteUpdatedAt, existing.RemoteUpdatedAt) {
		if err := s.discardInboundImport(ctx, prepared); err != nil {
			return fmt.Errorf("cannot discard stale Linear comment: %w", err)
		}

		return nil
	}

	if next, changed, mergeErr := mergedLinearAttachments(link.Metadata, files); mergeErr != nil {
		return mergeErr
	} else if changed {
		link.Metadata = next
		link.UpdatedAt = time.Now()

		if err := link.Update(ctx, tx, scope); err != nil {
			return fmt.Errorf("cannot store Linear attachments: %w", err)
		}
	}

	if existing != nil {
		if existing.ContentHash != nil && *existing.ContentHash == hash {
			existing.RemoteUpdatedAt = remoteUpdatedAt
			existing.UpdatedAt = time.Now()

			if err := existing.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot touch task comment external link: %w", err)
			}

			return nil
		}

		mentionActor := envelope.Actor
		mentionActor.Email = authorEmail

		actorProfileID, err := inboundActivityActorID(ctx, tx, scope, link.OrganizationID, mentionActor)
		if err != nil {
			return fmt.Errorf("cannot resolve Linear mention actor: %w", err)
		}

		return shared.UpdateImportedComment(
			ctx,
			tx,
			actorScope,
			existing,
			content,
			hash,
			remoteUpdatedAt,
			actorProfileID,
		)
	}

	ownerID, err := shared.ResolveCommentOwner(ctx, tx, scope, link, authorEmail)
	if err != nil {
		return fmt.Errorf("cannot resolve comment owner: %w", err)
	}

	now := time.Now()
	taskComment := &coredata.TaskComment{
		ID:             gid.New(scope.GetTenantID(), coredata.TaskCommentEntityType),
		OrganizationID: link.OrganizationID,
		TaskID:         link.TaskID,
		OwnerID:        ownerID,
		Content:        content,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	commentLink := &coredata.TaskCommentExternalLink{
		TaskCommentID:   taskComment.ID,
		OrganizationID:  link.OrganizationID,
		TaskID:          link.TaskID,
		ConnectorID:     link.ConnectorID,
		Provider:        coredata.ConnectorProviderLinearSync,
		ExternalID:      data.ID,
		RemoteUpdatedAt: remoteUpdatedAt,
		ContentHash:     &hash,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := taskComment.Insert(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot insert task comment: %w", err)
	}

	if err := commentLink.Insert(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot insert task comment external link: %w", err)
	}

	if err := insertSyncedCommentActivity(ctx, tx, scope, taskComment); err != nil {
		return fmt.Errorf("cannot record inbound comment activity: %w", err)
	}

	if err := webhook.InsertData(
		ctx,
		tx,
		actorScope,
		taskComment.OrganizationID,
		coredata.WebhookEventTypeTaskCommentCreated,
		webhooktypes.NewTaskComment(taskComment),
	); err != nil {
		return fmt.Errorf("cannot emit task comment created webhook: %w", err)
	}

	mentionActor := envelope.Actor
	mentionActor.Email = authorEmail

	actorProfileID, err := inboundActivityActorID(ctx, tx, scope, link.OrganizationID, mentionActor)
	if err != nil {
		return fmt.Errorf("cannot resolve Linear mention actor: %w", err)
	}

	if err := shared.NotifyCommentMentions(
		ctx,
		tx,
		scope,
		taskComment.TaskID,
		"",
		taskComment.Content,
		actorProfileID,
	); err != nil {
		return fmt.Errorf("cannot notify synced comment mentions: %w", err)
	}

	return nil
}

func insertSyncedCommentActivity(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	comment *coredata.TaskComment,
) error {
	commentID := comment.ID.String()
	activity := &coredata.TaskActivity{
		ID:             gid.New(scope.GetTenantID(), coredata.TaskActivityEntityType),
		OrganizationID: comment.OrganizationID,
		TaskID:         comment.TaskID,
		ActorID:        comment.OwnerID,
		ActivityType:   coredata.TaskActivityTypeCommented,
		NewValue:       &commentID,
		CreatedAt:      comment.CreatedAt,
	}

	if err := activity.Insert(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot insert task comment activity: %w", err)
	}

	return nil
}

func newCommentLink(
	link *coredata.TaskExternalLink,
	commentID gid.GID,
	remote *linear.Comment,
	hash string,
	now time.Time,
) *coredata.TaskCommentExternalLink {
	commentLink := &coredata.TaskCommentExternalLink{
		TaskCommentID:  commentID,
		OrganizationID: link.OrganizationID,
		TaskID:         link.TaskID,
		ConnectorID:    link.ConnectorID,
		Provider:       coredata.ConnectorProviderLinearSync,
		ExternalID:     remote.ID,
		ContentHash:    &hash,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if !remote.UpdatedAt.IsZero() {
		commentLink.RemoteUpdatedAt = &remote.UpdatedAt
	}

	return commentLink
}
