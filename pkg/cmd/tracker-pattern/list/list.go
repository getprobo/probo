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

package list

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const listQuery = `
query($id: ID!, $first: Int, $after: CursorKey) {
  node(id: $id) {
    __typename
    ... on CookieCategory {
      trackerPatterns(first: $first, after: $after) {
        totalCount
        edges {
          node {
            id
            pattern
            matchType
            trackerType
            source
            excluded
            commonTrackerPatternId
            attribution
          }
        }
        pageInfo {
          hasNextPage
          endCursor
        }
      }
    }
  }
}
`

type trackerPattern struct {
	ID                     string  `json:"id"`
	Pattern                string  `json:"pattern"`
	MatchType              string  `json:"matchType"`
	TrackerType            string  `json:"trackerType"`
	Source                 *string `json:"source"`
	Excluded               bool    `json:"excluded"`
	CommonTrackerPatternID *string `json:"commonTrackerPatternId"`
	Attribution            *string `json:"attribution"`
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagCategoryID string
		flagLimit      int
		flagAfter      string
		flagOutput     *string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List tracker patterns in a category",
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return err
			}

			if err := cmdutil.ValidateLimit(flagLimit); err != nil {
				return err
			}

			client, _, err := f.GraphQLClient(cmdutil.ConsoleEndpoint)
			if err != nil {
				return err
			}

			variables := map[string]any{"id": flagCategoryID}

			patterns, totalCount, nextCursor, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				flagAfter,
				api.NodeConnection[trackerPattern]("trackerPatterns", "CookieCategory", fmt.Sprintf("cookie category %s not found", flagCategoryID)),
			)
			if err != nil {
				return err
			}

			return cmdutil.PrintList(
				f.IOStreams,
				*flagOutput,
				patterns,
				totalCount,
				cmdutil.ListView[trackerPattern]{
					Empty:      "No tracker patterns found.",
					Noun:       "tracker patterns",
					NextCursor: nextCursor,
					Headers:    []string{"ID", "PATTERN", "MATCH TYPE", "TRACKER TYPE", "SOURCE", "EXCLUDED", "COMMON PATTERN ID", "ATTRIBUTION"},
					Row: func(p trackerPattern) ([]string, error) {
						excluded := ""
						if p.Excluded {
							excluded = "yes"
						}

						source := ""
						if p.Source != nil {
							source = *p.Source
						}

						commonPatternID := ""
						if p.CommonTrackerPatternID != nil {
							commonPatternID = *p.CommonTrackerPatternID
						}

						attribution := ""
						if p.Attribution != nil {
							attribution = *p.Attribution
						}

						return []string{p.ID, p.Pattern, p.MatchType, p.TrackerType, source, excluded, commonPatternID, attribution}, nil
					},
				},
			)
		},
	}

	cmd.Flags().StringVar(&flagCategoryID, "category-id", "", "Cookie category ID (required)")
	_ = cmd.MarkFlagRequired("category-id")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of items")
	cmdutil.AddAfterFlag(cmd, &flagAfter)
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
