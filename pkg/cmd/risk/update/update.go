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

package update

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const updateMutation = `
mutation($input: UpdateRiskInput!) {
  updateRisk(input: $input) {
    risk {
      id
      name
      category
      treatment
      inherentRiskScore
      residualRiskScore
    }
  }
}
`

type updateResponse struct {
	UpdateRisk struct {
		Risk struct {
			ID                string `json:"id"`
			Name              string `json:"name"`
			Category          string `json:"category"`
			Treatment         string `json:"treatment"`
			InherentRiskScore int    `json:"inherentRiskScore"`
			ResidualRiskScore int    `json:"residualRiskScore"`
		} `json:"risk"`
	} `json:"updateRisk"`
}

func NewCmdUpdate(f *cmdutil.Factory) *cobra.Command {
	var (
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
		Use:   "update <id>",
		Short: "Update a risk",
		Args:  cobra.ExactArgs(1),
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
			)

			input := map[string]any{
				"id": args[0],
			}

			if cmd.Flags().Changed("name") {
				input["name"] = flagName
			}
			if cmd.Flags().Changed("category") {
				input["category"] = flagCategory
			}
			if cmd.Flags().Changed("treatment") {
				input["treatment"] = flagTreatment
			}
			if cmd.Flags().Changed("inherent-likelihood") {
				input["inherentLikelihood"] = flagInherentLikelihood
			}
			if cmd.Flags().Changed("inherent-impact") {
				input["inherentImpact"] = flagInherentImpact
			}
			if cmd.Flags().Changed("residual-likelihood") {
				input["residualLikelihood"] = flagResidualLikelihood
			}
			if cmd.Flags().Changed("residual-impact") {
				input["residualImpact"] = flagResidualImpact
			}
			if cmd.Flags().Changed("description") {
				input["description"] = flagDescription
			}
			if cmd.Flags().Changed("note") {
				input["note"] = flagNote
			}
			if cmd.Flags().Changed("owner") {
				if flagOwner == "" {
					input["ownerId"] = nil
				} else {
					input["ownerId"] = flagOwner
				}
			}

			if len(input) == 1 {
				return fmt.Errorf("at least one field must be specified for update")
			}

			data, err := client.Do(
				updateMutation,
				map[string]any{"input": input},
			)
			if err != nil {
				return err
			}

			var resp updateResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			r := resp.UpdateRisk.Risk
			fmt.Fprintf(
				f.IOStreams.Out,
				"Updated risk %s (%s)\n",
				r.ID,
				r.Name,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagName, "name", "", "Risk name")
	cmd.Flags().StringVar(&flagCategory, "category", "", "Risk category")
	cmd.Flags().StringVar(&flagTreatment, "treatment", "", "Risk treatment: MITIGATED, ACCEPTED, AVOIDED, TRANSFERRED")
	cmd.Flags().IntVar(&flagInherentLikelihood, "inherent-likelihood", 0, "Inherent likelihood 1-5")
	cmd.Flags().IntVar(&flagInherentImpact, "inherent-impact", 0, "Inherent impact 1-5")
	cmd.Flags().IntVar(&flagResidualLikelihood, "residual-likelihood", 0, "Residual likelihood 1-5")
	cmd.Flags().IntVar(&flagResidualImpact, "residual-impact", 0, "Residual impact 1-5")
	cmd.Flags().StringVar(&flagDescription, "description", "", "Risk description")
	cmd.Flags().StringVar(&flagNote, "note", "", "Risk note")
	cmd.Flags().StringVar(&flagOwner, "owner", "", "Owner profile ID")

	return cmd
}
