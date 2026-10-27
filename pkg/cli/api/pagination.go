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

package api

import (
	"encoding/json"
	"fmt"
	"maps"
)

type (
	PageInfo struct {
		HasNextPage bool    `json:"hasNextPage"`
		EndCursor   *string `json:"endCursor"`
	}

	Edge[T any] struct {
		Node T `json:"node"`
	}

	Connection[T any] struct {
		TotalCount int       `json:"totalCount"`
		Edges      []Edge[T] `json:"edges"`
		PageInfo   PageInfo  `json:"pageInfo"`
	}
)

// Paginate collects up to limit items from a Relay connection.
// after is the cursor from a previous call. An empty after starts at the
// first page. The returned cursor continues the same list: pass it back as
// after with the same filters. It is empty when this window reaches the end,
// when a page has no edges, or when the server returned edges past limit.
// Those extra edges are omitted, and endCursor is withheld so they are not
// skipped. The extract function pulls the Connection out of the response data.
func Paginate[T any](
	client *Client,
	query string,
	variables map[string]any,
	limit int,
	after string,
	extract func(json.RawMessage) (*Connection[T], error),
) ([]T, int, string, error) {
	if limit <= 0 {
		return []T{}, 0, "", nil
	}

	vars := maps.Clone(variables)
	if after != "" {
		vars["after"] = after
	} else {
		delete(vars, "after")
	}

	var (
		nodes      = make([]T, 0)
		totalCount int
		nextCursor string
	)

	for len(nodes) < limit {
		vars["first"] = limit - len(nodes)

		data, err := client.Do(query, vars)
		if err != nil {
			return nil, 0, "", err
		}

		conn, err := extract(data)
		if err != nil {
			return nil, 0, "", fmt.Errorf("cannot parse response: %w", err)
		}

		totalCount = conn.TotalCount

		if len(conn.Edges) == 0 {
			nextCursor = ""

			break
		}

		take := min(len(conn.Edges), limit-len(nodes))

		for _, edge := range conn.Edges[:take] {
			nodes = append(nodes, edge.Node)
		}

		if take < len(conn.Edges) ||
			!conn.PageInfo.HasNextPage ||
			conn.PageInfo.EndCursor == nil ||
			*conn.PageInfo.EndCursor == "" {
			nextCursor = ""

			break
		}

		nextCursor = *conn.PageInfo.EndCursor
		vars["after"] = nextCursor
	}

	return nodes, totalCount, nextCursor, nil
}
