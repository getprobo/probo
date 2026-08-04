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

package breach

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const incidentFields = `
  id
  title
  description
  occurredAt
  discoveredAt
  awarenessAt
  affectedDataSubjects
  affectedDataRecords
  personalDataTypes
  affectedSystem
  likelyConsequences
  containmentActions
  potentialPhysicalHarm
  potentialFinancialLoss
  potentialCreditOrPropertyDamage
  potentialIllegalUse
  sensitivePersonalData
  potentialIdentityFraud
  significantHarm
  significantScale
  notificationRecommendation
  notificationReasons
  notificationDecision
  decisionRationale
  decisionEvidence
  assessedByProfileId
  assessedAt
  ruleVersion
  ruleSource
  commissionerNotificationDueAt
  commissionerNotificationOverdue
  commissionerNotifiedAt
  commissionerNotificationReference
  commissionerConfirmationReceivedAt
  commissionerConfirmationReference
  phasedInformationDueAt
  delayedNotificationReason
  delayedNotificationEvidence
  dataSubjectsNotificationDueAt
  dataSubjectsNotificationOverdue
  dataSubjectsNotifiedAt
  dataSubjectsNotificationEvidence
  status
  createdByProfileId
  createdAt
  updatedAt
`

const listQuery = `
query($id: ID!, $first: Int, $after: CursorKey) {
  node(id: $id) {
    ... on Organization {
      malaysiaPDPABreachIncidents(first: $first, after: $after, orderBy: {field: CREATED_AT, direction: DESC}) {
        totalCount
        edges { node { ` + incidentFields + ` } }
        pageInfo { hasNextPage endCursor }
      }
    }
  }
}`

const viewQuery = `
query($id: ID!) {
  node(id: $id) {
    ... on MalaysiaPDPABreachIncident { ` + incidentFields + ` }
  }
}`

const createMutation = `
mutation($input: CreateMalaysiaPDPABreachIncidentInput!) {
  createMalaysiaPDPABreachIncident(input: $input) {
    incidentEdge { node { ` + incidentFields + ` } }
  }
}`

const updateMutation = `
mutation($input: UpdateMalaysiaPDPABreachIncidentInput!) {
  updateMalaysiaPDPABreachIncident(input: $input) {
    incident { ` + incidentFields + ` }
  }
}`

const transitionMutation = `
mutation($input: TransitionMalaysiaPDPABreachStatusInput!) {
  transitionMalaysiaPDPABreachStatus(input: $input) {
    incident { ` + incidentFields + ` }
    historyEdge { node { id fromStatus toStatus changedByProfileId reason createdAt } }
  }
}`

const historyQuery = `
query($id: ID!, $first: Int, $after: CursorKey) {
  node(id: $id) {
    ... on MalaysiaPDPABreachIncident {
      statusHistory(first: $first, after: $after, orderBy: {field: CREATED_AT, direction: DESC}) {
        totalCount
        edges { node { id incidentId fromStatus toStatus changedByProfileId reason createdAt } }
        pageInfo { hasNextPage endCursor }
      }
    }
  }
}`

