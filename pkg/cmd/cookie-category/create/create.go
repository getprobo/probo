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
    cookieCategoryEdge {
      node {
        id
        name
        slug
      }
    }
    cookieBanner {
      id
    }
  }
}
`

type createResponse struct {
	CreateCookieCategory struct {
		CookieCategoryEdge struct {
			Node struct {
				ID   string `json:"id"`
				Name string `json:"name"`
				Slug string `json:"slug"`
			} `json:"node"`
		} `json:"cookieCategoryEdge"`
		CookieBanner struct {
			ID string `json:"id"`
		} `json:"cookieBanner"`
	} `json:"createCookieCategory"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagBannerID    string
		flagName        string
		flagSlug        string
		flagDescription string
		flagRank        int
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new cookie category",
		Args:  cobra.NoArgs,
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

			if f.IOStreams.IsInteractive() {
				if flagName == "" {
					if err := huh.NewInput().Title("Category name").Value(&flagName).Run(); err != nil {
						return err
					}
				}
				if flagSlug == "" {
					if err := huh.NewInput().Title("Category slug").Value(&flagSlug).Run(); err != nil {
						return err
					}
				}
				if flagDescription == "" {
					if err := huh.NewText().Title("Description").Value(&flagDescription).Run(); err != nil {
						return err
					}
				}
			}

			if flagName == "" {
				return fmt.Errorf("name is required; pass --name or run interactively")
			}
			if flagSlug == "" {
				return fmt.Errorf("slug is required; pass --slug or run interactively")
			}

			input := map[string]any{
				"cookieBannerId": flagBannerID,
				"name":           flagName,
				"slug":           flagSlug,
				"description":    flagDescription,
				"rank":           flagRank,
			}

			data, err := client.Do(createMutation, map[string]any{"input": input})
			if err != nil {
				return err
			}

			var resp createResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			c := resp.CreateCookieCategory.CookieCategoryEdge.Node
			_, _ = fmt.Fprintf(f.IOStreams.Out, "Created cookie category %s (%s)\n", c.ID, c.Name)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagBannerID, "banner-id", "", "Cookie banner ID (required)")
	cmd.Flags().StringVar(&flagName, "name", "", "Category name")
	cmd.Flags().StringVar(&flagSlug, "slug", "", "Category slug")
	cmd.Flags().StringVar(&flagDescription, "description", "", "Category description")
	cmd.Flags().IntVar(&flagRank, "rank", 10, "Display rank")

	_ = cmd.MarkFlagRequired("banner-id")

	return cmd
}
