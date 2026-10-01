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

package task

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/filevalidation"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/validator"
)

const (
	// MaxPictureBytes is the largest attachment a client may upload onto a task.
	// A picture is one kind of attachment; documents and other supported files
	// share the same limit.
	MaxPictureBytes = 10 << 20

	pendingPictureLimit = 20
)

var pictureValidator = filevalidation.NewValidator(
	filevalidation.WithCategories(
		filevalidation.CategoryDocument,
		filevalidation.CategorySpreadsheet,
		filevalidation.CategoryPresentation,
		filevalidation.CategoryText,
		filevalidation.CategoryImage,
		filevalidation.CategoryData,
		filevalidation.CategoryVideo,
	),
	filevalidation.WithMaxFileSize(MaxPictureBytes),
)

type UploadTaskPictureRequest struct {
	TaskID      gid.GID
	FileName    string
	ContentType string
	Content     io.Reader
}

func (s *Service) UploadPicture(
	ctx context.Context,
	scope coredata.Scoper,
	req UploadTaskPictureRequest,
) (*coredata.TaskPicture, error) {
	if s.files == nil || s.bucket == "" {
		return nil, fmt.Errorf("cannot upload picture: file storage is not configured")
	}

	body, err := io.ReadAll(io.LimitReader(req.Content, MaxPictureBytes+1))
	if err != nil {
		return nil, fmt.Errorf("cannot read picture: %w", err)
	}

	if err := validatePicture(req.TaskID, req.FileName, req.ContentType, body); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	var task coredata.Task

	err = s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := task.LoadByID(ctx, conn, scope, req.TaskID); err != nil {
				return fmt.Errorf("cannot load task: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	fileID := gid.New(scope.GetTenantID(), coredata.FileEntityType)
	pictureID := gid.New(scope.GetTenantID(), coredata.TaskPictureEntityType)

	objectKey, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("cannot generate object key: %w", err)
	}

	file := &coredata.File{
		ID:             fileID,
		OrganizationID: task.OrganizationID,
		BucketName:     s.bucket,
		MimeType:       req.ContentType,
		FileName:       req.FileName,
		FileKey:        objectKey.String(),
		FileSize:       int64(len(body)),
		Visibility:     coredata.FileVisibilityPrivate,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if _, err := s.files.PutFile(
		ctx,
		file,
		bytes.NewReader(body),
		map[string]string{
			"type":            "task-picture",
			"task-id":         task.ID.String(),
			"organization-id": task.OrganizationID.String(),
		},
	); err != nil {
		return nil, fmt.Errorf("cannot store picture: %w", err)
	}

	picture := &coredata.TaskPicture{
		ID:             pictureID,
		OrganizationID: task.OrganizationID,
		TaskID:         task.ID,
		FileID:         file.ID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	err = s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := file.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert picture file: %w", err)
			}

			if err := picture.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert task picture: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	uploaded, err := s.Sync.UploadPictureToLinkedIssue(
		ctx,
		scope,
		task.ID,
		req.FileName,
		req.ContentType,
		body,
	)
	if err != nil {
		s.discardPicture(ctx, scope, picture, file)

		return nil, err
	}

	if uploaded == nil {
		return picture, nil
	}

	picture.LinearAssetURL = &uploaded.AssetURL
	picture.LinearAttachmentID = &uploaded.AttachmentID
	picture.UpdatedAt = time.Now()

	err = s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := picture.UpdateLinear(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot store Linear picture attachment: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		if removeErr := s.Sync.RemoveLinkedPicture(ctx, scope, task.ID, uploaded.AttachmentID); removeErr != nil && s.logger != nil {
			s.logger.ErrorCtx(
				ctx,
				"cannot delete Linear picture after failed save",
				log.Error(removeErr),
			)
		}

		s.discardPicture(ctx, scope, picture, file)

		return nil, err
	}

	return picture, nil
}

func (s *Service) GetPicture(
	ctx context.Context,
	scope coredata.Scoper,
	pictureID gid.GID,
) (*coredata.TaskPicture, error) {
	var picture coredata.TaskPicture

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := picture.LoadByID(ctx, conn, scope, pictureID); err != nil {
				return fmt.Errorf("cannot load task picture: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &picture, nil
}

func (s *Service) ListPicturesForTaskID(
	ctx context.Context,
	scope coredata.Scoper,
	taskID gid.GID,
	cursor *page.Cursor[coredata.TaskPictureOrderField],
) (*page.Page[*coredata.TaskPicture, coredata.TaskPictureOrderField], error) {
	var pictures coredata.TaskPictures

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := pictures.LoadByTaskID(ctx, conn, scope, taskID, cursor); err != nil {
				return fmt.Errorf("cannot list task pictures: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(pictures, cursor), nil
}

func (s *Service) CountPicturesForTaskID(
	ctx context.Context,
	scope coredata.Scoper,
	taskID gid.GID,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var pictures coredata.TaskPictures

			loaded, err := pictures.CountByTaskID(ctx, conn, scope, taskID)
			if err != nil {
				return fmt.Errorf("cannot count task pictures: %w", err)
			}

			count = loaded

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Service) DeletePicture(
	ctx context.Context,
	scope coredata.Scoper,
	pictureID gid.GID,
) error {
	picture, err := s.GetPicture(ctx, scope, pictureID)
	if err != nil {
		return err
	}

	attachmentID := ""
	if picture.LinearAttachmentID != nil {
		attachmentID = *picture.LinearAttachmentID
	}

	if err := s.Sync.RemoveLinkedPicture(ctx, scope, picture.TaskID, attachmentID); err != nil {
		return err
	}

	file := &coredata.File{ID: picture.FileID}

	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := picture.Delete(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot delete task picture: %w", err)
			}

			if err := file.SoftDelete(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot delete picture file: %w", err)
			}

			return nil
		},
	)
}

// PublishToLinear publishes the task and then attaches any pictures that were
// uploaded before the Linear issue existed. A picture that cannot be attached
// stays on the task so a later publish can retry it.
func (s *Service) PublishToLinear(
	ctx context.Context,
	scope coredata.Scoper,
	taskID gid.GID,
	teamID string,
) (*coredata.TaskExternalLink, error) {
	link, err := s.Sync.PublishToLinear(ctx, scope, taskID, teamID)
	if err != nil {
		return nil, err
	}

	if err := s.pushPendingPictures(ctx, scope, taskID); err != nil && s.logger != nil {
		s.logger.ErrorCtx(
			ctx,
			"cannot attach pending pictures to Linear issue",
			log.Error(err),
		)
	}

	return link, nil
}

func (s *Service) pushPendingPictures(
	ctx context.Context,
	scope coredata.Scoper,
	taskID gid.GID,
) error {
	if s.files == nil {
		return nil
	}

	var pictures coredata.TaskPictures

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := pictures.LoadUnsyncedByTaskID(ctx, conn, scope, taskID, pendingPictureLimit); err != nil {
				return fmt.Errorf("cannot load pending pictures: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return err
	}

	var pushErr error

	for _, picture := range pictures {
		if err := s.pushPicture(ctx, scope, picture); err != nil && pushErr == nil {
			pushErr = err
		}
	}

	return pushErr
}

func (s *Service) pushPicture(
	ctx context.Context,
	scope coredata.Scoper,
	picture *coredata.TaskPicture,
) error {
	var file coredata.File

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := file.LoadActiveByID(ctx, conn, scope, picture.FileID); err != nil {
				return fmt.Errorf("cannot load picture file: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return err
	}

	body, err := s.files.GetFileBytes(ctx, &file)
	if err != nil {
		return fmt.Errorf("cannot read picture file: %w", err)
	}

	uploaded, err := s.Sync.UploadPictureToLinkedIssue(
		ctx,
		scope,
		picture.TaskID,
		file.FileName,
		file.MimeType,
		body,
	)
	if err != nil {
		return err
	}

	if uploaded == nil {
		return nil
	}

	picture.LinearAssetURL = &uploaded.AssetURL
	picture.LinearAttachmentID = &uploaded.AttachmentID
	picture.UpdatedAt = time.Now()

	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := picture.UpdateLinear(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot store Linear picture attachment: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) discardPicture(
	ctx context.Context,
	scope coredata.Scoper,
	picture *coredata.TaskPicture,
	file *coredata.File,
) {
	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := picture.Delete(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot delete task picture: %w", err)
			}

			if err := file.SoftDelete(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot delete picture file: %w", err)
			}

			return nil
		},
	)
	if err != nil && s.logger != nil {
		s.logger.ErrorCtx(ctx, "cannot discard picture after failed Linear upload", log.Error(err))
	}
}

func validatePicture(taskID gid.GID, filename, contentType string, body []byte) error {
	v := validator.New()
	v.Check(taskID, "task_id", validator.Required(), validator.GID(coredata.TaskEntityType))
	v.Check(filename, "file_name", validator.Required())
	v.Check(contentType, "content_type", validator.Required())
	v.Check(len(body), "file", validator.Min(1), validator.Max(MaxPictureBytes))

	if len(body) > 0 && len(body) <= MaxPictureBytes {
		if err := pictureValidator.Validate(filename, contentType, int64(len(body))); err != nil {
			v.Check(
				filename,
				"file",
				func(any) *validator.ValidationError {
					return &validator.ValidationError{
						Code:    validator.ErrorCodeInvalidFormat,
						Message: "must be a supported file type",
					}
				},
			)
		}
	}

	return v.Error()
}
