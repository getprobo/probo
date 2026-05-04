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
    ... on SCIMBridge {
      id
      state
      type
      excludedUserNames
      scimConfiguration {
        id
      }
      connector {
        id
        provider
      }
      createdAt
      updatedAt
    }
  }
}
`

type (
	connector struct {
		ID       string `json:"id"`
		Provider string `json:"provider"`
	}

	viewResponse struct {
		Node *struct {
			Typename          string   `json:"__typename"`
			ID                string   `json:"id"`
			State             string   `json:"state"`
			Type              string   `json:"type"`
			ExcludedUserNames []string `json:"excludedUserNames"`
			ScimConfiguration *struct {
				ID string `json:"id"`
			} `json:"scimConfiguration"`
			Connector *connector `json:"connector"`
			CreatedAt string     `json:"createdAt"`
			UpdatedAt string     `json:"updatedAt"`
		} `json:"node"`
	}
)

func NewCmdView(f *cmdutil.Factory) *cobra.Command {
	var flagOutput *string

	cmd := &cobra.Command{
		Use:   "view <id>",
		Short: "View a SCIM bridge",
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
				"/api/connect/v1/graphql",
				cfg.HTTPTimeoutDuration(),
				cmdutil.TokenRefreshOption(cfg, host, hc),
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
				return fmt.Errorf("SCIM bridge %s not found", args[0])
			}

			if resp.Node.Typename != "SCIMBridge" {
				return fmt.Errorf("expected SCIMBridge node, got %s", resp.Node.Typename)
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, resp.Node)
			}

			b := resp.Node
			out := f.IOStreams.Out
			bold := lipgloss.NewStyle().Bold(true)
			label := lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Width(22)

			_, _ = fmt.Fprintf(out, "%s\n\n", bold.Render("SCIM Bridge"))

			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("ID:"), b.ID)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("State:"), b.State)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Type:"), b.Type)

			if b.ScimConfiguration != nil {
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Configuration ID:"), b.ScimConfiguration.ID)
			}

			if b.Connector != nil {
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Connector ID:"), b.Connector.ID)
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Connector Provider:"), b.Connector.Provider)
			}

			if len(b.ExcludedUserNames) > 0 {
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Excluded Users:"), strings.Join(b.ExcludedUserNames, ", "))
			}

			_, _ = fmt.Fprintln(out)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Created:"), cmdutil.FormatTime(b.CreatedAt))
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Updated:"), cmdutil.FormatTime(b.UpdatedAt))

			return nil
		},
	}

	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
