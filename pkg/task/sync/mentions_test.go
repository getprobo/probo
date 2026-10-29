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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/prosemirror"
	"go.probo.inc/probo/pkg/task/sync/linear"
)

func TestMarkdownHasResolvableProfileLink(t *testing.T) {
	t.Parallel()

	profileURL := "https://linear.app/acme/profiles/jane"

	t.Run("profile link outside code is resolvable", func(t *testing.T) {
		t.Parallel()

		found, err := markdownHasResolvableProfileLink("Hi " + profileURL)
		require.NoError(t, err)
		assert.True(t, found)
	})

	t.Run("profile link inside a code span is not resolvable", func(t *testing.T) {
		t.Parallel()

		found, err := markdownHasResolvableProfileLink("See `" + profileURL + "`")
		require.NoError(t, err)
		assert.False(t, found)
	})

	t.Run("profile link inside a code block is not resolvable", func(t *testing.T) {
		t.Parallel()

		found, err := markdownHasResolvableProfileLink("```\n" + profileURL + "\n```")
		require.NoError(t, err)
		assert.False(t, found)
	})

	t.Run("an issue url is not resolvable", func(t *testing.T) {
		t.Parallel()

		found, err := markdownHasResolvableProfileLink("See https://linear.app/acme/issue/ENG-1/title")
		require.NoError(t, err)
		assert.False(t, found)
	})

	t.Run("text without a Linear link is not resolvable", func(t *testing.T) {
		t.Parallel()

		found, err := markdownHasResolvableProfileLink("Hello world")
		require.NoError(t, err)
		assert.False(t, found)
	})
}

func TestRewriteInboundMentionNodes_ProfileURLBecomesMention(t *testing.T) {
	t.Parallel()

	profileID := gid.New(gid.TenantID{}, 51)
	profileURL := "https://linear.app/acme/profiles/jane"

	t.Run("bare profile url", func(t *testing.T) {
		t.Parallel()

		content := rewriteTestMention(t, "Hi "+profileURL+" please", profileID, "Jane Doe")
		ids := mentionedIDs(t, content)
		assert.Equal(t, []gid.GID{profileID}, ids)

		markdown, err := ContentToMarkdown(content)
		require.NoError(t, err)
		assert.Contains(t, markdown, "@Jane Doe")
		assert.NotContains(t, markdown, "linear.app")
	})

	t.Run("markdown profile link", func(t *testing.T) {
		t.Parallel()

		content := rewriteTestMention(
			t,
			"Hi [Jane Doe]("+profileURL+") please",
			profileID,
			"Jane Doe",
		)
		assert.Equal(t, []gid.GID{profileID}, mentionedIDs(t, content))

		markdown, err := ContentToMarkdown(content)
		require.NoError(t, err)
		assert.Contains(t, markdown, "@Jane Doe")
		assert.NotContains(t, markdown, "linear.app")
	})

	t.Run("plain at name stays text", func(t *testing.T) {
		t.Parallel()

		content, err := MarkdownToContent("Hi @Jane Doe please")
		require.NoError(t, err)

		node, err := prosemirror.Parse(content)
		require.NoError(t, err)

		_, changed := rewriteInboundMentionNodes(node, func(string) (prosemirror.MentionAttrs, bool) {
			return prosemirror.MentionAttrs{ID: profileID.String(), Label: "Jane Doe"}, true
		})
		assert.False(t, changed)
		assert.Empty(t, mentionedIDs(t, content))
	})

	t.Run("issue url stays text", func(t *testing.T) {
		t.Parallel()

		content, err := MarkdownToContent("See https://linear.app/acme/issue/ENG-1/title")
		require.NoError(t, err)

		node, err := prosemirror.Parse(content)
		require.NoError(t, err)

		_, changed := rewriteInboundMentionNodes(node, func(string) (prosemirror.MentionAttrs, bool) {
			return prosemirror.MentionAttrs{ID: profileID.String(), Label: "Jane Doe"}, true
		})
		assert.False(t, changed)
	})

	t.Run("profile url inside code stays text", func(t *testing.T) {
		t.Parallel()

		content, err := MarkdownToContent("`" + profileURL + "`")
		require.NoError(t, err)

		node, err := prosemirror.Parse(content)
		require.NoError(t, err)

		_, changed := rewriteInboundMentionNodes(node, func(string) (prosemirror.MentionAttrs, bool) {
			return prosemirror.MentionAttrs{ID: profileID.String(), Label: "Jane Doe"}, true
		})
		assert.False(t, changed)
	})

	t.Run("unresolved profile url stays text", func(t *testing.T) {
		t.Parallel()

		content, err := MarkdownToContent("Hi " + profileURL)
		require.NoError(t, err)

		node, err := prosemirror.Parse(content)
		require.NoError(t, err)

		_, changed := rewriteInboundMentionNodes(node, func(string) (prosemirror.MentionAttrs, bool) {
			return prosemirror.MentionAttrs{}, false
		})
		assert.False(t, changed)
	})
}

