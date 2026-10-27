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

package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/cli/api"
)

type pageRow struct {
	ID string `json:"id"`
}

func TestPaginate(t *testing.T) {
	t.Parallel()

	extract := func(data json.RawMessage) (*api.Connection[pageRow], error) {
		var conn api.Connection[pageRow]
		if err := json.Unmarshal(data, &conn); err != nil {
			return nil, err
		}

		return &conn, nil
	}

	t.Run("returns the cursor for the next window", func(t *testing.T) {
		t.Parallel()

		var requests []map[string]any

		client := pageClient(t, &requests, map[string]string{
			"": `{
				"totalCount": 3,
				"edges": [{"node": {"id": "a"}}],
				"pageInfo": {"hasNextPage": true, "endCursor": "cur_b"}
			}`,
		})

		nodes, total, cursor, err := api.Paginate(client, "query", map[string]any{}, 1, "", extract)
		require.NoError(t, err)
		assert.Equal(t, []pageRow{{ID: "a"}}, nodes)
		assert.Equal(t, 3, total)
		assert.Equal(t, "cur_b", cursor)
		require.Len(t, requests, 1)
		assert.Equal(t, float64(1), requests[0]["first"])
		assert.NotContains(t, requests[0], "after")
	})

	t.Run("continues from after", func(t *testing.T) {
		t.Parallel()

		var requests []map[string]any

		client := pageClient(t, &requests, map[string]string{
			"cur_b": `{
				"totalCount": 3,
				"edges": [{"node": {"id": "b"}}, {"node": {"id": "c"}}],
				"pageInfo": {"hasNextPage": false, "endCursor": null}
			}`,
		})

		nodes, total, cursor, err := api.Paginate(client, "query", map[string]any{"id": "org_1"}, 2, "cur_b", extract)
		require.NoError(t, err)
		assert.Equal(t, []pageRow{{ID: "b"}, {ID: "c"}}, nodes)
		assert.Equal(t, 3, total)
		assert.Empty(t, cursor)
		require.Len(t, requests, 1)
		assert.Equal(t, "cur_b", requests[0]["after"])
		assert.Equal(t, "org_1", requests[0]["id"])
	})

	t.Run("fills the limit across server pages", func(t *testing.T) {
		t.Parallel()

		var requests []map[string]any

		client := pageClient(t, &requests, map[string]string{
			"": `{
				"totalCount": 4,
				"edges": [{"node": {"id": "a"}}, {"node": {"id": "b"}}],
				"pageInfo": {"hasNextPage": true, "endCursor": "cur_c"}
			}`,
			"cur_c": `{
				"totalCount": 4,
				"edges": [{"node": {"id": "c"}}],
				"pageInfo": {"hasNextPage": true, "endCursor": "cur_d"}
			}`,
		})

		nodes, total, cursor, err := api.Paginate(client, "query", map[string]any{}, 3, "", extract)
		require.NoError(t, err)
		assert.Equal(t, []pageRow{{ID: "a"}, {ID: "b"}, {ID: "c"}}, nodes)
		assert.Equal(t, 4, total)
		assert.Equal(t, "cur_d", cursor)
		require.Len(t, requests, 2)
		assert.Equal(t, float64(3), requests[0]["first"])
		assert.Equal(t, float64(1), requests[1]["first"])
		assert.Equal(t, "cur_c", requests[1]["after"])
	})

	t.Run("withholds the cursor when edges were left behind", func(t *testing.T) {
		t.Parallel()

		var requests []map[string]any

		client := pageClient(t, &requests, map[string]string{
			"": `{
				"totalCount": 5,
				"edges": [
					{"node": {"id": "a"}},
					{"node": {"id": "b"}},
					{"node": {"id": "c"}}
				],
				"pageInfo": {"hasNextPage": true, "endCursor": "cur_d"}
			}`,
		})

		nodes, total, cursor, err := api.Paginate(client, "query", map[string]any{}, 2, "", extract)
		require.NoError(t, err)
		assert.Equal(t, []pageRow{{ID: "a"}, {ID: "b"}}, nodes)
		assert.Equal(t, 5, total)
		assert.Empty(t, cursor)
		assert.Len(t, requests, 1)
	})

	t.Run("stops on an empty page", func(t *testing.T) {
		t.Parallel()

		var requests []map[string]any

		client := pageClient(t, &requests, map[string]string{
			"": `{
				"totalCount": 2,
				"edges": [{"node": {"id": "a"}}],
				"pageInfo": {"hasNextPage": true, "endCursor": "cur_b"}
			}`,
			"cur_b": `{
				"totalCount": 2,
				"edges": [],
				"pageInfo": {"hasNextPage": true, "endCursor": "cur_c"}
			}`,
		})

		nodes, total, cursor, err := api.Paginate(client, "query", map[string]any{}, 5, "", extract)
		require.NoError(t, err)
		assert.Equal(t, []pageRow{{ID: "a"}}, nodes)
		assert.Equal(t, 2, total)
		assert.Empty(t, cursor)
		assert.Len(t, requests, 2)
	})

	t.Run("non-positive limit does not call the server", func(t *testing.T) {
		t.Parallel()

		var requests []map[string]any

		client := pageClient(t, &requests, map[string]string{})

		nodes, total, cursor, err := api.Paginate(client, "query", map[string]any{}, 0, "cur_b", extract)
		require.NoError(t, err)
		assert.Empty(t, nodes)
		assert.Zero(t, total)
		assert.Empty(t, cursor)
		assert.Empty(t, requests)
	})
}

func pageClient(t *testing.T, requests *[]map[string]any, pages map[string]string) *api.Client {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Variables map[string]any `json:"variables"`
		}

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("cannot read request: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)

			return
		}

		*requests = append(*requests, body.Variables)

		after, _ := body.Variables["after"].(string)

		page, ok := pages[after]
		if !ok {
			t.Errorf("unexpected after %q", after)

			page = `{"totalCount":0,"edges":[],"pageInfo":{"hasNextPage":false}}`
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":` + page + `}`))
	}))
	t.Cleanup(server.Close)

	return api.NewClient(server.URL, "token", "", time.Second)
}
