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

func eventsListQuery(includePayload bool) string {
	payloadField := ""
	if includePayload {
		payloadField = "\n            payload"
	}

	return fmt.Sprintf(`
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: WebhookEventOrder, $filter: WebhookEventFilter) {
  node(id: $id) {
    __typename
    ... on WebhookSubscription {
      events(first: $first, after: $after, orderBy: $orderBy, filter: $filter) {
        totalCount
        edges {
          node {
            id
            status
            createdAt%s
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
`, payloadField)
}

type webhookEvent struct {
	ID        string  `json:"id"`
	Status    string  `json:"status"`
	CreatedAt string  `json:"createdAt"`
	Payload   *string `json:"payload,omitempty"`
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagLimit    int
		flagAfter    string
		flagOrderBy  string
		flagOrderDir string
		flagStatus   string
		flagOutput   *string
	)

	cmd := &cobra.Command{
		Use:     "list <webhook-id>",
		Short:   "List events for a webhook subscription",
		Aliases: []string{"ls"},
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
				"id": args[0],
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

			if flagStatus != "" {
				if err := cmdutil.ValidateEnum(
					"status",
					flagStatus,
					[]string{"PENDING", "SUCCEEDED", "FAILED"},
				); err != nil {
					return err
				}

				variables["filter"] = map[string]any{
					"status": flagStatus,
				}
			}

			events, totalCount, nextCursor, err := api.Paginate(
				client,
				eventsListQuery(*flagOutput == cmdutil.OutputJSON),
				variables,
				flagLimit,
				flagAfter,
				api.NodeConnection[webhookEvent]("events", "WebhookSubscription", fmt.Sprintf("webhook subscription %s not found", args[0])),
			)
			if err != nil {
				return err
			}

			return cmdutil.PrintList(
				f.IOStreams,
				*flagOutput,
				events,
				totalCount,
				cmdutil.ListView[webhookEvent]{
					Empty:      "No webhook events found.",
					Noun:       "events",
					NextCursor: nextCursor,
					Headers:    []string{"ID", "STATUS", "CREATED"},
					Row: func(e webhookEvent) ([]string, error) {
						return []string{
							e.ID,
							e.Status,
							cmdutil.FormatTime(e.CreatedAt),
						}, nil
					},
				},
			)
		},
	}

	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of events to list")
	cmdutil.AddAfterFlag(cmd, &flagAfter)
	cmd.Flags().StringVar(&flagOrderBy, "order-by", "", "Order by field (CREATED_AT)")
	cmdutil.AddOrderDirectionFlag(cmd, &flagOrderDir, "DESC")
	cmd.Flags().StringVar(&flagStatus, "status", "", "Filter by status (PENDING, SUCCEEDED, FAILED)")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
