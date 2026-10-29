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
	"strings"
	"time"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const listQuery = `
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: AuditLogEntryOrder, $filter: AuditLogEntryFilter) {
  node(id: $id) {
    __typename
    ... on Organization {
      auditLogEntries(first: $first, after: $after, orderBy: $orderBy, filter: $filter) {
        totalCount
        edges {
          node {
            id
            actorId
            actorType
            actorEmail
            actorName
            action
            resourceType
            resourceId
            outcome
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

type auditLogEntry struct {
	ID           string  `json:"id"`
	ActorID      string  `json:"actorId"`
	ActorType    string  `json:"actorType"`
	ActorEmail   *string `json:"actorEmail"`
	ActorName    *string `json:"actorName"`
	Action       string  `json:"action"`
	ResourceType string  `json:"resourceType"`
	ResourceID   string  `json:"resourceId"`
	Outcome      string  `json:"outcome"`
	CreatedAt    string  `json:"createdAt"`
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg          string
		flagLimit        int
		flagAfter        string
		flagOrderBy      string
		flagOrderDir     string
		flagAction       string
		flagActorID      string
		flagActorType    string
		flagResourceType string
		flagResourceID   string
		flagOutcome      string
		flagRequestID    string
		flagFrom         string
		flagTo           string
		flagOutput       *string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List audit log entries",
		Aliases: []string{"ls"},
		Example: `  prb audit-log list
  prb audit-log list --action core:thirdParty:create
  prb audit-log list --resource-type ThirdParty --limit 50`,
		Args: cobra.NoArgs,
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

			flagOrg, err = cmdutil.ResolveOrganization(flagOrg, hc.Organization)
			if err != nil {
				return err
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

			filter := map[string]any{}
			if flagAction != "" {
				if prefix, ok := strings.CutSuffix(flagAction, "*"); ok {
					if prefix != "" {
						filter["actionPrefix"] = prefix
					}
				} else {
					filter["action"] = flagAction
				}
			}

			if flagActorID != "" {
				filter["actorId"] = flagActorID
			}

			if flagResourceType != "" {
				filter["resourceType"] = flagResourceType
			}

			if flagResourceID != "" {
				filter["resourceId"] = flagResourceID
			}

			if flagActorType != "" {
				if err := cmdutil.ValidateEnum("actor-type", flagActorType, []string{"USER", "API_KEY", "SYSTEM"}); err != nil {
					return err
				}

				filter["actorType"] = flagActorType
			}

			if flagOutcome != "" {
				if err := cmdutil.ValidateEnum("outcome", flagOutcome, []string{"ALLOWED", "DENIED"}); err != nil {
					return err
				}

				filter["outcome"] = flagOutcome
			}

			if flagRequestID != "" {
				filter["requestId"] = flagRequestID
			}

			if flagFrom != "" {
				from, err := time.Parse(time.DateOnly, flagFrom)
				if err != nil {
					return fmt.Errorf("invalid --from %q: use YYYY-MM-DD", flagFrom)
				}

				filter["createdFrom"] = from.UTC().Format(time.RFC3339)
			}

			if flagTo != "" {
				to, err := time.Parse(time.DateOnly, flagTo)
				if err != nil {
					return fmt.Errorf("invalid --to %q: use YYYY-MM-DD", flagTo)
				}

				filter["createdTo"] = to.UTC().AddDate(0, 0, 1).Format(time.RFC3339)
			}

			if len(filter) > 0 {
				variables["filter"] = filter
			}

			entries, totalCount, nextCursor, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				flagAfter,
				api.NodeConnection[auditLogEntry]("auditLogEntries", "Organization", fmt.Sprintf("organization %s not found", flagOrg)),
			)
			if err != nil {
				return err
			}

			return cmdutil.PrintList(
				f.IOStreams,
				*flagOutput,
				entries,
				totalCount,
				cmdutil.ListView[auditLogEntry]{
					Empty:      "No audit log entries found.",
					Noun:       "audit log entries",
					NextCursor: nextCursor,
					Headers:    []string{"ID", "OUTCOME", "ACTOR TYPE", "ACTOR", "ACTION", "RESOURCE TYPE", "RESOURCE", "CREATED"},
					Row: func(e auditLogEntry) ([]string, error) {
						actor := e.ActorID
						if e.ActorEmail != nil && *e.ActorEmail != "" {
							actor = *e.ActorEmail
						} else if e.ActorName != nil && *e.ActorName != "" {
							actor = *e.ActorName
						}

						return []string{
							e.ID,
							e.Outcome,
							e.ActorType,
							actor,
							e.Action,
							e.ResourceType,
							e.ResourceID,
							cmdutil.FormatTime(e.CreatedAt),
						}, nil
					},
				},
			)
		},
	}

	cmdutil.AddOrganizationFlag(cmd, &flagOrg, "Organization ID")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of entries to list")
	cmdutil.AddAfterFlag(cmd, &flagAfter)
	cmd.Flags().StringVar(&flagOrderBy, "order-by", "", "Order by field (CREATED_AT)")
	cmdutil.AddOrderDirectionFlag(cmd, &flagOrderDir, "DESC")
	cmd.Flags().StringVar(&flagAction, "action", "", "Filter by action (e.g. core:thirdParty:create)")
	cmd.Flags().StringVar(&flagActorID, "actor-id", "", "Filter by actor ID")
	cmd.Flags().StringVar(&flagResourceType, "resource-type", "", "Filter by resource type (e.g. ThirdParty)")
	cmd.Flags().StringVar(&flagResourceID, "resource-id", "", "Filter by resource ID")
	cmd.Flags().StringVar(&flagActorType, "actor-type", "", "Filter by actor type (USER, API_KEY, SYSTEM)")
	cmd.Flags().StringVar(&flagOutcome, "outcome", "", "Filter by outcome (ALLOWED, DENIED)")
	cmd.Flags().StringVar(&flagRequestID, "request-id", "", "Filter by request ID")
	cmd.Flags().StringVar(&flagFrom, "from", "", "Include entries on or after this date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&flagTo, "to", "", "Include entries on or before this date (YYYY-MM-DD)")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
