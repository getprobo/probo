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

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/task/sync/internal"
	linear "go.probo.inc/probo/pkg/task/sync/linear/client"
)

func (s *Service) listTaskComments(
	ctx context.Context,
	scope coredata.Scoper,
	taskID gid.GID,
) ([]*coredata.TaskComment, error) {
	var comments []*coredata.TaskComment

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			loaded, err := loadTaskComments(ctx, conn, scope, taskID)
			if err != nil {
				return err
			}

			comments = loaded

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return comments, nil
}

func loadTaskComments(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	taskID gid.GID,
) ([]*coredata.TaskComment, error) {
	loaded, err := page.LoadAll(
		ctx,
		page.OrderBy[coredata.TaskCommentOrderField]{
			Field:     coredata.TaskCommentOrderFieldCreatedAt,
			Direction: page.OrderDirectionAsc,
		},
		func(ctx context.Context, cursor *page.Cursor[coredata.TaskCommentOrderField]) ([]*coredata.TaskComment, error) {
			var batch coredata.TaskComments
			if err := batch.LoadByTaskID(ctx, conn, scope, taskID, cursor); err != nil {
				return nil, err
			}

			return batch, nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot list task comments: %w", err)
	}

	return loaded, nil
}

func (s *Service) saveCommentLink(
	ctx context.Context,
	scope coredata.Scoper,
	link *coredata.TaskExternalLink,
	commentID gid.GID,
	remote *linear.Comment,
	hash string,
) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			return saveCommentLinkTx(ctx, tx, scope, link, commentID, remote, hash)
		},
	)
}

func saveCommentLinkTx(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	link *coredata.TaskExternalLink,
	commentID gid.GID,
	remote *linear.Comment,
	hash string,
) error {
	now := time.Now()
	existing := &coredata.TaskCommentExternalLink{}

	err := existing.LoadByCommentID(ctx, tx, scope, commentID)
	if errors.Is(err, coredata.ErrResourceNotFound) {
		created := newCommentLink(link, commentID, remote, hash, now)
		if err := created.Insert(ctx, tx, scope); err != nil && !errors.Is(err, coredata.ErrResourceAlreadyExists) {
			return fmt.Errorf("cannot insert task comment external link: %w", err)
		}

		return nil
	}

	if err != nil {
		return fmt.Errorf("cannot load task comment external link: %w", err)
	}

	existing.ExternalID = remote.ID
	existing.ContentHash = &hash

	if !remote.UpdatedAt.IsZero() {
		existing.RemoteUpdatedAt = &remote.UpdatedAt
	}

	existing.UpdatedAt = now

	if err := existing.Update(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot update task comment external link: %w", err)
	}

	return nil
}

func (s *Service) commentOwner(
	ctx context.Context,
	scope coredata.Scoper,
	link *coredata.TaskExternalLink,
	email string,
) (*gid.GID, error) {
	var ownerID *gid.GID

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			id, err := internal.ResolveCommentOwner(ctx, conn, scope, link, email)
			if err != nil {
				return err
			}

			ownerID = id

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return ownerID, nil
}

func (s *Service) commentAuthor(
	ctx context.Context,
	scope coredata.Scoper,
	ownerID *gid.GID,
) (linear.CommentAuthor, error) {
	if ownerID == nil {
		return linear.CommentAuthor{}, nil
	}

	var author linear.CommentAuthor

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			resolved, err := loadCommentAuthor(ctx, conn, scope, s.baseURL, *ownerID)
			if err != nil {
				return err
			}

			author = resolved

			return nil
		},
	)
	if err != nil {
		return linear.CommentAuthor{}, fmt.Errorf("cannot load comment author: %w", err)
	}

	return author, nil
}

func loadCommentAuthor(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	baseURL string,
	ownerID gid.GID,
) (linear.CommentAuthor, error) {
	profile := &coredata.MembershipProfile{}
	if err := profile.LoadByID(ctx, conn, scope, ownerID); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return linear.CommentAuthor{}, nil
		}

		return linear.CommentAuthor{}, fmt.Errorf("cannot load comment author profile: %w", err)
	}

	name := strings.TrimSpace(profile.FullName)
	if name == "" {
		return linear.CommentAuthor{}, nil
	}

	author := linear.CommentAuthor{Name: name}

	identity := &coredata.Identity{}
	if err := identity.LoadByID(ctx, conn, profile.IdentityID); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return author, nil
		}

		return linear.CommentAuthor{}, fmt.Errorf("cannot load comment author identity: %w", err)
	}

	if identity.AvatarFileID == nil {
		return author, nil
	}

	file := &coredata.File{}
	if err := file.LoadPublicByID(ctx, conn, *identity.AvatarFileID); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return author, nil
		}

		return linear.CommentAuthor{}, fmt.Errorf("cannot load comment author avatar: %w", err)
	}

	avatarURL, err := publicAvatarURL(baseURL, file.ID)
	if err != nil {
		return linear.CommentAuthor{}, fmt.Errorf("cannot resolve comment author avatar URL: %w", err)
	}

	author.AvatarURL = avatarURL

	return author, nil
}

func publicAvatarURL(baseURL string, fileID gid.GID) (string, error) {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		return "", nil
	}

	base, err := baseurl.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("cannot parse comment author avatar base URL: %w", err)
	}

	avatarURL, err := base.WithPath("/api/files/v1/public/" + fileID.String()).String()
	if err != nil {
		return "", fmt.Errorf("cannot build comment author avatar URL: %w", err)
	}

	return avatarURL, nil
}