func TestRewriteOutboundMentionNodes(t *testing.T) {
	t.Parallel()

	profileID := gid.New(gid.TenantID{}, 51)
	profileURL := "https://linear.app/acme/profiles/jane"
	content := mentionContent(t, profileID, "Jane Doe")

	t.Run("mention becomes a profile url", func(t *testing.T) {
		t.Parallel()

		markdown := outboundMentionMarkdown(t, content, profileURL)
		assert.Contains(t, markdown, profileURL)
		assert.NotContains(t, markdown, "@Jane Doe")
	})

	t.Run("unresolved mention stays plain", func(t *testing.T) {
		t.Parallel()

		markdown := outboundMentionMarkdown(t, content, "")
		assert.Contains(t, markdown, "@Jane Doe")
		assert.NotContains(t, markdown, "linear.app")
	})
}

func TestNewlyMentionedProfileIDs_SameDocumentOnly(t *testing.T) {
	t.Parallel()

	jane := gid.New(gid.TenantID{}, 51)
	ada := gid.New(gid.TenantID{}, 52)
	janeContent := mentionContent(t, jane, "Jane Doe")
	adaContent := mentionContent(t, ada, "Ada Lovelace")
	both := mentionContentWithIDs(t, jane, ada)

	t.Run("skips a profile already in this document", func(t *testing.T) {
		t.Parallel()

		ids, err := prosemirror.NewlyMentionedProfileIDs(janeContent, both)
		require.NoError(t, err)
		assert.Equal(t, []gid.GID{ada}, ids)
	})

	t.Run("notifies a profile that is new to this document", func(t *testing.T) {
		t.Parallel()

		ids, err := prosemirror.NewlyMentionedProfileIDs(adaContent, janeContent)
		require.NoError(t, err)
		assert.Equal(t, []gid.GID{jane}, ids)
	})

	t.Run("returns nothing when this document is unchanged", func(t *testing.T) {
		t.Parallel()

		ids, err := prosemirror.NewlyMentionedProfileIDs(janeContent, janeContent)
		require.NoError(t, err)
		assert.Empty(t, ids)
	})
}

func rewriteTestMention(t *testing.T, markdown string, profileID gid.GID, label string) string {
	t.Helper()

	content, err := MarkdownToContent(markdown)
	require.NoError(t, err)

	node, err := prosemirror.Parse(content)
	require.NoError(t, err)

	rewritten, changed := rewriteInboundMentionNodes(node, func(raw string) (prosemirror.MentionAttrs, bool) {
		if _, ok := linear.ProfileSlug(raw); !ok {
			return prosemirror.MentionAttrs{}, false
		}

		return prosemirror.MentionAttrs{ID: profileID.String(), Label: label}, true
	})
	require.True(t, changed)

	encoded, err := json.Marshal(rewritten)
	require.NoError(t, err)

	encodedJSON := string(encoded)
	sanitized, err := prosemirror.DefaultDocumentJSON(&encodedJSON)
	require.NoError(t, err)

	return sanitized
}

func outboundMentionMarkdown(t *testing.T, content string, profileURL string) string {
	t.Helper()

	node, err := prosemirror.Parse(content)
	require.NoError(t, err)

	rewritten := rewriteOutboundMentionNodes(node, func(_ string, _ string) (string, bool) {
		if profileURL == "" {
			return "", false
		}

		return profileURL, true
	})

	encoded, err := json.Marshal(rewritten)
	require.NoError(t, err)

	markdown, err := ContentToMarkdown(string(encoded))
	require.NoError(t, err)

	return markdown
}

func mentionContentWithIDs(t *testing.T, first gid.GID, second gid.GID) string {
	t.Helper()

	firstMention, err := newMentionNode(prosemirror.MentionAttrs{ID: first.String(), Label: "Jane Doe"})
	require.NoError(t, err)

	secondMention, err := newMentionNode(prosemirror.MentionAttrs{ID: second.String(), Label: "Ada Lovelace"})
	require.NoError(t, err)

	gap := " "
	document := prosemirror.Node{
		Type: prosemirror.NodeDoc,
		Content: []prosemirror.Node{
			{
				Type: prosemirror.NodeParagraph,
				Content: []prosemirror.Node{
					firstMention,
					{Type: prosemirror.NodeText, Text: &gap},
					secondMention,
				},
			},
		},
	}

	encoded, err := json.Marshal(document)
	require.NoError(t, err)

	encodedJSON := string(encoded)
	sanitized, err := prosemirror.DefaultDocumentJSON(&encodedJSON)
	require.NoError(t, err)

	return sanitized
}

func mentionContent(t *testing.T, profileID gid.GID, label string) string {
	t.Helper()

	mention, err := newMentionNode(prosemirror.MentionAttrs{ID: profileID.String(), Label: label})
	require.NoError(t, err)

	text := "Hi "
	document := prosemirror.Node{
		Type: prosemirror.NodeDoc,
		Content: []prosemirror.Node{
			{
				Type: prosemirror.NodeParagraph,
				Content: []prosemirror.Node{
					{Type: prosemirror.NodeText, Text: &text},
					mention,
				},
			},
		},
	}

	encoded, err := json.Marshal(document)
	require.NoError(t, err)

	encodedJSON := string(encoded)
	sanitized, err := prosemirror.DefaultDocumentJSON(&encodedJSON)
	require.NoError(t, err)

	return sanitized
}

func mentionedIDs(t *testing.T, content string) []gid.GID {
	t.Helper()

	ids, err := prosemirror.MentionedProfileIDs(content)
	require.NoError(t, err)

	return ids
}
