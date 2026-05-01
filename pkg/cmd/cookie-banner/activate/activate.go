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

package activate

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const activateMutation = `
mutation($input: ActivateCookieBannerInput!) {
  activateCookieBanner(input: $input) {
    cookieBanner {
      id
      name
      state
    }
  }
}
`

func NewCmdActivate(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "activate <id>",
		Short: "Activate a cookie banner",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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

			input := map[string]any{"cookieBannerId": args[0]}

			data, err := client.Do(activateMutation, map[string]any{"input": input})
			if err != nil {
				return err
			}

			var resp struct {
				ActivateCookieBanner struct {
					CookieBanner struct {
						ID    string `json:"id"`
						Name  string `json:"name"`
						State string `json:"state"`
					} `json:"cookieBanner"`
				} `json:"activateCookieBanner"`
			}
			if err := json.Unmarshal(data, &resp); err != nil {
				return err
			}

			b := resp.ActivateCookieBanner.CookieBanner
			_, _ = fmt.Fprintf(f.IOStreams.Out, "Activated cookie banner %s (%s)\n", b.Name, b.ID)

			return nil
		},
	}

	return cmd
}
