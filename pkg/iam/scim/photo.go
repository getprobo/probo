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
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/imageutil"
)

const (
	// Same limits as identity avatar uploads.
	avatarMaxEdge     = 512
	maxAvatarFileSize = 5 << 20
	photoFetchTimeout = 15 * time.Second
)

type (
	scimPhoto struct {
		Remove bool
		ETag   string
		Body   []byte
	}

	photoAction string
)

const (
	photoActionNone  photoAction = ""
	photoActionClear photoAction = "clear"
	photoActionApply photoAction = "apply"
)

func photoInputFromMap(attributes map[string]any) (uri *string, etag *string) {
	for key, value := range attributes {
		switch strings.ToLower(key) {
		case "photos":
			selected, ok := selectPhotoURI(value)
			if !ok {
				continue
			}

			uri = &selected
		case "photoetag":
			text, ok := value.(string)
			if !ok {
				continue
			}

			etag = &text
		}
	}

	return uri, etag
}

func selectPhotoURI(raw any) (string, bool) {
	items, ok := photoMaps(raw)
	if !ok {
		return "", false
	}

	var (
		first   string
		typed   string
		primary string
	)

	for _, item := range items {
		value := strings.TrimSpace(stringValue(item["value"]))
		if value == "" {
			continue
		}

		if first == "" {
			first = value
		}

		if strings.EqualFold(stringValue(item["type"]), "photo") && typed == "" {
			typed = value
		}

		if boolish(item["primary"]) {
			primary = value
		}
	}

	switch {
	case primary != "":
		return primary, true
	case typed != "":
		return typed, true
	default:
		return first, true
	}
}

func photoMaps(raw any) ([]map[string]any, bool) {
	switch items := raw.(type) {
	case nil:
		return nil, true
	case []any:
		photos := make([]map[string]any, 0, len(items))
		for _, item := range items {
			photo, ok := item.(map[string]any)
			if !ok {
				return nil, false
			}

			photos = append(photos, photo)
		}

		return photos, true
	case []map[string]any:
		return items, true
	case map[string]any:
		return []map[string]any{items}, true
	default:
		return nil, false
	}
}

func stringValue(value any) string {
	text, _ := value.(string)

	return text
}

func boolish(value any) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		return strings.EqualFold(typed, "true")
	default:
		return false
	}
}

func photoActionFor(stored string, photo *scimPhoto) photoAction {
	if photo == nil {
		return photoActionNone
	}

	if photo.Remove {
		if stored == "" {
			return photoActionNone
		}

		return photoActionClear
	}

	if len(photo.Body) == 0 || photo.ETag == stored {
		return photoActionNone
	}

	return photoActionApply
}

func (s *Service) prepareSCIMPhoto(ctx context.Context, uri *string, etag *string) (*scimPhoto, string) {
	if uri == nil && etag == nil {
		return nil, ""
	}

	etagValue := ""
	if etag != nil {
		etagValue = strings.TrimSpace(*etag)
	}

	if uri != nil && strings.TrimSpace(*uri) == "" {
		return &scimPhoto{Remove: true}, ""
	}

	if uri == nil {
		if etagValue == "" {
			return &scimPhoto{Remove: true}, ""
		}

		return &scimPhoto{ETag: etagValue}, ""
	}

	value := strings.TrimSpace(*uri)
	if strings.HasPrefix(strings.ToLower(value), "data:") {
		body, reason := decodePhotoDataURI(value)
		if reason != "" {
			return nil, reason
		}

		if etagValue == "" {
			etagValue = photoHash(body)
		}

		return &scimPhoto{ETag: etagValue, Body: body}, ""
	}

	photoURL, reason := parsePhotoURL(value)
	if reason != "" {
		return nil, reason
	}

	if etagValue == "" {
		etagValue = photoURL.String()
	}

	body, reason := s.fetchPhoto(ctx, photoURL)
	if reason != "" {
		return nil, reason
	}

	return &scimPhoto{ETag: etagValue, Body: body}, ""
}

func decodePhotoDataURI(value string) ([]byte, string) {
	comma := strings.Index(value, ",")
	if comma < 0 {
		return nil, "unsupported image"
	}

	meta := value[:comma]
	payload := strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' || r == ' ' || r == '\t' {
			return -1
		}

		return r
	}, value[comma+1:])
	if !strings.Contains(strings.ToLower(meta), ";base64") {
		return nil, "unsupported image"
	}

	if len(payload) > maxAvatarFileSize*2 {
		return nil, "unsupported image"
	}

	body, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		body, err = base64.RawStdEncoding.DecodeString(payload)
	}

	if err != nil {
		return nil, "unsupported image"
	}

	if len(body) == 0 || len(body) > maxAvatarFileSize {
		return nil, "unsupported image"
	}

	return body, ""
}

