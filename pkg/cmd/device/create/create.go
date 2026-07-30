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

package create

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const createMutation = `
mutation($input: CreateDeviceInput!) {
  createDevice(input: $input) {
    device {
      id
      state
    }
    enrollmentToken
    serverUrl
    enrollmentUrl
  }
}
`

type createResponse struct {
	CreateDevice struct {
		Device struct {
			ID    string `json:"id"`
			State string `json:"state"`
		} `json:"device"`
		EnrollmentToken string `json:"enrollmentToken"`
		ServerURL       string `json:"serverUrl"`
		EnrollmentURL   string `json:"enrollmentUrl"`
	} `json:"createDevice"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg   string
		flagOwner string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a PENDING ITAM device and enrollment token",
		Example: `  # Create a device for the default organization
  prb device create

  # Create a device with an owner
  prb device create --owner <profile-id>`,
		Args: cobra.NoArgs,
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

			if flagOrg == "" {
				flagOrg = hc.Organization
			}

			if flagOrg == "" {
				return fmt.Errorf("organization is required; pass --org or set a default with 'prb auth login'")
			}

			input := map[string]any{
				"organizationId": flagOrg,
			}

			if flagOwner != "" {
				input["ownerId"] = flagOwner
			}

			data, err := client.Do(
				createMutation,
				map[string]any{"input": input},
			)
			if err != nil {
				return err
			}

			var resp createResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			out := f.IOStreams.Out
			created := resp.CreateDevice

			_, _ = fmt.Fprintf(out, "Created device %s (%s)\n", created.Device.ID, created.Device.State)
			_, _ = fmt.Fprintf(out, "Server URL: %s\n", created.ServerURL)
			_, _ = fmt.Fprintf(out, "Enrollment URL: %s\n", created.EnrollmentURL)
			_, _ = fmt.Fprintf(out, "\nEnrollment Token (save this now — it will not be shown again and it expires):\n%s\n", created.EnrollmentToken)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().StringVar(&flagOwner, "owner", "", "Owner profile ID")

	return cmd
}
