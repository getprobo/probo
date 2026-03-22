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
      domain
      state
    }
  }
}
`

type updateResponse struct {
	UpdateCookieBanner struct {
		CookieBanner struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Domain string `json:"domain"`
			State  string `json:"state"`
		} `json:"cookieBanner"`
	} `json:"updateCookieBanner"`
}

func NewCmdUpdate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagName              string
		flagDomain            string
		flagTitle             string
		flagDescription       string
		flagAcceptAllLabel    string
		flagRejectAllLabel    string
		flagSavePrefsLabel    string
		flagPrivacyPolicyURL  string
		flagConsentExpiryDays int
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
			)

			input := map[string]any{
				"id": args[0],
			}

			if cmd.Flags().Changed("name") {
				input["name"] = flagName
			}
			if cmd.Flags().Changed("domain") {
				input["domain"] = flagDomain
			}
			if cmd.Flags().Changed("title") {
				input["title"] = flagTitle
			}
			if cmd.Flags().Changed("description") {
				input["description"] = flagDescription
			}
			if cmd.Flags().Changed("accept-all-label") {
				input["acceptAllLabel"] = flagAcceptAllLabel
			}
			if cmd.Flags().Changed("reject-all-label") {
				input["rejectAllLabel"] = flagRejectAllLabel
			}
			if cmd.Flags().Changed("save-preferences-label") {
				input["savePreferencesLabel"] = flagSavePrefsLabel
			}
			if cmd.Flags().Changed("privacy-policy-url") {
				input["privacyPolicyUrl"] = flagPrivacyPolicyURL
			}
			if cmd.Flags().Changed("consent-expiry-days") {
				input["consentExpiryDays"] = flagConsentExpiryDays
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

			b := resp.UpdateCookieBanner.CookieBanner
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Updated cookie banner %s (%s)\n",
				b.ID,
				b.Name,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagName, "name", "", "Banner name")
	cmd.Flags().StringVar(&flagDomain, "domain", "", "Domain")
	cmd.Flags().StringVar(&flagTitle, "title", "", "Banner title")
	cmd.Flags().StringVar(&flagDescription, "description", "", "Banner description")
	cmd.Flags().StringVar(&flagAcceptAllLabel, "accept-all-label", "", "Accept all button label")
	cmd.Flags().StringVar(&flagRejectAllLabel, "reject-all-label", "", "Reject all button label")
	cmd.Flags().StringVar(&flagSavePrefsLabel, "save-preferences-label", "", "Save preferences button label")
	cmd.Flags().StringVar(&flagPrivacyPolicyURL, "privacy-policy-url", "", "Privacy policy URL")
	cmd.Flags().IntVar(&flagConsentExpiryDays, "consent-expiry-days", 0, "Days until consent expires")

	return cmd
}
