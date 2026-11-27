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

package scim

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/avatar"
	"go.probo.inc/probo/pkg/imageutil"
)

var (
	errUnsupportedPhoto = errors.New("unsupported image")
	errPhotoFetch       = errors.New("photo fetch failed")
	errPhotoUpload      = errors.New("photo upload failed")
)

// selectPhotoURI returns the primary photo value of a SCIM photos attribute,
// else the first one, or an empty string when there is none.
func selectPhotoURI(raw any) string {
	items, _ := raw.([]any)

	var first string

	for _, item := range items {
		photo, _ := item.(map[string]any)
		value, _ := photo["value"].(string)

		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}

		if primary, _ := photo["primary"].(bool); primary {
			return value
		}

		if first == "" {
			first = value
		}
	}

	return first
}

func (s *Service) incomingPhoto(ctx context.Context, uri string) (*coredata.File, error) {
	if uri == "" {
		return nil, nil
	}

	if s.fm == nil || s.bucket == "" {
		return nil, nil
	}

	body, err := s.readPhoto(ctx, uri)
	if err != nil {
		s.logger.WarnCtx(ctx, "skipping SCIM avatar", log.Error(err))

		return nil, nil
	}

	staged, err := s.stageProfilePhoto(ctx, body)
	if errors.Is(err, errPhotoUpload) {
		return nil, err
	}

	if err != nil {
		s.logger.WarnCtx(ctx, "skipping SCIM avatar", log.Error(err))

		return nil, nil
	}

	return staged, nil
}

func (s *Service) syncProfilePhoto(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	profile *coredata.MembershipProfile,
	staged *coredata.File,
) (bool, error) {
	if staged == nil {
		return false, nil
	}

	identity := &coredata.Identity{}
	if err := identity.LoadByID(ctx, tx, profile.IdentityID); err != nil {
		return false, fmt.Errorf("cannot load identity: %w", err)
	}

	photo := &coredata.SCIMProfilePhoto{}
	if err := photo.LoadByMembershipProfileID(ctx, tx, scope, profile.ID); err != nil {
		if !errors.Is(err, coredata.ErrResourceNotFound) {
			return false, fmt.Errorf("cannot load scim profile photo: %w", err)
		}

		now := time.Now()
		photo = &coredata.SCIMProfilePhoto{
			MembershipProfileID: profile.ID,
			OrganizationID:      profile.OrganizationID,
			CreatedAt:           now,
			UpdatedAt:           now,
		}
	}

	// Only a missing avatar or the one SCIM set is replaced. A picture the
	// person uploaded is kept.
	if identity.AvatarFileID != nil &&
		(photo.AvatarFileID == nil || *identity.AvatarFileID != *photo.AvatarFileID) {
		return false, nil
	}

	if err := s.attachProfilePhoto(ctx, tx, scope, identity, photo, staged); err != nil {
		return false, err
	}

	return true, nil
}

func (s *Service) discardStagedPhoto(ctx context.Context, staged *coredata.File) {
	if staged == nil || s.fm == nil {
		return
	}

	if err := s.fm.DeleteFile(ctx, staged); err != nil && s.logger != nil {
		s.logger.WarnCtx(
			ctx,
			"cannot delete unused SCIM avatar",
			log.String("file_id", staged.ID.String()),
			log.Error(err),
		)
	}
}

// stageProfilePhoto uploads before the transaction opens, so it only links
// the file.
func (s *Service) stageProfilePhoto(ctx context.Context, body []byte) (*coredata.File, error) {
	normalized, err := imageutil.Downscale(bytes.NewReader(body), avatar.MaxEdge)
	if err != nil {
		return nil, fmt.Errorf("cannot downscale photo: %w", errUnsupportedPhoto)
	}

	now := time.Now()
	fileID := gid.New(gid.NilTenant, coredata.FileEntityType)
	fileName := "avatar.jpg"

	if normalized.ContentType == imageutil.ContentTypePNG {
		fileName = "avatar.png"
	}

	avatarFile := &coredata.File{
		ID:             fileID,
		OrganizationID: gid.Nil,
		BucketName:     s.bucket,
		MimeType:       normalized.ContentType,
		FileName:       fileName,
		FileKey:        uuid.MustNewV7().String(),
		FileSize:       int64(len(normalized.Bytes)),
		Visibility:     coredata.FileVisibilityPublic,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	fileSize, err := s.fm.PutFile(
		ctx,
		avatarFile,
		bytes.NewReader(normalized.Bytes),
		map[string]string{
			"file-id": fileID.String(),
			"type":    "identity-avatar",
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot upload photo: %w", errPhotoUpload)
	}

	avatarFile.FileSize = fileSize

	return avatarFile, nil
}

func (s *Service) attachProfilePhoto(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	identity *coredata.Identity,
	photo *coredata.SCIMProfilePhoto,
	avatarFile *coredata.File,
) error {
	fileScope := coredata.NewScope(gid.NilTenant, gid.Nil)

	if err := avatarFile.Insert(ctx, tx, fileScope); err != nil {
		return fmt.Errorf("cannot insert avatar file: %w", err)
	}

	previousAvatarID := identity.AvatarFileID

	identity.AvatarFileID = &avatarFile.ID
	identity.UpdatedAt = time.Now()

	if err := identity.Update(ctx, tx); err != nil {
		return fmt.Errorf("cannot update identity avatar: %w", err)
	}

	if previousAvatarID != nil {
		previous := coredata.File{ID: *previousAvatarID}
		if err := previous.SoftDelete(ctx, tx, fileScope); err != nil {
			return fmt.Errorf("cannot soft-delete previous avatar file: %w", err)
		}
	}

	photo.AvatarFileID = &avatarFile.ID
	photo.UpdatedAt = time.Now()

	if err := photo.Upsert(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot save scim avatar: %w", err)
	}

	return nil
}
