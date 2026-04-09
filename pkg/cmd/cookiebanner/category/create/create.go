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

package create

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const createMutation = `
mutation($input: CreateCookieCategoryInput!) {
  createCookieCategory(input: $input) {
    cookieCategory {
      id
      name
      rank
    }
  }
}
`

type createResponse struct {
	CreateCookieCategory struct {
		CookieCategory struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			Rank int    `json:"rank"`
		} `json:"cookieCategory"`
	} `json:"createCookieCategory"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagBanner      string
		flagName        string
		flagDescription string
		flagRank        int
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new cookie category",
		Example: `  # Create a cookie category
  prb cookie-banner category create --banner <banner-id> --name "Analytics" --description "Cookies used for analytics"`,
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

			if f.IOStreams.IsInteractive() {
				if flagName == "" {
					err := huh.NewInput().
						Title("Category name").
						Value(&flagName).
						Run()
					if err != nil {
						return err
					}
				}
			}

			if flagName == "" {
				return fmt.Errorf("name is required; pass --name or run interactively")
			}

			input := map[string]any{
				"cookieBannerId": flagBanner,
				"name":           flagName,
			}

			if flagDescription != "" {
				input["description"] = flagDescription
			}
			if cmd.Flags().Changed("rank") {
				input["rank"] = flagRank
			}

			data, err := client.Do(
				createMutation,
				map[string]any{"input": input},
			)
			if err != nil {
				return err
			}

			var resp createResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			c := resp.CreateCookieCategory.CookieCategory
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Created cookie category %s (%s)\n",
				c.ID,
				c.Name,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagBanner, "banner", "", "Cookie banner ID (required)")
	cmd.Flags().StringVar(&flagName, "name", "", "Category name (required)")
	cmd.Flags().StringVar(&flagDescription, "description", "", "Category description")
	cmd.Flags().IntVar(&flagRank, "rank", 0, "Display order rank")

	_ = cmd.MarkFlagRequired("banner")

	return cmd
}
