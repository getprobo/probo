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

// Paginate fetches all pages of a connection up to limit items. The extract
// function pulls the Connection out of the raw GraphQL response data.
func Paginate[T any](
	client *Client,
	query string,
	variables map[string]any,
	limit int,
	extract func(json.RawMessage) (*Connection[T], error),
) ([]T, int, error) {
	vars := maps.Clone(variables)

	var (
		nodes      = make([]T, 0)
		totalCount int
	)

	for {
		remaining := limit - len(nodes)
		if remaining <= 0 {
			break
		}

		vars["first"] = remaining

		data, err := client.Do(query, vars)
		if err != nil {
			return nil, 0, err
		}

		conn, err := extract(data)
		if err != nil {
			return nil, 0, fmt.Errorf("cannot parse response: %w", err)
		}

		totalCount = conn.TotalCount
		for _, edge := range conn.Edges {
			nodes = append(nodes, edge.Node)
		}

		if !conn.PageInfo.HasNextPage || conn.PageInfo.EndCursor == nil {
			break
		}

		vars["after"] = *conn.PageInfo.EndCursor
	}

	return nodes, totalCount, nil
}
