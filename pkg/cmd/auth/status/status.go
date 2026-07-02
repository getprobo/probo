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
  prb auth status`,
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
