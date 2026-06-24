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

package browse

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

func NewCmdBrowse(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg       string
		flagNoBrowser bool
	)

	cmd := &cobra.Command{
		Use:     "browse",
		Short:   "Open Probo in the browser",
		Long:    "Open the Probo console in the default web browser.",
		Aliases: []string{"open"},
		Example: `  # Open the Probo console in the browser
  prb browse

  # Print the URL without opening the browser
  prb browse --no-browser

  # Open a specific organization
  prb browse --org <org-id>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}

			host, hc, err := cfg.DefaultHost()
			if err != nil {
				return err
			}

			if flagOrg == "" {
				flagOrg = hc.Organization
			}

			lowerHost := strings.ToLower(host)
			if !strings.HasPrefix(lowerHost, "http://") && !strings.HasPrefix(lowerHost, "https://") {
				host = "https://" + host
			}

			var url string
			if flagOrg != "" {
				url = fmt.Sprintf("%s/organizations/%s", host, flagOrg)
			} else {
				url = host
			}

			if flagNoBrowser || f.IOStreams.ForceNonInteractive {
				_, _ = fmt.Fprintln(f.IOStreams.Out, url)
				return nil
			}

			browser := cfg.Browser
			if err := openBrowser(url, browser); err != nil {
				_, _ = fmt.Fprintln(f.IOStreams.Out, url)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(
		&flagNoBrowser,
		"no-browser",
		"n",
		false,
		"Print the URL instead of opening it",
	)

	cmd.Flags().StringVar(
		&flagOrg,
		"org",
		"",
		"Organization ID (defaults to the current organization)",
	)

	return cmd
}

func openBrowser(url, browser string) error {
	if browser != "" {
		return exec.Command("sh", "-c", browser+" \"$0\"", url).Start()
	}

	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Start()
	case "linux":
		return exec.Command("xdg-open", url).Start()
	case "windows":
		return exec.Command(
			"rundll32",
			"url.dll,FileProtocolHandler",
			url,
		).Start()
	default:
		return fmt.Errorf("unsupported platform")
	}
}
