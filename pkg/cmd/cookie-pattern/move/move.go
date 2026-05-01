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

package move

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const moveMutation = `
mutation($input: MoveCookiePatternToCategoryInput!) {
  moveCookiePatternToCategory(input: $input) {
    cookiePattern {
      id
      cookieCategory {
        id
        name
      }
    }
    cookieBanner {
      id
    }
  }
}
`

type moveResponse struct {
	MoveCookiePatternToCategory struct {
		CookiePattern struct {
			ID             string `json:"id"`
			CookieCategory struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"cookieCategory"`
		} `json:"cookiePattern"`
	} `json:"moveCookiePatternToCategory"`
}

func NewCmdMove(f *cmdutil.Factory) *cobra.Command {
	var flagTargetCategoryID string

	cmd := &cobra.Command{
		Use:   "move <id>",
		Short: "Move a cookie pattern to a different category",
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

			data, err := client.Do(moveMutation, map[string]any{
				"input": map[string]any{
					"cookiePatternId":        args[0],
					"targetCookieCategoryId": flagTargetCategoryID,
				},
			})
			if err != nil {
				return err
			}

			var resp moveResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			p := resp.MoveCookiePatternToCategory.CookiePattern
			_, _ = fmt.Fprintf(f.IOStreams.Out, "Moved cookie pattern %s to category %s\n", p.ID, p.CookieCategory.Name)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagTargetCategoryID, "target-category-id", "", "Target cookie category ID (required)")
	_ = cmd.MarkFlagRequired("target-category-id")

	return cmd
}
