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
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const listQuery = `
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: RiskAssessmentThreatOrder) {
  node(id: $id) {
    __typename
    ... on RiskAssessmentScope {
      threats(first: $first, after: $after, orderBy: $orderBy) {
        totalCount
        edges {
          node {
            id
            riskAssessmentScopeId
            processId
            name
            category
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

type riskAssessmentThreat struct {
	ID                    string `json:"id"`
	RiskAssessmentScopeId string `json:"riskAssessmentScopeId"`
	ProcessId             string `json:"processId"`
	Name                  string `json:"name"`
	Category              string `json:"category"`
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
		Short:   "List threats in a risk assessment scope",
		Aliases: []string{"ls"},
		Example: `  # List threats in a scope
  prb risk-assessment threat list --scope <id>

  # List threats as JSON
  prb risk-assessment threat ls --scope <id> --json`,
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

			threats, totalCount, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				func(data json.RawMessage) (*api.Connection[riskAssessmentThreat], error) {
					var resp struct {
						Node *struct {
							Typename string                               `json:"__typename"`
							Threats  api.Connection[riskAssessmentThreat] `json:"threats"`
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

					return &resp.Node.Threats, nil
				},
			)
			if err != nil {
				return err
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, threats)
			}

			if len(threats) == 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "No threats found.")
				return nil
			}

			rows := make([][]string, 0, len(threats))
			for _, t := range threats {
				rows = append(rows, []string{
					t.ID,
					t.Name,
					t.Category,
					t.ProcessId,
					cmdutil.FormatTime(t.CreatedAt),
				})
			}

			tbl := cmdutil.NewTable("ID", "NAME", "CATEGORY", "PROCESS", "CREATED AT").Rows(rows...)

			_, _ = fmt.Fprintln(f.IOStreams.Out, tbl)

			if totalCount > len(threats) {
				_, _ = fmt.Fprintf(
					f.IOStreams.ErrOut,
					"\nShowing %d of %d threats\n",
					len(threats),
					totalCount,
				)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flagScope, "scope", "", "Risk assessment scope ID (required)")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of threats to list")
	cmd.Flags().StringVar(&flagOrderBy, "order-by", "", "Order by field (CREATED_AT, NAME)")
	cmd.Flags().StringVar(&flagOrderDir, "order-direction", "DESC", "Sort direction (ASC, DESC)")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("scope")

	return cmd
}
