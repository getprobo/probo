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

package publish

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const publishMutation = `
mutation($input: PublishCookieBannerVersionInput!) {
  publishCookieBannerVersion(input: $input) {
    cookieBannerVersion {
      id
      version
      state
    }
  }
}
`

func NewCmdPublish(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "publish <id>",
		Short: "Publish the current draft version",
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

			data, err := client.Do(publishMutation, map[string]any{"input": input})
			if err != nil {
				return err
			}

			var resp struct {
				PublishCookieBannerVersion struct {
					CookieBannerVersion struct {
						ID      string `json:"id"`
						Version int    `json:"version"`
						State   string `json:"state"`
					} `json:"cookieBannerVersion"`
				} `json:"publishCookieBannerVersion"`
			}
			if err := json.Unmarshal(data, &resp); err != nil {
				return err
			}

			v := resp.PublishCookieBannerVersion.CookieBannerVersion
			_, _ = fmt.Fprintf(f.IOStreams.Out, "Published version %d for cookie banner %s\n", v.Version, args[0])

			return nil
		},
	}

	return cmd
}
