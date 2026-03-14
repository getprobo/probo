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

package browse

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

func NewCmdBrowse(f *cmdutil.Factory) *cobra.Command {
	var flagOrg string

	cmd := &cobra.Command{
		Use:     "browse",
		Short:   "Open Probo in the browser",
		Long:    "Open the Probo console in the default web browser.",
		Aliases: []string{"open"},
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

			var url string
			if flagOrg != "" {
				url = fmt.Sprintf("https://%s/organizations/%s", host, flagOrg)
			} else {
				url = fmt.Sprintf("https://%s", host)
			}

			noPrint, _ := cmd.Flags().GetBool("no-browser")
			if noPrint || f.IOStreams.ForceNonInteractive {
				fmt.Fprintln(f.IOStreams.Out, url)
				return nil
			}

			browser := cfg.Browser
			if err := openBrowser(url, browser); err != nil {
				fmt.Fprintln(f.IOStreams.Out, url)
			}

			return nil
		},
	}

	cmd.Flags().BoolP(
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
		return exec.Command(browser, url).Start()
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
