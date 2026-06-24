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

package update

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const updateMutation = `
mutation($input: UpdateTrackerPatternInput!) {
  updateTrackerPattern(input: $input) {
    trackerPattern {
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
	UpdateTrackerPattern struct {
		TrackerPattern struct {
			ID          string `json:"id"`
			DisplayName string `json:"displayName"`
		} `json:"trackerPattern"`
	} `json:"updateTrackerPattern"`
}

func NewCmdUpdate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagDescription string
		flagMaxAge      int
		flagExcluded    bool
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a tracker pattern",
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

			input := map[string]any{"trackerPatternId": args[0]}

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

			p := resp.UpdateTrackerPattern.TrackerPattern
			_, _ = fmt.Fprintf(f.IOStreams.Out, "Updated tracker pattern %s (%s)\n", p.ID, p.DisplayName)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagDescription, "description", "", "Description")
	cmd.Flags().IntVar(&flagMaxAge, "max-age-seconds", 0, "Maximum age in seconds")
	cmd.Flags().BoolVar(&flagExcluded, "excluded", false, "Exclude pattern from consent banner")

	return cmd
}
