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

package setowner

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const setOwnerMutation = `
mutation($input: SetDeviceOwnerInput!) {
  setDeviceOwner(input: $input) {
    device {
      id
      owner {
        id
      }
    }
  }
}
`

type setOwnerResponse struct {
	SetDeviceOwner struct {
		Device struct {
			ID    string `json:"id"`
			Owner *struct {
				ID string `json:"id"`
			} `json:"owner"`
		} `json:"device"`
	} `json:"setDeviceOwner"`
}

func NewCmdSetOwner(f *cmdutil.Factory) *cobra.Command {
	var flagOwner string

	cmd := &cobra.Command{
		Use:   "set-owner <id>",
		Short: "Set or clear the owner of an ITAM device",
		Example: `  # Assign an owner
  prb device set-owner <device-id> --owner <profile-id>

  # Clear the owner
  prb device set-owner <device-id> --owner ""`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmd.Flags().Changed("owner") {
				return fmt.Errorf("--owner is required (pass an empty value to clear)")
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
				"deviceId": args[0],
			}

			if flagOwner == "" {
				input["ownerId"] = nil
			} else {
				input["ownerId"] = flagOwner
			}

			data, err := client.Do(
				setOwnerMutation,
				map[string]any{"input": input},
			)
			if err != nil {
				return err
			}

			var resp setOwnerResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			ownerID := ""
			if resp.SetDeviceOwner.Device.Owner != nil {
				ownerID = resp.SetDeviceOwner.Device.Owner.ID
			}

			if ownerID == "" {
				_, _ = fmt.Fprintf(
					f.IOStreams.Out,
					"Cleared owner on device %s\n",
					resp.SetDeviceOwner.Device.ID,
				)
			} else {
				_, _ = fmt.Fprintf(
					f.IOStreams.Out,
					"Set owner of device %s to %s\n",
					resp.SetDeviceOwner.Device.ID,
					ownerID,
				)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOwner, "owner", "", "Owner profile ID (empty to clear)")

	return cmd
}
