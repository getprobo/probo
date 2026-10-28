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
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"go.probo.inc/probo/pkg/prosemirror"
	"go.probo.inc/probo/pkg/task/sync/asana"
)

func contentToAsanaHTML(content string) (string, error) {
	return contentToAsanaHTMLWithMentions(content, nil, "")
}

func asanaHTMLToContent(htmlText string) (string, error) {
	htmlText = strings.TrimSpace(htmlText)
	if htmlText == "" {
		return prosemirror.DefaultDocumentJSON(nil)
	}

	node, err := prosemirror.ParseMarkdown(htmlText)
	if err != nil {
		return "", fmt.Errorf("cannot parse asana html: %w", err)
	}

	if len(node.Content) == 0 {
		return prosemirror.DefaultDocumentJSON(nil)
	}

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

func withoutAttachments(nodes []prosemirror.Node) []prosemirror.Node {
	if len(nodes) == 0 {
		return nil
	}

	kept := make([]prosemirror.Node, 0, len(nodes))
	for _, node := range nodes {
		if node.Type == prosemirror.NodeFile || node.Type == prosemirror.NodeImage {
			continue
		}

		node.Content = withoutAttachments(node.Content)
		kept = append(kept, node)
	}

	return kept
}

func asanaSnapshotHash(task asana.Task, sectionGID, optionGID string) string {
	sum := sha256.Sum256(
		[]byte(strings.Join(
			[]string{
				task.Name,
				task.HTMLNotes,
				strconv.FormatBool(task.Completed),
				task.DueOn,
				task.AssigneeGID,
				sectionGID,
				optionGID,
			},
			"\n",
		)),
	)

	return hex.EncodeToString(sum[:])
}

func commentHTML(authorName, content string) (string, error) {
	body, err := contentToAsanaHTML(content)
	if err != nil {
		return "", err
	}

	return prefixAsanaAuthor(authorName, body), nil
}

func htmlEscape(value string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&#34;",
	)

	return replacer.Replace(value)
}
