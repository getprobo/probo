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
	"encoding/json"
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
            displayName
            source
            excluded
            lastMatchedAt
            commonTrackerPatternId
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
	DisplayName            string  `json:"displayName"`
	Source                 *string `json:"source"`
	Excluded               bool    `json:"excluded"`
	LastMatchedAt          *string `json:"lastMatchedAt"`
	CommonTrackerPatternID *string `json:"commonTrackerPatternId"`
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagCategoryID string
		flagLimit      int
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

			cfg, err := f.Config()
			if err != nil {
				return err
			}

			host, hc, err := cfg.DefaultHost()
			if err != nil {
				return err
			}

			client := api.NewClient(
				host,
				hc.Token,
				"/api/console/v1/graphql",
				cfg.HTTPTimeoutDuration(),
				cmdutil.TokenRefreshOption(cfg, host, hc),
			)

			variables := map[string]any{"id": flagCategoryID}

			patterns, totalCount, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				func(data json.RawMessage) (*api.Connection[trackerPattern], error) {
					var resp struct {
						Node *struct {
							Typename        string                         `json:"__typename"`
							TrackerPatterns api.Connection[trackerPattern] `json:"trackerPatterns"`
						} `json:"node"`
					}
					if err := json.Unmarshal(data, &resp); err != nil {
						return nil, err
					}

					if resp.Node == nil {
						return nil, fmt.Errorf("cookie category %s not found", flagCategoryID)
					}

					if resp.Node.Typename != "CookieCategory" {
						return nil, fmt.Errorf("expected CookieCategory node, got %s", resp.Node.Typename)
					}

					return &resp.Node.TrackerPatterns, nil
				},
			)
			if err != nil {
				return err
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, patterns)
			}

			if len(patterns) == 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "No tracker patterns found.")
				return nil
			}

			rows := make([][]string, 0, len(patterns))
			for _, p := range patterns {
				excluded := ""
				if p.Excluded {
					excluded = "yes"
				}

				source := ""
				if p.Source != nil {
					source = *p.Source
				}

				lastMatched := ""
				if p.LastMatchedAt != nil {
					lastMatched = cmdutil.FormatTime(*p.LastMatchedAt)
				}

				commonPatternID := ""
				if p.CommonTrackerPatternID != nil {
					commonPatternID = *p.CommonTrackerPatternID
				}

				rows = append(rows, []string{p.ID, p.Pattern, p.MatchType, p.TrackerType, p.DisplayName, source, excluded, lastMatched, commonPatternID})
			}

			t := cmdutil.NewTable("ID", "PATTERN", "MATCH TYPE", "TRACKER TYPE", "DISPLAY NAME", "SOURCE", "EXCLUDED", "LAST MATCHED", "COMMON PATTERN ID").Rows(rows...)
			_, _ = fmt.Fprintln(f.IOStreams.Out, t)

			if totalCount > len(patterns) {
				_, _ = fmt.Fprintf(f.IOStreams.ErrOut, "\nShowing %d of %d tracker patterns\n", len(patterns), totalCount)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flagCategoryID, "category-id", "", "Cookie category ID (required)")
	_ = cmd.MarkFlagRequired("category-id")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of items")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
