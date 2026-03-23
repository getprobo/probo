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

package list

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const listQuery = `
query($category: String) {
  cookieProviders(category: $category) {
    key
    name
    category
    cookies {
      name
    }
  }
}
`

type cookieProvider struct {
	Key      string `json:"key"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Cookies  []struct {
		Name string `json:"name"`
	} `json:"cookies"`
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagCategory string
		flagOutput   *string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List known cookie providers",
		Aliases: []string{"ls"},
		Example: `  # List all cookie providers
  prb cookie-banner provider list

  # List only analytics providers
  prb cookie-banner provider list --category Analytics`,
		Args: cobra.NoArgs,
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
			)

			variables := map[string]any{}
			if flagCategory != "" {
				variables["category"] = flagCategory
			}

			data, err := client.Do(listQuery, variables)
			if err != nil {
				return err
			}

			var resp struct {
				CookieProviders []cookieProvider `json:"cookieProviders"`
			}
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, resp.CookieProviders)
			}

			if len(resp.CookieProviders) == 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "No cookie providers found.")
				return nil
			}

			rows := make([][]string, 0, len(resp.CookieProviders))
			for _, p := range resp.CookieProviders {
				rows = append(rows, []string{
					p.Key,
					p.Name,
					p.Category,
					fmt.Sprintf("%d", len(p.Cookies)),
				})
			}

			t := cmdutil.NewTable("KEY", "NAME", "CATEGORY", "COOKIES").Rows(rows...)

			_, _ = fmt.Fprintln(f.IOStreams.Out, t)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagCategory, "category", "", "Filter by category (Necessary, Analytics, Marketing, Preferences)")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
