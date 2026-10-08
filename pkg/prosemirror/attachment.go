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

package prosemirror

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"go.probo.inc/probo/pkg/gid"
)

type (
	// AttachmentRef is a file id stored on an image or file node.
	AttachmentRef struct {
		ID    string
		Image bool
	}

	// StoredFile is the organization file record a content node points at.
	StoredFile struct {
		ID       string
		FileName string
		MimeType string
		Size     int64
	}
)

// AttachmentPath is the stable authenticated URL path for a stored file.
func AttachmentPath(fileID string) string {
	id := strings.TrimSpace(fileID)
	if id == "" {
		return ""
	}

	return "/api/files/v1/attachments/" + url.PathEscape(id)
}

// ParseAttachmentID returns the file id stored in an attachment URL.
// The URL may be the path alone or an absolute http(s) URL with that path.
func ParseAttachmentID(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", false
	}

	u, err := url.Parse(raw)
	if err != nil {
		return "", false
	}

	path := u.EscapedPath()

	const (
		attachmentPrefix = "/api/files/v1/attachments/"
		filePrefix       = "/api/files/v1/"
	)

	if after, ok := strings.CutPrefix(path, attachmentPrefix); ok {
		return fileIDFromPath(after)
	}

	if after, ok := strings.CutPrefix(path, filePrefix); ok {
		return fileIDFromPath(after)
	}

	return "", false
}

func fileIDFromPath(escaped string) (string, bool) {
	id, err := url.PathUnescape(escaped)
	if err != nil || id == "" || strings.Contains(id, "/") || !validFileID(id) {
		return "", false
	}

	if _, err := gid.ParseGID(id); err != nil {
		return "", false
	}

	return id, true
}

// promoteAttachmentBlocks turns a paragraph that is only a link to an uploaded
// file into a file block. That is how a file card round-trips through markdown.
func promoteAttachmentBlocks(nodes []Node) []Node {
	out := make([]Node, 0, len(nodes))

	for _, n := range nodes {
		if file, ok := attachmentFromParagraph(n); ok {
			out = append(out, file)

			continue
		}

		out = append(out, n)
	}

	return out
}

func attachmentFromParagraph(n Node) (Node, bool) {
	if n.Type != NodeParagraph || len(n.Content) == 0 {
		return Node{}, false
	}

	var name strings.Builder

	var fileID string

	for _, child := range n.Content {
		if child.Type == NodeText && child.Text != nil && len(child.Marks) == 0 && strings.TrimSpace(*child.Text) == "" {
			continue
		}

		if child.Type != NodeText || child.Text == nil || len(child.Marks) != 1 || child.Marks[0].Type != MarkLink {
			return Node{}, false
		}

		attrs, err := child.Marks[0].LinkAttrs()
		if err != nil {
			return Node{}, false
		}

		id, ok := ParseAttachmentID(attrs.Href)
		if !ok || (fileID != "" && id != fileID) {
			return Node{}, false
		}

		fileID = id

		name.WriteString(*child.Text)
	}

	if fileID == "" {
		return Node{}, false
	}

	fileName := strings.TrimSpace(name.String())
	if fileName == "" {
		fileName = "File"
	}

	if len(fileName) > 1024 {
		fileName = fileName[:1024]
	}

	raw, err := json.Marshal(FileAttrs{
		FileID:   fileID,
		FileName: fileName,
		MimeType: "application/octet-stream",
		Size:     0,
	})
	if err != nil {
		return Node{}, false
	}

	return Node{Type: NodeFile, Attrs: raw}, true
}

// ReferencedFiles returns file ids referenced by image and file nodes.
func ReferencedFiles(n Node) []AttachmentRef {
	var refs []AttachmentRef
	collectReferencedFiles(&refs, n)

	return refs
}

func collectReferencedFiles(refs *[]AttachmentRef, n Node) {
	switch n.Type {
	case NodeImage:
		attrs, err := n.ImageAttrs()
		if err == nil && attrs.FileID != nil {
			if id := strings.TrimSpace(*attrs.FileID); id != "" {
				*refs = append(*refs, AttachmentRef{ID: id, Image: true})
			}
		}
	case NodeFile:
		attrs, err := n.FileAttrs()
		if err == nil {
			if id := strings.TrimSpace(attrs.FileID); id != "" {
				*refs = append(*refs, AttachmentRef{ID: id, Image: false})
			}
		}
	}

	for _, child := range n.Content {
		collectReferencedFiles(refs, child)
	}
}

