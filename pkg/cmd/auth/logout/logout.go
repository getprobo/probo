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

package logout

import (
	"fmt"
	"maps"
	"slices"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/config"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

func NewCmdLogout(f *cmdutil.Factory) *cobra.Command {
	var flagHost string

	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Log out of a Probo host",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			if flagHost == "" {
				hosts := slices.Sorted(maps.Keys(cfg.Hosts))
				switch {
				case len(hosts) == 0:
					return fmt.Errorf("not logged in to any host")
				case len(hosts) == 1:
					flagHost = hosts[0]
				case f.IOStreams.IsInteractive():
					options := make([]huh.Option[string], len(hosts))
					for i, h := range hosts {
						options[i] = huh.NewOption(h, h)
					}
					err := huh.NewSelect[string]().
						Title("Select a host to log out of").
						Options(options...).
						Value(&flagHost).
						Run()
					if err != nil {
						return err
					}
				default:
					return fmt.Errorf("multiple hosts configured; use --hostname to specify which one")
				}
			}

			if _, ok := cfg.Hosts[flagHost]; !ok {
				return fmt.Errorf("not logged in to %s", flagHost)
			}

			delete(cfg.Hosts, flagHost)

			if err := cfg.Save(); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(
				f.IOStreams.ErrOut,
				"Logged out of %s\n",
				flagHost,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagHost, "hostname", "", "Probo hostname to log out of")

	return cmd
}
