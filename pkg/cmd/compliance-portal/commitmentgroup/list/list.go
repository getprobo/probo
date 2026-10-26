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
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: CompliancePortalCommitmentGroupOrder) {
  node(id: $id) {
    __typename
    ... on Organization {
      trustCenter {
        commitmentGroups(first: $first, after: $after, orderBy: $orderBy) {
          totalCount
          edges {
            node {
              id
              title
              description
              rank
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
}
`

type commitmentGroup struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Rank        int    `json:"rank"`
	CreatedAt   string `json:"createdAt"`
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg      string
		flagLimit    int
		flagOrderBy  string
		flagOrderDir string
		flagOutput   *string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List compliance portal commitment groups",
		Aliases: []string{"ls"},
		Example: `  # List commitment groups in the default organization
  prb trust-center commitment-group list

  # List commitment groups sorted by rank
  prb trust-center cg ls --order-by RANK`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return err
			}

			client, hc, err := f.GraphQLClient(cmdutil.ConsoleEndpoint)
			if err != nil {
				return err
			}

			flagOrg, err = cmdutil.ResolveOrganization(flagOrg, hc.Organization)
			if err != nil {
				return err
			}

			variables := map[string]any{
				"id": flagOrg,
			}

			if flagOrderBy != "" {
				if err := cmdutil.ValidateEnum("order-by", flagOrderBy, []string{"RANK", "CREATED_AT", "UPDATED_AT"}); err != nil {
					return err
				}

				variables["orderBy"] = map[string]any{
					"field":     flagOrderBy,
					"direction": flagOrderDir,
				}
			}

			groups, totalCount, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				api.NodeNestedConnection[commitmentGroup](
					"trustCenter",
					"commitmentGroups",
					"Organization",
					fmt.Sprintf("organization %s not found", flagOrg),
					fmt.Sprintf("trust center not found for organization %s", flagOrg),
				),
			)
			if err != nil {
				return err
			}

			return cmdutil.PrintList(
				f.IOStreams,
				*flagOutput,
				groups,
				totalCount,
				cmdutil.ListView[commitmentGroup]{
					Empty:   "No commitment groups found.",
					Noun:    "commitment groups",
					Headers: []string{"ID", "TITLE", "RANK"},
					Row: func(g commitmentGroup) ([]string, error) {
						return []string{
							g.ID,
							g.Title,
							fmt.Sprintf("%d", g.Rank),
						}, nil
					},
				},
			)
		},
	}

	cmdutil.AddOrganizationFlag(cmd, &flagOrg, "Organization ID")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of commitment groups to list")
	cmd.Flags().StringVar(&flagOrderBy, "order-by", "", "Order by field (RANK, CREATED_AT, UPDATED_AT)")
	cmd.Flags().StringVar(&flagOrderDir, "order-direction", "DESC", "Sort direction (ASC, DESC)")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
