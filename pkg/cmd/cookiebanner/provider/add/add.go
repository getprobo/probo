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

package add

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const addMutation = `
mutation($input: AddCookiesFromProviderInput!) {
  addCookiesFromProvider(input: $input) {
    cookieCategory {
      id
      name
      cookies {
        name
      }
    }
  }
}
`

func NewCmdAdd(f *cmdutil.Factory) *cobra.Command {
	var (
		flagCategoryID string
		flagProvider   string
	)

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add cookies from a known provider to a category",
		Example: `  # Add Google Analytics cookies to a category
  prb cookie-banner provider add --category-id <id> --provider google-analytics`,
		Args: cobra.NoArgs,
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

			variables := map[string]any{
				"input": map[string]any{
					"cookieCategoryId": flagCategoryID,
					"providerKey":      flagProvider,
				},
			}

			data, err := client.Do(addMutation, variables)
			if err != nil {
				return err
			}

			var resp struct {
				AddCookiesFromProvider struct {
					CookieCategory struct {
						ID      string `json:"id"`
						Name    string `json:"name"`
						Cookies []struct {
							Name string `json:"name"`
						} `json:"cookies"`
					} `json:"cookieCategory"`
				} `json:"addCookiesFromProvider"`
			}
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			cat := resp.AddCookiesFromProvider.CookieCategory
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Added cookies from %s to category %s (%s), now has %d cookies\n",
				flagProvider,
				cat.Name,
				cat.ID,
				len(cat.Cookies),
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagCategoryID, "category-id", "", "Cookie category ID (required)")
	cmd.Flags().StringVar(&flagProvider, "provider", "", "Provider key (required)")

	_ = cmd.MarkFlagRequired("category-id")
	_ = cmd.MarkFlagRequired("provider")

	return cmd
}
