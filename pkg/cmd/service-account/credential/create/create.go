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
	"time"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
	"go.probo.inc/probo/pkg/cmd/service-account/shared"
)

const createMutation = `
mutation($input: CreateServiceAccountCredentialInput!) {
  createServiceAccountCredential(input: $input) {
    serviceAccountCredentialEdge {
      node {
        id
        serviceAccountId
        name
        scopes
        expiresAt
        lastUsedAt
        revokedAt
        createdAt
      }
    }
    token
  }
}
`

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagName      string
		flagScopes    []string
		flagExpiresAt string
	)

	cmd := &cobra.Command{
		Use:   "create <service-account-id>",
		Short: "Create a service account credential",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			expiresAt, err := time.Parse(time.RFC3339, flagExpiresAt)
			if err != nil {
				return fmt.Errorf("--expires-at must be an RFC3339 timestamp: %w", err)
			}

			cfg, err := f.Config()
			if err != nil {
				return err
			}

			host, hc, err := cfg.DefaultHost()
			if err != nil {
				return err
			}

			data, err := shared.NewClient(cfg, host, hc).Do(
				createMutation,
				map[string]any{
					"input": map[string]any{
						"serviceAccountId": args[0],
						"name":             flagName,
						"scopes":           flagScopes,
						"expiresAt":        expiresAt.Format(time.RFC3339Nano),
					},
				},
			)
			if err != nil {
				return err
			}

			var resp struct {
				CreateServiceAccountCredential struct {
					ServiceAccountCredentialEdge struct {
						Node shared.Credential `json:"node"`
					} `json:"serviceAccountCredentialEdge"`
					Token string `json:"token"`
				} `json:"createServiceAccountCredential"`
			}
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			credential := resp.CreateServiceAccountCredential
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Created service account credential %s (%s)\n\n",
				credential.ServiceAccountCredentialEdge.Node.Name,
				credential.ServiceAccountCredentialEdge.Node.ID,
			)
			_, _ = fmt.Fprintln(f.IOStreams.Out, "Credential token (save it now; it will not be shown again):")
			_, _ = fmt.Fprintln(f.IOStreams.Out, credential.Token)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagName, "name", "", "Credential name")
	cmd.Flags().StringSliceVar(&flagScopes, "scope", nil, "OAuth scope (repeat or separate with commas)")
	cmd.Flags().StringVar(&flagExpiresAt, "expires-at", "", "Credential expiration as an RFC3339 timestamp")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("scope")
	_ = cmd.MarkFlagRequired("expires-at")

	return cmd
}
