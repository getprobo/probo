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

package sendinstallationemails

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const sendInstallationEmailsMutation = `
mutation($input: SendDeviceInstallationEmailsInput!) {
  sendDeviceInstallationEmails(input: $input) {
    recipientCount
  }
}
`

type sendInstallationEmailsResponse struct {
	SendDeviceInstallationEmails struct {
		RecipientCount int `json:"recipientCount"`
	} `json:"sendDeviceInstallationEmails"`
}

func NewCmdSendInstallationEmails(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrganization string
		flagOutput       *string
	)

	cmd := &cobra.Command{
		Use:   "send-installation-emails",
		Short: "Email employees who have no active device agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			if flagOrganization == "" {
				return fmt.Errorf("organization is required")
			}

			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return err
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
				sendInstallationEmailsMutation,
				map[string]any{
					"input": map[string]any{
						"organizationId": flagOrganization,
					},
				},
			)
			if err != nil {
				return err
			}

			var resp sendInstallationEmailsResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, resp.SendDeviceInstallationEmails)
			}

			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Queued installation emails for %d employees\n",
				resp.SendDeviceInstallationEmails.RecipientCount,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrganization, "organization", "", "Organization ID (required)")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
