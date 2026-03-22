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
mutation($input: CreateCookieBannerInput!) {
  createCookieBanner(input: $input) {
    cookieBanner {
      id
      name
      domain
      state
    }
  }
}
`

type createResponse struct {
	CreateCookieBanner struct {
		CookieBanner struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Domain string `json:"domain"`
			State  string `json:"state"`
		} `json:"cookieBanner"`
	} `json:"createCookieBanner"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg    string
		flagName   string
		flagDomain string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new cookie banner",
		Example: `  # Create a cookie banner interactively
  prb cookie-banner create

  # Create a cookie banner non-interactively
  prb cookie-banner create --name "Main Site" --domain "example.com"`,
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
						Title("Banner name").
						Value(&flagName).
						Run()
					if err != nil {
						return err
					}
				}

				if flagDomain == "" {
					err := huh.NewInput().
						Title("Domain").
						Value(&flagDomain).
						Run()
					if err != nil {
						return err
					}
				}
			}

			if flagName == "" {
				return fmt.Errorf("name is required; pass --name or run interactively")
			}

			input := map[string]any{
				"organizationId": flagOrg,
				"name":           flagName,
			}

			if flagDomain != "" {
				input["domain"] = flagDomain
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

			b := resp.CreateCookieBanner.CookieBanner
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Created cookie banner %s (%s)\n",
				b.ID,
				b.Name,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().StringVar(&flagName, "name", "", "Banner name (required)")
	cmd.Flags().StringVar(&flagDomain, "domain", "", "Domain the banner is deployed on")

	return cmd
}
