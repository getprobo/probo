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
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: RiskAssessmentBoundaryOrder) {
  node(id: $id) {
    __typename
    ... on RiskAssessmentScope {
      boundaries(first: $first, after: $after, orderBy: $orderBy) {
        totalCount
        edges {
          node {
            id
            riskAssessmentScopeId
            parentBoundaryId
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

type riskAssessmentBoundary struct {
	ID                    string  `json:"id"`
	RiskAssessmentScopeId string  `json:"riskAssessmentScopeId"`
	ParentBoundaryId      *string `json:"parentBoundaryId"`
	Name                  string  `json:"name"`
	CreatedAt             string  `json:"createdAt"`
	UpdatedAt             string  `json:"updatedAt"`
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
		Short:   "List boundaries in a risk assessment scope",
		Aliases: []string{"ls"},
		Example: `  # List boundaries in a scope
  prb risk-assessment boundary list --scope <id>

  # List boundaries as JSON
  prb risk-assessment boundary ls --scope <id> --json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return err
			}

			if err := cmdutil.ValidateLimit(flagLimit); err != nil {
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

			boundaries, totalCount, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				func(data json.RawMessage) (*api.Connection[riskAssessmentBoundary], error) {
					var resp struct {
						Node *struct {
							Typename   string                                 `json:"__typename"`
							Boundaries api.Connection[riskAssessmentBoundary] `json:"boundaries"`
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

					return &resp.Node.Boundaries, nil
				},
			)
			if err != nil {
				return err
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, boundaries)
			}

			if len(boundaries) == 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "No boundaries found.")
				return nil
			}

			rows := make([][]string, 0, len(boundaries))
			for _, b := range boundaries {
				parent := ""
				if b.ParentBoundaryId != nil {
					parent = *b.ParentBoundaryId
				}

				rows = append(rows, []string{
					b.ID,
					b.Name,
					parent,
					cmdutil.FormatTime(b.CreatedAt),
				})
			}

			t := cmdutil.NewTable("ID", "NAME", "PARENT", "CREATED AT").Rows(rows...)

			_, _ = fmt.Fprintln(f.IOStreams.Out, t)

			if totalCount > len(boundaries) {
				_, _ = fmt.Fprintf(
					f.IOStreams.ErrOut,
					"\nShowing %d of %d boundaries\n",
					len(boundaries),
					totalCount,
				)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flagScope, "scope", "", "Risk assessment scope ID (required)")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of boundaries to list")
	cmd.Flags().StringVar(&flagOrderBy, "order-by", "", "Order by field (CREATED_AT, NAME)")
	cmd.Flags().StringVar(&flagOrderDir, "order-direction", "DESC", "Sort direction (ASC, DESC)")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("scope")

	return cmd
}
