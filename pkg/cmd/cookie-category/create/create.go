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
