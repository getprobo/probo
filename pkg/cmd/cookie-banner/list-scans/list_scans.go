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

package listscans

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const listScansQuery = `
query($id: ID!, $first: Int, $after: CursorKey, $direction: OrderDirection!, $state: CookieBannerScanState) {
  node(id: $id) {
    __typename
    ... on CookieBanner {
      scans(
        first: $first
        after: $after
        orderBy: { field: CREATED_AT, direction: $direction }
        filter: { state: $state }
      ) {
        totalCount
        edges {
          node {
            id
            state
            pagesScanned
            cookiesFound
            storageFound
            resourcesFound
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

type scan struct {
	ID             string `json:"id"`
	State          string `json:"state"`
	PagesScanned   int    `json:"pagesScanned"`
	CookiesFound   int    `json:"cookiesFound"`
	StorageFound   int    `json:"storageFound"`
	ResourcesFound int    `json:"resourcesFound"`
	CreatedAt      string `json:"createdAt"`
}

func NewCmdListScans(f *cmdutil.Factory) *cobra.Command {
	var (
		flagLimit     int
		flagAfter     string
		flagDirection string
		flagState     string
		flagOutput    *string
	)

	cmd := &cobra.Command{
		Use:     "list-scans <id>",
		Short:   "List cookie banner scans",
		Aliases: []string{"ls-scans"},
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

			variables := map[string]any{
				"id":        args[0],
				"direction": flagDirection,
			}
			if flagState != "" {
				if err := cmdutil.ValidateEnum("state", flagState, []string{"PENDING", "RUNNING", "PAUSED", "COMPLETED", "FAILED"}); err != nil {
					return err
				}

				variables["state"] = flagState
			}

			scans, totalCount, nextCursor, err := api.Paginate(
				client,
				listScansQuery,
				variables,
				flagLimit,
				flagAfter,
				api.NodeConnectionOrMissing[scan]("scans", "CookieBanner", fmt.Sprintf("cookie banner %s not found", args[0])),
			)
			if err != nil {
				return err
			}

			return cmdutil.PrintList(
				f.IOStreams,
				*flagOutput,
				scans,
				totalCount,
				cmdutil.ListView[scan]{
					Empty:      "No cookie banner scans found.",
					Noun:       "scans",
					NextCursor: nextCursor,
					Headers:    []string{"ID", "STATE", "REQUESTED", "PAGES", "COOKIES", "STORAGE", "RESOURCES"},
					Row: func(s scan) ([]string, error) {
						return []string{
							s.ID,
							s.State,
							cmdutil.FormatTime(s.CreatedAt),
							fmt.Sprintf("%d", s.PagesScanned),
							fmt.Sprintf("%d", s.CookiesFound),
							fmt.Sprintf("%d", s.StorageFound),
							fmt.Sprintf("%d", s.ResourcesFound),
						}, nil
					},
				},
			)
		},
	}

	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of items")
	cmd.Flags().StringVar(&flagState, "state", "", "Filter by state (PENDING, RUNNING, PAUSED, COMPLETED, FAILED)")
	cmdutil.AddAfterFlag(cmd, &flagAfter)
	cmdutil.AddOrderDirectionFlag(cmd, &flagDirection, "DESC")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
