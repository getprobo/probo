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
	"fmt"
	"maps"
	"sort"
	"strings"
	"sync"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/attachment"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/prosemirror"
	"go.probo.inc/probo/pkg/task/sync/linear"
)

type (
	// Attachment is the bytes of a rich-text image or attachment.
	Attachment struct {
		Name     string
		MimeType string
		Body     []byte
	}

	// AttachmentReader loads a stored attachment for upload to Linear.
	AttachmentReader interface {
		ReadAttachment(
			ctx context.Context,
			scope coredata.Scoper,
			organizationID gid.GID,
			fileID gid.GID,
		) (Attachment, error)
	}

	linearAssetCache struct {
		mu   sync.Mutex
		urls map[string]string
	}
)

const linearAttachmentsKey = "attachments"

// embedLinearFiles replaces stored file URLs in markdown with Linear asset
// URLs. known maps a file id to an asset URL already stored for this task.
// The returned map includes known entries plus any file uploaded by this call.
// changed reports that the map gained an entry.
func (s *Service) embedLinearFiles(
	ctx context.Context,
	client *linear.Client,
	scope coredata.Scoper,
	connectorID gid.GID,
	organizationID gid.GID,
	markdown string,
	content string,
	known map[string]string,
) (string, map[string]string, bool, error) {
	ids, err := attachment.FileIDs(content)
	if err != nil {
		return "", nil, false, err
	}

	if len(ids) == 0 {
		return markdown, known, false, nil
	}

	if s.attachments == nil {
		return "", nil, false, fmt.Errorf("cannot upload attachment to Linear: storage is not configured")
	}

	files := make(map[string]string, len(known)+len(ids))
	maps.Copy(files, known)

	remote := markdown
	changed := false

	for _, id := range ids {
		fileID := id.String()
		assetURL := files[fileID]

		if assetURL == "" {
			assetURL = s.cachedLinearAsset(connectorID, fileID)
		}

		if assetURL == "" {
			assetURL, err = s.uploadLinearFile(ctx, client, scope, connectorID, organizationID, id)
			if err != nil {
				return "", nil, false, err
			}
		}

		if files[fileID] == "" {
			changed = true
		}

		files[fileID] = assetURL
		path := prosemirror.AttachmentPath(fileID)

		if path == "" || !strings.Contains(remote, path) {
			return "", nil, false, fmt.Errorf("cannot embed attachment %q: missing from markdown", fileID)
		}

		remote = strings.ReplaceAll(remote, path, assetURL)
	}

	return remote, files, changed, nil
}

func (s *Service) uploadLinearFile(
	ctx context.Context,
	client *linear.Client,
	scope coredata.Scoper,
	connectorID gid.GID,
	organizationID gid.GID,
	fileID gid.GID,
) (string, error) {
	file, err := s.attachments.ReadAttachment(ctx, scope, organizationID, fileID)
	if err != nil {
		return "", fmt.Errorf("cannot read attachment %q: %w", fileID, err)
	}

	assetURL, err := client.UploadFile(ctx, file.Name, file.MimeType, file.Body)
	if err != nil {
		return "", fmt.Errorf("cannot upload attachment %q to Linear: %w", fileID, err)
	}

	s.storeLinearAsset(connectorID, fileID.String(), assetURL)

	return assetURL, nil
}

func linearFilesMissing(known map[string]string, content string) (bool, error) {
	ids, err := attachment.FileIDs(content)
	if err != nil {
		return false, err
	}

	for _, id := range ids {
		if known[id.String()] == "" {
			return true, nil
		}
	}

	return false, nil
}

func (s *Service) linearAttachments(
	ctx context.Context,
	scope coredata.Scoper,
	taskID gid.GID,
) (map[string]string, error) {
	link := &coredata.TaskExternalLink{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := link.LoadByTaskID(ctx, conn, scope, taskID); err != nil {
				return fmt.Errorf("cannot load task external link: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load Linear attachments: %w", err)
	}

	return attachmentsFromMetadata(link.Metadata), nil
}

func (s *Service) saveLinearAttachments(
	ctx context.Context,
	scope coredata.Scoper,
	taskID gid.GID,
	files map[string]string,
) error {
	if len(files) == 0 {
		return nil
	}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			link := &coredata.TaskExternalLink{}
			if err := link.LoadByTaskIDForUpdate(ctx, tx, scope, taskID); err != nil {
				return fmt.Errorf("cannot load task external link: %w", err)
			}

			merged := attachmentsFromMetadata(link.Metadata)
			if merged == nil {
				merged = map[string]string{}
			}

			changed := false

			for id, assetURL := range files {
				if assetURL == "" || merged[id] == assetURL {
					continue
				}

				merged[id] = assetURL
				changed = true
			}

			if !changed {
				return nil
			}

			metadata, err := mergeAttachments(link.Metadata, merged)
			if err != nil {
				return err
			}

			link.Metadata = metadata
			link.UpdatedAt = time.Now()

			if err := link.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update task external link: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return fmt.Errorf("cannot store Linear attachments: %w", err)
	}

	return nil
}

func attachmentsFromMetadata(metadata json.RawMessage) map[string]string {
	if len(metadata) == 0 {
		return nil
	}

	var payload struct {
		Attachments map[string]string `json:"attachments"`
	}

	if err := json.Unmarshal(metadata, &payload); err != nil {
		return nil
	}

	return payload.Attachments
}

func mergeAttachments(metadata json.RawMessage, files map[string]string) (json.RawMessage, error) {
	payload := map[string]any{}

	if len(metadata) > 0 && string(metadata) != "null" {
		if err := json.Unmarshal(metadata, &payload); err != nil {
			return nil, fmt.Errorf("cannot read Linear link metadata: %w", err)
		}
	}

	payload[linearAttachmentsKey] = files

	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("cannot write Linear link metadata: %w", err)
	}

	return encoded, nil
}

// restoreLinearFiles rewrites Linear asset URLs back to stored file paths so
// an unchanged echo keeps the file id.
func restoreLinearFiles(markdown string, files map[string]string) string {
	if len(files) == 0 || markdown == "" {
		return markdown
	}

	type pair struct {
		asset string
		path  string
	}

	pairs := make([]pair, 0, len(files))

	for id, assetURL := range files {
		if id == "" || assetURL == "" {
			continue
		}

		pairs = append(pairs, pair{asset: assetURL, path: prosemirror.AttachmentPath(id)})
	}

	sort.Slice(pairs, func(i, j int) bool {
		return len(pairs[i].asset) > len(pairs[j].asset)
	})

	for _, item := range pairs {
		markdown = strings.ReplaceAll(markdown, item.asset, item.path)
	}

	return markdown
}

func (s *Service) cachedLinearAsset(connectorID gid.GID, fileID string) string {
	s.linearAssets.mu.Lock()
	defer s.linearAssets.mu.Unlock()

	if s.linearAssets.urls == nil {
		return ""
	}

	return s.linearAssets.urls[linearAssetKey(connectorID, fileID)]
}

func (s *Service) storeLinearAsset(connectorID gid.GID, fileID, assetURL string) {
	s.linearAssets.mu.Lock()
	defer s.linearAssets.mu.Unlock()

	if s.linearAssets.urls == nil {
		s.linearAssets.urls = map[string]string{}
	}

	s.linearAssets.urls[linearAssetKey(connectorID, fileID)] = assetURL
}

func linearAssetKey(connectorID gid.GID, fileID string) string {
	return connectorID.String() + "\x00" + fileID
}
