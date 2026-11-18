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

package search

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const searchQuery = `
query SearchResources(
  $organizationId: ID!
  $query: String!
  $first: Int
  $after: CursorKey
) {
  search(
    organizationId: $organizationId
    query: $query
    first: $first
    after: $after
  ) {
    totalCount
    edges {
      node {
        id
        __typename
      }
    }
    pageInfo {
      hasNextPage
      endCursor
    }
  }
}
`

type searchHit struct {
	ID       string `json:"id"`
	Typename string `json:"__typename"`
}

func NewCmdSearch(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg    string
		flagLimit  int
		flagAfter  string
		flagOutput *string
	)

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search resources with a query string",
		Long: `Search resources with a query string.

Qualifiers:
  tag:name        Resource must carry this tag. Repeat to require every tag (AND).
  type:name       Restrict results to this resource type, for example document or risk.
                  Repeat to accept several types (OR).

Quote a name that contains spaces, and escape a quote or backslash inside
a quoted value with a backslash.`,
		Example: `  # Find resources tagged Production
  prb search 'tag:Production'

  # Find documents that have every listed tag
  prb search 'tag:Production tag:Compliance type:document'`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return err
			}

			if err := cmdutil.ValidateLimit(flagLimit); err != nil {
				return err
			}

			query := strings.TrimSpace(args[0])
			if query == "" {
				return fmt.Errorf("query is required")
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

			if flagOrg == "" {
				flagOrg = hc.Organization
			}

			if f.IOStreams.IsInteractive() {
				if flagOrg == "" {
					err := huh.NewInput().
						Title("Organization ID").
						Value(&flagOrg).
						Run()
					if err != nil {
						return err
					}
				}
			}

			if flagOrg == "" {
				return fmt.Errorf("organization ID is required; pass --org or set a default with 'prb auth login'")
			}

			variables := map[string]any{
				"organizationId": flagOrg,
				"query":          query,
			}

			hits, totalCount, nextCursor, err := api.Paginate(
				client,
				searchQuery,
				variables,
				flagLimit,
				flagAfter,
				api.RootConnection[searchHit]("search"),
			)
			if err != nil {
				return err
			}

			return cmdutil.PrintList(
				f.IOStreams,
				*flagOutput,
				hits,
				totalCount,
				cmdutil.ListView[searchHit]{
					Empty:      "No resources found.",
					Noun:       "resources",
					NextCursor: nextCursor,
					Headers:    []string{"ID", "TYPE"},
					Row: func(hit searchHit) ([]string, error) {
						return []string{hit.ID, hit.Typename}, nil
					},
				},
			)
		},
	}

	cmdutil.AddOrganizationFlag(cmd, &flagOrg, "Organization ID")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of resources to list")
	cmd.Flags().StringVar(&flagAfter, "after", "", "Cursor to continue listing from")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
