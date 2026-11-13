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

package client

import (
	"bytes"
	"encoding/json"
	"net/url"
	"path"
	"strings"
)

const descriptionWalkLimit = 64

type descriptionFile struct {
	URL   string
	Name  string
	Image bool
}

// MergeDescriptionFiles appends uploads.linear.app links found in a Linear
// issue's ProseMirror descriptionData so the importer can store those files.
func MergeDescriptionFiles(description string, descriptionData json.RawMessage) string {
	return appendDescriptionFiles(description, filesFromDescriptionData(descriptionData))
}

func appendMissingUploadLinks(description, extra string) string {
	if strings.TrimSpace(extra) == "" {
		return description
	}

	return appendDescriptionFiles(description, filesFromMarkdown(extra))
}

func appendDescriptionFiles(description string, files []descriptionFile) string {
	if len(files) == 0 {
		return description
	}

	present := map[string]struct{}{}
	for _, assetURL := range AssetURLs(description) {
		present[assetURL] = struct{}{}
	}

	var (
		b     strings.Builder
		added bool
		seen  = map[string]struct{}{}
	)

	b.WriteString(description)

	for _, file := range files {
		if file.URL == "" {
			continue
		}

		if _, ok := seen[file.URL]; ok {
			continue
		}

		if _, ok := present[file.URL]; ok {
			continue
		}

		seen[file.URL] = struct{}{}

		if b.Len() > 0 && !strings.HasSuffix(b.String(), "\n") {
			b.WriteByte('\n')
		}

		b.WriteString(markdownFileLink(file))
		b.WriteByte('\n')

		added = true
	}

	if !added {
		return description
	}

	return b.String()
}

func filesFromDescriptionData(raw json.RawMessage) []descriptionFile {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return nil
	}

	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil
	}

	if text, ok := value.(string); ok {
		text = strings.TrimSpace(text)
		if text == "" {
			return nil
		}

		var nested any
		if err := json.Unmarshal([]byte(text), &nested); err != nil {
			return filesFromMarkdown(text)
		}

		value = nested
	}

	var files []descriptionFile
	collectDescriptionFiles(value, 0, &files)

	return files
}

func collectDescriptionFiles(value any, depth int, files *[]descriptionFile) {
	if value == nil || depth > descriptionWalkLimit {
		return
	}

	switch node := value.(type) {
	case []any:
		for _, child := range node {
			collectDescriptionFiles(child, depth+1, files)
		}
	case map[string]any:
		if file, ok := descriptionFileFromNode(node); ok {
			*files = append(*files, file)
		}

		if content, ok := node["content"]; ok {
			collectDescriptionFiles(content, depth+1, files)
		}

		if marks, ok := node["marks"]; ok {
			collectDescriptionFiles(marks, depth+1, files)
		}
	}
}

func descriptionFileFromNode(node map[string]any) (descriptionFile, bool) {
	nodeType, _ := node["type"].(string)

	known, image := mediaNode(nodeType)
	if !known {
		return descriptionFile{}, false
	}

	attrs, _ := node["attrs"].(map[string]any)
	if attrs == nil {
		return descriptionFile{}, false
	}

	assetURL := firstString(attrs, "href", "src", "url")
	if !UploadURL(assetURL) {
		return descriptionFile{}, false
	}

	if !image && strings.HasPrefix(strings.ToLower(firstString(attrs, "mimetype", "mimeType", "contentType")), "image/") {
		image = true
	}

	name := sanitizeLabel(firstString(attrs, "name", "fileName", "filename", "alt", "title"))
	if name == "" {
		name = fileNameFromURL(assetURL)
	}

	return descriptionFile{
		URL:   assetURL,
		Name:  name,
		Image: image,
	}, true
}

func mediaNode(nodeType string) (bool, bool) {
	switch nodeType {
	case "image":
		return true, true
	case "file", "video", "audio", "embed", "link":
		return true, false
	default:
		return false, false
	}
}

