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
      trackerResources(first: $first, after: $after) {
        totalCount
        edges {
          node {
            id
            type
            origin
            path
            displayName
            excluded
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

type trackerResource struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Origin      string `json:"origin"`
	Path        string `json:"path"`
	DisplayName string `json:"displayName"`
	Excluded    bool   `json:"excluded"`
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagCategoryID string
		flagLimit      int
		flagOutput     *string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List tracker resources in a category",
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return err
			}

			client, _, err := f.GraphQLClient(cmdutil.ConsoleEndpoint)
			if err != nil {
				return err
			}

			variables := map[string]any{"id": flagCategoryID}

			resources, totalCount, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				api.NodeConnection[trackerResource]("trackerResources", "CookieCategory", fmt.Sprintf("cookie category %s not found", flagCategoryID)),
			)
			if err != nil {
				return err
			}

			return cmdutil.PrintList(
				f.IOStreams,
				*flagOutput,
				resources,
				totalCount,
				cmdutil.ListView[trackerResource]{
					Empty:   "No tracker resources found.",
					Noun:    "tracker resources",
					Headers: []string{"ID", "TYPE", "ORIGIN", "PATH", "DISPLAY NAME", "EXCLUDED"},
					Row: func(r trackerResource) ([]string, error) {
						excluded := ""
						if r.Excluded {
							excluded = "yes"
						}

						return []string{r.ID, r.Type, r.Origin, r.Path, r.DisplayName, excluded}, nil
					},
				},
			)
		},
	}

	cmd.Flags().StringVar(&flagCategoryID, "category-id", "", "Cookie category ID (required)")
	_ = cmd.MarkFlagRequired("category-id")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of items")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
