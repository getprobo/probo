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

package unlinkframework

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const unlinkMutation = `
mutation($input: DeleteAuditFrameworkMappingInput!) {
  deleteAuditFrameworkMapping(input: $input) {
    deletedFrameworkId
    audit { id }
  }
}
`

func NewCmdUnlinkFramework(f *cmdutil.Factory) *cobra.Command {
	var flagFramework string

	cmd := &cobra.Command{
		Use:   "unlink-framework <audit-id>",
		Short: "Unlink a framework from an audit",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if flagFramework == "" {
				return fmt.Errorf("framework is required; pass --framework")
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

			_, err = client.Do(
				unlinkMutation,
				map[string]any{
					"input": map[string]any{
						"auditId":     args[0],
						"frameworkId": flagFramework,
					},
				},
			)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Unlinked framework %s from audit %s\n",
				flagFramework,
				args[0],
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagFramework, "framework", "", "Framework ID")

	return cmd
}
