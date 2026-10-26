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
		flagThirdPartyIDs   []string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new asset",
		Example: `  # Create an asset interactively
  prb asset create

  # Create an asset non-interactively
  prb asset create --name "Production Database" --asset-type VIRTUAL --amount 1 --owner <profile-id> --data-types-stored "Customer records"`,
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

			flagOrg, err = cmdutil.ResolveOrganization(flagOrg, hc.Organization)
			if err != nil {
				return err
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

			if flagAmount < 1 {
				return fmt.Errorf("amount is required; pass --amount")
			}

			if flagOwner == "" {
				return fmt.Errorf("owner is required; pass --owner or run interactively")
			}

			if flagDataTypesStored == "" {
				return fmt.Errorf("data types stored is required; pass --data-types-stored or run interactively")
			}

			input := map[string]any{
				"organizationId":  flagOrg,
				"name":            flagName,
				"assetType":       flagAssetType,
				"amount":          flagAmount,
				"ownerId":         flagOwner,
				"dataTypesStored": flagDataTypesStored,
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

	cmdutil.AddOrganizationFlag(cmd, &flagOrg, "Organization ID")
	cmd.Flags().StringVar(&flagName, "name", "", "Asset name (required)")
	cmd.Flags().StringVar(&flagAssetType, "asset-type", "", "Asset type: PHYSICAL, VIRTUAL (required)")
	cmd.Flags().IntVar(&flagAmount, "amount", 0, "Asset amount (required, at least 1)")
	cmd.Flags().StringVar(&flagOwner, "owner", "", "Owner profile ID (required)")
	cmd.Flags().StringVar(&flagDataTypesStored, "data-types-stored", "", "Data types stored (required)")
	cmd.Flags().StringSliceVar(&flagThirdPartyIDs, "third-party-ids", nil, "Third party IDs (comma-separated)")

	return cmd
}
