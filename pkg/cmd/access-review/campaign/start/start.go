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

package start

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const startMutation = `
mutation($input: StartAccessReviewCampaignInput!) {
  startAccessReviewCampaign(input: $input) {
    accessReviewCampaign {
      id
      name
      status
    }
  }
}
`

type startResponse struct {
	StartAccessReviewCampaign struct {
		AccessReviewCampaign struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"accessReviewCampaign"`
	} `json:"startAccessReviewCampaign"`
}

func NewCmdStart(f *cmdutil.Factory) *cobra.Command {
	var flagYes bool

	cmd := &cobra.Command{
		Use:   "start <id>",
		Short: "Start an access review campaign",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !flagYes {
				if !f.IOStreams.IsInteractive() {
					return fmt.Errorf("cannot start campaign: confirmation required, use --yes to confirm")
				}

				var confirmed bool

				err := huh.NewConfirm().
					Title(fmt.Sprintf("Start access review campaign %s?", args[0])).
					Value(&confirmed).
					Run()
				if err != nil {
					return err
				}

				if !confirmed {
					return nil
				}
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
				cmdutil.TokenRefreshOption(cfg, host, hc),
			)

			input := map[string]any{
				"accessReviewCampaignId": args[0],
			}

			data, err := client.Do(
				startMutation,
				map[string]any{"input": input},
			)
			if err != nil {
				return err
			}

			var resp startResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			c := resp.StartAccessReviewCampaign.AccessReviewCampaign
			out := f.IOStreams.Out
			_, _ = fmt.Fprintf(out, "Started access review campaign %s\n", c.ID)
			_, _ = fmt.Fprintf(out, "Name: %s\n", c.Name)
			_, _ = fmt.Fprintf(out, "Status: %s\n", c.Status)

			return nil
		},
	}

	cmd.Flags().BoolVarP(&flagYes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}
