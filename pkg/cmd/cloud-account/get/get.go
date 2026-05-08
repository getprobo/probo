// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package get

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const getQuery = `
query($id: ID!) {
  node(id: $id) {
    __typename
    ... on CloudAccount {
      id
      label
      provider
      status
      credentialKind
      enabledAuditModules
      scope { kind identifier }
      lastProbeAt
      lastProbeError
      lastVerifiedAt
      createdAt
      updatedAt
    }
  }
}
`

// cloudAccountResponse mirrors the GraphQL response shape. We expose
// every field including external_id-equivalent (scope.identifier),
// last_probe_error, and last_verified_at. The recovery path for AWS
// operators ("I lost the install-assets response") relies on this
// command surfacing the persisted external_id; in the GraphQL schema
// the external_id is part of the AWS install assets payload and is
// not duplicated on CloudAccount, so the recovery path is to rerun
// `prb cloud-account install-assets` for the same scope, which
// returns the persisted value when it already exists on the row.
type cloudAccountResponse struct {
	Node *struct {
		Typename            string   `json:"__typename"`
		ID                  string   `json:"id"`
		Label               string   `json:"label"`
		Provider            string   `json:"provider"`
		Status              string   `json:"status"`
		CredentialKind      string   `json:"credentialKind"`
		EnabledAuditModules []string `json:"enabledAuditModules"`
		Scope               struct {
			Kind       string  `json:"kind"`
			Identifier *string `json:"identifier"`
		} `json:"scope"`
		LastProbeAt    *string `json:"lastProbeAt"`
		LastProbeError *string `json:"lastProbeError"`
		LastVerifiedAt *string `json:"lastVerifiedAt"`
		CreatedAt      string  `json:"createdAt"`
		UpdatedAt      string  `json:"updatedAt"`
	} `json:"node"`
}

func NewCmdGet(f *cmdutil.Factory) *cobra.Command {
	var flagOutput *string

	cmd := &cobra.Command{
		Use:     "get <id>",
		Short:   "Get a cloud account",
		Aliases: []string{"view"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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
				getQuery,
				map[string]any{"id": args[0]},
			)
			if err != nil {
				return err
			}

			var resp cloudAccountResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			if resp.Node == nil {
				return fmt.Errorf("cloud account %s not found", args[0])
			}

			if resp.Node.Typename != "CloudAccount" {
				return fmt.Errorf("expected CloudAccount node, got %s", resp.Node.Typename)
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, resp.Node)
			}

			a := resp.Node
			out := f.IOStreams.Out

			bold := lipgloss.NewStyle().Bold(true)
			label := lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Width(22)

			_, _ = fmt.Fprintf(out, "%s\n\n", bold.Render("Cloud Account"))

			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("ID:"), a.ID)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Label:"), a.Label)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Provider:"), a.Provider)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Status:"), a.Status)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Credential Kind:"), a.CredentialKind)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Scope Kind:"), a.Scope.Kind)
			if a.Scope.Identifier != nil {
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Scope Identifier:"), *a.Scope.Identifier)
			}

			if len(a.EnabledAuditModules) > 0 {
				_, _ = fmt.Fprintf(out, "%s%v\n", label.Render("Audit Modules:"), a.EnabledAuditModules)
			}

			if a.LastVerifiedAt != nil {
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Last Verified:"), cmdutil.FormatTime(*a.LastVerifiedAt))
			}
			if a.LastProbeAt != nil {
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Last Probed:"), cmdutil.FormatTime(*a.LastProbeAt))
			}
			if a.LastProbeError != nil {
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Last Probe Error:"), *a.LastProbeError)
			}

			_, _ = fmt.Fprintln(out)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Created:"), cmdutil.FormatTime(a.CreatedAt))
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Updated:"), cmdutil.FormatTime(a.UpdatedAt))

			return nil
		},
	}

	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
