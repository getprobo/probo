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

package view

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const viewQuery = `
query($id: ID!) {
  node(id: $id) {
    __typename
    ... on CookieConsentRecord {
      id
      visitorId
      ipAddress
      userAgent
      consentData
      action
      sdkVersion
      regulation
      countryCode
      createdAt
    }
  }
}
`

type viewResponse struct {
	Node *struct {
		Typename    string  `json:"__typename"`
		ID          string  `json:"id"`
		VisitorID   string  `json:"visitorId"`
		IPAddress   *string `json:"ipAddress"`
		UserAgent   *string `json:"userAgent"`
		ConsentData string  `json:"consentData"`
		Action      string  `json:"action"`
		SdkVersion  string  `json:"sdkVersion"`
		Regulation  string  `json:"regulation"`
		CountryCode string  `json:"countryCode"`
		CreatedAt   string  `json:"createdAt"`
	} `json:"node"`
}

func NewCmdView(f *cmdutil.Factory) *cobra.Command {
	var flagOutput *string

	cmd := &cobra.Command{
		Use:   "view <id>",
		Short: "View a consent record",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return err
			}

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

			data, err := client.Do(viewQuery, map[string]any{"id": args[0]})
			if err != nil {
				return err
			}

			var resp viewResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			if resp.Node == nil || resp.Node.Typename != "CookieConsentRecord" {
				return fmt.Errorf("consent record %s not found", args[0])
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, resp.Node)
			}

			v := resp.Node
			out := f.IOStreams.Out

			bold := lipgloss.NewStyle().Bold(true)
			label := lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Width(22)

			_, _ = fmt.Fprintf(out, "%s\n\n", bold.Render("Consent Record"))
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("ID:"), v.ID)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Visitor ID:"), v.VisitorID)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Action:"), v.Action)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("SDK Version:"), v.SdkVersion)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Regulation:"), v.Regulation)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Country Code:"), v.CountryCode)
			if v.IPAddress != nil && *v.IPAddress != "" {
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("IP Address:"), *v.IPAddress)
			}
			if v.UserAgent != nil && *v.UserAgent != "" {
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("User Agent:"), *v.UserAgent)
			}
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Consent Data:"), v.ConsentData)
			_, _ = fmt.Fprintln(out)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Created:"), cmdutil.FormatTime(v.CreatedAt))

			return nil
		},
	}

	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