type incident struct {
	ID                                 string   `json:"id"`
	Title                              string   `json:"title"`
	Description                        *string  `json:"description"`
	OccurredAt                         *string  `json:"occurredAt"`
	DiscoveredAt                       string   `json:"discoveredAt"`
	AwarenessAt                        string   `json:"awarenessAt"`
	AffectedDataSubjects               int64    `json:"affectedDataSubjects"`
	AffectedDataRecords                int64    `json:"affectedDataRecords"`
	PersonalDataTypes                  string   `json:"personalDataTypes"`
	AffectedSystem                     *string  `json:"affectedSystem"`
	LikelyConsequences                 *string  `json:"likelyConsequences"`
	ContainmentActions                 *string  `json:"containmentActions"`
	PotentialPhysicalHarm              bool     `json:"potentialPhysicalHarm"`
	PotentialFinancialLoss             bool     `json:"potentialFinancialLoss"`
	PotentialCreditOrPropertyDamage    bool     `json:"potentialCreditOrPropertyDamage"`
	PotentialIllegalUse                bool     `json:"potentialIllegalUse"`
	SensitivePersonalData              bool     `json:"sensitivePersonalData"`
	PotentialIdentityFraud             bool     `json:"potentialIdentityFraud"`
	SignificantHarm                    bool     `json:"significantHarm"`
	SignificantScale                   bool     `json:"significantScale"`
	NotificationRecommendation         string   `json:"notificationRecommendation"`
	NotificationReasons                []string `json:"notificationReasons"`
	NotificationDecision               string   `json:"notificationDecision"`
	DecisionRationale                  *string  `json:"decisionRationale"`
	DecisionEvidence                   *string  `json:"decisionEvidence"`
	AssessedByProfileID                string   `json:"assessedByProfileId"`
	AssessedAt                         string   `json:"assessedAt"`
	RuleVersion                        string   `json:"ruleVersion"`
	RuleSource                         string   `json:"ruleSource"`
	CommissionerNotificationDueAt      *string  `json:"commissionerNotificationDueAt"`
	CommissionerNotificationOverdue    bool     `json:"commissionerNotificationOverdue"`
	CommissionerNotifiedAt             *string  `json:"commissionerNotifiedAt"`
	CommissionerNotificationReference  *string  `json:"commissionerNotificationReference"`
	CommissionerConfirmationReceivedAt *string  `json:"commissionerConfirmationReceivedAt"`
	CommissionerConfirmationReference  *string  `json:"commissionerConfirmationReference"`
	PhasedInformationDueAt             *string  `json:"phasedInformationDueAt"`
	DelayedNotificationReason          *string  `json:"delayedNotificationReason"`
	DelayedNotificationEvidence        *string  `json:"delayedNotificationEvidence"`
	DataSubjectsNotificationDueAt      *string  `json:"dataSubjectsNotificationDueAt"`
	DataSubjectsNotificationOverdue    bool     `json:"dataSubjectsNotificationOverdue"`
	DataSubjectsNotifiedAt             *string  `json:"dataSubjectsNotifiedAt"`
	DataSubjectsNotificationEvidence   *string  `json:"dataSubjectsNotificationEvidence"`
	Status                             string   `json:"status"`
	CreatedByProfileID                 string   `json:"createdByProfileId"`
	CreatedAt                          string   `json:"createdAt"`
	UpdatedAt                          string   `json:"updatedAt"`
}

type statusHistory struct {
	ID                 string  `json:"id"`
	IncidentID         string  `json:"incidentId"`
	FromStatus         *string `json:"fromStatus"`
	ToStatus           string  `json:"toStatus"`
	ChangedByProfileID string  `json:"changedByProfileId"`
	Reason             *string `json:"reason"`
	CreatedAt          string  `json:"createdAt"`
}

type createFlags struct {
	organizationID                  string
	title                           string
	description                     string
	occurredAt                      string
	discoveredAt                    string
	awarenessAt                     string
	affectedDataSubjects            int64
	affectedDataRecords             int64
	personalDataTypes               string
	affectedSystem                  string
	likelyConsequences              string
	containmentActions              string
	potentialPhysicalHarm           bool
	potentialFinancialLoss          bool
	potentialCreditOrPropertyDamage bool
	potentialIllegalUse             bool
	sensitivePersonalData           bool
	potentialIdentityFraud          bool
	notificationDecision            string
	decisionRationale               string
	decisionEvidence                string
}

func NewCmdBreach(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "breach <command>",
		Short: "Manage Malaysia PDPA personal data breaches",
	}

	cmd.AddCommand(newCmdList(f))
	cmd.AddCommand(newCmdView(f))
	cmd.AddCommand(newCmdCreate(f))
	cmd.AddCommand(newCmdUpdate(f))
	cmd.AddCommand(newCmdTransition(f))
	cmd.AddCommand(newCmdHistory(f))

	return cmd
}

func newClient(f *cmdutil.Factory) (*api.Client, string, error) {
	cfg, err := f.Config()
	if err != nil {
		return nil, "", err
	}

	host, hc, err := cfg.DefaultHost()
	if err != nil {
		return nil, "", err
	}

	client := api.NewClient(
		host,
		hc.Token,
		"/api/console/v1/graphql",
		cfg.HTTPTimeoutDuration(),
		cmdutil.TokenRefreshOption(cfg, host, hc),
	)

	return client, hc.Organization, nil
}

