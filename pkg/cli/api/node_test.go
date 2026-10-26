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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/cli/api"
)

func TestNodeConnection(t *testing.T) {
	t.Parallel()

	extract := api.NodeConnection[struct {
		ID string `json:"id"`
	}]("frameworks", "Organization", "organization org_1 not found")

	conn, err := extract(json.RawMessage(`{
		"node": {
			"__typename": "Organization",
			"frameworks": {
				"totalCount": 1,
				"edges": [{"node": {"id": "fw_1"}}],
				"pageInfo": {"hasNextPage": false}
			}
		}
	}`))
	require.NoError(t, err)
	require.Len(t, conn.Edges, 1)
	assert.Equal(t, "fw_1", conn.Edges[0].Node.ID)
	assert.Equal(t, 1, conn.TotalCount)

	_, err = extract(json.RawMessage(`{"node": null}`))
	assert.EqualError(t, err, "organization org_1 not found")

	_, err = extract(json.RawMessage(`{"node": {"__typename": "Framework", "frameworks": {}}}`))
	assert.EqualError(t, err, "expected Organization node, got Framework")
}

func TestNodeConnectionOrMissing(t *testing.T) {
	t.Parallel()

	extract := api.NodeConnectionOrMissing[struct {
		ID string `json:"id"`
	}]("risks", "Organization", "Organization org_1 not found")

	_, err := extract(json.RawMessage(`{"node": {"__typename": "Risk", "risks": {}}}`))
	assert.EqualError(t, err, "Organization org_1 not found")
}

func TestNodeNestedConnection(t *testing.T) {
	t.Parallel()

	extract := api.NodeNestedConnection[struct {
		ID string `json:"id"`
	}](
		"trustCenter",
		"commitmentGroups",
		"Organization",
		"organization org_1 not found",
		"trust center not found for organization org_1",
	)

	conn, err := extract(json.RawMessage(`{
		"node": {
			"__typename": "Organization",
			"trustCenter": {
				"commitmentGroups": {
					"totalCount": 1,
					"edges": [{"node": {"id": "grp_1"}}]
				}
			}
		}
	}`))
	require.NoError(t, err)
	require.Len(t, conn.Edges, 1)
	assert.Equal(t, "grp_1", conn.Edges[0].Node.ID)

	_, err = extract(json.RawMessage(`{"node": {"__typename": "Organization", "trustCenter": null}}`))
	assert.EqualError(t, err, "trust center not found for organization org_1")
}

func TestViewerConnection(t *testing.T) {
	t.Parallel()

	extract := api.ViewerConnection[struct {
		ID string `json:"id"`
	}]("profiles")

	conn, err := extract(json.RawMessage(`{"viewer": {"profiles": {"totalCount": 0, "edges": []}}}`))
	require.NoError(t, err)
	assert.Empty(t, conn.Edges)

	conn, err = extract(json.RawMessage(`{"viewer": null}`))
	require.NoError(t, err)
	assert.Empty(t, conn.Edges)
}

func TestRootConnection(t *testing.T) {
	t.Parallel()

	extract := api.RootConnection[struct {
		ID string `json:"id"`
	}]("commonGVLVendors")

	conn, err := extract(json.RawMessage(`{"commonGVLVendors": {"totalCount": 1, "edges": [{"node": {"id": "v_1"}}]}}`))
	require.NoError(t, err)
	require.Len(t, conn.Edges, 1)
	assert.Equal(t, "v_1", conn.Edges[0].Node.ID)
}

func TestNodeSlice(t *testing.T) {
	t.Parallel()

	items, err := api.NodeSlice[struct {
		ID string `json:"id"`
	}](
		json.RawMessage(`{"node": {"__typename": "Organization", "connectors": [{"id": "c_1"}]}}`),
		"connectors",
		"Organization",
		"organization org_1 not found",
	)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "c_1", items[0].ID)

	_, err = api.NodeSlice[struct {
		ID string `json:"id"`
	}](json.RawMessage(`{"node": null}`), "connectors", "Organization", "organization org_1 not found")
	assert.EqualError(t, err, "organization org_1 not found")
}
