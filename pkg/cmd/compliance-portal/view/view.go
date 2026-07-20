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
    ... on Organization {
      compliancePortal {
        id
        active
        searchEngineIndexing
        logoFileUrl
        darkLogoFileUrl
        ndaFileName
        ndaFileUrl
        createdAt
        updatedAt
      }
    }
  }
}
`

type viewResponse struct {
	Node *struct {
		Typename         string `json:"__typename"`
		CompliancePortal *struct {
			ID                   string  `json:"id"`
			Active               bool    `json:"active"`
			SearchEngineIndexing string  `json:"searchEngineIndexing"`
			LogoFileUrl          *string `json:"logoFileUrl"`
			DarkLogoFileUrl      *string `json:"darkLogoFileUrl"`
			NdaFileName          *string `json:"ndaFileName"`
			NdaFileUrl           *string `json:"ndaFileUrl"`
			CreatedAt            string  `json:"createdAt"`
			UpdatedAt            string  `json:"updatedAt"`
		} `json:"compliancePortal"`
	} `json:"node"`
}

func NewCmdView(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg    string
		flagOutput *string
	)

	cmd := &cobra.Command{
		Use:   "view",
		Short: "View compliance portal settings",
		Args:  cobra.NoArgs,
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

			if resp.Node.CompliancePortal == nil {
				return fmt.Errorf("compliance portal not found for organization %s", flagOrg)
			}

			tc := resp.Node.CompliancePortal

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, tc)
			}

			out := f.IOStreams.Out

			bold := lipgloss.NewStyle().Bold(true)
			label := lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Width(28)

			_, _ = fmt.Fprintf(out, "%s\n\n", bold.Render("Compliance Portal"))

			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("ID:"), tc.ID)
			_, _ = fmt.Fprintf(out, "%s%v\n", label.Render("Active:"), tc.Active)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Search Engine Indexing:"), tc.SearchEngineIndexing)

			if tc.NdaFileName != nil && *tc.NdaFileName != "" {
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("NDA File:"), *tc.NdaFileName)
			}

			if tc.LogoFileUrl != nil && *tc.LogoFileUrl != "" {
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Logo URL:"), *tc.LogoFileUrl)
			}

			if tc.DarkLogoFileUrl != nil && *tc.DarkLogoFileUrl != "" {
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Dark Logo URL:"), *tc.DarkLogoFileUrl)
			}

			_, _ = fmt.Fprintln(out)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Created:"), cmdutil.FormatTime(tc.CreatedAt))
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Updated:"), cmdutil.FormatTime(tc.UpdatedAt))

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
