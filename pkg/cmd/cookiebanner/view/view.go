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
    ... on CookieBanner {
      id
      name
      domain
      state
      title
      description
      acceptAllLabel
      rejectAllLabel
      savePreferencesLabel
      privacyPolicyUrl
      consentExpiryDays
      consentMode
      version
      embedSnippet
      createdAt
      updatedAt
    }
  }
}
`

type viewResponse struct {
	Node *struct {
		Typename             string `json:"__typename"`
		ID                   string `json:"id"`
		Name                 string `json:"name"`
		Domain               string `json:"domain"`
		State                string `json:"state"`
		Title                string `json:"title"`
		Description          string `json:"description"`
		AcceptAllLabel       string `json:"acceptAllLabel"`
		RejectAllLabel       string `json:"rejectAllLabel"`
		SavePreferencesLabel string `json:"savePreferencesLabel"`
		PrivacyPolicyURL     string `json:"privacyPolicyUrl"`
		ConsentExpiryDays    int    `json:"consentExpiryDays"`
		ConsentMode          string `json:"consentMode"`
		Version              int    `json:"version"`
		EmbedSnippet         string `json:"embedSnippet"`
		CreatedAt            string `json:"createdAt"`
		UpdatedAt            string `json:"updatedAt"`
	} `json:"node"`
}

func NewCmdView(f *cmdutil.Factory) *cobra.Command {
	var flagOutput *string

	cmd := &cobra.Command{
		Use:   "view <id>",
		Short: "View a cookie banner",
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
			)

			data, err := client.Do(
				viewQuery,
				map[string]any{"id": args[0]},
			)
			if err != nil {
				return err
			}

			var resp viewResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			if resp.Node == nil {
				return fmt.Errorf("cookie banner %s not found", args[0])
			}

			if resp.Node.Typename != "CookieBanner" {
				return fmt.Errorf("expected CookieBanner node, got %s", resp.Node.Typename)
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, resp.Node)
			}

			b := resp.Node
			out := f.IOStreams.Out

			bold := lipgloss.NewStyle().Bold(true)
			label := lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Width(22)

			_, _ = fmt.Fprintf(out, "%s\n\n", bold.Render(b.Name))

			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("ID:"), b.ID)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Domain:"), b.Domain)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("State:"), b.State)
			_, _ = fmt.Fprintf(out, "%s%d\n", label.Render("Version:"), b.Version)

			_, _ = fmt.Fprintln(out)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Title:"), b.Title)
			if b.Description != "" {
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Description:"), b.Description)
			}
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Privacy Policy URL:"), b.PrivacyPolicyURL)
			_, _ = fmt.Fprintf(out, "%s%d days\n", label.Render("Consent Expiry:"), b.ConsentExpiryDays)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Consent Mode:"), b.ConsentMode)

			_, _ = fmt.Fprintln(out)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Accept All Label:"), b.AcceptAllLabel)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Reject All Label:"), b.RejectAllLabel)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Save Prefs Label:"), b.SavePreferencesLabel)

			_, _ = fmt.Fprintln(out)
			_, _ = fmt.Fprintf(out, "%s\n%s\n", label.Render("Embed Snippet:"), b.EmbedSnippet)

			_, _ = fmt.Fprintln(out)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Created:"), cmdutil.FormatTime(b.CreatedAt))
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Updated:"), cmdutil.FormatTime(b.UpdatedAt))

			return nil
		},
	}

	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
