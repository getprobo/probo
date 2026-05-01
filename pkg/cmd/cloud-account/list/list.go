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

package list

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const listQuery = `
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: CloudAccountOrder, $filter: CloudAccountFilter) {
  node(id: $id) {
    __typename
    ... on Organization {
      cloudAccounts(first: $first, after: $after, orderBy: $orderBy, filter: $filter) {
        edges {
          node {
            id
            label
            provider
            status
            credentialKind
            scope { kind }
            lastVerifiedAt
            createdAt
          }
        }
        pageInfo {
          hasNextPage
          endCursor
        }
      }
    }
  }
}
`

type cloudAccountNode struct {
	ID             string `json:"id"`
	Label          string `json:"label"`
	Provider       string `json:"provider"`
	Status         string `json:"status"`
	CredentialKind string `json:"credentialKind"`
	Scope          struct {
		Kind string `json:"kind"`
	} `json:"scope"`
	LastVerifiedAt *string `json:"lastVerifiedAt"`
	CreatedAt      string  `json:"createdAt"`
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg       string
		flagLimit     int
		flagOrderBy   string
		flagOrderDir  string
		flagProvider  string
		flagStatus    string
		flagScopeKind string
		flagOutput    *string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List cloud accounts",
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
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

			if flagOrg == "" {
				flagOrg = hc.Organization
			}

			if flagOrg == "" {
				return fmt.Errorf("cannot determine organization, use --org or 'prb auth login'")
			}

			if err := cmdutil.ValidateEnum("order-direction", flagOrderDir, []string{"ASC", "DESC"}); err != nil {
				return err
			}

			variables := map[string]any{
				"id": flagOrg,
			}

			if flagOrderBy != "" {
				if err := cmdutil.ValidateEnum("order-by", flagOrderBy, []string{"CREATED_AT", "STATUS", "PROVIDER"}); err != nil {
					return err
				}
				variables["orderBy"] = map[string]any{
					"field":     flagOrderBy,
					"direction": flagOrderDir,
				}
			}

			filter := map[string]any{}
			if flagProvider != "" {
				if err := cmdutil.ValidateEnum("provider", flagProvider, []string{"AWS", "GCP", "AZURE"}); err != nil {
					return err
				}
				filter["provider"] = flagProvider
			}
			if flagStatus != "" {
				if err := cmdutil.ValidateEnum("status", flagStatus, []string{"PENDING_VERIFICATION", "VERIFIED", "ERRORED", "DISCONNECTED"}); err != nil {
					return err
				}
				filter["status"] = flagStatus
			}
			if flagScopeKind != "" {
				if err := cmdutil.ValidateEnum("scope-kind", flagScopeKind, []string{"AWS_ACCOUNT", "GCP_PROJECT", "GCP_ORGANIZATION", "AZURE_SUBSCRIPTION", "AZURE_MANAGEMENT_GROUP", "AZURE_TENANT"}); err != nil {
					return err
				}
				filter["scopeKind"] = flagScopeKind
			}
			if len(filter) > 0 {
				variables["filter"] = filter
			}

			accounts, _, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				func(data json.RawMessage) (*api.Connection[cloudAccountNode], error) {
					var resp struct {
						Node *struct {
							Typename      string                           `json:"__typename"`
							CloudAccounts api.Connection[cloudAccountNode] `json:"cloudAccounts"`
						} `json:"node"`
					}
					if err := json.Unmarshal(data, &resp); err != nil {
						return nil, err
					}
					if resp.Node == nil {
						return nil, fmt.Errorf("organization %s not found", flagOrg)
					}
					if resp.Node.Typename != "Organization" {
						return nil, fmt.Errorf("expected Organization node, got %s", resp.Node.Typename)
					}
					return &resp.Node.CloudAccounts, nil
				},
			)
			if err != nil {
				return err
			}

			if *flagOutput == cmdutil.OutputJSON {
				if accounts == nil {
					accounts = []cloudAccountNode{}
				}
				return cmdutil.PrintJSON(f.IOStreams.Out, accounts)
			}

			if len(accounts) == 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "No cloud accounts found.")
				return nil
			}

			rows := make([][]string, 0, len(accounts))
			for _, a := range accounts {
				lastVerified := ""
				if a.LastVerifiedAt != nil {
					lastVerified = cmdutil.FormatTime(*a.LastVerifiedAt)
				}
				rows = append(rows, []string{
					a.ID,
					a.Label,
					a.Provider,
					a.Status,
					a.Scope.Kind,
					lastVerified,
				})
			}

			t := cmdutil.NewTable("ID", "LABEL", "PROVIDER", "STATUS", "SCOPE", "LAST VERIFIED").Rows(rows...)

			_, _ = fmt.Fprintln(f.IOStreams.Out, t)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of cloud accounts to list")
	cmd.Flags().StringVar(&flagOrderBy, "order-by", "", "Order by field (CREATED_AT, STATUS, PROVIDER)")
	cmd.Flags().StringVar(&flagOrderDir, "order-direction", "DESC", "Sort direction (ASC, DESC)")
	cmd.Flags().StringVar(&flagProvider, "provider", "", "Filter by provider (AWS, GCP, AZURE)")
	cmd.Flags().StringVar(&flagStatus, "status", "", "Filter by status (PENDING_VERIFICATION, VERIFIED, ERRORED, DISCONNECTED)")
	cmd.Flags().StringVar(&flagScopeKind, "scope-kind", "", "Filter by scope kind")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
