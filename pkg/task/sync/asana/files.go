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
	"maps"
	"strings"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/prosemirror"
	"go.probo.inc/probo/pkg/richtext"
	"go.probo.inc/probo/pkg/richtext/attachment"
	client "go.probo.inc/probo/pkg/task/sync/asana/client"
	"go.probo.inc/probo/pkg/task/sync/job"
)

type asanaFileMeta struct {
	Files        map[string]string            `json:"asana_files,omitempty"`
	CommentFiles map[string]map[string]string `json:"asana_comment_files,omitempty"`
}

func readAsanaFiles(raw json.RawMessage) (asanaFileMeta, error) {
	var meta asanaFileMeta
	if len(raw) == 0 {
		return meta, nil
	}

	if err := json.Unmarshal(raw, &meta); err != nil {
		return asanaFileMeta{}, fmt.Errorf("cannot read Asana file metadata: %w", err)
	}

	if meta.Files == nil {
		meta.Files = map[string]string{}
	}

	if meta.CommentFiles == nil {
		meta.CommentFiles = map[string]map[string]string{}
	}

	return meta, nil
}

func writeAsanaFiles(meta asanaFileMeta) (json.RawMessage, error) {
	if meta.Files == nil {
		meta.Files = map[string]string{}
	}

	if meta.CommentFiles == nil {
		meta.CommentFiles = map[string]map[string]string{}
	}

	encoded, err := json.Marshal(meta)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal Asana file metadata: %w", err)
	}

	return encoded, nil
}

func fileIDs(content string) ([]string, error) {
	if strings.TrimSpace(content) == "" {
		return nil, nil
	}

	node, err := prosemirror.Parse(content)
	if err != nil {
		return nil, fmt.Errorf("cannot parse task content: %w", err)
	}

	ids := make([]string, 0)

	for _, ref := range attachment.Referenced(node) {
		if ref.ID != "" {
			ids = append(ids, ref.ID)
		}
	}

	return ids, nil
}

func (s *Service) pushAsanaFiles(
	ctx context.Context,
	scope coredata.Scoper,
	account asanaAccount,
	organizationID gid.GID,
	parentGID string,
	content string,
	known map[string]string,
) (map[string]string, error) {
	ids, err := fileIDs(content)
	if err != nil {
		return nil, fmt.Errorf("cannot list task attachments: %w", err)
	}

	if known == nil {
		known = map[string]string{}
	}

	uploaded := make([]string, 0)
	wanted := map[string]struct{}{}

	for _, id := range ids {
		wanted[id] = struct{}{}
		if known[id] != "" {
			continue
		}

		fileGID, err := gid.ParseGID(id)
		if err != nil {
			return nil, rollbackAsanaUploads(
				ctx,
				account,
				uploaded,
				fmt.Errorf("cannot parse attachment id %q: %w", id, err),
			)
		}

		file, err := s.attachments.ReadAttachment(ctx, scope, organizationID, fileGID)
		if err != nil {
			return nil, rollbackAsanaUploads(
				ctx,
				account,
				uploaded,
				fmt.Errorf("cannot read attachment %q: %w", id, err),
			)
		}

		created, err := account.client.UploadAttachment(ctx, parentGID, file.Name, file.Body)
		if err != nil {
			return nil, rollbackAsanaUploads(
				ctx,
				account,
				uploaded,
				fmt.Errorf("cannot upload attachment %q to Asana: %w", id, err),
			)
		}

		known[id] = created.GID
		uploaded = append(uploaded, created.GID)
	}

	for id, attachmentGID := range known {
		if _, ok := wanted[id]; ok || attachmentGID == "" {
			continue
		}

		err := account.client.DeleteAttachment(ctx, attachmentGID)
		if err != nil && !client.IsNotFound(err) {
			return nil, rollbackAsanaUploads(
				ctx,
				account,
				uploaded,
				fmt.Errorf("cannot delete Asana attachment: %w", err),
			)
		}

		delete(known, id)
	}

	return known, nil
}