func filesFromMarkdown(markdown string) []descriptionFile {
	urls := AssetURLs(markdown)
	files := make([]descriptionFile, 0, len(urls))

	for _, assetURL := range urls {
		files = append(files, descriptionFile{
			URL:   assetURL,
			Name:  uploadLabel(markdown, assetURL),
			Image: uploadLinkIsImage(markdown, assetURL),
		})
	}

	return files
}

// uploadURLTrailing is punctuation that can follow a bare upload URL in
// markdown. It is not a scan stop: dots also appear inside the host and the
// file extension.
const uploadURLTrailing = ".,;:!?_*~"

// AssetURLs returns uploads.linear.app URLs in markdown, in appearance order.
func AssetURLs(markdown string) []string {
	const marker = "https://uploads.linear.app/"

	var urls []string

	seen := map[string]struct{}{}
	rest := markdown

	for {
		index := strings.Index(rest, marker)
		if index < 0 {
			return urls
		}

		raw := rest[index:]
		end := 0

		for end < len(raw) {
			char := raw[end]
			if char <= ' ' || strings.ContainsRune(")>\"'],", rune(char)) {
				break
			}

			end++
		}

		candidate := raw[:end]

		rest = raw[end:]
		if len(rest) == 0 || rest[0] != ')' {
			candidate = strings.TrimRight(candidate, uploadURLTrailing)
		}

		if _, ok := seen[candidate]; ok || !UploadURL(candidate) {
			continue
		}

		seen[candidate] = struct{}{}
		urls = append(urls, candidate)
	}
}

// UploadURL reports whether raw is an uploads.linear.app file URL.
func UploadURL(raw string) bool {
	if strings.ContainsAny(raw, " \t\r\n)>\"'],") {
		return false
	}

	parsed, err := url.Parse(raw)
	if err != nil || parsed.User != nil {
		return false
	}

	if parsed.Scheme != "https" || !strings.EqualFold(parsed.Hostname(), "uploads.linear.app") {
		return false
	}

	return parsed.Path != "" && parsed.Path != "/"
}

// AssetLabel returns the markdown link text for assetURL.
func AssetLabel(markdown, assetURL string) string {
	needle := "](" + assetURL + ")"
	index := strings.Index(markdown, needle)

	if index < 0 {
		return ""
	}

	start := strings.LastIndex(markdown[:index], "[")
	if start < 0 || start+1 >= index {
		return ""
	}

	label := strings.TrimSpace(markdown[start+1 : index])
	if label == "" {
		return ""
	}

	return label
}

func uploadLabel(markdown, assetURL string) string {
	label := sanitizeLabel(AssetLabel(markdown, assetURL))
	if label == "" {
		return fileNameFromURL(assetURL)
	}

	return label
}

func uploadLinkIsImage(markdown, assetURL string) bool {
	needle := "](" + assetURL + ")"
	before, _, ok := strings.Cut(markdown, needle)

	if !ok {
		return false
	}

	start := strings.LastIndex(before, "[")

	return start > 0 && markdown[start-1] == '!'
}

func markdownFileLink(file descriptionFile) string {
	name := file.Name
	if name == "" {
		name = "file"
	}

	link := "[" + name + "](" + file.URL + ")"
	if file.Image {
		return "!" + link
	}

	return link
}

func fileNameFromURL(assetURL string) string {
	parsed, err := url.Parse(assetURL)
	if err != nil {
		return "file"
	}

	name := sanitizeLabel(path.Base(parsed.Path))
	if name == "" || name == "." || name == "/" {
		return "file"
	}

	return name
}

func sanitizeLabel(name string) string {
	name = strings.ReplaceAll(name, "\n", " ")
	name = strings.ReplaceAll(name, "\r", " ")
	name = strings.ReplaceAll(name, "[", "")
	name = strings.ReplaceAll(name, "]", "")
	name = strings.TrimSpace(name)

	if len(name) <= 1024 {
		return name
	}

	runes := []rune(name)
	if len(runes) > 200 {
		runes = runes[:200]
	}

	return strings.TrimSpace(string(runes))
}

func firstString(obj map[string]any, keys ...string) string {
	for _, key := range keys {
		value, ok := obj[key].(string)
		if !ok {
			continue
		}

		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}

	return ""
}
