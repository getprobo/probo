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
mutation($input: CreateAssetInput!) {
  createAsset(input: $input) {
    assetEdge {
      node {
        id
        name
        assetType
        amount
      }
    }
  }
}
`

type createResponse struct {
	CreateAsset struct {
		AssetEdge struct {
			Node struct {
				ID        string `json:"id"`
				Name      string `json:"name"`
				AssetType string `json:"assetType"`
				Amount    int    `json:"amount"`
			} `json:"node"`
		} `json:"assetEdge"`
	} `json:"createAsset"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg             string
		flagName            string
		flagAssetType       string
		flagAmount          int
		flagOwner           string
		flagDataTypesStored string
		flagVendorIDs       []string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new asset",
		Example: `  # Create an asset interactively
  prb asset create

  # Create an asset non-interactively
  prb asset create --name "Production Database" --asset-type VIRTUAL --amount 50000`,
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
						Title("Asset name").
						Value(&flagName).
						Run()
					if err != nil {
						return err
					}
				}

				if flagAssetType == "" {
					err := huh.NewSelect[string]().
						Title("Asset type").
						Options(
							huh.NewOption("Physical", "PHYSICAL"),
							huh.NewOption("Virtual", "VIRTUAL"),
						).
						Value(&flagAssetType).
						Run()
					if err != nil {
						return err
					}
				}
			}

			if flagName == "" {
				return fmt.Errorf("name is required; pass --name or run interactively")
			}
			if flagAssetType == "" {
				return fmt.Errorf("asset type is required; pass --asset-type or run interactively")
			}

			input := map[string]any{
				"organizationId": flagOrg,
				"name":           flagName,
				"assetType":      flagAssetType,
			}

			if cmd.Flags().Changed("amount") {
				input["amount"] = flagAmount
			}
			if flagOwner != "" {
				input["ownerId"] = flagOwner
			}
			if flagDataTypesStored != "" {
				input["dataTypesStored"] = flagDataTypesStored
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

			a := resp.CreateAsset.AssetEdge.Node
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Created asset %s (%s)\n",
				a.ID,
				a.Name,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().StringVar(&flagName, "name", "", "Asset name (required)")
	cmd.Flags().StringVar(&flagAssetType, "asset-type", "", "Asset type: PHYSICAL, VIRTUAL (required)")
	cmd.Flags().IntVar(&flagAmount, "amount", 0, "Asset amount")
	cmd.Flags().StringVar(&flagOwner, "owner", "", "Owner profile ID")
	cmd.Flags().StringVar(&flagDataTypesStored, "data-types-stored", "", "Data types stored")
	cmd.Flags().StringSliceVar(&flagVendorIDs, "vendor-ids", nil, "Vendor IDs (comma-separated)")

	return cmd
}