func parsePhotoURL(value string) (*url.URL, string) {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" || parsed.User != nil || !strings.EqualFold(parsed.Scheme, "https") {
		return nil, "unsupported image"
	}

	return parsed, ""
}

func photoHash(body []byte) string {
	sum := sha256.Sum256(body)

	return hex.EncodeToString(sum[:])
}

func (s *Service) fetchPhoto(ctx context.Context, photoURL *url.URL) ([]byte, string) {
	if s.httpClient == nil {
		return nil, "photo fetch failed"
	}

	ctx, cancel := context.WithTimeout(ctx, photoFetchTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, photoURL.String(), nil)
	if err != nil {
		return nil, "photo fetch failed"
	}

	req.Header.Set("Accept", "image/jpeg, image/png, */*")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, "photo fetch failed"
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, "photo fetch failed"
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxAvatarFileSize+1))
	if err != nil || len(body) == 0 || len(body) > maxAvatarFileSize {
		return nil, "unsupported image"
	}

	return body, ""
}

func (s *Service) syncProfilePhoto(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	profile *coredata.MembershipProfile,
	photo *scimPhoto,
) error {
	stored := ""
	if profile.SCIMPhotoETag != nil {
		stored = *profile.SCIMPhotoETag
	}

	switch photoActionFor(stored, photo) {
	case photoActionNone:
		return nil
	case photoActionClear:
		return s.savePhotoETag(ctx, tx, scope, profile, nil)
	case photoActionApply:
		return s.applyProfilePhoto(ctx, tx, scope, profile, photo)
	default:
		return nil
	}
}

func (s *Service) applyProfilePhoto(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	profile *coredata.MembershipProfile,
	photo *scimPhoto,
) error {
	if s.fm == nil || s.bucket == "" {
		return nil
	}

	normalized, err := imageutil.Downscale(bytes.NewReader(photo.Body), avatarMaxEdge)
	if err != nil {
		s.warnPhoto(ctx, profile, "unsupported image")

		return s.savePhotoETag(ctx, tx, scope, profile, &photo.ETag)
	}

	identity := &coredata.Identity{}
	if err := identity.LoadByID(ctx, tx, profile.IdentityID); err != nil {
		return fmt.Errorf("cannot load identity: %w", err)
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
			"file-id":     fileID.String(),
			"identity-id": profile.IdentityID.String(),
			"type":        "identity-avatar",
		},
	)
	if err != nil {
		s.warnPhoto(ctx, profile, "photo upload failed")

		return nil
	}

	avatarFile.FileSize = fileSize
	fileScope := coredata.NewScope(gid.NilTenant, gid.Nil)
	if err := avatarFile.Insert(ctx, tx, fileScope); err != nil {
		return fmt.Errorf("cannot insert avatar file: %w", err)
	}

	if identity.AvatarFileID != nil {
		previous := coredata.File{ID: *identity.AvatarFileID}
		if err := previous.SoftDelete(ctx, tx, fileScope); err != nil {
			return fmt.Errorf("cannot soft-delete previous avatar file: %w", err)
		}
	}

	identity.AvatarFileID = &avatarFile.ID
	identity.UpdatedAt = now
	if err := identity.Update(ctx, tx); err != nil {
		return fmt.Errorf("cannot update identity avatar: %w", err)
	}

	return s.savePhotoETag(ctx, tx, scope, profile, &photo.ETag)
}

func (s *Service) savePhotoETag(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	profile *coredata.MembershipProfile,
	etag *string,
) error {
	next := ""
	if etag != nil {
		next = *etag
	}

	current := ""
	if profile.SCIMPhotoETag != nil {
		current = *profile.SCIMPhotoETag
	}

	if current == next {
		return nil
	}

	if next == "" {
		profile.SCIMPhotoETag = nil
	} else {
		value := next
		profile.SCIMPhotoETag = &value
	}

	profile.UpdatedAt = time.Now()
	if err := profile.Update(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot update profile photo etag: %w", err)
	}

	return nil
}

func (s *Service) warnPhoto(ctx context.Context, profile *coredata.MembershipProfile, reason string) {
	if s.logger == nil {
		return
	}

	identityID := ""
	if profile != nil {
		identityID = profile.IdentityID.String()
	}

	s.logger.WarnCtx(
		ctx,
		"skipping SCIM avatar",
		log.String("reason", reason),
		log.String("identity_id", identityID),
	)
}
