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

package sync

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const syncMutation = `
mutation AccessReviewSourceSync($input: SyncAccessInventoryInput!) {
  syncAccessInventory(input: $input) {
    accessReviewSource {
      id
      name
      lastSuccessfulSyncAt
      nextSyncAt
      syncConsecutiveFailures
    }
    accessSourceSyncRun {
      id
      status
      createdAt
    }
  }
}
`

type syncResponse struct {
	SyncAccessInventory struct {
		AccessReviewSource struct {
			ID                      string  `json:"id"`
			Name                    string  `json:"name"`
			LastSuccessfulSyncAt    *string `json:"lastSuccessfulSyncAt"`
			NextSyncAt              *string `json:"nextSyncAt"`
			SyncConsecutiveFailures int     `json:"syncConsecutiveFailures"`
		} `json:"accessReviewSource"`
		AccessSourceSyncRun struct {
			ID        string `json:"id"`
			Status    string `json:"status"`
			CreatedAt string `json:"createdAt"`
		} `json:"accessSourceSyncRun"`
	} `json:"syncAccessInventory"`
}

func NewCmdSync(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "sync <source-id>",
		Short: "Enqueue an inventory sync for an access source",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, _, err := f.GraphQLClient(cmdutil.ConsoleEndpoint)
			if err != nil {
				return err
			}

			data, err := client.Do(
				syncMutation,
				map[string]any{
					"input": map[string]any{
						"accessReviewSourceId": args[0],
					},
				},
			)
			if err != nil {
				return err
			}

			var resp syncResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			payload := resp.SyncAccessInventory
			out := f.IOStreams.Out
			_, _ = fmt.Fprintf(
				out,
				"Requested inventory sync for source %s (%s): run %s [%s]\n",
				payload.AccessReviewSource.ID,
				payload.AccessReviewSource.Name,
				payload.AccessSourceSyncRun.ID,
				payload.AccessSourceSyncRun.Status,
			)

			return nil
		},
	}
}
