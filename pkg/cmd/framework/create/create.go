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
mutation($input: CreateFrameworkInput!) {
  createFramework(input: $input) {
    frameworkEdge {
      node {
        id
        name
        description
      }
    }
  }
}
`

type createResponse struct {
	CreateFramework struct {
		FrameworkEdge struct {
			Node struct {
				ID          string  `json:"id"`
				Name        string  `json:"name"`
				Description *string `json:"description"`
			} `json:"node"`
		} `json:"frameworkEdge"`
	} `json:"createFramework"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg         string
		flagName        string
		flagDescription string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new framework",
		Example: `  # Create a framework
  prb framework create --name "ISO 27001:2022" --description "Information security standard"`,
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
				"name":           flagName,
			}

			if flagDescription != "" {
				input["description"] = flagDescription
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

			fw := resp.CreateFramework.FrameworkEdge.Node
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Created framework %s (%s)\n",
				fw.ID,
				fw.Name,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().StringVar(&flagName, "name", "", "Framework name (required)")
	cmd.Flags().StringVar(&flagDescription, "description", "", "Framework description")

	_ = cmd.MarkFlagRequired("name")

	return cmd
}
