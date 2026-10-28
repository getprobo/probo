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
	"strings"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/prosemirror"
	"go.probo.inc/probo/pkg/task/sync/asana"
)

type asanaFileMeta struct {
	Files        map[string]string            `json:"asana_files,omitempty"`
	CommentFiles map[string]map[string]string `json:"asana_comment_files,omitempty"`
}

func readAsanaFiles(raw json.RawMessage) asanaFileMeta {
	var meta asanaFileMeta
	if len(raw) == 0 {
		return meta
	}

	_ = json.Unmarshal(raw, &meta)
	if meta.Files == nil {
		meta.Files = map[string]string{}
	}

	if meta.CommentFiles == nil {
		meta.CommentFiles = map[string]map[string]string{}
	}

	return meta
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

	for _, ref := range prosemirror.ReferencedFiles(node) {
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
		return nil, err
	}

	if known == nil {
		known = map[string]string{}
	}

	wanted := map[string]struct{}{}
	for _, id := range ids {
		wanted[id] = struct{}{}
		if known[id] != "" {
			continue
		}

		fileGID, err := gid.ParseGID(id)
		if err != nil {
			return nil, fmt.Errorf("cannot parse attachment id %q: %w", id, err)
		}

		file, err := s.attachments.ReadAttachment(ctx, scope, organizationID, fileGID)
		if err != nil {
			return nil, fmt.Errorf("cannot read attachment %q: %w", id, err)
		}

		uploaded, err := account.client.UploadAttachment(ctx, parentGID, file.Name, file.Body)
		if err != nil {
			return nil, fmt.Errorf("cannot upload attachment %q to Asana: %w", id, err)
		}

		known[id] = uploaded.GID
	}

	for id, attachmentGID := range known {
		if _, ok := wanted[id]; ok || attachmentGID == "" {
			continue
		}

		if err := account.client.DeleteAttachment(ctx, attachmentGID); err != nil && !asana.IsNotFound(err) {
			return nil, fmt.Errorf("cannot delete Asana attachment: %w", err)
		}

		delete(known, id)
	}

	return known, nil
}

func (s *Service) pullAsanaFiles(
	ctx context.Context,
	scope coredata.Scoper,
	account asanaAccount,
	organizationID gid.GID,
	content string,
	known map[string]string,
	attachments []asana.Attachment,
	parentGID string,
	parentType string,
) (string, map[string]string, error) {
	if known == nil {
		known = map[string]string{}
	}

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
			Attachment{Name: name, MimeType: mimeType, Body: body},
		)
		if err != nil {
			return content, known, fmt.Errorf("cannot store Asana attachment: %w", err)
		}

		next, err := appendFile(content, stored)
		if err != nil {
			return content, known, err
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

	if len(removed) == 0 {
		return content, known, nil
	}

	next, err := dropFiles(content, removed)
	if err != nil {
		return content, known, err
	}

	return next, known, nil
}

func attachmentParent(attachment asana.Attachment, parentGID, parentType string) bool {
	if attachment.ParentType == "" || attachment.ParentGID == "" {
		return parentType == "task"
	}

	return attachment.ParentType == parentType && attachment.ParentGID == parentGID
}

func appendFile(content string, stored StoredAttachment) (string, error) {
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
		return prosemirror.DefaultDocumentJSON(nil)
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

	content, err := prosemirror.DefaultDocumentJSON(&encodedJSON)
	if err != nil {
		return "", fmt.Errorf("cannot sanitize task content: %w", err)
	}

	return content, nil
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
