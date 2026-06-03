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
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: RiskAssessmentProcessOrder) {
  node(id: $id) {
    __typename
    ... on RiskAssessmentScope {
      processes(first: $first, after: $after, orderBy: $orderBy) {
        totalCount
        edges {
          node {
            id
            riskAssessmentScopeId
            sourceNodeId
            targetNodeId
            name
            createdAt
            updatedAt
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

type riskAssessmentProcess struct {
	ID                    string `json:"id"`
	RiskAssessmentScopeId string `json:"riskAssessmentScopeId"`
	SourceNodeId          string `json:"sourceNodeId"`
	TargetNodeId          string `json:"targetNodeId"`
	Name                  string `json:"name"`
	CreatedAt             string `json:"createdAt"`
	UpdatedAt             string `json:"updatedAt"`
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagScope    string
		flagLimit    int
		flagOrderBy  string
		flagOrderDir string
		flagOutput   *string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List processes in a risk assessment scope",
		Aliases: []string{"ls"},
		Example: `  # List processes in a scope
  prb risk-assessment process list --scope <id>

  # List processes as JSON
  prb risk-assessment process ls --scope <id> --json`,
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

			if flagScope == "" {
				return fmt.Errorf("scope is required; pass --scope")
			}

			variables := map[string]any{
				"id": flagScope,
			}

			if flagOrderBy != "" {
				if err := cmdutil.ValidateEnum("order-by", flagOrderBy, []string{"CREATED_AT", "NAME"}); err != nil {
					return err
				}

				variables["orderBy"] = map[string]any{
					"field":     flagOrderBy,
					"direction": flagOrderDir,
				}
			}

			processes, totalCount, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				func(data json.RawMessage) (*api.Connection[riskAssessmentProcess], error) {
					var resp struct {
						Node *struct {
							Typename  string                                `json:"__typename"`
							Processes api.Connection[riskAssessmentProcess] `json:"processes"`
						} `json:"node"`
					}
					if err := json.Unmarshal(data, &resp); err != nil {
						return nil, err
					}

					if resp.Node == nil {
						return nil, fmt.Errorf("scope %s not found", flagScope)
					}

					if resp.Node.Typename != "RiskAssessmentScope" {
						return nil, fmt.Errorf("expected RiskAssessmentScope node, got %s", resp.Node.Typename)
					}

					return &resp.Node.Processes, nil
				},
			)
			if err != nil {
				return err
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, processes)
			}

			if len(processes) == 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "No processes found.")
				return nil
			}

			rows := make([][]string, 0, len(processes))
			for _, p := range processes {
				rows = append(rows, []string{
					p.ID,
					p.Name,
					p.SourceNodeId,
					p.TargetNodeId,
					cmdutil.FormatTime(p.CreatedAt),
				})
			}

			t := cmdutil.NewTable("ID", "NAME", "SOURCE NODE", "TARGET NODE", "CREATED AT").Rows(rows...)

			_, _ = fmt.Fprintln(f.IOStreams.Out, t)

			if totalCount > len(processes) {
				_, _ = fmt.Fprintf(
					f.IOStreams.ErrOut,
					"\nShowing %d of %d processes\n",
					len(processes),
					totalCount,
				)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flagScope, "scope", "", "Risk assessment scope ID (required)")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of processes to list")
	cmd.Flags().StringVar(&flagOrderBy, "order-by", "", "Order by field (CREATED_AT, NAME)")
	cmd.Flags().StringVar(&flagOrderDir, "order-direction", "DESC", "Sort direction (ASC, DESC)")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("scope")

	return cmd
}
