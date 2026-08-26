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

package create

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const createMutation = `
mutation($input: CreateRiskAnalysisInput!) {
  createRiskAnalysis(input: $input) {
    riskAnalysisEdge {
      node {
        id
        name
        description
        period {
          start
          end
        }
        matrixSize {
          rows
          cols
        }
        createdAt
        updatedAt
      }
    }
  }
}
`

type createResponse struct {
	CreateRiskAnalysis struct {
		RiskAnalysisEdge struct {
			Node struct {
				ID          string  `json:"id"`
				Name        string  `json:"name"`
				Description *string `json:"description"`
				Period      *struct {
					Start *string `json:"start"`
					End   *string `json:"end"`
				} `json:"period"`
				MatrixSize struct {
					Rows int `json:"rows"`
					Cols int `json:"cols"`
				} `json:"matrixSize"`
				CreatedAt string `json:"createdAt"`
				UpdatedAt string `json:"updatedAt"`
			} `json:"node"`
		} `json:"riskAnalysisEdge"`
	} `json:"createRiskAnalysis"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg         string
		flagName        string
		flagDescription string
		flagPeriodStart string
		flagPeriodEnd   string
		flagMatrixRows  int
		flagMatrixCols  int
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new risk analysis",
		Example: `  # Create a risk analysis interactively
  prb risk-analysis create

  # Create a risk analysis non-interactively
  prb risk-analysis create --name "Annual Risk Analysis" --matrix-rows 5 --matrix-cols 5 --period-start 2026-01-01 --period-end 2026-12-31`,
		RunE: func(cmd *cobra.Command, args []string) error {
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

			if flagOrg == "" {
				flagOrg = hc.Organization
			}

			if flagOrg == "" {
				return fmt.Errorf("organization is required; pass --org or set a default with 'prb auth login'")
			}

			if f.IOStreams.IsInteractive() {
				if flagName == "" {
					err := huh.NewInput().
						Title("Risk analysis name").
						Value(&flagName).
						Run()
					if err != nil {
						return err
					}
				}

				if !cmd.Flags().Changed("matrix-rows") && !cmd.Flags().Changed("matrix-cols") {
					matrixSize := "5"

					err := huh.NewSelect[string]().
						Title("Matrix size").
						Options(
							huh.NewOption("3×3", "3"),
							huh.NewOption("4×4", "4"),
							huh.NewOption("5×5", "5"),
						).
						Value(&matrixSize).
						Run()
					if err != nil {
						return err
					}

					switch matrixSize {
					case "3":
						flagMatrixRows, flagMatrixCols = 3, 3
					case "4":
						flagMatrixRows, flagMatrixCols = 4, 4
					case "5":
						flagMatrixRows, flagMatrixCols = 5, 5
					}
				}
			}

			if flagName == "" {
				return fmt.Errorf("name is required; pass --name or run interactively")
			}

			input := map[string]any{
				"organizationId": flagOrg,
				"name":           flagName,
			}

			if flagDescription != "" {
				input["description"] = flagDescription
			}

			if flagPeriodStart != "" || flagPeriodEnd != "" {
				period := map[string]any{}
				if flagPeriodStart != "" {
					period["start"] = flagPeriodStart
				}

				if flagPeriodEnd != "" {
					period["end"] = flagPeriodEnd
				}

				input["period"] = period
			}

			rowsChanged := cmd.Flags().Changed("matrix-rows")
			colsChanged := cmd.Flags().Changed("matrix-cols")

			switch {
			case rowsChanged && !colsChanged:
				flagMatrixCols = flagMatrixRows
			case colsChanged && !rowsChanged:
				flagMatrixRows = flagMatrixCols
			case flagMatrixRows != flagMatrixCols:
				return fmt.Errorf("--matrix-rows and --matrix-cols must match")
			}

			input["matrixSize"] = map[string]any{
				"rows": flagMatrixRows,
				"cols": flagMatrixCols,
			}

			data, err := client.Do(
				createMutation,
				map[string]any{"input": input},
			)
			if err != nil {
				return err
			}

			var resp createResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			r := resp.CreateRiskAnalysis.RiskAnalysisEdge.Node
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Created risk analysis %s (%s)\n",
				r.ID,
				r.Name,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().StringVar(&flagName, "name", "", "Risk analysis name (required)")
	cmd.Flags().StringVar(&flagDescription, "description", "", "Risk analysis description")
	cmd.Flags().StringVar(&flagPeriodStart, "period-start", "", "Period start date (e.g. 2026-01-01)")
	cmd.Flags().StringVar(&flagPeriodEnd, "period-end", "", "Period end date (e.g. 2026-12-31)")
	cmd.Flags().IntVar(&flagMatrixRows, "matrix-rows", 5, "Matrix rows (3, 4, or 5; default 5)")
	cmd.Flags().IntVar(&flagMatrixCols, "matrix-cols", 5, "Matrix cols (3, 4, or 5; default 5, must match --matrix-rows)")

	return cmd
}