func newCmdList(f *cmdutil.Factory) *cobra.Command {
	var organizationID string
	var limit int
	var output *string

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List Malaysia PDPA breach incidents",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(output); err != nil {
				return err
			}
			if err := cmdutil.ValidateLimit(limit); err != nil {
				return err
			}

			client, defaultOrganizationID, err := newClient(f)
			if err != nil {
				return err
			}
			if organizationID == "" {
				organizationID = defaultOrganizationID
			}
			if organizationID == "" {
				return fmt.Errorf("organization ID is required: pass --org or run `prb auth login`")
			}

			incidents, totalCount, err := api.Paginate(client, listQuery, map[string]any{"id": organizationID}, limit, func(data json.RawMessage) (*api.Connection[incident], error) {
				var response struct {
					Node *struct {
						Incidents api.Connection[incident] `json:"malaysiaPDPABreachIncidents"`
					} `json:"node"`
				}
				if err := json.Unmarshal(data, &response); err != nil {
					return nil, err
				}
				if response.Node == nil {
					return nil, fmt.Errorf("organization %s not found", organizationID)
				}
				return &response.Node.Incidents, nil
			})
			if err != nil {
				return err
			}

			if *output == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, incidents)
			}
			if len(incidents) == 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "No Malaysia PDPA breach incidents found.")
				return nil
			}

			rows := make([][]string, len(incidents))
			for index, current := range incidents {
				rows[index] = []string{
					current.ID,
					current.Title,
					current.Status,
					current.NotificationRecommendation,
					cmdutil.FormatTime(current.AwarenessAt),
				}
			}
			table := cmdutil.NewTable("ID", "TITLE", "STATUS", "RECOMMENDATION", "AWARENESS AT").Rows(rows...)
			_, _ = fmt.Fprintln(f.IOStreams.Out, table)
			_, _ = fmt.Fprintf(f.IOStreams.Out, "\nShowing %d of %d incident(s).\n", len(incidents), totalCount)
			return nil
		},
	}

	cmd.Flags().StringVar(&organizationID, "org", "", "Organization ID")
	cmd.Flags().IntVar(&limit, "limit", 100, "Maximum number of incidents to return")
	output = cmdutil.AddOutputFlag(cmd)
	return cmd
}

func newCmdView(f *cmdutil.Factory) *cobra.Command {
	var output *string
	cmd := &cobra.Command{
		Use:   "view <incident-id>",
		Short: "View a Malaysia PDPA breach incident",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(output); err != nil {
				return err
			}
			client, _, err := newClient(f)
			if err != nil {
				return err
			}
			data, err := client.Do(viewQuery, map[string]any{"id": args[0]})
			if err != nil {
				return err
			}
			var response struct {
				Node *incident `json:"node"`
			}
			if err := json.Unmarshal(data, &response); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}
			if response.Node == nil {
				return fmt.Errorf("incident %s not found", args[0])
			}
			if *output == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, response.Node)
			}
			return renderIncident(f, response.Node)
		},
	}
	output = cmdutil.AddOutputFlag(cmd)
	return cmd
}

