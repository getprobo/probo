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

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const createMutation = `
mutation($input: CreateRiskInput!) {
  createRisk(input: $input) {
    riskEdge {
      node {
        id
        name
        category
        treatment
        inherentLikelihood
        inherentImpact
        inherentRiskScore
        residualLikelihood
        residualImpact
        residualRiskScore
      }
    }
  }
}
`

type createResponse struct {
	CreateRisk struct {
		RiskEdge struct {
			Node struct {
				ID                  string `json:"id"`
				Name                string `json:"name"`
				Category            string `json:"category"`
				Treatment           string `json:"treatment"`
				InherentLikelihood  int    `json:"inherentLikelihood"`
				InherentImpact      int    `json:"inherentImpact"`
				InherentRiskScore   int    `json:"inherentRiskScore"`
				ResidualLikelihood  int    `json:"residualLikelihood"`
				ResidualImpact      int    `json:"residualImpact"`
				ResidualRiskScore   int    `json:"residualRiskScore"`
			} `json:"node"`
		} `json:"riskEdge"`
	} `json:"createRisk"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg                string
		flagName               string
		flagCategory           string
		flagTreatment          string
		flagInherentLikelihood int
		flagInherentImpact     int
		flagResidualLikelihood int
		flagResidualImpact     int
		flagDescription        string
		flagNote               string
		flagOwner              string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new risk",
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
				cfg.HTTPTimeoutDuration(),
			)

			if flagOrg == "" {
				flagOrg = hc.Organization
			}

			if flagOrg == "" {
				return fmt.Errorf("organization is required; pass --org or set a default with 'proboctl auth login'")
			}

			input := map[string]any{
				"organizationId":     flagOrg,
				"name":               flagName,
				"category":           flagCategory,
				"treatment":          flagTreatment,
				"inherentLikelihood": flagInherentLikelihood,
				"inherentImpact":     flagInherentImpact,
			}

			if flagDescription != "" {
				input["description"] = flagDescription
			}
			if flagNote != "" {
				input["note"] = flagNote
			}
			if flagOwner != "" {
				input["ownerId"] = flagOwner
			}
			if cmd.Flags().Changed("residual-likelihood") {
				input["residualLikelihood"] = flagResidualLikelihood
			}
			if cmd.Flags().Changed("residual-impact") {
				input["residualImpact"] = flagResidualImpact
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

			r := resp.CreateRisk.RiskEdge.Node
			fmt.Fprintf(
				f.IOStreams.Out,
				"Created risk %s (%s)\n",
				r.ID,
				r.Name,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().StringVar(&flagName, "name", "", "Risk name (required)")
	cmd.Flags().StringVar(&flagCategory, "category", "", "Risk category (required)")
	cmd.Flags().StringVar(&flagTreatment, "treatment", "", "Risk treatment: MITIGATED, ACCEPTED, AVOIDED, TRANSFERRED (required)")
	cmd.Flags().IntVar(&flagInherentLikelihood, "inherent-likelihood", 0, "Inherent likelihood 1-5 (required)")
	cmd.Flags().IntVar(&flagInherentImpact, "inherent-impact", 0, "Inherent impact 1-5 (required)")
	cmd.Flags().IntVar(&flagResidualLikelihood, "residual-likelihood", 0, "Residual likelihood 1-5")
	cmd.Flags().IntVar(&flagResidualImpact, "residual-impact", 0, "Residual impact 1-5")
	cmd.Flags().StringVar(&flagDescription, "description", "", "Risk description")
	cmd.Flags().StringVar(&flagNote, "note", "", "Risk note")
	cmd.Flags().StringVar(&flagOwner, "owner", "", "Owner profile ID")

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("category")
	_ = cmd.MarkFlagRequired("treatment")
	_ = cmd.MarkFlagRequired("inherent-likelihood")
	_ = cmd.MarkFlagRequired("inherent-impact")

	return cmd
}
