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

package listchangeevents

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const listQuery = `
query AccessInventoryChangeEventList(
  $id: ID!,
  $first: Int,
  $after: CursorKey,
  $orderBy: AccessInventoryChangeEventOrder,
  $filter: AccessInventoryChangeEventFilter
) {
  node(id: $id) {
    __typename
    ... on Organization {
      accessInventoryChangeEvents(
        first: $first,
        after: $after,
        orderBy: $orderBy,
        filter: $filter
      ) {
        totalCount
        edges {
          node {
            id
            changeType
            principalKey
            resourceKey
            entitlementKey
            observedAt
            providerOccurredAt
            effectiveAt
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

type changeEventNode struct {
	ID                 string  `json:"id"`
	ChangeType         string  `json:"changeType"`
	PrincipalKey       string  `json:"principalKey"`
	ResourceKey        string  `json:"resourceKey"`
	EntitlementKey     string  `json:"entitlementKey"`
	ObservedAt         string  `json:"observedAt"`
	ProviderOccurredAt *string `json:"providerOccurredAt"`
	EffectiveAt        string  `json:"effectiveAt"`
	CreatedAt          string  `json:"createdAt"`
}

func NewCmdListChangeEvents(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg        string
		flagLimit      int
		flagAfter      string
		flagOrderBy    string
		flagOrderDir   string
		flagSourceID   string
		flagChangeType []string
		flagOutput     *string
	)

	cmd := &cobra.Command{
		Use:     "list-change-events",
		Short:   "List the always-on access inventory diff feed",
		Aliases: []string{"change-events", "ls-change-events"},
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
				orderBy = "EFFECTIVE_AT"
			}

			if err := cmdutil.ValidateEnum(
				"order-by",
				orderBy,
				[]string{"CREATED_AT", "OBSERVED_AT", "EFFECTIVE_AT"},
			); err != nil {
				return err
			}

			for _, changeType := range flagChangeType {
				if err := cmdutil.ValidateEnum(
					"change-type",
					changeType,
					[]string{"APPEARED", "DISAPPEARED", "ATTRIBUTE_CHANGED"},
				); err != nil {
					return err
				}
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
			if len(flagChangeType) > 0 {
				filter["changeTypes"] = flagChangeType
			}
			if len(filter) > 0 {
				variables["filter"] = filter
			}

			events, totalCount, nextCursor, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				flagAfter,
				api.NodeConnection[changeEventNode](
					"accessInventoryChangeEvents",
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
				events,
				totalCount,
				cmdutil.ListView[changeEventNode]{
					Empty:      "No access inventory change events found.",
					Noun:       "access inventory change events",
					NextCursor: nextCursor,
					Headers:    []string{"ID", "TYPE", "PRINCIPAL", "RESOURCE", "ENTITLEMENT", "EFFECTIVE_AT"},
					Row: func(e changeEventNode) ([]string, error) {
						return []string{
							e.ID,
							e.ChangeType,
							e.PrincipalKey,
							e.ResourceKey,
							e.EntitlementKey,
							e.EffectiveAt,
						}, nil
					},
				},
			)
		},
	}

	cmdutil.AddOrganizationFlag(cmd, &flagOrg, "Organization ID")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of items")
	cmdutil.AddAfterFlag(cmd, &flagAfter)
	cmd.Flags().StringVar(&flagOrderBy, "order-by", "EFFECTIVE_AT", "Order field")
	cmdutil.AddOrderDirectionFlag(cmd, &flagOrderDir, "DESC")
	cmd.Flags().StringVar(&flagSourceID, "source-id", "", "Filter by access review source ID")
	cmd.Flags().StringSliceVar(
		&flagChangeType,
		"change-type",
		nil,
		"Filter by change type (APPEARED, DISAPPEARED, ATTRIBUTE_CHANGED)",
	)
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
