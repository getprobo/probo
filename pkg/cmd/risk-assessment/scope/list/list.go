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
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: RiskAssessmentScopeOrder) {
  node(id: $id) {
    __typename
    ... on RiskAssessment {
      scopes(first: $first, after: $after, orderBy: $orderBy) {
        totalCount
        edges {
          node {
            id
            riskAssessmentId
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

type riskAssessmentScope struct {
	ID               string `json:"id"`
	RiskAssessmentId string `json:"riskAssessmentId"`
	Name             string `json:"name"`
	CreatedAt        string `json:"createdAt"`
	UpdatedAt        string `json:"updatedAt"`
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagRiskAssessment string
		flagLimit          int
		flagOrderBy        string
		flagOrderDir       string
		flagOutput         *string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List scopes in a risk assessment",
		Aliases: []string{"ls"},
		Example: `  # List scopes for a risk assessment
  prb risk-assessment scope list --risk-assessment <id>

  # List scopes as JSON
  prb risk-assessment scope ls --risk-assessment <id> --json`,
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

			if flagRiskAssessment == "" {
				return fmt.Errorf("risk assessment is required; pass --risk-assessment")
			}

			variables := map[string]any{
				"id": flagRiskAssessment,
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

			scopes, totalCount, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				func(data json.RawMessage) (*api.Connection[riskAssessmentScope], error) {
					var resp struct {
						Node *struct {
							Typename string                              `json:"__typename"`
							Scopes   api.Connection[riskAssessmentScope] `json:"scopes"`
						} `json:"node"`
					}
					if err := json.Unmarshal(data, &resp); err != nil {
						return nil, err
					}

					if resp.Node == nil {
						return nil, fmt.Errorf("risk assessment %s not found", flagRiskAssessment)
					}

					if resp.Node.Typename != "RiskAssessment" {
						return nil, fmt.Errorf("expected RiskAssessment node, got %s", resp.Node.Typename)
					}

					return &resp.Node.Scopes, nil
				},
			)
			if err != nil {
				return err
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, scopes)
			}

			if len(scopes) == 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "No scopes found.")
				return nil
			}

			rows := make([][]string, 0, len(scopes))
			for _, s := range scopes {
				rows = append(rows, []string{
					s.ID,
					s.Name,
					cmdutil.FormatTime(s.CreatedAt),
				})
			}

			t := cmdutil.NewTable("ID", "NAME", "CREATED AT").Rows(rows...)

			_, _ = fmt.Fprintln(f.IOStreams.Out, t)

			if totalCount > len(scopes) {
				_, _ = fmt.Fprintf(
					f.IOStreams.ErrOut,
					"\nShowing %d of %d scopes\n",
					len(scopes),
					totalCount,
				)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flagRiskAssessment, "risk-assessment", "", "Risk assessment ID (required)")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of scopes to list")
	cmd.Flags().StringVar(&flagOrderBy, "order-by", "", "Order by field (CREATED_AT, NAME)")
	cmd.Flags().StringVar(&flagOrderDir, "order-direction", "DESC", "Sort direction (ASC, DESC)")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("risk-assessment")

	return cmd
}
