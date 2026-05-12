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
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: FindingOrder, $filter: FindingFilter) {
  node(id: $id) {
    __typename
    ... on Organization {
      findings(first: $first, after: $after, orderBy: $orderBy, filter: $filter) {
        totalCount
        edges {
          node {
            id
            referenceId
            kind
            status
            priority
            dueDate
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

type finding struct {
	ID          string  `json:"id"`
	ReferenceID string  `json:"referenceId"`
	Kind        string  `json:"kind"`
	Status      string  `json:"status"`
	Priority    string  `json:"priority"`
	DueDate     *string `json:"dueDate"`
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrganization string
		flagLimit        int
		flagOrderBy      string
		flagOrderDir     string
		flagKind         string
		flagOutput       *string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List findings in an organization",
		Aliases: []string{"ls"},
		Example: `  # List findings in an organization
  prb finding list --organization <organization-id>

  # Filter by kind and output as JSON
  prb finding ls --organization <organization-id> --kind MINOR_NONCONFORMITY --json`,
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
				cmdutil.TokenRefreshOption(cfg, host, hc),
			)

			variables := map[string]any{
				"id": flagOrganization,
			}

			if flagOrderBy != "" {
				if err := cmdutil.ValidateEnum("order-by", flagOrderBy, []string{"CREATED_AT", "REFERENCE_ID", "IDENTIFIED_ON", "DUE_DATE", "STATUS", "PRIORITY", "KIND"}); err != nil {
					return err
				}
				variables["orderBy"] = map[string]any{
					"field":     flagOrderBy,
					"direction": flagOrderDir,
				}
			}

			filter := map[string]any{}
			if flagKind != "" {
				if err := cmdutil.ValidateEnum("kind", flagKind, []string{"MINOR_NONCONFORMITY", "MAJOR_NONCONFORMITY", "OBSERVATION", "EXCEPTION"}); err != nil {
					return err
				}
				filter["kind"] = flagKind
			}
			if len(filter) > 0 {
				variables["filter"] = filter
			}

			findings, totalCount, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				func(data json.RawMessage) (*api.Connection[finding], error) {
					var resp struct {
						Node *struct {
							Typename string                  `json:"__typename"`
							Findings api.Connection[finding] `json:"findings"`
						} `json:"node"`
					}
					if err := json.Unmarshal(data, &resp); err != nil {
						return nil, err
					}
					if resp.Node == nil {
						return nil, fmt.Errorf("organization %s not found", flagOrganization)
					}
					if resp.Node.Typename != "Organization" {
						return nil, fmt.Errorf("expected Organization node, got %s", resp.Node.Typename)
					}
					return &resp.Node.Findings, nil
				},
			)
			if err != nil {
				return err
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, findings)
			}

			if len(findings) == 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "No findings found.")
				return nil
			}

			rows := make([][]string, 0, len(findings))
			for _, fi := range findings {
				dueDate := "-"
				if fi.DueDate != nil {
					dueDate = cmdutil.FormatTime(*fi.DueDate)
				}
				rows = append(rows, []string{
					fi.ID,
					fi.ReferenceID,
					fi.Kind,
					fi.Status,
					fi.Priority,
					dueDate,
				})
			}

			t := cmdutil.NewTable("ID", "REFERENCE", "KIND", "STATUS", "PRIORITY", "DUE DATE").Rows(rows...)

			_, _ = fmt.Fprintln(f.IOStreams.Out, t)

			if totalCount > len(findings) {
				_, _ = fmt.Fprintf(
					f.IOStreams.ErrOut,
					"\nShowing %d of %d findings\n",
					len(findings),
					totalCount,
				)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrganization, "organization", "", "Organization ID (required)")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of findings to list")
	cmd.Flags().StringVar(&flagOrderBy, "order-by", "", "Order by field (CREATED_AT, REFERENCE_ID, IDENTIFIED_ON, DUE_DATE, STATUS, PRIORITY, KIND)")
	cmd.Flags().StringVar(&flagOrderDir, "order-direction", "DESC", "Sort direction (ASC, DESC)")
	cmd.Flags().StringVar(&flagKind, "kind", "", "Filter by kind (MINOR_NONCONFORMITY, MAJOR_NONCONFORMITY, OBSERVATION, EXCEPTION)")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("organization")

	return cmd
}
