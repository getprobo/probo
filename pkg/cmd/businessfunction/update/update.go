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
mutation($input: UpdateBusinessFunctionInput!) {
  updateBusinessFunction(input: $input) {
    businessFunction {
      id
      referenceId
      name
      classification
      mtdMinutes
      rtoMinutes
      rpoMinutes
    }
  }
}
`

type updateResponse struct {
	UpdateBusinessFunction struct {
		BusinessFunction struct {
			ID             string `json:"id"`
			ReferenceID    string `json:"referenceId"`
			Name           string `json:"name"`
			Classification string `json:"classification"`
			MTDMinutes     int    `json:"mtdMinutes"`
			RTOMinutes     int    `json:"rtoMinutes"`
			RPOMinutes     int    `json:"rpoMinutes"`
		} `json:"businessFunction"`
	} `json:"updateBusinessFunction"`
}

func NewCmdUpdate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagReferenceID     string
		flagName            string
		flagClassification  string
		flagMTDMinutes      int
		flagRTOMinutes      int
		flagRPOMinutes      int
		flagImpactTolerance string
		flagNotes           string
		flagOwner           string
		flagAssetIDs        []string
		flagThirdPartyIDs   []string
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a business function",
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

			if cmd.Flags().Changed("reference-id") {
				input["referenceId"] = flagReferenceID
			}

			if cmd.Flags().Changed("name") {
				input["name"] = flagName
			}

			if cmd.Flags().Changed("classification") {
				if err := cmdutil.ValidateEnum(
					"classification",
					flagClassification,
					[]string{"CRITICAL", "IMPORTANT", "SECONDARY", "STANDARD"},
				); err != nil {
					return err
				}

				input["classification"] = flagClassification
			}

			if cmd.Flags().Changed("mtd-minutes") {
				input["mtdMinutes"] = flagMTDMinutes
			}

			if cmd.Flags().Changed("rto-minutes") {
				input["rtoMinutes"] = flagRTOMinutes
			}

			if cmd.Flags().Changed("rpo-minutes") {
				input["rpoMinutes"] = flagRPOMinutes
			}

			if cmd.Flags().Changed("impact-tolerance") {
				input["impactTolerance"] = flagImpactTolerance
			}

			if cmd.Flags().Changed("notes") {
				input["notes"] = flagNotes
			}

			if cmd.Flags().Changed("owner") {
				if flagOwner == "" {
					input["ownerId"] = nil
				} else {
					input["ownerId"] = flagOwner
				}
			}

			if cmd.Flags().Changed("asset") {
				input["assetIds"] = flagAssetIDs
			}

			if cmd.Flags().Changed("third-party") {
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

			bf := resp.UpdateBusinessFunction.BusinessFunction
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Updated business function %s\n",
				bf.ID,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagReferenceID, "reference-id", "", "Reference ID")
	cmd.Flags().StringVar(&flagName, "name", "", "Business function name")
	cmd.Flags().StringVar(&flagClassification, "classification", "", "Classification: CRITICAL, IMPORTANT, SECONDARY, STANDARD")
	cmd.Flags().IntVar(&flagMTDMinutes, "mtd-minutes", 0, "Maximum tolerable downtime in minutes")
	cmd.Flags().IntVar(&flagRTOMinutes, "rto-minutes", 0, "Recovery time objective in minutes")
	cmd.Flags().IntVar(&flagRPOMinutes, "rpo-minutes", 0, "Recovery point objective in minutes")
	cmd.Flags().StringVar(&flagImpactTolerance, "impact-tolerance", "", "Impact tolerance")
	cmd.Flags().StringVar(&flagNotes, "notes", "", "Notes")
	cmd.Flags().StringVar(&flagOwner, "owner", "", "Owner profile ID")
	cmd.Flags().StringArrayVar(&flagAssetIDs, "asset", nil, "Asset ID (can be repeated)")
	cmd.Flags().StringArrayVar(&flagThirdPartyIDs, "third-party", nil, "Third party ID (can be repeated)")

	return cmd
}
