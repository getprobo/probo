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

package update

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const updateMutation = `
mutation RegistryAssetUpdate($input: UpdateRegistryAssetInput!) {
  updateRegistryAsset(input: $input) {
    registryAsset {
      id
      name
      registryAssetType
      amount
    }
  }
}
`

type updateResponse struct {
	UpdateRegistryAsset struct {
		RegistryAsset struct {
			ID                string `json:"id"`
			Name              string `json:"name"`
			RegistryAssetType string `json:"registryAssetType"`
			Amount            int    `json:"amount"`
		} `json:"registryAsset"`
	} `json:"updateRegistryAsset"`
}

func NewCmdUpdate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagName              string
		flagRegistryAssetType string
		flagAmount            int
		flagOwner             string
		flagDataTypesStored   string
		flagThirdPartyIDs     []string
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

			if cmd.Flags().Changed("registry-asset-type") {
				input["registryAssetType"] = flagRegistryAssetType
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

			if cmd.Flags().Changed("third-party-ids") {
				input["thirdPartyIds"] = flagThirdPartyIDs
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

			a := resp.UpdateRegistryAsset.RegistryAsset
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Updated asset %s (%s)\n",
				a.ID,
				a.Name,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagName, "name", "", "Registry asset name")
	cmd.Flags().StringVar(&flagRegistryAssetType, "registry-asset-type", "", "Registry asset type: PHYSICAL, VIRTUAL")
	cmd.Flags().IntVar(&flagAmount, "amount", 0, "Registry asset amount")
	cmd.Flags().StringVar(&flagOwner, "owner", "", "Owner profile ID")
	cmd.Flags().StringVar(&flagDataTypesStored, "data-types-stored", "", "Data types stored")
	cmd.Flags().StringSliceVar(&flagThirdPartyIDs, "third-party-ids", nil, "Third party IDs (comma-separated)")

	return cmd
}