func newCmdCreate(f *cmdutil.Factory) *cobra.Command {
	flags := &createFlags{}
	var output *string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Record and assess a Malaysia PDPA breach incident",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(output); err != nil {
				return err
			}
			if err := cmdutil.ValidateEnum("notification-decision", flags.notificationDecision, []string{"PENDING", "NOT_REQUIRED", "COMMISSIONER_ONLY", "COMMISSIONER_AND_DATA_SUBJECTS"}); err != nil {
				return err
			}

			client, defaultOrganizationID, err := newClient(f)
			if err != nil {
				return err
			}
			if flags.organizationID == "" {
				flags.organizationID = defaultOrganizationID
			}
			if flags.organizationID == "" {
				return fmt.Errorf("organization ID is required: pass --org or run `prb auth login`")
			}

			discoveredAt, err := parseRFC3339("discovered-at", flags.discoveredAt)
			if err != nil {
				return err
			}
			awarenessAt, err := parseRFC3339("awareness-at", flags.awarenessAt)
			if err != nil {
				return err
			}

			input := map[string]any{
				"organizationId":                  flags.organizationID,
				"title":                           flags.title,
				"discoveredAt":                    discoveredAt,
				"awarenessAt":                     awarenessAt,
				"affectedDataSubjects":            flags.affectedDataSubjects,
				"affectedDataRecords":             flags.affectedDataRecords,
				"personalDataTypes":               flags.personalDataTypes,
				"potentialPhysicalHarm":           flags.potentialPhysicalHarm,
				"potentialFinancialLoss":          flags.potentialFinancialLoss,
				"potentialCreditOrPropertyDamage": flags.potentialCreditOrPropertyDamage,
				"potentialIllegalUse":             flags.potentialIllegalUse,
				"sensitivePersonalData":           flags.sensitivePersonalData,
				"potentialIdentityFraud":          flags.potentialIdentityFraud,
				"notificationDecision":            flags.notificationDecision,
			}
			setOptionalString(input, "description", flags.description)
			setOptionalString(input, "affectedSystem", flags.affectedSystem)
			setOptionalString(input, "likelyConsequences", flags.likelyConsequences)
			setOptionalString(input, "containmentActions", flags.containmentActions)
			setOptionalString(input, "decisionRationale", flags.decisionRationale)
			setOptionalString(input, "decisionEvidence", flags.decisionEvidence)
			if flags.occurredAt != "" {
				occurredAt, err := parseRFC3339("occurred-at", flags.occurredAt)
				if err != nil {
					return err
				}
				input["occurredAt"] = occurredAt
			}

			data, err := client.Do(createMutation, map[string]any{"input": input})
			if err != nil {
				return err
			}
			var response struct {
				Create struct {
					IncidentEdge struct {
						Node *incident `json:"node"`
					} `json:"incidentEdge"`
				} `json:"createMalaysiaPDPABreachIncident"`
			}
			if err := json.Unmarshal(data, &response); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}
			if *output == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, response.Create.IncidentEdge.Node)
			}
			return renderIncident(f, response.Create.IncidentEdge.Node)
		},
	}

	addCreateFlags(cmd, flags)
	output = cmdutil.AddOutputFlag(cmd)
	return cmd
}

