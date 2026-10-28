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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/prosemirror"
	client "go.probo.inc/probo/pkg/task/sync/asana/client"
)

func TestContentToAsanaHTMLWrapsTheBody(t *testing.T) {
	t.Parallel()

	content := prosemirror.FromPlainText("Ship the sync")

	html, err := contentToAsanaHTML(content)
	require.NoError(t, err)
	assert.Equal(t, "<body>Ship the sync</body>", html)
}

func TestContentToAsanaHTMLWritesMentionsAsNames(t *testing.T) {
	t.Parallel()

	profileID := gid.New(gid.TenantID{}, 3).String()
	content := `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Hi "},{"type":"mention","attrs":{"id":"` + profileID + `","label":"Jane <Doe>"}}]}]}`

	html, err := contentToAsanaHTML(content)
	require.NoError(t, err)
	assert.Contains(t, html, "Hi @Jane &lt;Doe&gt;")
	assert.NotContains(t, html, "data-type")
}

func TestContentToAsanaHTMLWritesResolvedMentions(t *testing.T) {
	t.Parallel()

	profileID := gid.New(gid.TenantID{}, 3).String()
	content := `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Hi "},{"type":"mention","attrs":{"id":"` + profileID + `","label":"Jane"}}]}]}`

	html, err := contentToAsanaHTMLWithMentions(content, map[string]string{profileID: "99"}, "https://client.example.test")
	require.NoError(t, err)
	assert.Contains(t, html, `data-asana-gid="99"`)
	assert.Contains(t, html, `href="https://client.example.test/0/profile/99"`)
	assert.Contains(t, html, "@Jane")
	assert.NotContains(t, html, "data-type")
}

func TestAsanaMentionAnchorBecomesAProfileURL(t *testing.T) {
	t.Parallel()

	html := asanaMentionAnchorsToURLs(`<body>Hi <a data-asana-gid="99">Jane</a></body>`, "https://client.example.test")
	assert.Contains(t, html, "https://client.example.test/0/profile/99")
	assert.NotContains(t, html, "data-asana-gid")
}

func TestAsanaMentionAnchorsToURLs_SelfClosingDoesNotSwallowTheNextTag(t *testing.T) {
	t.Parallel()

	html := asanaMentionAnchorsToURLs(
		`<body><a data-asana-gid="99"/> kept <a href="https://example.test">link</a></body>`,
		"https://client.example.test",
	)
	assert.Contains(t, html, "https://client.example.test/0/profile/99")
	assert.Contains(t, html, "kept")
	assert.Contains(t, html, `href="https://example.test"`)
	assert.Contains(t, html, ">link</a>")
}

func TestAsanaPageOffsets_SingleAccountKeepsTheRawCursor(t *testing.T) {
	t.Parallel()

	offsets, err := asanaPageOffsets([]string{"account-a"}, "raw-offset")
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"account-a": "raw-offset"}, offsets)

	offsets, err = asanaPageOffsets([]string{"account-a"}, "")
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"account-a": ""}, offsets)
}

func TestAsanaPageOffsets_MultipleAccountsUsePerAccountCursors(t *testing.T) {
	t.Parallel()

	cursor, err := encodeAsanaAccountCursors(
		map[string]string{
			"account-b": "cursor-b",
			"account-a": "cursor-a",
		},
	)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(cursor, asanaAccountCursorPrefix))

	offsets, err := asanaPageOffsets([]string{"account-a", "account-b", "account-c"}, cursor)
	require.NoError(t, err)
	assert.Equal(
		t,
		map[string]string{
			"account-a": "cursor-a",
			"account-b": "cursor-b",
		},
		offsets,
	)
}

func TestAsanaPageOffsets_MultipleAccountsRejectASharedCursor(t *testing.T) {
	t.Parallel()

	_, err := asanaPageOffsets([]string{"account-a", "account-b"}, "shared-offset")
	require.Error(t, err)
}

func TestAsanaPageOffsets_FirstCombinedPageStartsEveryAccount(t *testing.T) {
	t.Parallel()

	offsets, err := asanaPageOffsets([]string{"account-a", "account-b"}, "")
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"account-a": "", "account-b": ""}, offsets)
}

func TestAsanaPageOffsets_EmptyOffsetStaysInTheCursor(t *testing.T) {
	t.Parallel()

	cursor, err := encodeAsanaAccountCursors(
		map[string]string{
			"account-a": "cursor-a",
			"account-b": "",
		},
	)
	require.NoError(t, err)

	offsets, err := asanaPageOffsets([]string{"account-a", "account-b", "account-c"}, cursor)
	require.NoError(t, err)
	assert.Equal(
		t,
		map[string]string{
			"account-a": "cursor-a",
			"account-b": "",
		},
		offsets,
	)
}

