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
mutation($input: UpdateThirdPartyInput!) {
  updateThirdParty(input: $input) {
    third_party {
      id
      name
      category
    }
  }
}
`

type updateResponse struct {
	UpdateThirdParty struct {
		ThirdParty struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Category string `json:"category"`
		} `json:"third_party"`
	} `json:"updateThirdParty"`
}

func NewCmdUpdate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagName        string
		flagDescription string
		flagCategory    string
		flagLegalName   string
		flagAddress     string
		flagWebsite     string
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a thirdParty",
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

			if cmd.Flags().Changed("description") {
				input["description"] = flagDescription
			}

			if cmd.Flags().Changed("category") {
				input["category"] = flagCategory
			}

			if cmd.Flags().Changed("legal-name") {
				input["legalName"] = flagLegalName
			}

			if cmd.Flags().Changed("address") {
				input["headquarterAddress"] = flagAddress
			}

			if cmd.Flags().Changed("website") {
				input["websiteUrl"] = flagWebsite
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

			v := resp.UpdateThirdParty.ThirdParty
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Updated thirdParty %s (%s)\n",
				v.ID,
				v.Name,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagName, "name", "", "ThirdParty name")
	cmd.Flags().StringVar(&flagDescription, "description", "", "ThirdParty description")
	cmd.Flags().StringVar(&flagCategory, "category", "", "ThirdParty category")
	cmd.Flags().StringVar(&flagLegalName, "legal-name", "", "Legal name")
	cmd.Flags().StringVar(&flagAddress, "address", "", "Headquarter address")
	cmd.Flags().StringVar(&flagWebsite, "website", "", "Website URL")

	return cmd
}
