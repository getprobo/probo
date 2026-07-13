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
mutation($input: CreateDatumInput!) {
  createDatum(input: $input) {
    datumEdge {
      node {
        id
        name
        dataClassification
      }
    }
  }
}
`

type createResponse struct {
	CreateDatum struct {
		DatumEdge struct {
			Node struct {
				ID                 string `json:"id"`
				Name               string `json:"name"`
				DataClassification string `json:"dataClassification"`
			} `json:"node"`
		} `json:"datumEdge"`
	} `json:"createDatum"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg            string
		flagName           string
		flagClassification string
		flagOwner          string
		flagThirdPartyIDs  []string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new datum",
		Example: `  # Create a datum interactively
  prb datum create

  # Create a datum non-interactively
  prb datum create --name "Customer PII" --data-classification CONFIDENTIAL`,
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
						Title("Datum name").
						Value(&flagName).
						Run()
					if err != nil {
						return err
					}
				}

				if flagClassification == "" {
					err := huh.NewSelect[string]().
						Title("Data classification").
						Options(
							huh.NewOption("Public", "PUBLIC"),
							huh.NewOption("Internal", "INTERNAL"),
							huh.NewOption("Confidential", "CONFIDENTIAL"),
							huh.NewOption("Secret", "SECRET"),
						).
						Value(&flagClassification).
						Run()
					if err != nil {
						return err
					}
				}
			}

			if flagName == "" {
				return fmt.Errorf("name is required; pass --name or run interactively")
			}

			if flagClassification == "" {
				return fmt.Errorf("data classification is required; pass --data-classification or run interactively")
			}

			input := map[string]any{
				"organizationId":     flagOrg,
				"name":               flagName,
				"dataClassification": flagClassification,
			}

			if flagOwner != "" {
				input["ownerId"] = flagOwner
			}

			if len(flagThirdPartyIDs) > 0 {
				input["thirdPartyIds"] = flagThirdPartyIDs
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

			d := resp.CreateDatum.DatumEdge.Node
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Created datum %s (%s)\n",
				d.ID,
				d.Name,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().StringVar(&flagName, "name", "", "Datum name (required)")
	cmd.Flags().StringVar(&flagClassification, "data-classification", "", "Data classification: PUBLIC, INTERNAL, CONFIDENTIAL, SECRET (required)")
	cmd.Flags().StringVar(&flagOwner, "owner", "", "Owner profile ID")
	cmd.Flags().StringSliceVar(&flagThirdPartyIDs, "third-party-ids", nil, "ThirdParty IDs (comma-separated)")

	return cmd
}
