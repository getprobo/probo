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

package fork

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const forkMutation = `
mutation($input: ForkRiskAnalysisInput!) {
  forkRiskAnalysis(input: $input) {
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

type forkResponse struct {
	ForkRiskAnalysis struct {
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
	} `json:"forkRiskAnalysis"`
}

func NewCmdFork(f *cmdutil.Factory) *cobra.Command {
	var (
		flagName        string
		flagDescription string
		flagPeriodStart string
		flagPeriodEnd   string
	)

	cmd := &cobra.Command{
		Use:   "fork <id>",
		Short: "Fork a risk analysis",
		Example: `  # Fork a risk analysis interactively
  prb risk-analysis fork <id>

  # Fork a risk analysis non-interactively
  prb risk-analysis fork <id> --name "Annual Risk Analysis" --period-start 2026-01-01T00:00:00Z --period-end 2026-12-31T00:00:00Z`,
		Args: cobra.ExactArgs(1),
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
			}

			if flagName == "" {
				return fmt.Errorf("name is required; pass --name or run interactively")
			}

			input := map[string]any{
				"riskAnalysisId": args[0],
				"name":           flagName,
			}

			if flagDescription != "" {
				input["description"] = flagDescription
			}

			if flagPeriodStart != "" || flagPeriodEnd != "" {
				period := map[string]any{}

				if flagPeriodStart != "" {
					if _, err := time.Parse(time.RFC3339, flagPeriodStart); err != nil {
						return fmt.Errorf(
							"--period-start must be RFC3339 (e.g. 2026-01-01T00:00:00Z): %w",
							err,
						)
					}

					period["start"] = flagPeriodStart
				}

				if flagPeriodEnd != "" {
					if _, err := time.Parse(time.RFC3339, flagPeriodEnd); err != nil {
						return fmt.Errorf(
							"--period-end must be RFC3339 (e.g. 2026-12-31T00:00:00Z): %w",
							err,
						)
					}

					period["end"] = flagPeriodEnd
				}

				input["period"] = period
			}

			data, err := client.Do(
				forkMutation,
				map[string]any{"input": input},
			)
			if err != nil {
				return err
			}

			var resp forkResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			r := resp.ForkRiskAnalysis.RiskAnalysisEdge.Node
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Forked risk analysis %s (%s)\n",
				r.ID,
				r.Name,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagName, "name", "", "Risk analysis name (required)")
	cmd.Flags().StringVar(&flagDescription, "description", "", "Risk analysis description")
	cmd.Flags().StringVar(
		&flagPeriodStart,
		"period-start",
		"",
		"Period start in RFC3339 format (e.g. 2026-01-01T00:00:00Z)",
	)
	cmd.Flags().StringVar(
		&flagPeriodEnd,
		"period-end",
		"",
		"Period end in RFC3339 format (e.g. 2026-12-31T00:00:00Z)",
	)

	return cmd
}
