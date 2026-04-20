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

package listapprovaldecisions

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const listQuery = `
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: DocumentVersionApprovalDecisionOrder) {
  node(id: $id) {
    __typename
    ... on DocumentVersionApprovalQuorum {
      decisions(first: $first, after: $after, orderBy: $orderBy) {
        totalCount
        edges {
          node {
            id
            approver {
              fullName
            }
            state
            comment
            decidedAt
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

type approvalDecision struct {
	ID       string `json:"id"`
	Approver struct {
		FullName string `json:"fullName"`
	} `json:"approver"`
	State     string  `json:"state"`
	Comment   *string `json:"comment"`
	DecidedAt *string `json:"decidedAt"`
}

func NewCmdListApprovalDecisions(f *cmdutil.Factory) *cobra.Command {
	var (
		flagLimit    int
		flagOrderBy  string
		flagOrderDir string
		flagOutput   *string
	)

	cmd := &cobra.Command{
		Use:   "list-approval-decisions <quorum-id>",
		Short: "List approval decisions for a quorum",
		Example: `  # List approval decisions for a quorum
  prb document list-approval-decisions <quorum-id>`,
		Args: cobra.ExactArgs(1),
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

			variables := map[string]any{
				"id": args[0],
			}

			if flagOrderBy != "" {
				if err := cmdutil.ValidateEnum(
					"order-by",
					flagOrderBy,
					[]string{"CREATED_AT"},
				); err != nil {
					return err
				}
				if err := cmdutil.ValidateEnum(
					"order-direction",
					flagOrderDir,
					[]string{"ASC", "DESC"},
				); err != nil {
					return err
				}
				variables["orderBy"] = map[string]any{
					"field":     flagOrderBy,
					"direction": flagOrderDir,
				}
			}

			decisions, totalCount, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				func(data json.RawMessage) (*api.Connection[approvalDecision], error) {
					var resp struct {
						Node *struct {
							Typename  string                           `json:"__typename"`
							Decisions api.Connection[approvalDecision] `json:"decisions"`
						} `json:"node"`
					}
					if err := json.Unmarshal(data, &resp); err != nil {
						return nil, err
					}
					if resp.Node == nil {
						return nil, fmt.Errorf("approval quorum %s not found", args[0])
					}
					if resp.Node.Typename != "DocumentVersionApprovalQuorum" {
						return nil, fmt.Errorf("expected DocumentVersionApprovalQuorum node, got %s", resp.Node.Typename)
					}
					return &resp.Node.Decisions, nil
				},
			)
			if err != nil {
				return err
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, decisions)
			}

			if len(decisions) == 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "No approval decisions found.")
				return nil
			}

			rows := make([][]string, 0, len(decisions))
			for _, d := range decisions {
				comment := ""
				if d.Comment != nil {
					comment = *d.Comment
				}

				decidedAt := ""
				if d.DecidedAt != nil {
					decidedAt = cmdutil.FormatTime(*d.DecidedAt)
				}

				rows = append(rows, []string{
					d.ID,
					d.Approver.FullName,
					d.State,
					comment,
					decidedAt,
				})
			}

			t := cmdutil.NewTable("ID", "APPROVER", "STATE", "COMMENT", "DECIDED AT").Rows(rows...)

			_, _ = fmt.Fprintln(f.IOStreams.Out, t)

			if totalCount > len(decisions) {
				_, _ = fmt.Fprintf(
					f.IOStreams.ErrOut,
					"\nShowing %d of %d approval decisions\n",
					len(decisions),
					totalCount,
				)
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of approval decisions to list")
	cmd.Flags().StringVar(&flagOrderBy, "order-by", "", "Order by field (CREATED_AT)")
	cmd.Flags().StringVar(&flagOrderDir, "order-direction", "DESC", "Sort direction (ASC, DESC)")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