func TestListAsanaNamed_MultipleAccountsShareOnePageBudget(t *testing.T) {
	t.Parallel()

	idA := gid.New(gid.TenantID{}, 1)
	idB := gid.New(gid.TenantID{}, 1)
	idC := gid.New(gid.TenantID{}, 1)

	var calls []asanaNamedCall

	page, err := (&Service{}).listAsanaNamed(
		context.Background(),
		asanaNamedAccounts(idA, idB, idC),
		2,
		"",
		func(_ context.Context, account asanaAccount, limit int, offset string) (client.Page, error) {
			calls = append(calls, asanaNamedCall{
				connectorID: account.connector.ID.String(),
				limit:       limit,
				offset:      offset,
			})

			return client.Page{
				Items:       []client.Named{{GID: "a1"}, {GID: "a2"}},
				EndCursor:   "cursor-a",
				HasNextPage: true,
			}, nil
		},
	)
	require.NoError(t, err)
	require.Len(t, calls, 1)
	assert.Equal(t, asanaNamedCall{connectorID: idA.String(), limit: 2}, calls[0])
	assert.Equal(t, []client.Named{{GID: "a1"}, {GID: "a2"}}, page.Items)
	assert.True(t, page.HasNextPage)

	offsets, err := asanaPageOffsets([]string{idA.String(), idB.String(), idC.String()}, page.EndCursor)
	require.NoError(t, err)
	assert.Equal(
		t,
		map[string]string{
			idA.String(): "cursor-a",
			idB.String(): "",
			idC.String(): "",
		},
		offsets,
	)
}

func TestListAsanaNamed_PartialAccountContinuesWithinTheBudget(t *testing.T) {
	t.Parallel()

	idA := gid.New(gid.TenantID{}, 1)
	idB := gid.New(gid.TenantID{}, 1)
	idC := gid.New(gid.TenantID{}, 1)

	var calls []asanaNamedCall

	page, err := (&Service{}).listAsanaNamed(
		context.Background(),
		asanaNamedAccounts(idA, idB, idC),
		3,
		"",
		func(_ context.Context, account asanaAccount, limit int, offset string) (client.Page, error) {
			calls = append(calls, asanaNamedCall{
				connectorID: account.connector.ID.String(),
				limit:       limit,
				offset:      offset,
			})

			if account.connector.ID == idA {
				return client.Page{Items: []client.Named{{GID: "a1"}}}, nil
			}

			if account.connector.ID == idB {
				return client.Page{
					Items:       []client.Named{{GID: "b1"}, {GID: "b2"}},
					EndCursor:   "cursor-b",
					HasNextPage: true,
				}, nil
			}

			return client.Page{Items: []client.Named{{GID: "unexpected"}}}, nil
		},
	)
	require.NoError(t, err)
	require.Len(t, calls, 2)
	assert.Equal(t, asanaNamedCall{connectorID: idA.String(), limit: 3}, calls[0])
	assert.Equal(t, asanaNamedCall{connectorID: idB.String(), limit: 2}, calls[1])
	assert.Equal(t, []client.Named{{GID: "a1"}, {GID: "b1"}, {GID: "b2"}}, page.Items)

	offsets, err := asanaPageOffsets([]string{idA.String(), idB.String(), idC.String()}, page.EndCursor)
	require.NoError(t, err)
	assert.Equal(
		t,
		map[string]string{
			idB.String(): "cursor-b",
			idC.String(): "",
		},
		offsets,
	)
}

func TestListAsanaNamed_ResumesUnconsumedAccountsAtTheirOffset(t *testing.T) {
	t.Parallel()

	idA := gid.New(gid.TenantID{}, 1)
	idB := gid.New(gid.TenantID{}, 1)
	idC := gid.New(gid.TenantID{}, 1)

	cursor, err := encodeAsanaAccountCursors(
		map[string]string{
			idA.String(): "cursor-a",
			idB.String(): "",
			idC.String(): "cursor-c",
		},
	)
	require.NoError(t, err)

	var calls []asanaNamedCall

	page, err := (&Service{}).listAsanaNamed(
		context.Background(),
		asanaNamedAccounts(idA, idB, idC),
		2,
		cursor,
		func(_ context.Context, account asanaAccount, limit int, offset string) (client.Page, error) {
			calls = append(calls, asanaNamedCall{
				connectorID: account.connector.ID.String(),
				limit:       limit,
				offset:      offset,
			})

			if account.connector.ID == idA {
				return client.Page{Items: []client.Named{{GID: "a3"}}}, nil
			}

			return client.Page{
				Items:       []client.Named{{GID: "b1"}},
				EndCursor:   "cursor-b",
				HasNextPage: true,
			}, nil
		},
	)
	require.NoError(t, err)
	require.Len(t, calls, 2)
	assert.Equal(t, asanaNamedCall{connectorID: idA.String(), limit: 2, offset: "cursor-a"}, calls[0])
	assert.Equal(t, asanaNamedCall{connectorID: idB.String(), limit: 1}, calls[1])
	assert.Equal(t, []client.Named{{GID: "a3"}, {GID: "b1"}}, page.Items)

	offsets, err := asanaPageOffsets([]string{idA.String(), idB.String(), idC.String()}, page.EndCursor)
	require.NoError(t, err)
	assert.Equal(
		t,
		map[string]string{
			idB.String(): "cursor-b",
			idC.String(): "cursor-c",
		},
		offsets,
	)
}

