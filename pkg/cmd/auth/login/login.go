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

package login

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/config"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

func NewCmdLogin(f *cmdutil.Factory) *cobra.Command {
	var (
		flagHost         string
		flagToken        string
		flagOrganization string
	)

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with a Probo host",
		Example: `  # Interactive login (prompts for hostname, token, and org)
  proboctl auth login

  # Non-interactive login
  proboctl auth login --hostname app.getprobo.com --token <token> --org <org-id>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if f.IOStreams.IsInteractive() {
				if flagHost == "" {
					err := huh.NewInput().
						Title("Probo hostname").
						Placeholder("app.getprobo.com").
						Value(&flagHost).
						Run()
					if err != nil {
						return err
					}
					if flagHost == "" {
						flagHost = "app.getprobo.com"
					}
				}

				if flagToken == "" {
					err := huh.NewInput().
						Title("API token").
						EchoMode(huh.EchoModePassword).
						Value(&flagToken).
						Run()
					if err != nil {
						return err
					}
				}

				if flagOrganization == "" {
					err := huh.NewInput().
						Title("Default organization ID").
						Placeholder("optional").
						Value(&flagOrganization).
						Run()
					if err != nil {
						return err
					}
				}
			}

			if flagHost == "" {
				flagHost = "app.getprobo.com"
			}

			if flagToken == "" {
				return fmt.Errorf("token is required; pass --token or run interactively")
			}

			cfg, err := f.Config()
			if err != nil {
				return err
			}

			cfg.Hosts[flagHost] = &config.HostConfig{
				Token:        flagToken,
				Organization: flagOrganization,
			}
			cfg.ActiveHost = flagHost

			if err := cfg.Save(); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(
				f.IOStreams.ErrOut,
				"Logged in to %s\n",
				flagHost,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagHost, "hostname", "", "Probo hostname (default: app.getprobo.com)")
	cmd.Flags().StringVar(&flagToken, "token", "", "API token")
	cmd.Flags().StringVar(&flagOrganization, "org", "", "Default organization ID")

	return cmd
}
