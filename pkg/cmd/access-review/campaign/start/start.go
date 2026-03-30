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
