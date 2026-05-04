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
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const viewQuery = `
query($id: ID!) {
  node(id: $id) {
    __typename
    ... on Organization {
      scimConfiguration {
        id
        endpointUrl
        bridge {
          id
          state
          type
          excludedUserNames
          connector {
            id
            provider
          }
        }
        createdAt
        updatedAt
      }
    }
  }
}
`

type (
	connector struct {
		ID       string `json:"id"`
		Provider string `json:"provider"`
	}

	scimBridge struct {
		ID                string     `json:"id"`
		State             string     `json:"state"`
		Type              string     `json:"type"`
		ExcludedUserNames []string   `json:"excludedUserNames"`
		Connector         *connector `json:"connector"`
	}

	scimConfiguration struct {
		ID          string      `json:"id"`
		EndpointURL string      `json:"endpointUrl"`
		Bridge      *scimBridge `json:"bridge"`
		CreatedAt   string      `json:"createdAt"`
		UpdatedAt   string      `json:"updatedAt"`
	}

	viewResponse struct {
		Node *struct {
			Typename          string             `json:"__typename"`
			ScimConfiguration *scimConfiguration `json:"scimConfiguration"`
		} `json:"node"`
	}
)

func NewCmdView(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg    string
		flagOutput *string
	)

	cmd := &cobra.Command{
		Use:   "view",
		Short: "View SCIM configuration for an organization",
		Example: `  # View SCIM configuration for the default organization
  prb scim view

  # View as JSON
  prb scim view --json`,
		Args: cobra.NoArgs,
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
				"/api/connect/v1/graphql",
				cfg.HTTPTimeoutDuration(),
				cmdutil.TokenRefreshOption(cfg, host, hc),
			)

			if flagOrg == "" {
				flagOrg = hc.Organization
			}

			if flagOrg == "" {
				return fmt.Errorf("organization is required; pass --org or set a default with 'prb auth login'")
			}

			data, err := client.Do(
				viewQuery,
				map[string]any{"id": flagOrg},
			)
			if err != nil {
				return err
			}

			var resp viewResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			if resp.Node == nil {
				return fmt.Errorf("organization %s not found", flagOrg)
			}

			if resp.Node.Typename != "Organization" {
				return fmt.Errorf("expected Organization node, got %s", resp.Node.Typename)
			}

			if resp.Node.ScimConfiguration == nil {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "No SCIM configuration found.")
				return nil
			}

			sc := resp.Node.ScimConfiguration

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, sc)
			}

			out := f.IOStreams.Out
			bold := lipgloss.NewStyle().Bold(true)
			label := lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Width(22)

			_, _ = fmt.Fprintf(out, "%s\n\n", bold.Render("SCIM Configuration"))

			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("ID:"), sc.ID)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Endpoint URL:"), sc.EndpointURL)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Created:"), cmdutil.FormatTime(sc.CreatedAt))
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Updated:"), cmdutil.FormatTime(sc.UpdatedAt))

			if sc.Bridge != nil {
				_, _ = fmt.Fprintln(out)
				_, _ = fmt.Fprintf(out, "%s\n\n", bold.Render("Bridge"))
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Bridge ID:"), sc.Bridge.ID)
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("State:"), sc.Bridge.State)
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Type:"), sc.Bridge.Type)

				if sc.Bridge.Connector != nil {
					_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Connector ID:"), sc.Bridge.Connector.ID)
					_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Connector Provider:"), sc.Bridge.Connector.Provider)
				}

				if len(sc.Bridge.ExcludedUserNames) > 0 {
					_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Excluded Users:"), strings.Join(sc.Bridge.ExcludedUserNames, ", "))
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
