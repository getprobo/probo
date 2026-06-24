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

package addsource

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const addSourceMutation = `
mutation($input: AddAccessReviewCampaignSourceInput!) {
  addAccessReviewCampaignSource(input: $input) {
    accessReviewCampaign {
      id
      name
      status
    }
  }
}
`

type addSourceResponse struct {
	AddAccessReviewCampaignSource struct {
		AccessReviewCampaign struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"accessReviewCampaign"`
	} `json:"addAccessReviewCampaignSource"`
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
				"accessReviewSourceId":   flagSourceID,
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

			c := resp.AddAccessReviewCampaignSource.AccessReviewCampaign
			out := f.IOStreams.Out
			_, _ = fmt.Fprintf(out, "Added source %s to campaign %s\n", flagSourceID, c.ID)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagSourceID, "source-id", "", "Access source ID to add (required)")

	_ = cmd.MarkFlagRequired("source-id")

	return cmd
}
