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

package close

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const closeMutation = `
mutation($input: CloseAccessReviewCampaignInput!) {
  closeAccessReviewCampaign(input: $input) {
    accessReviewCampaign {
      id
      name
      status
    }
  }
}
`

type closeResponse struct {
	CloseAccessReviewCampaign struct {
		AccessReviewCampaign struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"accessReviewCampaign"`
	} `json:"closeAccessReviewCampaign"`
}

func NewCmdClose(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "close <id>",
		Short: "Close an access review campaign",
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
			)

			data, err := client.Do(
				closeMutation,
				map[string]any{
					"input": map[string]any{
						"accessReviewCampaignId": args[0],
					},
				},
			)
			if err != nil {
				return err
			}

			var resp closeResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			c := resp.CloseAccessReviewCampaign.AccessReviewCampaign
			out := f.IOStreams.Out
			_, _ = fmt.Fprintf(out, "Closed access review campaign %s\n", c.ID)
			_, _ = fmt.Fprintf(out, "Name: %s\n", c.Name)
			_, _ = fmt.Fprintf(out, "Status: %s\n", c.Status)

			return nil
		},
	}

	return cmd
}
