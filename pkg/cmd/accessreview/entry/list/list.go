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
query(
  $id: ID!,
  $first: Int,
  $after: CursorKey,
  $orderBy: AccessEntryOrder,
  $accessSourceId: ID,
  $filter: AccessEntryFilter
) {
  node(id: $id) {
    __typename
    ... on AccessReviewCampaign {
      entries(
        first: $first,
        after: $after,
        orderBy: $orderBy,
        accessSourceId: $accessSourceId,
        filter: $filter
      ) {
        totalCount
        edges {
          node {
            id
            email
            fullName
            role
            jobTitle
            isAdmin
            mfaStatus
            authMethod
            accountType
            lastLogin
            externalId
            incrementalTag
            flag
            flagReason
            decision
            decisionNote
            accessSource {
              id
              name
            }
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

type entryNode struct {
	ID             string  `json:"id"`
	Email          string  `json:"email"`
	FullName       string  `json:"fullName"`
	Role           string  `json:"role"`
	JobTitle       string  `json:"jobTitle"`
	IsAdmin        bool    `json:"isAdmin"`
	MfaStatus      string  `json:"mfaStatus"`
	AuthMethod     string  `json:"authMethod"`
	AccountType    string  `json:"accountType"`
	LastLogin      *string `json:"lastLogin"`
	ExternalID     string  `json:"externalId"`
	IncrementalTag string  `json:"incrementalTag"`
	Flag           string  `json:"flag"`
	FlagReason     *string `json:"flagReason"`
	Decision       string  `json:"decision"`
	DecisionNote   *string `json:"decisionNote"`
	AccessSource   struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"accessSource"`
	CreatedAt string `json:"createdAt"`
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagLimit      int
		flagOrderBy    string
		flagOrderDir   string
		flagSourceID   string
		flagDecision   string
		flagFlag       string
		flagIncTag     string
		flagIsAdmin    *bool
		flagAuthMethod  string
		flagAccountType string
		flagOutput      *string
	)

	cmd := &cobra.Command{
		Use:   "list <campaign-id>",
		Short: "List access entries for a campaign",
		Args:  cobra.ExactArgs(1),
		Example: `  # List all entries for a campaign
  prb access-review entry list <campaign-id>

  # List entries for a specific source
  prb access-review entry list <campaign-id> --source-id <source-id>

  # List only pending entries
  prb access-review entry list <campaign-id> --decision PENDING

  # List flagged entries
  prb access-review entry list <campaign-id> --flag ORPHANED`,
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
			)

			variables := map[string]any{
				"id": args[0],
			}

			if flagOrderBy != "" {
				if err := cmdutil.ValidateEnum("order-by", flagOrderBy, []string{"CREATED_AT"}); err != nil {
					return err
				}
				variables["orderBy"] = map[string]any{
					"field":     flagOrderBy,
					"direction": flagOrderDir,
				}
			}

			if flagSourceID != "" {
				variables["accessSourceId"] = flagSourceID
			}

			filter := map[string]any{}
			if flagDecision != "" {
				if err := cmdutil.ValidateEnum(
					"decision",
					flagDecision,
					[]string{"PENDING", "APPROVED", "REVOKE", "DEFER", "ESCALATE"},
				); err != nil {
					return err
				}
				filter["decision"] = flagDecision
			}
			if flagFlag != "" {
				if err := cmdutil.ValidateEnum(
					"flag",
					flagFlag,
					[]string{"NONE", "ORPHANED", "INACTIVE", "EXCESSIVE", "ROLE_MISMATCH", "NEW"},
				); err != nil {
					return err
				}
				filter["flag"] = flagFlag
			}
			if flagIncTag != "" {
				if err := cmdutil.ValidateEnum(
					"incremental-tag",
					flagIncTag,
					[]string{"NEW", "REMOVED", "UNCHANGED"},
				); err != nil {
					return err
				}
				filter["incrementalTag"] = flagIncTag
			}
			if cmd.Flags().Changed("is-admin") {
				filter["isAdmin"] = *flagIsAdmin
			}
			if flagAuthMethod != "" {
				if err := cmdutil.ValidateEnum(
					"auth-method",
					flagAuthMethod,
					[]string{"SSO", "PASSWORD", "API_KEY", "SERVICE_ACCOUNT", "UNKNOWN"},
				); err != nil {
					return err
				}
				filter["authMethod"] = flagAuthMethod
			}
			if flagAccountType != "" {
				if err := cmdutil.ValidateEnum(
					"account-type",
					flagAccountType,
					[]string{"USER", "SERVICE_ACCOUNT"},
				); err != nil {
					return err
				}
				filter["accountType"] = flagAccountType
			}
			if len(filter) > 0 {
				variables["filter"] = filter
			}

			entries, totalCount, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				func(data json.RawMessage) (*api.Connection[entryNode], error) {
					var resp struct {
						Node *struct {
							Typename string                    `json:"__typename"`
							Entries  api.Connection[entryNode] `json:"entries"`
						} `json:"node"`
					}
					if err := json.Unmarshal(data, &resp); err != nil {
						return nil, err
					}
					if resp.Node == nil {
						return nil, fmt.Errorf("campaign %s not found", args[0])
					}
					if resp.Node.Typename != "AccessReviewCampaign" {
						return nil, fmt.Errorf("expected AccessReviewCampaign node, got %s", resp.Node.Typename)
					}
					return &resp.Node.Entries, nil
				},
			)
			if err != nil {
				return err
			}

			if *flagOutput == cmdutil.OutputJSON {
				if entries == nil {
					entries = []entryNode{}
				}
				return cmdutil.PrintJSON(f.IOStreams.Out, entries)
			}

			if len(entries) == 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "No access entries found.")
				return nil
			}

			rows := make([][]string, 0, len(entries))
			for _, e := range entries {
				admin := ""
				if e.IsAdmin {
					admin = "yes"
				}
				rows = append(rows, []string{
					e.ID,
					e.Email,
					e.FullName,
					e.AccessSource.Name,
					e.Decision,
					e.Flag,
					admin,
				})
			}

			t := cmdutil.NewTable("ID", "EMAIL", "NAME", "SOURCE", "DECISION", "FLAG", "ADMIN").Rows(rows...)

			_, _ = fmt.Fprintln(f.IOStreams.Out, t)

			if totalCount > len(entries) {
				_, _ = fmt.Fprintf(
					f.IOStreams.ErrOut,
					"\nShowing %d of %d entries\n",
					len(entries),
					totalCount,
				)
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of entries to list")
	cmd.Flags().StringVar(&flagOrderBy, "order-by", "", "Order by field (CREATED_AT)")
	cmd.Flags().StringVar(&flagOrderDir, "order-direction", "DESC", "Sort direction (ASC, DESC)")
	cmd.Flags().StringVar(&flagSourceID, "source-id", "", "Filter by access source ID")
	cmd.Flags().StringVar(&flagDecision, "decision", "", "Filter by decision (PENDING, APPROVED, REVOKE, DEFER, ESCALATE)")
	cmd.Flags().StringVar(&flagFlag, "flag", "", "Filter by flag (NONE, ORPHANED, INACTIVE, EXCESSIVE, ROLE_MISMATCH, NEW)")
	cmd.Flags().StringVar(&flagIncTag, "incremental-tag", "", "Filter by incremental tag (NEW, REMOVED, UNCHANGED)")
	flagIsAdmin = cmd.Flags().Bool("is-admin", false, "Filter by admin status")
	cmd.Flags().StringVar(&flagAuthMethod, "auth-method", "", "Filter by auth method (SSO, PASSWORD, API_KEY, SERVICE_ACCOUNT, UNKNOWN)")
	cmd.Flags().StringVar(&flagAccountType, "account-type", "", "Filter by account type (USER, SERVICE_ACCOUNT)")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
