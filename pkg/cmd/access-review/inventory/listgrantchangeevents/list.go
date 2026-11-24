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

package listgrantchangeevents

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const listQuery = `
query AccessInventoryGrantChangeEventList(
  $id: ID!,
  $first: Int,
  $after: CursorKey,
  $orderBy: AccessInventoryChangeEventOrder,
  $filter: AccessInventoryChangeEventFilter
) {
  node(id: $id) {
    __typename
    ... on AccessInventoryGrant {
      changeEvents(
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
            observedAt
            providerOccurredAt
            effectiveAt
            createdAt
            before {
              entitlementDisplayName
              presence
            }
            after {
              entitlementDisplayName
              presence
            }
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

type attributeSnapshot struct {
	EntitlementDisplayName string `json:"entitlementDisplayName"`
	Presence               string `json:"presence"`
}

type changeEventNode struct {
	ID                 string             `json:"id"`
	ChangeType         string             `json:"changeType"`
	ObservedAt         string             `json:"observedAt"`
	ProviderOccurredAt *string            `json:"providerOccurredAt"`
	EffectiveAt        string             `json:"effectiveAt"`
	CreatedAt          string             `json:"createdAt"`
	Before             *attributeSnapshot `json:"before"`
	After              *attributeSnapshot `json:"after"`
}

func formatSnapshot(s *attributeSnapshot) string {
	if s == nil {
		return "-"
	}

	return fmt.Sprintf("%s/%s", s.Presence, s.EntitlementDisplayName)
}

func NewCmdListGrantChangeEvents(f *cmdutil.Factory) *cobra.Command {
	var (
		flagLimit                int
		flagAfter                string
		flagOrderBy              string
		flagOrderDir             string
		flagChangeType           []string
		flagObservedAtOnOrAfter  string
		flagObservedAtOnOrBefore string
		flagOutput               *string
	)

	cmd := &cobra.Command{
		Use:   "list-grant-change-events <grant-id>",
		Short: "List the changelog for one access inventory grant",
		Args:  cobra.ExactArgs(1),
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
				"id": args[0],
				"orderBy": map[string]any{
					"field":     orderBy,
					"direction": flagOrderDir,
				},
			}

			filter := map[string]any{}
			if len(flagChangeType) > 0 {
				filter["changeTypes"] = flagChangeType
			}
			if flagObservedAtOnOrAfter != "" {
				if _, err := time.Parse(time.RFC3339, flagObservedAtOnOrAfter); err != nil {
					return fmt.Errorf("invalid --observed-at-on-or-after %q: use RFC3339", flagObservedAtOnOrAfter)
				}

				filter["observedAtOnOrAfter"] = flagObservedAtOnOrAfter
			}
			if flagObservedAtOnOrBefore != "" {
				if _, err := time.Parse(time.RFC3339, flagObservedAtOnOrBefore); err != nil {
					return fmt.Errorf("invalid --observed-at-on-or-before %q: use RFC3339", flagObservedAtOnOrBefore)
				}

				filter["observedAtOnOrBefore"] = flagObservedAtOnOrBefore
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
					"changeEvents",
					"AccessInventoryGrant",
					fmt.Sprintf("access inventory grant %s not found", args[0]),
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
					Empty:      "No change events found for this grant.",
					Noun:       "grant change events",
					NextCursor: nextCursor,
					Headers:    []string{"ID", "TYPE", "BEFORE", "AFTER", "EFFECTIVE_AT"},
					Row: func(e changeEventNode) ([]string, error) {
						return []string{
							e.ID,
							e.ChangeType,
							formatSnapshot(e.Before),
							formatSnapshot(e.After),
							e.EffectiveAt,
						}, nil
					},
				},
			)
		},
	}

	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of items")
	cmdutil.AddAfterFlag(cmd, &flagAfter)
	cmd.Flags().StringVar(&flagOrderBy, "order-by", "EFFECTIVE_AT", "Order field")
	cmdutil.AddOrderDirectionFlag(cmd, &flagOrderDir, "DESC")
	cmd.Flags().StringSliceVar(
		&flagChangeType,
		"change-type",
		nil,
		"Filter by change type (APPEARED, DISAPPEARED, ATTRIBUTE_CHANGED)",
	)
	cmd.Flags().StringVar(
		&flagObservedAtOnOrAfter,
		"observed-at-on-or-after",
		"",
		"Include events observed at or after this RFC3339 timestamp",
	)
	cmd.Flags().StringVar(
		&flagObservedAtOnOrBefore,
		"observed-at-on-or-before",
		"",
		"Include events observed at or before this RFC3339 timestamp",
	)
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