// BindStoredFiles rewrites image and file nodes from the organization file
// records. An image file id must belong to an image. A missing id is an error.
func BindStoredFiles(content string, files []StoredFile) (string, error) {
	node, err := Parse(content)
	if err != nil {
		return "", fmt.Errorf("cannot parse attachments: %w", err)
	}

	byID := make(map[string]StoredFile, len(files))
	for _, file := range files {
		byID[file.ID] = file
	}

	if err := bindStoredFiles(&node, byID); err != nil {
		return "", err
	}

	out, err := json.Marshal(node)
	if err != nil {
		return "", fmt.Errorf("cannot marshal attachments: %w", err)
	}

	return string(out), nil
}

func bindStoredFiles(n *Node, byID map[string]StoredFile) error {
	switch n.Type {
	case NodeImage:
		attrs, err := n.ImageAttrs()
		if err != nil {
			return fmt.Errorf("cannot bind image file: %w", err)
		}

		if attrs.FileID != nil {
			id := strings.TrimSpace(*attrs.FileID)

			file, ok := byID[id]
			if !ok {
				return fmt.Errorf("cannot bind image file %s: file not found", id)
			}

			if !isImageMime(file.MimeType) {
				return fmt.Errorf("cannot bind image file %s: not an image", id)
			}

			attrs.FileID = &file.ID
			attrs.Src = ""

			raw, err := json.Marshal(attrs)
			if err != nil {
				return fmt.Errorf("cannot marshal image attrs: %w", err)
			}

			n.Attrs = raw
		}
	case NodeFile:
		attrs, err := n.FileAttrs()
		if err != nil {
			return fmt.Errorf("cannot bind file: %w", err)
		}

		id := strings.TrimSpace(attrs.FileID)

		file, ok := byID[id]
		if !ok {
			return fmt.Errorf("cannot bind file %s: file not found", id)
		}

		attrs.FileID = file.ID
		attrs.FileName = file.FileName
		attrs.MimeType = file.MimeType
		attrs.Size = file.Size

		raw, err := json.Marshal(attrs)
		if err != nil {
			return fmt.Errorf("cannot marshal file attrs: %w", err)
		}

		n.Attrs = raw
	}

	for i := range n.Content {
		if err := bindStoredFiles(&n.Content[i], byID); err != nil {
			return err
		}
	}

	return nil
}

// DropRemoteImageSrcs clears http and https image sources. PDF export uses this
// so the renderer does not fetch arbitrary hosts. Uploaded images are already
// inlined as data URLs.
func DropRemoteImageSrcs(n *Node) {
	if n.Type == NodeImage {
		attrs, err := n.ImageAttrs()
		if err == nil && isRemoteImageSrc(attrs.Src) {
			attrs.Src = ""

			if raw, marshalErr := json.Marshal(attrs); marshalErr == nil {
				n.Attrs = raw
			}
		}
	}

	for i := range n.Content {
		DropRemoteImageSrcs(&n.Content[i])
	}
}

func isRemoteImageSrc(src string) bool {
	u, err := url.Parse(strings.TrimSpace(src))
	if err != nil || u.Scheme == "" {
		return false
	}

	switch strings.ToLower(u.Scheme) {
	case "http", "https":
		return true
	default:
		return false
	}
}

// EmbedImageDataURLs replaces uploaded images with in-memory data URLs so a
// PDF renderer can draw them without a browser session.
func EmbedImageDataURLs(n *Node, dataURLByFileID map[string]string) {
	if n.Type == NodeImage {
		attrs, err := n.ImageAttrs()
		if err == nil && attrs.FileID != nil {
			id := strings.TrimSpace(*attrs.FileID)
			if dataURL, ok := dataURLByFileID[id]; ok {
				attrs.FileID = nil
				attrs.Src = dataURL

				if raw, marshalErr := json.Marshal(attrs); marshalErr == nil {
					n.Attrs = raw
				}
			}
		}
	}

	for i := range n.Content {
		EmbedImageDataURLs(&n.Content[i], dataURLByFileID)
	}
}

func imageSrc(attrs ImageAttrs) string {
	if attrs.FileID != nil {
		if path := AttachmentPath(*attrs.FileID); path != "" {
			return path
		}
	}

	return safeImageSrc(attrs.Src)
}

func isImageMime(mime string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(mime)), "image/")
}
