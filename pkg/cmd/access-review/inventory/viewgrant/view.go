// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package viewgrant

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const viewQuery = `
query AccessInventoryGrantView($id: ID!) {
  node(id: $id) {
    __typename
    ... on AccessInventoryGrant {
      id
      principalKey
      resourceKey
      entitlementKey
      entitlementDisplayName
      presence
      consecutiveMissingSyncs
      firstSeenAt
      lastSeenAt
      providerChangedAt
      createdAt
      updatedAt
      source {
        id
        name
      }
      principal {
        id
        email
        fullName
      }
      resource {
        id
        displayName
        kind
      }
    }
  }
}
`

type viewResponse struct {
	Node *struct {
		Typename                string  `json:"__typename"`
		ID                      string  `json:"id"`
		PrincipalKey            string  `json:"principalKey"`
		ResourceKey             string  `json:"resourceKey"`
		EntitlementKey          string  `json:"entitlementKey"`
		EntitlementDisplayName  string  `json:"entitlementDisplayName"`
		Presence                string  `json:"presence"`
		ConsecutiveMissingSyncs int     `json:"consecutiveMissingSyncs"`
		FirstSeenAt             string  `json:"firstSeenAt"`
		LastSeenAt              string  `json:"lastSeenAt"`
		ProviderChangedAt       *string `json:"providerChangedAt"`
		CreatedAt               string  `json:"createdAt"`
		UpdatedAt               string  `json:"updatedAt"`
		Source                  struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"source"`
		Principal struct {
			ID       string `json:"id"`
			Email    string `json:"email"`
			FullName string `json:"fullName"`
		} `json:"principal"`
		Resource struct {
			ID          string `json:"id"`
			DisplayName string `json:"displayName"`
			Kind        string `json:"kind"`
		} `json:"resource"`
	} `json:"node"`
}

func NewCmdViewGrant(f *cmdutil.Factory) *cobra.Command {
	var flagOutput *string

	cmd := &cobra.Command{
		Use:   "view-grant <id>",
		Short: "View a live access inventory grant",
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

			data, err := client.Do(
				viewQuery,
				map[string]any{"id": args[0]},
			)
			if err != nil {
				return err
			}

			var result viewResponse
			if err := json.Unmarshal(data, &result); err != nil {
				return fmt.Errorf("cannot decode response: %w", err)
			}

			if result.Node == nil || result.Node.Typename != "AccessInventoryGrant" {
				return fmt.Errorf("access inventory grant %s not found", args[0])
			}

			node := result.Node

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, node)
			}

			out := f.IOStreams.Out
			bold := lipgloss.NewStyle().Bold(true)
			label := lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Width(22)

			_, _ = fmt.Fprintf(out, "%s\n\n", bold.Render("Access Inventory Grant"))
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("ID:"), node.ID)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Principal:"), node.PrincipalKey)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Resource:"), node.ResourceKey)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Entitlement:"), node.EntitlementDisplayName)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Presence:"), node.Presence)
			_, _ = fmt.Fprintf(out, "%s%d\n", label.Render("Missing syncs:"), node.ConsecutiveMissingSyncs)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Source:"), node.Source.Name)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Last seen:"), cmdutil.FormatTime(node.LastSeenAt))

			return nil
		},
	}

	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
