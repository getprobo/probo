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

package verify

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const verifyMutation = `
mutation($input: VerifyCloudAccountInput!) {
  verifyCloudAccount(input: $input) {
    cloudAccount {
      id
      status
    }
    status
    lastProbeError
  }
}
`

type verifyResponse struct {
	VerifyCloudAccount struct {
		CloudAccount struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"cloudAccount"`
		Status         string  `json:"status"`
		LastProbeError *string `json:"lastProbeError"`
	} `json:"verifyCloudAccount"`
}

func NewCmdVerify(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify <id>",
		Short: "Synchronously probe a cloud account and write the transition",
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

			data, err := client.Do(
				verifyMutation,
				map[string]any{
					"input": map[string]any{
						"cloudAccountId": args[0],
					},
				},
			)
			if err != nil {
				return err
			}

			var resp verifyResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			out := f.IOStreams.Out
			_, _ = fmt.Fprintf(out, "Verified cloud account %s\n", resp.VerifyCloudAccount.CloudAccount.ID)
			_, _ = fmt.Fprintf(out, "Status: %s\n", resp.VerifyCloudAccount.Status)
			if resp.VerifyCloudAccount.LastProbeError != nil {
				_, _ = fmt.Fprintf(out, "Last Probe Error: %s\n", *resp.VerifyCloudAccount.LastProbeError)
			}

			return nil
		},
	}

	return cmd
}
