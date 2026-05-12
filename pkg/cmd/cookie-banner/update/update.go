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
mutation($input: UpdateCookieBannerInput!) {
  updateCookieBanner(input: $input) {
    cookieBanner {
      id
      name
    }
  }
}
`

type updateResponse struct {
	UpdateCookieBanner struct {
		CookieBanner struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"cookieBanner"`
	} `json:"updateCookieBanner"`
}

func NewCmdUpdate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagName             string
		flagCookiePolicyUrl  string
		flagPrivacyPolicyUrl string
		flagConsentExpiry    int
		flagConsentMode      string
		flagDefaultLanguage  string
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a cookie banner",
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

			input := map[string]any{"cookieBannerId": args[0]}

			if cmd.Flags().Changed("name") {
				input["name"] = flagName
			}
			if cmd.Flags().Changed("cookie-policy-url") {
				input["cookiePolicyUrl"] = flagCookiePolicyUrl
			}
			if cmd.Flags().Changed("privacy-policy-url") {
				input["privacyPolicyUrl"] = flagPrivacyPolicyUrl
			}
			if cmd.Flags().Changed("consent-expiry-days") {
				input["consentExpiryDays"] = flagConsentExpiry
			}
			if cmd.Flags().Changed("consent-mode") {
				input["consentMode"] = flagConsentMode
			}
			if cmd.Flags().Changed("default-language") {
				input["defaultLanguage"] = flagDefaultLanguage
			}

			if len(input) == 1 {
				return fmt.Errorf("at least one field must be specified for update")
			}

			data, err := client.Do(updateMutation, map[string]any{"input": input})
			if err != nil {
				return err
			}

			var resp updateResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			b := resp.UpdateCookieBanner.CookieBanner
			_, _ = fmt.Fprintf(f.IOStreams.Out, "Updated cookie banner %s (%s)\n", b.ID, b.Name)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagName, "name", "", "Banner name")
	cmd.Flags().StringVar(&flagCookiePolicyUrl, "cookie-policy-url", "", "Cookie policy URL")
	cmd.Flags().StringVar(&flagPrivacyPolicyUrl, "privacy-policy-url", "", "Privacy policy URL")
	cmd.Flags().IntVar(&flagConsentExpiry, "consent-expiry-days", 0, "Days until consent expires")
	cmd.Flags().StringVar(&flagConsentMode, "consent-mode", "", "Consent mode: OPT_IN or OPT_OUT")
	cmd.Flags().StringVar(&flagDefaultLanguage, "default-language", "", "Default language code")

	return cmd
}