func newCmdUpdate(f *cmdutil.Factory) *cobra.Command {
	var notificationDecision string
	var decisionRationale string
	var decisionEvidence string
	var commissionerNotifiedAt string
	var commissionerNotificationReference string
	var commissionerConfirmationReceivedAt string
	var commissionerConfirmationReference string
	var delayedNotificationReason string
	var delayedNotificationEvidence string
	var dataSubjectsNotifiedAt string
	var dataSubjectsNotificationEvidence string
	var clearCommissionerNotification bool
	var clearCommissionerConfirmation bool
	var clearDataSubjectsNotification bool
	var output *string

	cmd := &cobra.Command{
		Use:   "update <incident-id>",
		Short: "Update a breach notification decision and evidence",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(output); err != nil {
				return err
			}
			input := map[string]any{"id": args[0]}
			if cmd.Flags().Changed("notification-decision") {
				if err := cmdutil.ValidateEnum("notification-decision", notificationDecision, []string{"PENDING", "NOT_REQUIRED", "COMMISSIONER_ONLY", "COMMISSIONER_AND_DATA_SUBJECTS"}); err != nil {
					return err
				}
				input["notificationDecision"] = notificationDecision
			}
			copyChangedString(cmd, input, "decision-rationale", "decisionRationale", decisionRationale)
			copyChangedString(cmd, input, "decision-evidence", "decisionEvidence", decisionEvidence)
			copyChangedString(cmd, input, "commissioner-notification-reference", "commissionerNotificationReference", commissionerNotificationReference)
			copyChangedString(cmd, input, "commissioner-confirmation-reference", "commissionerConfirmationReference", commissionerConfirmationReference)
			copyChangedString(cmd, input, "delayed-notification-reason", "delayedNotificationReason", delayedNotificationReason)
			copyChangedString(cmd, input, "delayed-notification-evidence", "delayedNotificationEvidence", delayedNotificationEvidence)
			copyChangedString(cmd, input, "data-subjects-notification-evidence", "dataSubjectsNotificationEvidence", dataSubjectsNotificationEvidence)

			for _, candidate := range []struct{ flagName, inputName, value string }{
				{"commissioner-notified-at", "commissionerNotifiedAt", commissionerNotifiedAt},
				{"commissioner-confirmation-received-at", "commissionerConfirmationReceivedAt", commissionerConfirmationReceivedAt},
				{"data-subjects-notified-at", "dataSubjectsNotifiedAt", dataSubjectsNotifiedAt},
			} {
				if cmd.Flags().Changed(candidate.flagName) {
					parsed, err := parseRFC3339(candidate.flagName, candidate.value)
					if err != nil {
						return err
					}
					input[candidate.inputName] = parsed
				}
			}

			if clearCommissionerNotification {
				input["commissionerNotifiedAt"] = nil
				input["commissionerNotificationReference"] = nil
			}
			if clearCommissionerConfirmation {
				input["commissionerConfirmationReceivedAt"] = nil
				input["commissionerConfirmationReference"] = nil
			}
			if clearDataSubjectsNotification {
				input["dataSubjectsNotifiedAt"] = nil
				input["dataSubjectsNotificationEvidence"] = nil
			}
			if len(input) == 1 {
				return fmt.Errorf("at least one update flag is required")
			}

			client, _, err := newClient(f)
			if err != nil {
				return err
			}
			data, err := client.Do(updateMutation, map[string]any{"input": input})
			if err != nil {
				return err
			}
			var response struct {
				Update struct {
					Incident *incident `json:"incident"`
				} `json:"updateMalaysiaPDPABreachIncident"`
			}
			if err := json.Unmarshal(data, &response); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}
			if *output == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, response.Update.Incident)
			}
			return renderIncident(f, response.Update.Incident)
		},
	}

	cmd.Flags().StringVar(&notificationDecision, "notification-decision", "", "Human notification decision")
	cmd.Flags().StringVar(&decisionRationale, "decision-rationale", "", "Reason for the human notification decision")
	cmd.Flags().StringVar(&decisionEvidence, "decision-evidence", "", "Evidence supporting the decision")
	cmd.Flags().StringVar(&commissionerNotifiedAt, "commissioner-notified-at", "", "Commissioner notification time in RFC3339 format")
	cmd.Flags().StringVar(&commissionerNotificationReference, "commissioner-notification-reference", "", "Commissioner notification reference")
	cmd.Flags().StringVar(&commissionerConfirmationReceivedAt, "commissioner-confirmation-received-at", "", "Commissioner confirmation time in RFC3339 format")
	cmd.Flags().StringVar(&commissionerConfirmationReference, "commissioner-confirmation-reference", "", "Commissioner confirmation reference")
	cmd.Flags().StringVar(&delayedNotificationReason, "delayed-notification-reason", "", "Reason for a late Commissioner notification")
	cmd.Flags().StringVar(&delayedNotificationEvidence, "delayed-notification-evidence", "", "Evidence supporting the late-notification reason")
	cmd.Flags().StringVar(&dataSubjectsNotifiedAt, "data-subjects-notified-at", "", "Data-subject notification time in RFC3339 format")
	cmd.Flags().StringVar(&dataSubjectsNotificationEvidence, "data-subjects-notification-evidence", "", "Data-subject notification evidence")
	cmd.Flags().BoolVar(&clearCommissionerNotification, "clear-commissioner-notification", false, "Clear Commissioner notification time and reference")
	cmd.Flags().BoolVar(&clearCommissionerConfirmation, "clear-commissioner-confirmation", false, "Clear Commissioner confirmation time and reference")
	cmd.Flags().BoolVar(&clearDataSubjectsNotification, "clear-data-subjects-notification", false, "Clear data-subject notification time and evidence")
	cmd.MarkFlagsMutuallyExclusive("commissioner-notified-at", "clear-commissioner-notification")
	cmd.MarkFlagsMutuallyExclusive("commissioner-confirmation-received-at", "clear-commissioner-confirmation")
	cmd.MarkFlagsMutuallyExclusive("data-subjects-notified-at", "clear-data-subjects-notification")
	output = cmdutil.AddOutputFlag(cmd)
	return cmd
}

