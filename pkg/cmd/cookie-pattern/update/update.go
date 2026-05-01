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

package update

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const updateMutation = `
mutation($input: UpdateCookiePatternInput!) {
  updateCookiePattern(input: $input) {
    cookiePattern {
      id
      displayName
    }
    cookieBanner {
      id
    }
  }
}
`

type updateResponse struct {
	UpdateCookiePattern struct {
		CookiePattern struct {
			ID          string `json:"id"`
			DisplayName string `json:"displayName"`
		} `json:"cookiePattern"`
	} `json:"updateCookiePattern"`
}

func NewCmdUpdate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagDisplayName string
		flagDescription string
		flagMaxAge      int
		flagExcluded    bool
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a cookie pattern",
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

			input := map[string]any{"cookiePatternId": args[0]}

			if cmd.Flags().Changed("display-name") {
				input["displayName"] = flagDisplayName
			}
			if cmd.Flags().Changed("description") {
				input["description"] = flagDescription
			}
			if cmd.Flags().Changed("max-age-seconds") {
				input["maxAgeSeconds"] = flagMaxAge
			}
			if cmd.Flags().Changed("excluded") {
				input["excluded"] = flagExcluded
			}

			if len(input) == 1 {
				return fmt.Errorf("at least one field must be specified for update")
			}

			data, err := client.Do(updateMutation, map[string]any{"input": input})
			if err != nil {
				return err
			}

			var resp updateResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			p := resp.UpdateCookiePattern.CookiePattern
			_, _ = fmt.Fprintf(f.IOStreams.Out, "Updated cookie pattern %s (%s)\n", p.ID, p.DisplayName)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagDisplayName, "display-name", "", "Display name")
	cmd.Flags().StringVar(&flagDescription, "description", "", "Description")
	cmd.Flags().IntVar(&flagMaxAge, "max-age-seconds", 0, "Maximum age in seconds")
	cmd.Flags().BoolVar(&flagExcluded, "excluded", false, "Exclude pattern from consent banner")

	return cmd
}
