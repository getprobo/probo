// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package list

import (
	"encoding/json"
	"fmt"

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
            action
            resourceType
            resourceId
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
	ID           string `json:"id"`
	ActorID      string `json:"actorId"`
	ActorType    string `json:"actorType"`
	Action       string `json:"action"`
	ResourceType string `json:"resourceType"`
	ResourceID   string `json:"resourceId"`
	CreatedAt    string `json:"createdAt"`
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg          string
		flagLimit        int
		flagOrderBy      string
		flagOrderDir     string
		flagAction       string
		flagActorID      string
		flagResourceType string
		flagResourceID   string
		flagOutput       *string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List audit log entries",
		Aliases: []string{"ls"},
		Example: `  prb audit-log list
  prb audit-log list --action core:vendor:create
  prb audit-log list --resource-type Vendor --limit 50`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return err
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
			)

			if flagOrg == "" {
				flagOrg = hc.Organization
			}

			if flagOrg == "" {
				return fmt.Errorf("organization is required; pass --org or set a default with 'prb auth login'")
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
				filter["action"] = flagAction
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
			if len(filter) > 0 {
				variables["filter"] = filter
			}

			entries, totalCount, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				func(data json.RawMessage) (*api.Connection[auditLogEntry], error) {
					var resp struct {
						Node *struct {
							Typename        string                        `json:"__typename"`
							AuditLogEntries api.Connection[auditLogEntry] `json:"auditLogEntries"`
						} `json:"node"`
					}
					if err := json.Unmarshal(data, &resp); err != nil {
						return nil, err
					}
					if resp.Node == nil {
						return nil, fmt.Errorf("organization %s not found", flagOrg)
					}
					if resp.Node.Typename != "Organization" {
						return nil, fmt.Errorf("expected Organization node, got %s", resp.Node.Typename)
					}
					return &resp.Node.AuditLogEntries, nil
				},
			)
			if err != nil {
				return err
			}

			if *flagOutput == cmdutil.OutputJSON {
				if entries == nil {
					entries = []auditLogEntry{}
				}
				return cmdutil.PrintJSON(f.IOStreams.Out, entries)
			}

			if len(entries) == 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "No audit log entries found.")
				return nil
			}

			rows := make([][]string, 0, len(entries))
			for _, e := range entries {
				rows = append(rows, []string{
					e.ID,
					e.ActorType,
					e.ActorID,
					e.Action,
					e.ResourceType,
					e.ResourceID,
					cmdutil.FormatTime(e.CreatedAt),
				})
			}

			t := cmdutil.NewTable("ID", "ACTOR TYPE", "ACTOR", "ACTION", "RESOURCE TYPE", "RESOURCE", "CREATED").Rows(rows...)

			_, _ = fmt.Fprintln(f.IOStreams.Out, t)

			if totalCount > len(entries) {
				_, _ = fmt.Fprintf(
					f.IOStreams.ErrOut,
					"\nShowing %d of %d audit log entries\n",
					len(entries),
					totalCount,
				)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of entries to list")
	cmd.Flags().StringVar(&flagOrderBy, "order-by", "", "Order by field (CREATED_AT)")
	cmd.Flags().StringVar(&flagOrderDir, "order-direction", "DESC", "Sort direction (ASC, DESC)")
	cmd.Flags().StringVar(&flagAction, "action", "", "Filter by action (e.g. core:vendor:create)")
	cmd.Flags().StringVar(&flagActorID, "actor-id", "", "Filter by actor ID")
	cmd.Flags().StringVar(&flagResourceType, "resource-type", "", "Filter by resource type (e.g. Vendor)")
	cmd.Flags().StringVar(&flagResourceID, "resource-id", "", "Filter by resource ID")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