func newCmdTransition(f *cmdutil.Factory) *cobra.Command {
	var toStatus string
	var reason string
	var output *string
	cmd := &cobra.Command{
		Use:   "transition <incident-id>",
		Short: "Move a breach incident to an allowed status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(output); err != nil {
				return err
			}
			if err := cmdutil.ValidateEnum("to", toStatus, []string{"OPEN", "ASSESSING", "CONTAINED", "CLOSED"}); err != nil {
				return err
			}
			input := map[string]any{"id": args[0], "toStatus": toStatus}
			setOptionalString(input, "reason", reason)
			client, _, err := newClient(f)
			if err != nil {
				return err
			}
			data, err := client.Do(transitionMutation, map[string]any{"input": input})
			if err != nil {
				return err
			}
			var response struct {
				Transition struct {
					Incident    *incident `json:"incident"`
					HistoryEdge struct {
						Node *statusHistory `json:"node"`
					} `json:"historyEdge"`
				} `json:"transitionMalaysiaPDPABreachStatus"`
			}
			if err := json.Unmarshal(data, &response); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}
			if *output == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, response.Transition)
			}
			return renderIncident(f, response.Transition.Incident)
		},
	}
	cmd.Flags().StringVar(&toStatus, "to", "", "Target status (OPEN, ASSESSING, CONTAINED, CLOSED)")
	cmd.Flags().StringVar(&reason, "reason", "", "Reason for the status change")
	_ = cmd.MarkFlagRequired("to")
	output = cmdutil.AddOutputFlag(cmd)
	return cmd
}

func newCmdHistory(f *cmdutil.Factory) *cobra.Command {
	var limit int
	var output *string
	cmd := &cobra.Command{
		Use:   "history <incident-id>",
		Short: "List immutable breach status history",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(output); err != nil {
				return err
			}
			if err := cmdutil.ValidateLimit(limit); err != nil {
				return err
			}
			client, _, err := newClient(f)
			if err != nil {
				return err
			}
			history, _, err := api.Paginate(client, historyQuery, map[string]any{"id": args[0]}, limit, func(data json.RawMessage) (*api.Connection[statusHistory], error) {
				var response struct {
					Node *struct {
						History api.Connection[statusHistory] `json:"statusHistory"`
					} `json:"node"`
				}
				if err := json.Unmarshal(data, &response); err != nil {
					return nil, err
				}
				if response.Node == nil {
					return nil, fmt.Errorf("incident %s not found", args[0])
				}
				return &response.Node.History, nil
			})
			if err != nil {
				return err
			}
			if *output == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, history)
			}
			rows := make([][]string, len(history))
			for index, entry := range history {
				fromStatus := "-"
				if entry.FromStatus != nil {
					fromStatus = *entry.FromStatus
				}
				reason := "-"
				if entry.Reason != nil {
					reason = *entry.Reason
				}
				rows[index] = []string{fromStatus, entry.ToStatus, reason, cmdutil.FormatTime(entry.CreatedAt)}
			}
			table := cmdutil.NewTable("FROM", "TO", "REASON", "CHANGED AT").Rows(rows...)
			_, _ = fmt.Fprintln(f.IOStreams.Out, table)
			return nil
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 100, "Maximum number of history entries to return")
	output = cmdutil.AddOutputFlag(cmd)
	return cmd
}

