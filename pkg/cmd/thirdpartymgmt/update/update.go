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
mutation($input: UpdateVendorInput!) {
  updateVendor(input: $input) {
    vendor {
      id
      name
      category
    }
  }
}
`

type updateResponse struct {
	UpdateVendor struct {
		Vendor struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Category string `json:"category"`
		} `json:"vendor"`
	} `json:"updateVendor"`
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
		Short: "Update a vendor",
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

			v := resp.UpdateVendor.Vendor
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Updated vendor %s (%s)\n",
				v.ID,
				v.Name,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagName, "name", "", "Vendor name")
	cmd.Flags().StringVar(&flagDescription, "description", "", "Vendor description")
	cmd.Flags().StringVar(&flagCategory, "category", "", "Vendor category")
	cmd.Flags().StringVar(&flagLegalName, "legal-name", "", "Legal name")
	cmd.Flags().StringVar(&flagAddress, "address", "", "Headquarter address")
	cmd.Flags().StringVar(&flagWebsite, "website", "", "Website URL")

	return cmd
}