func rollbackAsanaUploads(
	ctx context.Context,
	account asanaAccount,
	uploaded []string,
	cause error,
) error {
	var deleteErr error

	for _, attachmentGID := range uploaded {
		err := account.client.DeleteAttachment(ctx, attachmentGID)
		if err != nil && !client.IsNotFound(err) {
			deleteErr = errors.Join(deleteErr, fmt.Errorf("cannot delete Asana attachment: %w", err))
		}
	}

	if deleteErr == nil {
		return fmt.Errorf("cannot upload Asana attachments: %w", cause)
	}

	return fmt.Errorf("%w: cannot delete uploaded Asana attachments: %w", cause, deleteErr)
}

func (s *Service) pullAsanaFiles(
	ctx context.Context,
	scope coredata.Scoper,
	account asanaAccount,
	organizationID gid.GID,
	content string,
	previous string,
	known map[string]string,
	attachments []client.Attachment,
	parentGID string,
	parentType string,
) (string, map[string]string, error) {
	known = copyStringMap(known)

	byAttachment := map[string]string{}
	for fileID, attachmentGID := range known {
		byAttachment[attachmentGID] = fileID
	}

	seen := map[string]struct{}{}

	for _, attachment := range attachments {
		if !attachmentParent(attachment, parentGID, parentType) {
			continue
		}

		seen[attachment.GID] = struct{}{}
		if _, ok := byAttachment[attachment.GID]; ok {
			continue
		}

		if s.stores == nil || attachment.DownloadURL == "" {
			continue
		}

		body, mimeType, err := account.client.Download(ctx, attachment.DownloadURL)
		if err != nil {
			return content, known, fmt.Errorf("cannot download Asana attachment: %w", err)
		}

		if mimeType == "" {
			mimeType = "application/octet-stream"
		}

		name := attachment.Name
		if name == "" {
			name = "file"
		}

		stored, err := s.stores.StoreAttachment(
			ctx,
			scope,
			organizationID,
			job.Attachment{Name: name, MimeType: mimeType, Body: body},
		)
		if err != nil {
			return content, known, fmt.Errorf("cannot store Asana attachment: %w", err)
		}

		next, err := appendFile(content, stored)
		if err != nil {
			return content, known, fmt.Errorf("cannot append Asana attachment: %w", err)
		}

		content = next
		known[stored.ID.String()] = attachment.GID
	}

	removed := map[string]struct{}{}

	for fileID, attachmentGID := range known {
		if _, ok := seen[attachmentGID]; ok {
			continue
		}

		removed[fileID] = struct{}{}
		delete(known, fileID)
	}

	restored, err := restoreKnownFileNodes(content, previous, known)
	if err != nil {
		return content, known, fmt.Errorf("cannot restore Asana attachments: %w", err)
	}

	content = restored

	if len(removed) == 0 {
		return content, known, nil
	}

	next, err := dropFiles(content, removed)
	if err != nil {
		return content, known, fmt.Errorf("cannot drop removed Asana attachments: %w", err)
	}

	return next, known, nil
}

func attachmentParent(attachment client.Attachment, parentGID, parentType string) bool {
	if attachment.ParentType == "" || attachment.ParentGID == "" {
		return parentType == "task"
	}

	return attachment.ParentType == parentType && attachment.ParentGID == parentGID
}

func appendFile(content string, stored job.Stored) (string, error) {
	node, err := prosemirror.Parse(content)
	if err != nil {
		return "", fmt.Errorf("cannot parse task content: %w", err)
	}

	raw, err := json.Marshal(prosemirror.FileAttrs{
		FileID:   stored.ID.String(),
		FileName: stored.Name,
		MimeType: stored.MimeType,
		Size:     stored.Size,
	})
	if err != nil {
		return "", fmt.Errorf("cannot marshal file node: %w", err)
	}

	node.Content = append(node.Content, prosemirror.Node{Type: prosemirror.NodeFile, Attrs: raw})

	return sanitizeNode(node)
}

func dropFiles(content string, removed map[string]struct{}) (string, error) {
	node, err := prosemirror.Parse(content)
	if err != nil {
		return "", fmt.Errorf("cannot parse task content: %w", err)
	}

	node.Content = dropFileNodes(node.Content, removed)
	if len(node.Content) == 0 {
		return richtext.DefaultDocumentJSON(nil)
	}

	return sanitizeNode(node)
}