func TestListAsanaNamed_SingleAccountKeepsTheRawCursor(t *testing.T) {
	t.Parallel()

	idA := gid.New(gid.TenantID{}, 1)

	var calls []asanaNamedCall

	page, err := (&Service{}).listAsanaNamed(
		context.Background(),
		asanaNamedAccounts(idA),
		20,
		"raw-in",
		func(_ context.Context, account asanaAccount, limit int, offset string) (client.Page, error) {
			calls = append(calls, asanaNamedCall{
				connectorID: account.connector.ID.String(),
				limit:       limit,
				offset:      offset,
			})

			return client.Page{
				Items:       []client.Named{{GID: "only"}},
				EndCursor:   "raw-out",
				HasNextPage: true,
			}, nil
		},
	)
	require.NoError(t, err)
	require.Len(t, calls, 1)
	assert.Equal(t, asanaNamedCall{connectorID: idA.String(), limit: 20, offset: "raw-in"}, calls[0])
	assert.Equal(t, "raw-out", page.EndCursor)
	assert.True(t, page.HasNextPage)
	assert.False(t, strings.HasPrefix(page.EndCursor, asanaAccountCursorPrefix))
}

type asanaNamedCall struct {
	connectorID string
	limit       int
	offset      string
}

func asanaNamedAccounts(ids ...gid.GID) []asanaAccount {
	accounts := make([]asanaAccount, 0, len(ids))
	for _, id := range ids {
		accounts = append(accounts, asanaAccount{connector: &coredata.Connector{ID: id}})
	}

	return accounts
}

func TestRewriteAsanaProfileURLBecomesAMention(t *testing.T) {
	t.Parallel()

	content, err := asanaHTMLToContent(asanaMentionAnchorsToURLs(
		`<body>Hi <a data-asana-gid="99">Jane</a></body>`,
		"https://client.example.test",
	))
	require.NoError(t, err)

	node, err := prosemirror.Parse(content)
	require.NoError(t, err)

	node, changed := rewriteAsanaMentionNodes(node, "https://client.example.test", func(userGID string) (prosemirror.MentionAttrs, bool) {
		if userGID != "99" {
			return prosemirror.MentionAttrs{}, false
		}

		return prosemirror.MentionAttrs{ID: "profile", Label: "Jane"}, true
	})
	require.True(t, changed)

	encoded, err := json.Marshal(node)
	require.NoError(t, err)
	assert.Contains(t, string(encoded), `"type":"mention"`)
	assert.Contains(t, string(encoded), `"label":"Jane"`)
}

func TestAsanaTaskURLUsesTheAPIHost(t *testing.T) {
	t.Parallel()

	taskURL, err := asanaTaskURL("https://client.example.test/api/1.0", "1200", "3400")
	require.NoError(t, err)
	assert.Equal(t, "https://client.example.test/0/1200/3400", taskURL)
}

func TestAsanaSnapshotHashIncludesDueAt(t *testing.T) {
	t.Parallel()

	without := client.Task{
		Name:      "Ship",
		HTMLNotes: "<body></body>",
		DueOn:     "2026-05-01",
	}
	withDueAt := without
	withDueAt.DueAt = "2026-05-01T15:00:00.000Z"

	assert.NotEqual(
		t,
		asanaSnapshotHash(without, "section", "option"),
		asanaSnapshotHash(withDueAt, "section", "option"),
	)
}

func TestCommentHTMLPrefixesTheAuthor(t *testing.T) {
	t.Parallel()

	content := prosemirror.FromPlainText("Looks good")

	html, err := commentHTML("Ada <Admin>", content)
	require.NoError(t, err)
	assert.Equal(t, "<body><strong>Ada &lt;Admin&gt;</strong>\nLooks good</body>", html)
}
