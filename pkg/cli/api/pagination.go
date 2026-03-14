// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package api

import (
	"encoding/json"
	"fmt"
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
	var (
		nodes      []T
		totalCount int
	)

	for {
		remaining := limit - len(nodes)
		if remaining <= 0 {
			break
		}
		variables["first"] = remaining

		data, err := client.Do(query, variables)
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

		variables["after"] = *conn.PageInfo.EndCursor
	}

	delete(variables, "after")

	return nodes, totalCount, nil
}
