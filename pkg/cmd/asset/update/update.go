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
mutation($input: UpdateAssetInput!) {
  updateAsset(input: $input) {
    asset {
      id
      name
      assetType
      amount
    }
  }
}
`

type updateResponse struct {
	UpdateAsset struct {
		Asset struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			AssetType string `json:"assetType"`
			Amount    int    `json:"amount"`
		} `json:"asset"`
	} `json:"updateAsset"`
}

func NewCmdUpdate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagName            string
		flagAssetType       string
		flagAmount          int
		flagOwner           string
		flagDataTypesStored string
		flagVendorIDs       []string
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an asset",
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
				cmdutil.TokenRefreshOption(cfg, host, hc),
			)

			input := map[string]any{
				"id": args[0],
			}

			if cmd.Flags().Changed("name") {
				input["name"] = flagName
			}
			if cmd.Flags().Changed("asset-type") {
				input["assetType"] = flagAssetType
			}
			if cmd.Flags().Changed("amount") {
				input["amount"] = flagAmount
			}
			if cmd.Flags().Changed("owner") {
				if flagOwner == "" {
					input["ownerId"] = nil
				} else {
					input["ownerId"] = flagOwner
				}
			}
			if cmd.Flags().Changed("data-types-stored") {
				input["dataTypesStored"] = flagDataTypesStored
			}
			if cmd.Flags().Changed("vendor-ids") {
				input["vendorIds"] = flagVendorIDs
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

			a := resp.UpdateAsset.Asset
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Updated asset %s (%s)\n",
				a.ID,
				a.Name,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagName, "name", "", "Asset name")
	cmd.Flags().StringVar(&flagAssetType, "asset-type", "", "Asset type: PHYSICAL, VIRTUAL")
	cmd.Flags().IntVar(&flagAmount, "amount", 0, "Asset amount")
	cmd.Flags().StringVar(&flagOwner, "owner", "", "Owner profile ID")
	cmd.Flags().StringVar(&flagDataTypesStored, "data-types-stored", "", "Data types stored")
	cmd.Flags().StringSliceVar(&flagVendorIDs, "vendor-ids", nil, "Vendor IDs (comma-separated)")

	return cmd
}
