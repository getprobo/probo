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

package publish

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const publishMutation = `
mutation($input: PublishCookieBannerInput!) {
  publishCookieBanner(input: $input) {
    cookieBanner {
      id
      name
      state
    }
  }
}
`

type publishResponse struct {
	PublishCookieBanner struct {
		CookieBanner struct {
			ID    string `json:"id"`
			Name  string `json:"name"`
			State string `json:"state"`
		} `json:"cookieBanner"`
	} `json:"publishCookieBanner"`
}

func NewCmdPublish(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "publish <id>",
		Short: "Publish a cookie banner",
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
			)

			data, err := client.Do(
				publishMutation,
				map[string]any{
					"input": map[string]any{
						"id": args[0],
					},
				},
			)
			if err != nil {
				return err
			}

			var resp publishResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			b := resp.PublishCookieBanner.CookieBanner
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Published cookie banner %s (%s)\n",
				b.ID,
				b.Name,
			)

			return nil
		},
	}

	return cmd
}
