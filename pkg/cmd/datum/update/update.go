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
mutation($input: UpdateDatumInput!) {
  updateDatum(input: $input) {
    datum {
      id
      name
      dataClassification
    }
  }
}
`

type updateResponse struct {
	UpdateDatum struct {
		Datum struct {
			ID                 string `json:"id"`
			Name               string `json:"name"`
			DataClassification string `json:"dataClassification"`
		} `json:"datum"`
	} `json:"updateDatum"`
}

func NewCmdUpdate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagName           string
		flagClassification string
		flagOwner          string
		flagThirdPartyIDs  []string
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a datum",
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

			if cmd.Flags().Changed("data-classification") {
				input["dataClassification"] = flagClassification
			}

			if cmd.Flags().Changed("owner") {
				if flagOwner == "" {
					input["ownerId"] = nil
				} else {
					input["ownerId"] = flagOwner
				}
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

			d := resp.UpdateDatum.Datum
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Updated datum %s (%s)\n",
				d.ID,
				d.Name,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagName, "name", "", "Datum name")
	cmd.Flags().StringVar(&flagClassification, "data-classification", "", "Data classification: PUBLIC, INTERNAL, CONFIDENTIAL, SECRET")
	cmd.Flags().StringVar(&flagOwner, "owner", "", "Owner profile ID")
	cmd.Flags().StringSliceVar(&flagThirdPartyIDs, "third-party-ids", nil, "ThirdParty IDs (comma-separated)")

	return cmd
}
