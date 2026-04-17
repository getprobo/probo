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

package addsource

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const addSourceMutation = `
mutation($input: AddAccessReviewCampaignScopeSourceInput!) {
  addAccessReviewCampaignScopeSource(input: $input) {
    accessReviewCampaign {
      id
      name
      status
    }
  }
}
`

type addSourceResponse struct {
	AddAccessReviewCampaignScopeSource struct {
		AccessReviewCampaign struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"accessReviewCampaign"`
	} `json:"addAccessReviewCampaignScopeSource"`
}

func NewCmdAddSource(f *cmdutil.Factory) *cobra.Command {
	var flagSourceID string

	cmd := &cobra.Command{
		Use:   "add-source <campaign-id>",
		Short: "Add a scope source to an access review campaign",
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

			input := map[string]any{
				"accessReviewCampaignId": args[0],
				"accessSourceId":         flagSourceID,
			}

			data, err := client.Do(
				addSourceMutation,
				map[string]any{"input": input},
			)
			if err != nil {
				return err
			}

			var resp addSourceResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			c := resp.AddAccessReviewCampaignScopeSource.AccessReviewCampaign
			out := f.IOStreams.Out
			_, _ = fmt.Fprintf(out, "Added source %s to campaign %s\n", flagSourceID, c.ID)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagSourceID, "source-id", "", "Access source ID to add (required)")

	_ = cmd.MarkFlagRequired("source-id")

	return cmd
}
