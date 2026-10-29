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

package attach

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const attachMutation = `
mutation($input: AttachTagInput!) {
  attachTag(input: $input) {
    tag {
      id
      key
      value
    }
    tagAssignmentEdge {
      node {
        id
      }
    }
    replacedTagAssignmentId
  }
}
`

type attachResponse struct {
	AttachTag struct {
		Tag struct {
			ID    string `json:"id"`
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"tag"`
		TagAssignmentEdge struct {
			Node struct {
				ID string `json:"id"`
			} `json:"node"`
		} `json:"tagAssignmentEdge"`
		ReplacedTagAssignmentID *string `json:"replacedTagAssignmentId"`
	} `json:"attachTag"`
}

func NewCmdAttach(f *cmdutil.Factory) *cobra.Command {
	var (
		flagResourceID string
		flagTagID      string
	)

	cmd := &cobra.Command{
		Use:   "attach",
		Short: "Attach a tag to a resource",
		Long: "Attach a tag to a resource. A resource holds one value per key, so " +
			"attaching env=staging replaces env=prod on that resource.",
		Example: `  # Attach a tag interactively
  prb tag attach --resource-id prbdoc_...

  # Attach a tag non-interactively
  prb tag attach --resource-id prbdoc_... --tag-id prbrtg_...`,
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

			if f.IOStreams.IsInteractive() {
				if flagResourceID == "" {
					err := huh.NewInput().
						Title("Resource ID").
						Value(&flagResourceID).
						Run()
					if err != nil {
						return err
					}
				}

				if flagTagID == "" {
					err := huh.NewInput().
						Title("Tag ID").
						Value(&flagTagID).
						Run()
					if err != nil {
						return err
					}
				}
			}

			if flagResourceID == "" {
				return fmt.Errorf("resource ID is required; pass --resource-id or run interactively")
			}

			if flagTagID == "" {
				return fmt.Errorf("tag ID is required; pass --tag-id or run interactively")
			}

			data, err := client.Do(
				attachMutation,
				map[string]any{
					"input": map[string]any{
						"resourceId": flagResourceID,
						"tagId":      flagTagID,
					},
				},
			)
			if err != nil {
				return err
			}

			var resp attachResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			a := resp.AttachTag
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Attached tag %s=%s to resource %s\n",
				a.Tag.Key,
				a.Tag.Value,
				flagResourceID,
			)

			if a.ReplacedTagAssignmentID != nil {
				_, _ = fmt.Fprintf(
					f.IOStreams.Out,
					"Replaced the previous %s value on that resource\n",
					a.Tag.Key,
				)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flagResourceID, "resource-id", "", "ID of the resource to tag")
	cmd.Flags().StringVar(&flagTagID, "tag-id", "", "ID of the tag to attach")

	return cmd
}
