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

package attachment

import (
	"fmt"
	"net/url"
	"strings"

	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/prosemirror"
)

// Ref is a file id stored on an image or file node.
type Ref struct {
	ID    string
	Image bool
}

// Path is the stable authenticated URL path for a stored file.
func Path(fileID string) string {
	id := strings.TrimSpace(fileID)
	if id == "" {
		return ""
	}

	return "/api/files/v1/" + url.PathEscape(id)
}

// ParseID returns the file id stored in an attachment URL.
// The URL may be the path alone or an absolute http(s) URL with that path.
// Other schemes and protocol-relative URLs are ignored.
func ParseID(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", false
	}

	u, err := url.Parse(raw)
	if err != nil {
		return "", false
	}

	if !urlAllowed(u) {
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

func urlAllowed(u *url.URL) bool {
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

func fileIDFromPath(escaped string) (string, bool) {
	id, err := url.PathUnescape(escaped)
	if err != nil || id == "" || strings.Contains(id, "/") {
		return "", false
	}

	if _, err := gid.ParseGID(id); err != nil {
		return "", false
	}

	return id, true
}

// FileIDs returns the distinct file ids referenced by content.
func FileIDs(content string) ([]gid.GID, error) {
	if strings.TrimSpace(content) == "" {
		return nil, nil
	}

	node, err := prosemirror.Parse(content)
	if err != nil {
		return nil, fmt.Errorf("cannot parse content: %w", err)
	}

	refs := Referenced(node)
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

// Referenced returns file ids referenced by image and file nodes.
func Referenced(n prosemirror.Node) []Ref {
	var refs []Ref

	prosemirror.Walk(&n, func(node *prosemirror.Node) bool {
		switch node.Type {
		case prosemirror.NodeImage:
			attrs, err := node.ImageAttrs()
			if err == nil && attrs.FileID != nil {
				if id := strings.TrimSpace(*attrs.FileID); id != "" {
					refs = append(refs, Ref{ID: id, Image: true})
				}
			}
		case prosemirror.NodeFile:
			attrs, err := node.FileAttrs()
			if err == nil {
				if id := strings.TrimSpace(attrs.FileID); id != "" {
					refs = append(refs, Ref{ID: id, Image: false})
				}
			}
		}

		return true
	})

	return refs
}

func validFileID(id string) bool {
	_, err := gid.ParseGID(id)

	return err == nil
}

func imageSrc(attrs prosemirror.ImageAttrs) string {
	if attrs.FileID != nil {
		if path := Path(*attrs.FileID); path != "" {
			return path
		}
	}

	return prosemirror.SafeImageSrc(attrs.Src)
}
