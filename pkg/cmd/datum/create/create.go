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
		flagVendorIDs      []string
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
			if len(flagVendorIDs) > 0 {
				input["vendorIds"] = flagVendorIDs
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
	cmd.Flags().StringSliceVar(&flagVendorIDs, "vendor-ids", nil, "Vendor IDs (comma-separated)")

	return cmd
}
