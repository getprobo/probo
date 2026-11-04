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

package listscanresources

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const listScanResourcesQuery = `
query CookieBannerListScanResources($id: ID!, $first: Int, $after: CursorKey, $direction: OrderDirection!) {
  node(id: $id) {
    __typename
    ... on CookieBannerScan {
      trackerResources(
        first: $first
        after: $after
        orderBy: { field: CREATED_AT, direction: $direction }
      ) {
        totalCount
        edges {
          node {
            id
            type
            origin
            path
            createdAt
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

type resource struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Origin    string `json:"origin"`
	Path      string `json:"path"`
	CreatedAt string `json:"createdAt"`
}

func NewCmdListScanResources(f *cmdutil.Factory) *cobra.Command {
	var (
		flagLimit     int
		flagAfter     string
		flagDirection string
		flagOutput    *string
	)

	cmd := &cobra.Command{
		Use:     "list-scan-resources <scan-id>",
		Short:   "List resources from a cookie banner scan",
		Aliases: []string{"ls-scan-resources"},
		Args:    cobra.ExactArgs(1),
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

			resources, totalCount, nextCursor, err := api.Paginate(
				client,
				listScanResourcesQuery,
				map[string]any{
					"id":        args[0],
					"direction": flagDirection,
				},
				flagLimit,
				flagAfter,
				api.NodeConnectionOrMissing[resource]("trackerResources", "CookieBannerScan", fmt.Sprintf("cookie banner scan %s not found", args[0])),
			)
			if err != nil {
				return err
			}

			return cmdutil.PrintList(
				f.IOStreams,
				*flagOutput,
				resources,
				totalCount,
				cmdutil.ListView[resource]{
					Empty:      "No tracker resources found.",
					Noun:       "tracker resources",
					NextCursor: nextCursor,
					Headers:    []string{"ID", "TYPE", "ORIGIN", "PATH", "CREATED"},
					Row: func(item resource) ([]string, error) {
						return []string{
							item.ID,
							item.Type,
							item.Origin,
							item.Path,
							cmdutil.FormatTime(item.CreatedAt),
						}, nil
					},
				},
			)
		},
	}

	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of items")
	cmdutil.AddAfterFlag(cmd, &flagAfter)
	cmdutil.AddOrderDirectionFlag(cmd, &flagDirection, "DESC")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
