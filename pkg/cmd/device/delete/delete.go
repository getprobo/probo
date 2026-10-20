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

package delete

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const deleteMutation = `
mutation($input: DeleteDevicesInput!) {
  deleteDevices(input: $input) {
    deletedDeviceIds
  }
}
`

type deleteResponse struct {
	DeleteDevices struct {
		DeletedDeviceIds []string `json:"deletedDeviceIds"`
	} `json:"deleteDevices"`
}

func NewCmdDelete(f *cmdutil.Factory) *cobra.Command {
	var flagYes bool

	cmd := &cobra.Command{
		Use:   "delete <id>...",
		Short: "Delete one or more revoked ITAM devices",
		Example: `  # Revoke devices, then delete them
  prb device revoke <device-id> [<device-id>...]
  prb device delete <device-id> [<device-id>...]`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !flagYes {
				if !f.IOStreams.IsInteractive() {
					return fmt.Errorf("cannot delete device: confirmation required, use --yes to confirm")
				}

				var confirmed bool

				title := fmt.Sprintf("Delete device %s?", args[0])
				if len(args) > 1 {
					title = fmt.Sprintf("Delete %d devices?", len(args))
				}

				err := huh.NewConfirm().
					Title(title).
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

			data, err := client.Do(
				deleteMutation,
				map[string]any{
					"input": map[string]any{
						"deviceIds": args,
					},
				},
			)
			if err != nil {
				return err
			}

			var resp deleteResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			for _, deviceID := range resp.DeleteDevices.DeletedDeviceIds {
				_, _ = fmt.Fprintf(
					f.IOStreams.Out,
					"Deleted device %s\n",
					deviceID,
				)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&flagYes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}
