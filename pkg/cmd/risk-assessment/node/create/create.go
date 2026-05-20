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
mutation($input: CreateRiskAssessmentNodeInput!) {
  createRiskAssessmentNode(input: $input) {
    riskAssessmentNodeEdge {
      node {
        id
        riskAssessmentScopeId
        nodeType
        name
        createdAt
        updatedAt
      }
    }
  }
}
`

type createResponse struct {
	CreateRiskAssessmentNode struct {
		RiskAssessmentNodeEdge struct {
			Node struct {
				ID                    string `json:"id"`
				RiskAssessmentScopeId string `json:"riskAssessmentScopeId"`
				NodeType              string `json:"nodeType"`
				Name                  string `json:"name"`
				CreatedAt             string `json:"createdAt"`
				UpdatedAt             string `json:"updatedAt"`
			} `json:"node"`
		} `json:"riskAssessmentNodeEdge"`
	} `json:"createRiskAssessmentNode"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagScopeId  string
		flagNodeType string
		flagName     string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new risk assessment node",
		Example: `  # Create a node interactively
  prb risk-assessment node create --scope-id <id>

  # Create a node non-interactively
  prb risk-assessment node create --scope-id <id> --node-type ASSET --name "Database server"`,
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
						Title("Node name").
						Value(&flagName).
						Run()
					if err != nil {
						return err
					}
				}

				if flagNodeType == "" {
					err := huh.NewSelect[string]().
						Title("Node type").
						Options(
							huh.NewOption("Entity", "ENTITY"),
							huh.NewOption("Boundary", "BOUNDARY"),
							huh.NewOption("Asset", "ASSET"),
							huh.NewOption("Data", "DATA"),
						).
						Value(&flagNodeType).
						Run()
					if err != nil {
						return err
					}
				}
			}

			if flagName == "" {
				return fmt.Errorf("name is required; pass --name or run interactively")
			}

			if flagNodeType == "" {
				return fmt.Errorf("node type is required; pass --node-type or run interactively")
			}

			if err := cmdutil.ValidateEnum("node-type", flagNodeType, []string{"ENTITY", "BOUNDARY", "ASSET", "DATA"}); err != nil {
				return err
			}

			input := map[string]any{
				"riskAssessmentScopeId": flagScopeId,
				"nodeType":              flagNodeType,
				"name":                  flagName,
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

			r := resp.CreateRiskAssessmentNode.RiskAssessmentNodeEdge.Node
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Created risk assessment node %s (%s)\n",
				r.ID,
				r.Name,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagScopeId, "scope-id", "", "Risk assessment scope ID (required)")
	cmd.Flags().StringVar(&flagNodeType, "node-type", "", "Node type: ENTITY, BOUNDARY, ASSET, DATA (required)")
	cmd.Flags().StringVar(&flagName, "name", "", "Node name (required)")

	_ = cmd.MarkFlagRequired("scope-id")

	return cmd
}
