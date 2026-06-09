// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
