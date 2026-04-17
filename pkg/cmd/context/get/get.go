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

package get

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const getQuery = `
query($id: ID!) {
  node(id: $id) {
    ... on Organization {
      id
      name
      context {
        product
        architecture
        team
        processes
        customers
      }
    }
  }
}
`

type getResponse struct {
	Node *struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Context *struct {
			Product      *string `json:"product"`
			Architecture *string `json:"architecture"`
			Team         *string `json:"team"`
			Processes    *string `json:"processes"`
			Customers    *string `json:"customers"`
		} `json:"context"`
	} `json:"node"`
}

func NewCmdGet(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg    string
		flagOutput *string
	)

	cmd := &cobra.Command{
		Use:     "get",
		Short:   "Get organization context",
		Example: `  prb context get --org <org-id>`,
		Args:    cobra.NoArgs,
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

			orgID := flagOrg
			if orgID == "" {
				orgID = hc.Organization
			}
			if orgID == "" {
				return fmt.Errorf("organization ID is required: pass --org or run `prb auth login`")
			}

			client := api.NewClient(
				host,
				hc.Token,
				"/api/console/v1/graphql",
				cfg.HTTPTimeoutDuration(),
				cmdutil.TokenRefreshOption(cfg, host, hc),
			)

			data, err := client.Do(
				getQuery,
				map[string]any{"id": orgID},
			)
			if err != nil {
				return err
			}

			var resp getResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			if resp.Node == nil {
				return fmt.Errorf("organization %s not found", orgID)
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, resp.Node.Context)
			}

			ctx := resp.Node.Context
			out := f.IOStreams.Out

			bold := lipgloss.NewStyle().Bold(true)
			label := lipgloss.NewStyle().Foreground(lipgloss.Color("242"))

			sections := []struct {
				title string
				value *string
			}{
				{"Product", ctx.Product},
				{"Architecture", ctx.Architecture},
				{"Team", ctx.Team},
				{"Processes", ctx.Processes},
				{"Customers", ctx.Customers},
			}

			for _, s := range sections {
				_, _ = fmt.Fprintf(out, "%s\n", bold.Render(s.title))
				if s.value != nil && *s.value != "" {
					_, _ = fmt.Fprintf(out, "%s\n\n", *s.value)
				} else {
					_, _ = fmt.Fprintf(out, "%s\n\n", label.Render("(empty)"))
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
