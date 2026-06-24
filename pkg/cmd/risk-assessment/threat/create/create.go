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
mutation($input: CreateRiskAssessmentThreatInput!) {
  createRiskAssessmentThreat(input: $input) {
    riskAssessmentThreatEdge {
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
  }
}
`

type createResponse struct {
	CreateRiskAssessmentThreat struct {
		RiskAssessmentThreatEdge struct {
			Node struct {
				ID                    string `json:"id"`
				RiskAssessmentScopeId string `json:"riskAssessmentScopeId"`
				ProcessId             string `json:"processId"`
				Name                  string `json:"name"`
				Category              string `json:"category"`
				CreatedAt             string `json:"createdAt"`
				UpdatedAt             string `json:"updatedAt"`
			} `json:"node"`
		} `json:"riskAssessmentThreatEdge"`
	} `json:"createRiskAssessmentThreat"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagScopeId   string
		flagProcessId string
		flagName      string
		flagCategory  string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new risk assessment threat",
		Example: `  # Create a threat interactively
  prb risk-assessment threat create --scope-id <id> --process-id <id>

  # Create a threat non-interactively
  prb risk-assessment threat create --scope-id <id> --process-id <id> --name "SQL injection" --category "Application"`,
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
						Title("Threat name").
						Value(&flagName).
						Run()
					if err != nil {
						return err
					}
				}

				if flagCategory == "" {
					err := huh.NewInput().
						Title("Threat category").
						Value(&flagCategory).
						Run()
					if err != nil {
						return err
					}
				}
			}

			if flagName == "" {
				return fmt.Errorf("name is required; pass --name or run interactively")
			}

			if flagCategory == "" {
				return fmt.Errorf("category is required; pass --category or run interactively")
			}

			input := map[string]any{
				"riskAssessmentScopeId": flagScopeId,
				"processId":             flagProcessId,
				"name":                  flagName,
				"category":              flagCategory,
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

			r := resp.CreateRiskAssessmentThreat.RiskAssessmentThreatEdge.Node
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Created risk assessment threat %s (%s)\n",
				r.ID,
				r.Name,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagScopeId, "scope-id", "", "Risk assessment scope ID (required)")
	cmd.Flags().StringVar(&flagProcessId, "process-id", "", "Process ID (required)")
	cmd.Flags().StringVar(&flagName, "name", "", "Threat name (required)")
	cmd.Flags().StringVar(&flagCategory, "category", "", "Threat category (required)")

	_ = cmd.MarkFlagRequired("scope-id")
	_ = cmd.MarkFlagRequired("process-id")

	return cmd
}
