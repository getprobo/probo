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

package listprincipals

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const listQuery = `
query AccessInventoryPrincipalList(
  $id: ID!,
  $first: Int,
  $after: CursorKey,
  $orderBy: AccessInventoryPrincipalOrder,
  $filter: AccessInventoryPrincipalFilter
) {
  node(id: $id) {
    __typename
    ... on Organization {
      accessInventoryPrincipals(
        first: $first,
        after: $after,
        orderBy: $orderBy,
        filter: $filter
      ) {
        totalCount
        edges {
          node {
            id
            principalKey
            email
            fullName
            accountType
            lastSeenAt
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

type principalNode struct {
	ID           string `json:"id"`
	PrincipalKey string `json:"principalKey"`
	Email        string `json:"email"`
	FullName     string `json:"fullName"`
	AccountType  string `json:"accountType"`
	LastSeenAt   string `json:"lastSeenAt"`
	CreatedAt    string `json:"createdAt"`
}

func NewCmdListPrincipals(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg      string
		flagLimit    int
		flagAfter    string
		flagOrderBy  string
		flagOrderDir string
		flagSourceID string
		flagOutput   *string
	)

	cmd := &cobra.Command{
		Use:     "list-principals",
		Short:   "List live access inventory principals",
		Aliases: []string{"principals", "ls-principals"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return err
			}

			if err := cmdutil.ValidateLimit(flagLimit); err != nil {
				return err
			}

			client, hc, err := f.GraphQLClient(cmdutil.ConsoleEndpoint)
			if err != nil {
				return err
			}

			if flagOrg == "" {
				flagOrg = hc.Organization
			}

			if flagOrg == "" {
				return fmt.Errorf("cannot determine organization, use --org or 'prb auth login'")
			}

			variables := map[string]any{
				"id": flagOrg,
			}

			if flagOrderBy != "" {
				if err := cmdutil.ValidateEnum("order-by", flagOrderBy, []string{"CREATED_AT"}); err != nil {
					return err
				}

				variables["orderBy"] = map[string]any{
					"field":     flagOrderBy,
					"direction": flagOrderDir,
				}
			}

			if flagSourceID != "" {
				variables["filter"] = map[string]any{
					"accessReviewSourceId": flagSourceID,
				}
			}

			principals, totalCount, nextCursor, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				flagAfter,
				api.NodeConnection[principalNode](
					"accessInventoryPrincipals",
					"Organization",
					fmt.Sprintf("organization %s not found", flagOrg),
				),
			)
			if err != nil {
				return err
			}

			return cmdutil.PrintList(
				f.IOStreams,
				*flagOutput,
				principals,
				totalCount,
				cmdutil.ListView[principalNode]{
					Empty:      "No access inventory principals found.",
					Noun:       "access inventory principals",
					NextCursor: nextCursor,
					Headers:    []string{"ID", "KEY", "EMAIL", "NAME", "TYPE"},
					Row: func(p principalNode) ([]string, error) {
						return []string{
							p.ID,
							p.PrincipalKey,
							p.Email,
							p.FullName,
							p.AccountType,
						}, nil
					},
				},
			)
		},
	}

	cmdutil.AddOrganizationFlag(cmd, &flagOrg, "Organization ID")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of principals to list")
	cmdutil.AddAfterFlag(cmd, &flagAfter)
	cmd.Flags().StringVar(&flagOrderBy, "order-by", "", "Order by field (CREATED_AT)")
	cmdutil.AddOrderDirectionFlag(cmd, &flagOrderDir, "DESC")
	cmd.Flags().StringVar(&flagSourceID, "source-id", "", "Filter by access review source ID")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