func dropFileNodes(nodes []prosemirror.Node, removed map[string]struct{}) []prosemirror.Node {
	kept := make([]prosemirror.Node, 0, len(nodes))
	for _, node := range nodes {
		if node.Type == prosemirror.NodeFile || node.Type == prosemirror.NodeImage {
			id := ""

			if node.Type == prosemirror.NodeFile {
				attrs, err := node.FileAttrs()
				if err == nil {
					id = attrs.FileID
				}
			}

			if node.Type == prosemirror.NodeImage {
				attrs, err := node.ImageAttrs()
				if err == nil && attrs.FileID != nil {
					id = *attrs.FileID
				}
			}

			if _, ok := removed[id]; ok && id != "" {
				continue
			}
		}

		node.Content = dropFileNodes(node.Content, removed)
		kept = append(kept, node)
	}

	return kept
}

func sanitizeNode(node prosemirror.Node) (string, error) {
	encoded, err := json.Marshal(node)
	if err != nil {
		return "", fmt.Errorf("cannot marshal task content: %w", err)
	}

	encodedJSON := string(encoded)

	content, err := richtext.DefaultDocumentJSON(&encodedJSON)
	if err != nil {
		return "", fmt.Errorf("cannot sanitize task content: %w", err)
	}

	return content, nil
}

func restoreKnownFileNodes(content, previous string, kept map[string]string) (string, error) {
	if len(kept) == 0 || previous == "" || previous == content {
		return content, nil
	}

	current, err := prosemirror.Parse(content)
	if err != nil {
		return "", fmt.Errorf("cannot parse task content: %w", err)
	}

	previousNode, err := prosemirror.Parse(previous)
	if err != nil {
		return "", fmt.Errorf("cannot parse previous task content: %w", err)
	}

	present := map[string]struct{}{}
	collectFileIDs(current, present)

	missing := missingFileNodes(previousNode, kept, present)
	if len(missing) == 0 {
		return content, nil
	}

	current.Content = append(current.Content, missing...)

	restored, err := sanitizeNode(current)
	if err != nil {
		return "", fmt.Errorf("cannot restore Asana file nodes: %w", err)
	}

	return restored, nil
}

func missingFileNodes(node prosemirror.Node, kept map[string]string, present map[string]struct{}) []prosemirror.Node {
	var missing []prosemirror.Node

	collectMissingFileNodes(&missing, node, kept, present)

	return missing
}

func collectMissingFileNodes(
	missing *[]prosemirror.Node,
	node prosemirror.Node,
	kept map[string]string,
	present map[string]struct{},
) {
	if id, ok := referencedFileID(node); ok {
		if _, keep := kept[id]; keep {
			if _, already := present[id]; !already {
				*missing = append(*missing, node)
				present[id] = struct{}{}
			}
		}

		return
	}

	for _, child := range node.Content {
		collectMissingFileNodes(missing, child, kept, present)
	}
}

func collectFileIDs(node prosemirror.Node, ids map[string]struct{}) {
	if id, ok := referencedFileID(node); ok {
		ids[id] = struct{}{}

		return
	}

	for _, child := range node.Content {
		collectFileIDs(child, ids)
	}
}

func referencedFileID(node prosemirror.Node) (string, bool) {
	switch node.Type {
	case prosemirror.NodeFile:
		attrs, err := node.FileAttrs()
		if err != nil || attrs.FileID == "" {
			return "", false
		}

		return attrs.FileID, true
	case prosemirror.NodeImage:
		attrs, err := node.ImageAttrs()
		if err != nil || attrs.FileID == nil || *attrs.FileID == "" {
			return "", false
		}

		return *attrs.FileID, true
	default:
		return "", false
	}
}

func copyStringMap(known map[string]string) map[string]string {
	copied := make(map[string]string, len(known))

	maps.Copy(copied, known)

	return copied
}

func sameStringMap(left, right map[string]string) bool {
	if len(left) != len(right) {
		return false
	}

	for key, value := range left {
		if right[key] != value {
			return false
		}
	}

	return true
}
