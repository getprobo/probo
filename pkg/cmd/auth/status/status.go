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

package status

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

func NewCmdStatus(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "View authentication status",
		Example: `  # Show all configured hosts and their authentication status
  proboctl auth status`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}

			if len(cfg.Hosts) == 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "You are not logged in to any Probo hosts.")
				return nil
			}

			bold := lipgloss.NewStyle().Bold(true)

			for host, hc := range cfg.Hosts {
				label := host
				if host == cfg.ActiveHost {
					label += " (active)"
				}
				_, _ = fmt.Fprintf(
					f.IOStreams.Out,
					"%s\n",
					bold.Render(label),
				)

				tokenStatus := "not set"
				if hc.Token != "" {
					tokenStatus = "set"
				}
				_, _ = fmt.Fprintf(
					f.IOStreams.Out,
					"  Token: %s\n",
					tokenStatus,
				)

				if hc.Organization != "" {
					_, _ = fmt.Fprintf(
						f.IOStreams.Out,
						"  Organization: %s\n",
						hc.Organization,
					)
				}
			}

			return nil
		},
	}
}
