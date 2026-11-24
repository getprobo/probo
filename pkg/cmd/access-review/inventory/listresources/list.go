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

package listresources

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const listQuery = `
query AccessInventoryResourceList(
  $id: ID!,
  $first: Int,
  $after: CursorKey,
  $orderBy: AccessInventoryResourceOrder,
  $filter: AccessInventoryResourceFilter
) {
  node(id: $id) {
    __typename
    ... on Organization {
      accessInventoryResources(
        first: $first,
        after: $after,
        orderBy: $orderBy,
        filter: $filter
      ) {
        totalCount
        edges {
          node {
            id
            resourceKey
            displayName
            kind
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

type resourceNode struct {
	ID          string `json:"id"`
	ResourceKey string `json:"resourceKey"`
	DisplayName string `json:"displayName"`
	Kind        string `json:"kind"`
	LastSeenAt  string `json:"lastSeenAt"`
	CreatedAt   string `json:"createdAt"`
}

func NewCmdListResources(f *cmdutil.Factory) *cobra.Command {
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
		Use:     "list-resources",
		Short:   "List live access inventory resources",
		Aliases: []string{"resources", "ls-resources"},
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

			orderBy := flagOrderBy
			if orderBy == "" {
				orderBy = "CREATED_AT"
			}

			if err := cmdutil.ValidateEnum("order-by", orderBy, []string{"CREATED_AT"}); err != nil {
				return err
			}

			variables := map[string]any{
				"id": flagOrg,
				"orderBy": map[string]any{
					"field":     orderBy,
					"direction": flagOrderDir,
				},
			}

			if flagSourceID != "" {
				variables["filter"] = map[string]any{
					"accessReviewSourceId": flagSourceID,
				}
			}

			resources, totalCount, nextCursor, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				flagAfter,
				api.NodeConnection[resourceNode](
					"accessInventoryResources",
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
				resources,
				totalCount,
				cmdutil.ListView[resourceNode]{
					Empty:      "No access inventory resources found.",
					Noun:       "access inventory resources",
					NextCursor: nextCursor,
					Headers:    []string{"ID", "KEY", "NAME", "KIND"},
					Row: func(r resourceNode) ([]string, error) {
						return []string{
							r.ID,
							r.ResourceKey,
							r.DisplayName,
							r.Kind,
						}, nil
					},
				},
			)
		},
	}

	cmdutil.AddOrganizationFlag(cmd, &flagOrg, "Organization ID")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of resources to list")
	cmdutil.AddAfterFlag(cmd, &flagAfter)
	cmd.Flags().StringVar(&flagOrderBy, "order-by", "", "Order by field (CREATED_AT)")
	cmdutil.AddOrderDirectionFlag(cmd, &flagOrderDir, "DESC")
	cmd.Flags().StringVar(&flagSourceID, "source-id", "", "Filter by access review source ID")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