func addCreateFlags(cmd *cobra.Command, flags *createFlags) {
	cmd.Flags().StringVar(&flags.organizationID, "org", "", "Organization ID")
	cmd.Flags().StringVar(&flags.title, "title", "", "Incident title")
	cmd.Flags().StringVar(&flags.description, "description", "", "Incident description")
	cmd.Flags().StringVar(&flags.occurredAt, "occurred-at", "", "Occurrence time in RFC3339 format")
	cmd.Flags().StringVar(&flags.discoveredAt, "discovered-at", "", "Discovery time in RFC3339 format")
	cmd.Flags().StringVar(&flags.awarenessAt, "awareness-at", "", "Awareness time in RFC3339 format")
	cmd.Flags().Int64Var(&flags.affectedDataSubjects, "affected-data-subjects", 0, "Estimated affected data subjects")
	cmd.Flags().Int64Var(&flags.affectedDataRecords, "affected-data-records", 0, "Estimated affected data records")
	cmd.Flags().StringVar(&flags.personalDataTypes, "personal-data-types", "", "Affected personal data types")
	cmd.Flags().StringVar(&flags.affectedSystem, "affected-system", "", "Affected system")
	cmd.Flags().StringVar(&flags.likelyConsequences, "likely-consequences", "", "Likely consequences")
	cmd.Flags().StringVar(&flags.containmentActions, "containment-actions", "", "Containment actions")
	cmd.Flags().BoolVar(&flags.potentialPhysicalHarm, "potential-physical-harm", false, "Potential physical harm")
	cmd.Flags().BoolVar(&flags.potentialFinancialLoss, "potential-financial-loss", false, "Potential financial loss")
	cmd.Flags().BoolVar(&flags.potentialCreditOrPropertyDamage, "potential-credit-or-property-damage", false, "Potential credit or property damage")
	cmd.Flags().BoolVar(&flags.potentialIllegalUse, "potential-illegal-use", false, "Potential illegal use")
	cmd.Flags().BoolVar(&flags.sensitivePersonalData, "sensitive-personal-data", false, "Sensitive personal data involved")
	cmd.Flags().BoolVar(&flags.potentialIdentityFraud, "potential-identity-fraud", false, "Potential identity fraud")
	cmd.Flags().StringVar(&flags.notificationDecision, "notification-decision", "PENDING", "Human notification decision")
	cmd.Flags().StringVar(&flags.decisionRationale, "decision-rationale", "", "Reason for the human notification decision")
	cmd.Flags().StringVar(&flags.decisionEvidence, "decision-evidence", "", "Evidence supporting the decision")
	_ = cmd.MarkFlagRequired("title")
	_ = cmd.MarkFlagRequired("discovered-at")
	_ = cmd.MarkFlagRequired("awareness-at")
	_ = cmd.MarkFlagRequired("personal-data-types")
}

func renderIncident(f *cmdutil.Factory, current *incident) error {
	if current == nil {
		return fmt.Errorf("empty incident response")
	}

	rows := [][]string{
		{"ID", current.ID},
		{"Title", current.Title},
		{"Status", current.Status},
		{"Recommendation", current.NotificationRecommendation},
		{"Decision", current.NotificationDecision},
		{"Significant harm", fmt.Sprintf("%t", current.SignificantHarm)},
		{"Significant scale", fmt.Sprintf("%t", current.SignificantScale)},
		{"Awareness at", cmdutil.FormatTime(current.AwarenessAt)},
	}
	if current.CommissionerNotificationDueAt != nil {
		rows = append(rows, []string{"Commissioner due", cmdutil.FormatTime(*current.CommissionerNotificationDueAt)})
	}
	if current.DataSubjectsNotificationDueAt != nil {
		rows = append(rows, []string{"Data subjects due", cmdutil.FormatTime(*current.DataSubjectsNotificationDueAt)})
	}
	table := cmdutil.NewTable("FIELD", "VALUE").Rows(rows...)
	_, _ = fmt.Fprintln(f.IOStreams.Out, table)
	return nil
}

func parseRFC3339(flagName, value string) (string, error) {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return "", fmt.Errorf("invalid --%s: use RFC3339 format: %w", flagName, err)
	}
	return parsed.Format(time.RFC3339), nil
}

func setOptionalString(input map[string]any, name, value string) {
	if value != "" {
		input[name] = value
	}
}

func copyChangedString(cmd *cobra.Command, input map[string]any, flagName, inputName, value string) {
	if cmd.Flags().Changed(flagName) {
		input[inputName] = value
	}
}
