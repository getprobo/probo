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

package cmdutil

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/page"
)

// defaultPageSize is the forward page size used when neither --first nor
// --last is provided.
const defaultPageSize = 50

type (
	// PageFlags holds the cursor-pagination flag values registered by
	// AddPageFlags. They mirror the GraphQL connection arguments: --first
	// with --after walks forward, --last with --before walks backward.
	PageFlags struct {
		First  int
		Last   int
		After  string
		Before string
	}

	// PageInfo is the proboctl-side counterpart of the API's PageInfo. The
	// cursors are the base64 page.CursorKey scalars, so they round-trip
	// straight back into --after / --before.
	PageInfo struct {
		HasNextPage     bool    `json:"hasNextPage"`
		HasPreviousPage bool    `json:"hasPreviousPage"`
		StartCursor     *string `json:"startCursor"`
		EndCursor       *string `json:"endCursor"`
	}

	// PageOutput wraps a flat slice of results with its page info for JSON
	// output.
	PageOutput struct {
		Items    any      `json:"items"`
		PageInfo PageInfo `json:"pageInfo"`
	}
)

// AddPageFlags registers the cursor-pagination flags on cmd and returns a
// pointer to the bound values.
func AddPageFlags(cmd *cobra.Command) *PageFlags {
	pf := &PageFlags{}

	cmd.Flags().IntVar(&pf.First, "first", 0, fmt.Sprintf("Return the first N rows after --after (default %d)", defaultPageSize))
	cmd.Flags().IntVar(&pf.Last, "last", 0, fmt.Sprintf("Return the last N rows before --before (default %d)", defaultPageSize))
	cmd.Flags().StringVar(&pf.After, "after", "", "Cursor to page forward from (use with --first)")
	cmd.Flags().StringVar(&pf.Before, "before", "", "Cursor to page backward from (use with --last)")

	return pf
}

// NewCursorFromFlags builds a keyset cursor from the pagination flags,
// mirroring the GraphQL API's first/after vs last/before semantics.
func NewCursorFromFlags[F page.OrderField](
	pf *PageFlags,
	orderBy page.OrderBy[F],
) (*page.Cursor[F], error) {
	if pf.First > 0 && pf.Last > 0 {
		return nil, fmt.Errorf("--first and --last are mutually exclusive")
	}

	if pf.After != "" && pf.Before != "" {
		return nil, fmt.Errorf("--after and --before are mutually exclusive")
	}

	if pf.After != "" && pf.Last > 0 {
		return nil, fmt.Errorf("--after cannot be combined with --last")
	}

	if pf.Before != "" && pf.First > 0 {
		return nil, fmt.Errorf("--before cannot be combined with --first")
	}

	var (
		size     int
		from     *page.CursorKey
		position page.Position
	)

	switch {
	case pf.Last > 0 || pf.Before != "":
		size = pf.Last
		if size == 0 {
			size = defaultPageSize
		}

		position = page.Tail

		if pf.Before != "" {
			ck, err := page.ParseCursorKey(pf.Before)
			if err != nil {
				return nil, fmt.Errorf("invalid --before cursor: %w", err)
			}

			from = &ck
		}
	default:
		size = pf.First
		if size == 0 {
			size = defaultPageSize
		}

		position = page.Head

		if pf.After != "" {
			ck, err := page.ParseCursorKey(pf.After)
			if err != nil {
				return nil, fmt.Errorf("invalid --after cursor: %w", err)
			}

			from = &ck
		}
	}

	return page.NewCursor(size, from, position, orderBy), nil
}

// FetchPage loads a single keyset page using cursor and wraps the rows into a
// page.Page, which trims the over-fetch and computes HasNext / HasPrev. It is
// the proboctl-side counterpart of the API's connection resolvers.
func FetchPage[E page.Paginable[F], F page.OrderField](
	ctx context.Context,
	cursor *page.Cursor[F],
	load func(ctx context.Context, cursor *page.Cursor[F]) ([]E, error),
) (*page.Page[E, F], error) {
	rows, err := load(ctx, cursor)
	if err != nil {
		return nil, err
	}

	return page.NewPage(rows, cursor), nil
}

// NewPageInfo derives the proboctl PageInfo (cursors as base64 strings) from a
// loaded page.
func NewPageInfo[T page.Paginable[F], F page.OrderField](p *page.Page[T, F]) PageInfo {
	pi := PageInfo{
		HasNextPage:     p.Info.HasNext,
		HasPreviousPage: p.Info.HasPrev,
	}

	if len(p.Data) > 0 {
		start := p.First().CursorKey(p.Cursor.OrderBy.Field).String()
		end := p.Last().CursorKey(p.Cursor.OrderBy.Field).String()
		pi.StartCursor = &start
		pi.EndCursor = &end
	}

	return pi
}

// PrintPageInfo writes the page info as a footer below a rendered table.
func PrintPageInfo(out io.Writer, pi PageInfo) {
	var start, end string
	if pi.StartCursor != nil {
		start = *pi.StartCursor
	}

	if pi.EndCursor != nil {
		end = *pi.EndCursor
	}

	_, _ = fmt.Fprintf(
		out,
		"hasNextPage: %t  hasPreviousPage: %t\nstartCursor: %s\nendCursor: %s\n",
		pi.HasNextPage,
		pi.HasPreviousPage,
		start,
		end,
	)
}
