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
	"unicode/utf8"

	"go.probo.inc/probo/pkg/gid"
)

// AttachmentRef is a file id stored on an image or file node.
type AttachmentRef struct {
	ID    string
	Image bool
}

// AttachmentPath is the stable authenticated URL path for a stored file.
func AttachmentPath(fileID string) string {
	id := strings.TrimSpace(fileID)
	if id == "" {
		return ""
	}

	return "/api/files/v1/" + url.PathEscape(id)
}

// ParentAttachmentPath is the authenticated URL for a file embedded in a parent.
func ParentAttachmentPath(parentID, fileID string) string {
	parent := strings.TrimSpace(parentID)
	file := strings.TrimSpace(fileID)

	if parent == "" || file == "" {
		return ""
	}

	return "/api/files/v1/attachments/" + url.PathEscape(parent) + "/" + url.PathEscape(file)
}

// ParseAttachmentID returns the file id stored in an attachment URL.
// The URL may be the path alone or an absolute http(s) URL with that path.
// Other schemes and protocol-relative URLs are ignored.
func ParseAttachmentID(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", false
	}

	u, err := url.Parse(raw)
	if err != nil {
		return "", false
	}

	if !attachmentURLAllowed(u) {
		return "", false
	}

	path := u.EscapedPath()

	const (
		attachmentPrefix = "/api/files/v1/attachments/"
		filePrefix       = "/api/files/v1/"
	)

	if after, ok := strings.CutPrefix(path, attachmentPrefix); ok {
		return attachmentFileID(after)
	}

	if after, ok := strings.CutPrefix(path, filePrefix); ok {
		return fileIDFromPath(after)
	}

	return "", false
}

func attachmentURLAllowed(u *url.URL) bool {
	scheme := strings.ToLower(u.Scheme)
	switch scheme {
	case "", "http", "https":
	default:
		return false
	}

	if u.Host != "" && scheme != "http" && scheme != "https" {
		return false
	}

	return true
}

func attachmentFileID(escaped string) (string, bool) {
	parent, file, ok := strings.Cut(escaped, "/")
	if !ok {
		return fileIDFromPath(escaped)
	}

	if strings.Contains(file, "/") {
		return "", false
	}

	if _, ok := fileIDFromPath(parent); !ok {
		return "", false
	}

	return fileIDFromPath(file)
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
			name.WriteString(*child.Text)

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

	if len(fileName) > MaxFileNameBytes {
		end := MaxFileNameBytes
		for end > 0 && !utf8.RuneStart(fileName[end]) {
			end--
		}

		fileName = fileName[:end]
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

// FileIDs returns the distinct file ids referenced by content.
func FileIDs(content string) ([]gid.GID, error) {
	if strings.TrimSpace(content) == "" {
		return nil, nil
	}

	node, err := Parse(content)
	if err != nil {
		return nil, fmt.Errorf("cannot parse content: %w", err)
	}

	refs := ReferencedFiles(node)
	ids := make([]gid.GID, 0, len(refs))
	seen := make(map[string]struct{}, len(refs))

	for _, ref := range refs {
		if _, ok := seen[ref.ID]; ok {
			continue
		}

		seen[ref.ID] = struct{}{}

		id, err := gid.ParseGID(ref.ID)
		if err != nil {
			return nil, fmt.Errorf("invalid file id %q", ref.ID)
		}

		ids = append(ids, id)
	}

	return ids, nil
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

// IsImageMime reports whether mime is an image media type.
func IsImageMime(mime string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(mime)), "image/")
}
