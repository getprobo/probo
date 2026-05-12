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
    cookieBannerEdge {
      node {
        id
        name
        origin
      }
    }
  }
}
`

type createResponse struct {
	CreateCookieBanner struct {
		CookieBannerEdge struct {
			Node struct {
				ID     string `json:"id"`
				Name   string `json:"name"`
				Origin string `json:"origin"`
			} `json:"node"`
		} `json:"cookieBannerEdge"`
	} `json:"createCookieBanner"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg              string
		flagName             string
		flagOrigin           string
		flagCookiePolicyUrl  string
		flagPrivacyPolicyUrl string
		flagConsentExpiry    int
		flagConsentMode      string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new cookie banner",
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
					if err := huh.NewInput().Title("Banner name").Value(&flagName).Run(); err != nil {
						return err
					}
				}
				if flagOrigin == "" {
					if err := huh.NewInput().Title("Website origin (e.g. https://example.com)").Value(&flagOrigin).Run(); err != nil {
						return err
					}
				}
				if flagCookiePolicyUrl == "" {
					if err := huh.NewInput().Title("Cookie policy URL").Value(&flagCookiePolicyUrl).Run(); err != nil {
						return err
					}
				}
				if flagConsentMode == "" {
					if err := huh.NewSelect[string]().
						Title("Consent mode").
						Options(
							huh.NewOption("Opt-In", "OPT_IN"),
							huh.NewOption("Opt-Out", "OPT_OUT"),
						).
						Value(&flagConsentMode).Run(); err != nil {
						return err
					}
				}
			}

			if flagName == "" {
				return fmt.Errorf("name is required; pass --name or run interactively")
			}
			if flagOrigin == "" {
				return fmt.Errorf("origin is required; pass --origin or run interactively")
			}
			if flagCookiePolicyUrl == "" {
				return fmt.Errorf("cookie-policy-url is required; pass --cookie-policy-url or run interactively")
			}
			if flagConsentMode == "" {
				return fmt.Errorf("consent-mode is required; pass --consent-mode or run interactively")
			}

			input := map[string]any{
				"organizationId":    flagOrg,
				"name":              flagName,
				"origin":            flagOrigin,
				"cookiePolicyUrl":   flagCookiePolicyUrl,
				"consentExpiryDays": flagConsentExpiry,
				"consentMode":       flagConsentMode,
			}
			if flagPrivacyPolicyUrl != "" {
				input["privacyPolicyUrl"] = flagPrivacyPolicyUrl
			}

			data, err := client.Do(createMutation, map[string]any{"input": input})
			if err != nil {
				return err
			}

			var resp createResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			b := resp.CreateCookieBanner.CookieBannerEdge.Node
			_, _ = fmt.Fprintf(f.IOStreams.Out, "Created cookie banner %s (%s)\n", b.ID, b.Name)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().StringVar(&flagName, "name", "", "Banner name (required)")
	cmd.Flags().StringVar(&flagOrigin, "origin", "", "Website origin (required)")
	cmd.Flags().StringVar(&flagCookiePolicyUrl, "cookie-policy-url", "", "Cookie policy URL (required)")
	cmd.Flags().StringVar(&flagPrivacyPolicyUrl, "privacy-policy-url", "", "Privacy policy URL")
	cmd.Flags().IntVar(&flagConsentExpiry, "consent-expiry-days", 365, "Days until consent expires")
	cmd.Flags().StringVar(&flagConsentMode, "consent-mode", "", "Consent mode: OPT_IN or OPT_OUT (required)")

	return cmd
}
