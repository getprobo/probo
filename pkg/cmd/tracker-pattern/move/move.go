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

package move

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const moveMutation = `
mutation($input: MoveTrackerPatternToCategoryInput!) {
  moveTrackerPatternToCategory(input: $input) {
    trackerPattern {
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
	MoveTrackerPatternToCategory struct {
		TrackerPattern struct {
			ID             string `json:"id"`
			CookieCategory struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"cookieCategory"`
		} `json:"trackerPattern"`
	} `json:"moveTrackerPatternToCategory"`
}

func NewCmdMove(f *cmdutil.Factory) *cobra.Command {
	var flagTargetCategoryID string

	cmd := &cobra.Command{
		Use:   "move <id>",
		Short: "Move a tracker pattern to a different category",
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
					"trackerPatternId":       args[0],
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

			p := resp.MoveTrackerPatternToCategory.TrackerPattern
			_, _ = fmt.Fprintf(f.IOStreams.Out, "Moved tracker pattern %s to category %s\n", p.ID, p.CookieCategory.Name)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagTargetCategoryID, "target-category-id", "", "Target cookie category ID (required)")
	_ = cmd.MarkFlagRequired("target-category-id")

	return cmd
}
