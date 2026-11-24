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

package listgrants

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const listQuery = `
query AccessInventoryGrantList(
  $id: ID!,
  $first: Int,
  $after: CursorKey,
  $orderBy: AccessInventoryGrantOrder,
  $filter: AccessInventoryGrantFilter
) {
  node(id: $id) {
    __typename
    ... on Organization {
      accessInventoryGrants(
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
            resourceKey
            entitlementKey
            entitlementDisplayName
            presence
            consecutiveMissingSyncs
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

type grantNode struct {
	ID                      string `json:"id"`
	PrincipalKey            string `json:"principalKey"`
	ResourceKey             string `json:"resourceKey"`
	EntitlementKey          string `json:"entitlementKey"`
	EntitlementDisplayName  string `json:"entitlementDisplayName"`
	Presence                string `json:"presence"`
	ConsecutiveMissingSyncs int    `json:"consecutiveMissingSyncs"`
	LastSeenAt              string `json:"lastSeenAt"`
	CreatedAt               string `json:"createdAt"`
}

func NewCmdListGrants(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg                string
		flagLimit              int
		flagAfter              string
		flagOrderBy            string
		flagOrderDir           string
		flagSourceID           string
		flagIncludeDisappeared bool
		flagOutput             *string
	)

	cmd := &cobra.Command{
		Use:     "list-grants",
		Short:   "List live access inventory grants",
		Aliases: []string{"grants", "ls-grants"},
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

			filter := map[string]any{}
			if flagSourceID != "" {
				filter["accessReviewSourceId"] = flagSourceID
			}
			if flagIncludeDisappeared {
				filter["includeDisappeared"] = true
			}
			if len(filter) > 0 {
				variables["filter"] = filter
			}

			grants, totalCount, nextCursor, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				flagAfter,
				api.NodeConnection[grantNode](
					"accessInventoryGrants",
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
				grants,
				totalCount,
				cmdutil.ListView[grantNode]{
					Empty:      "No access inventory grants found.",
					Noun:       "access inventory grants",
					NextCursor: nextCursor,
					Headers:    []string{"ID", "PRINCIPAL", "RESOURCE", "ENTITLEMENT", "PRESENCE"},
					Row: func(g grantNode) ([]string, error) {
						entitlement := g.EntitlementDisplayName
						if entitlement == "" {
							entitlement = g.EntitlementKey
						}

						return []string{
							g.ID,
							g.PrincipalKey,
							g.ResourceKey,
							entitlement,
							g.Presence,
						}, nil
					},
				},
			)
		},
	}

	cmdutil.AddOrganizationFlag(cmd, &flagOrg, "Organization ID")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of grants to list")
	cmdutil.AddAfterFlag(cmd, &flagAfter)
	cmd.Flags().StringVar(&flagOrderBy, "order-by", "", "Order by field (CREATED_AT)")
	cmdutil.AddOrderDirectionFlag(cmd, &flagOrderDir, "DESC")
	cmd.Flags().StringVar(&flagSourceID, "source-id", "", "Filter by access review source ID")
	cmd.Flags().BoolVar(
		&flagIncludeDisappeared,
		"include-disappeared",
		false,
		"Include grants past the hide-after-X threshold",
	)
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
