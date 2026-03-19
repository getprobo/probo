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
mutation($input: UpdateOrganizationContextInput!) {
  updateOrganizationContext(input: $input) {
    context {
      organizationId
      product
      architecture
      team
      processes
      customers
    }
  }
}
`

type updateResponse struct {
	UpdateOrganizationContext struct {
		Context struct {
			OrganizationID string  `json:"organizationId"`
			Product        *string `json:"product"`
			Architecture   *string `json:"architecture"`
			Team           *string `json:"team"`
			Processes      *string `json:"processes"`
			Customers      *string `json:"customers"`
		} `json:"context"`
	} `json:"updateOrganizationContext"`
}

func NewCmdUpdate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg          string
		flagProduct      string
		flagArchitecture string
		flagTeam         string
		flagProcesses    string
		flagCustomers    string
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update organization context",
		Example: `  prb context update --org <org-id> --product "We build compliance software"
  prb context update --org <org-id> --architecture "Monolith deployed on AWS ECS"`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}

			host, hc, err := cfg.DefaultHost()
			if err != nil {
				return err
			}

			orgID := flagOrg
			if orgID == "" {
				orgID = hc.Organization
			}
			if orgID == "" {
				return fmt.Errorf("organization ID is required: pass --org or run `prb auth login`")
			}

			input := map[string]any{
				"organizationId": orgID,
			}

			if cmd.Flags().Changed("product") {
				input["product"] = flagProduct
			}
			if cmd.Flags().Changed("architecture") {
				input["architecture"] = flagArchitecture
			}
			if cmd.Flags().Changed("team") {
				input["team"] = flagTeam
			}
			if cmd.Flags().Changed("processes") {
				input["processes"] = flagProcesses
			}
			if cmd.Flags().Changed("customers") {
				input["customers"] = flagCustomers
			}

			if len(input) == 1 {
				return fmt.Errorf("at least one section flag is required (--product, --architecture, --team, --processes, --customers)")
			}

			client := api.NewClient(
				host,
				hc.Token,
				"/api/console/v1/graphql",
				cfg.HTTPTimeoutDuration(),
			)

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

			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Updated context for organization %s\n",
				resp.UpdateOrganizationContext.Context.OrganizationID,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().StringVar(&flagProduct, "product", "", "Product description (markdown)")
	cmd.Flags().StringVar(&flagArchitecture, "architecture", "", "Architecture description (markdown)")
	cmd.Flags().StringVar(&flagTeam, "team", "", "Team description (markdown)")
	cmd.Flags().StringVar(&flagProcesses, "processes", "", "Processes description (markdown)")
	cmd.Flags().StringVar(&flagCustomers, "customers", "", "Customers description (markdown)")

	return cmd
}
