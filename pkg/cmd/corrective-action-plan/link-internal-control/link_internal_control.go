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

package linkinternalcontrol

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const mutation = `
mutation CorrectiveActionPlanLinkInternalControl($input: CreateCorrectiveActionPlanInternalControlMappingInput!) {
  createCorrectiveActionPlanInternalControlMapping(input: $input) {
    correctiveActionPlanEdge {
      node { id }
    }
    internalControlEdge {
      node { id }
    }
  }
}
`

func NewCmdLinkInternalControl(f *cmdutil.Factory) *cobra.Command {
	var (
		flagCorrectiveActionPlanID string
		flagInternalControlID      string
	)

	cmd := &cobra.Command{
		Use:   "link-internal-control",
		Short: "Link an internal control to a corrective action plan",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, _, err := f.GraphQLClient(cmdutil.ConsoleEndpoint)
			if err != nil {
				return err
			}

			_, err = client.Do(
				mutation,
				map[string]any{
					"input": map[string]any{
						"correctiveActionPlanId": flagCorrectiveActionPlanID,
						"internalControlId":      flagInternalControlID,
					},
				},
			)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Linked internal control %s to corrective action plan %s\n",
				flagInternalControlID,
				flagCorrectiveActionPlanID,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagCorrectiveActionPlanID, "corrective-action-plan-id", "", "Corrective action plan ID (required)")
	cmd.Flags().StringVar(&flagInternalControlID, "internal-control-id", "", "Internal control ID (required)")

	_ = cmd.MarkFlagRequired("corrective-action-plan-id")
	_ = cmd.MarkFlagRequired("internal-control-id")

	return cmd
}
