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

package list

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const listQuery = `
query(
  $id: ID!,
  $first: Int,
  $after: CursorKey,
  $orderBy: AccessReviewEntryOrder,
  $filter: AccessReviewEntryFilter
) {
  node(id: $id) {
    __typename
    ... on AccessReviewCampaign {
      entries(
        first: $first,
        after: $after,
        orderBy: $orderBy,
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
            active
            mfaStatus
            authMethod
            accountType
            lastLogin
            externalId
            flags
            flagReasons
            decision
            decisionNote
            campaignSource {
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

const listBySourceQuery = `
query(
  $id: ID!,
  $first: Int,
  $after: CursorKey,
  $orderBy: AccessReviewEntryOrder,
  $filter: AccessReviewEntryFilter
) {
  node(id: $id) {
    __typename
    ... on AccessReviewCampaignSource {
      entries(
        first: $first,
        after: $after,
        orderBy: $orderBy,
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
            active
            mfaStatus
            authMethod
            accountType
            lastLogin
            externalId
            flags
            flagReasons
            decision
            decisionNote
            campaignSource {
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
	ID             string   `json:"id"`
	Email          string   `json:"email"`
	FullName       string   `json:"fullName"`
	Role           string   `json:"role"`
	JobTitle       string   `json:"jobTitle"`
	IsAdmin        *bool    `json:"isAdmin"`
	Active         *bool    `json:"active"`
	MfaStatus      string   `json:"mfaStatus"`
	AuthMethod     string   `json:"authMethod"`
	AccountType    string   `json:"accountType"`
	LastLogin      *string  `json:"lastLogin"`
	ExternalID     string   `json:"externalId"`
	Flags          []string `json:"flags"`
	FlagReasons    []string `json:"flagReasons"`
	Decision       string   `json:"decision"`
	DecisionNote   *string  `json:"decisionNote"`
	CampaignSource struct {
		Name string `json:"name"`
	} `json:"campaignSource"`
	CreatedAt string `json:"createdAt"`
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagLimit            int
		flagAfter            string
		flagOrderBy          string
		flagOrderDir         string
		flagCampaignSourceID string
		flagDecision         string
		flagFlag             string
		flagIsAdmin          *bool
		flagActive           *bool
		flagAuthMethod       string
		flagAccountType      string
		flagOutput           *string
	)

	cmd := &cobra.Command{
		Use:     "list [<campaign-id>]",
		Short:   "List access entries for a campaign or campaign source",
		Aliases: []string{"ls"},
		Args:    cobra.MaximumNArgs(1),
		Example: `  # List all entries for a campaign
  prb access-review entry list <campaign-id>

  # List entries for a specific campaign source
  prb access-review entry list --source-id <campaign-source-id>

  # List only pending entries
  prb access-review entry list <campaign-id> --decision PENDING

  # List flagged entries
  prb access-review entry list <campaign-id> --flag ORPHANED`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return err
			}

			if err := cmdutil.ValidateLimit(flagLimit); err != nil {
				return err
			}

			if err := cmdutil.ValidateEnum("order-direction", flagOrderDir, []string{"ASC", "DESC"}); err != nil {
				return err
			}

			if flagCampaignSourceID == "" && len(args) == 0 {
				return fmt.Errorf("campaign ID is required when --source-id is not set")
			}

			client, _, err := f.GraphQLClient(cmdutil.ConsoleEndpoint)
			if err != nil {
				return err
			}

			variables := map[string]any{}

			if len(args) > 0 {
				variables["id"] = args[0]
			}

			if flagCampaignSourceID != "" {
				variables["id"] = flagCampaignSourceID
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
					[]string{
						"NONE", "ORPHANED", "INACTIVE", "EXCESSIVE", "ROLE_MISMATCH",
						"NEW", "DORMANT", "TERMINATED_USER", "CONTRACTOR_EXPIRED",
						"SOD_CONFLICT", "PRIVILEGED_ACCESS", "ROLE_CREEP",
						"NO_BUSINESS_JUSTIFICATION", "OUT_OF_DEPARTMENT", "SHARED_ACCOUNT",
					},
				); err != nil {
					return err
				}

				filter["flag"] = flagFlag
			}

			if cmd.Flags().Changed("is-admin") {
				filter["isAdmin"] = *flagIsAdmin
			}

			if cmd.Flags().Changed("active") {
				filter["active"] = *flagActive
			}

			if flagAuthMethod != "" {
				if err := cmdutil.ValidateEnum(
					"auth-method",
					flagAuthMethod,
					[]string{"SSO", "PASSWORD", "API_KEY", "OAUTH2", "SSH", "SERVICE_ACCOUNT", "UNKNOWN"},
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

			query := listQuery
			expectedTypename := "AccessReviewCampaign"
			notFoundLabel := "campaign"

			if flagCampaignSourceID != "" {
				query = listBySourceQuery
				expectedTypename = "AccessReviewCampaignSource"
				notFoundLabel = "campaign source"
			}

			entries, totalCount, nextCursor, err := api.Paginate(
				client,
				query,
				variables,
				flagLimit,
				flagAfter,
				api.NodeConnection[entryNode](
					"entries",
					expectedTypename,
					fmt.Sprintf("%s %s not found", notFoundLabel, variables["id"]),
				),
			)
			if err != nil {
				return err
			}

			return cmdutil.PrintList(
				f.IOStreams,
				*flagOutput,
				entries,
				totalCount,
				cmdutil.ListView[entryNode]{
					Empty:      "No access entries found.",
					Noun:       "entries",
					NextCursor: nextCursor,
					Headers:    []string{"ID", "EMAIL", "NAME", "SOURCE", "DECISION", "FLAGS", "ADMIN", "ACTIVE"},
					Row: func(e entryNode) ([]string, error) {
						admin := "unknown"

						if e.IsAdmin != nil {
							if *e.IsAdmin {
								admin = "yes"
							} else {
								admin = "no"
							}
						}

						active := "unknown"

						if e.Active != nil {
							if *e.Active {
								active = "active"
							} else {
								active = "disabled"
							}
						}

						return []string{
							e.ID,
							e.Email,
							e.FullName,
							e.CampaignSource.Name,
							e.Decision,
							strings.Join(e.Flags, ","),
							admin,
							active,
						}, nil
					},
				},
			)
		},
	}

	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of entries to list")
	cmdutil.AddAfterFlag(cmd, &flagAfter)
	cmd.Flags().StringVar(&flagOrderBy, "order-by", "", "Order by field (CREATED_AT)")
	cmd.Flags().StringVar(&flagOrderDir, "order-direction", "DESC", "Sort direction (ASC, DESC)")
	cmd.Flags().StringVar(&flagCampaignSourceID, "source-id", "", "Campaign source ID to list entries for")
	cmd.Flags().StringVar(&flagDecision, "decision", "", "Filter by decision (PENDING, APPROVED, REVOKE, DEFER, ESCALATE)")
	cmd.Flags().StringVar(&flagFlag, "flag", "", "Filter by flag (NONE, ORPHANED, INACTIVE, EXCESSIVE, ROLE_MISMATCH, NEW)")
	flagIsAdmin = cmd.Flags().Bool("is-admin", false, "Filter by admin status")
	flagActive = cmd.Flags().Bool("active", false, "Filter by active status at the source")
	cmd.Flags().StringVar(&flagAuthMethod, "auth-method", "", "Filter by auth method (SSO, PASSWORD, API_KEY, OAUTH2, SSH, SERVICE_ACCOUNT, UNKNOWN)")
	cmd.Flags().StringVar(&flagAccountType, "account-type", "", "Filter by account type (USER, SERVICE_ACCOUNT)")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
