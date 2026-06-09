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
mutation($input: CreateRiskAssessmentBoundaryInput!) {
  createRiskAssessmentBoundary(input: $input) {
    riskAssessmentBoundaryEdge {
      node {
        id
        riskAssessmentScopeId
        parentBoundaryId
        name
        createdAt
        updatedAt
      }
    }
  }
}
`

type createResponse struct {
	CreateRiskAssessmentBoundary struct {
		RiskAssessmentBoundaryEdge struct {
			Node struct {
				ID                    string  `json:"id"`
				RiskAssessmentScopeId string  `json:"riskAssessmentScopeId"`
				ParentBoundaryId      *string `json:"parentBoundaryId"`
				Name                  string  `json:"name"`
				CreatedAt             string  `json:"createdAt"`
				UpdatedAt             string  `json:"updatedAt"`
			} `json:"node"`
		} `json:"riskAssessmentBoundaryEdge"`
	} `json:"createRiskAssessmentBoundary"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagScopeId  string
		flagParentId string
		flagName     string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new risk assessment boundary",
		Example: `  # Create a boundary interactively
  prb risk-assessment boundary create --scope-id <id>

  # Create a boundary non-interactively
  prb risk-assessment boundary create --scope-id <id> --name "Production environment"

  # Create a boundary nested inside another boundary
  prb risk-assessment boundary create --scope-id <id> --name "Database tier" --parent-id <boundary-id>`,
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
						Title("Boundary name").
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
				"riskAssessmentScopeId": flagScopeId,
				"name":                  flagName,
			}

			if flagParentId != "" {
				input["parentBoundaryId"] = flagParentId
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

			r := resp.CreateRiskAssessmentBoundary.RiskAssessmentBoundaryEdge.Node
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Created risk assessment boundary %s (%s)\n",
				r.ID,
				r.Name,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagScopeId, "scope-id", "", "Risk assessment scope ID (required)")
	cmd.Flags().StringVar(&flagParentId, "parent-id", "", "Parent boundary ID (optional, for nested boundaries)")
	cmd.Flags().StringVar(&flagName, "name", "", "Boundary name (required)")

	_ = cmd.MarkFlagRequired("scope-id")

	return cmd
}
