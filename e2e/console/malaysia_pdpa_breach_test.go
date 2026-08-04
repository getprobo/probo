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

package console_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/testutil"
)

const (
	malaysiaPDPABreachFields = `
		id
		title
		awarenessAt
		affectedDataSubjects
		significantHarm
		significantScale
		notificationRecommendation
		notificationReasons
		notificationDecision
		commissionerNotificationDueAt
		commissionerNotifiedAt
		phasedInformationDueAt
		dataSubjectsNotificationDueAt
		status
		createdAt
		updatedAt
	`

	createMalaysiaPDPABreachMutation = `
		mutation CreateMalaysiaPDPABreach($input: CreateMalaysiaPDPABreachIncidentInput!) {
			createMalaysiaPDPABreachIncident(input: $input) {
				incidentEdge { node { ` + malaysiaPDPABreachFields + ` } }
			}
		}
	`

	transitionMalaysiaPDPABreachMutation = `
		mutation TransitionMalaysiaPDPABreach($input: TransitionMalaysiaPDPABreachStatusInput!) {
			transitionMalaysiaPDPABreachStatus(input: $input) {
				incident { ` + malaysiaPDPABreachFields + ` }
				historyEdge { node { id fromStatus toStatus reason createdAt } }
			}
		}
	`

	listMalaysiaPDPABreachQuery = `
		query ListMalaysiaPDPABreaches($id: ID!) {
			node(id: $id) {
				... on Organization {
					malaysiaPDPABreachIncidents(first: 100) {
						totalCount
						edges { node { ` + malaysiaPDPABreachFields + ` } }
					}
				}
			}
		}
	`

	malaysiaPDPABreachHistoryQuery = `
		query MalaysiaPDPABreachHistory($id: ID!) {
			node(id: $id) {
				... on MalaysiaPDPABreachIncident {
					id
					statusHistory(first: 100, orderBy: {field: CREATED_AT, direction: ASC}) {
						totalCount
						edges { node { id fromStatus toStatus reason createdAt } }
					}
				}
			}
		}
	`
)

type malaysiaPDPABreachResult struct {
	ID                            string     `json:"id"`
	Title                         string     `json:"title"`
	AwarenessAt                   time.Time  `json:"awarenessAt"`
	AffectedDataSubjects          int64      `json:"affectedDataSubjects"`
	SignificantHarm               bool       `json:"significantHarm"`
	SignificantScale              bool       `json:"significantScale"`
	NotificationRecommendation    string     `json:"notificationRecommendation"`
	NotificationReasons           []string   `json:"notificationReasons"`
	NotificationDecision          string     `json:"notificationDecision"`
	CommissionerNotificationDueAt *time.Time `json:"commissionerNotificationDueAt"`
	CommissionerNotifiedAt        *time.Time `json:"commissionerNotifiedAt"`
	PhasedInformationDueAt        *time.Time `json:"phasedInformationDueAt"`
	DataSubjectsNotificationDueAt *time.Time `json:"dataSubjectsNotificationDueAt"`
	Status                        string     `json:"status"`
	CreatedAt                     time.Time  `json:"createdAt"`
	UpdatedAt                     time.Time  `json:"updatedAt"`
}

type malaysiaPDPABreachHistoryResult struct {
	ID         string    `json:"id"`
	FromStatus *string   `json:"fromStatus"`
	ToStatus   string    `json:"toStatus"`
	Reason     *string   `json:"reason"`
	CreatedAt  time.Time `json:"createdAt"`
}

func TestMalaysiaPDPABreach_AssessmentAndDeadlines(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	awarenessAt := time.Date(2026, time.August, 4, 8, 0, 0, 0, time.UTC)

	scaleIncident := createMalaysiaPDPABreach(t, owner, map[string]any{
		"organizationId":                  owner.GetOrganizationID().String(),
		"title":                           "Large mailing-list exposure",
		"discoveredAt":                    awarenessAt.Add(-time.Hour).Format(time.RFC3339),
		"awarenessAt":                     awarenessAt.Format(time.RFC3339),
		"affectedDataSubjects":            1_001,
		"affectedDataRecords":             1_001,
		"personalDataTypes":               "Names and email addresses",
		"potentialPhysicalHarm":           false,
		"potentialFinancialLoss":          false,
		"potentialCreditOrPropertyDamage": false,
		"potentialIllegalUse":             false,
		"sensitivePersonalData":           false,
		"potentialIdentityFraud":          false,
		"notificationDecision":            "PENDING",
	})

	assert.True(t, scaleIncident.SignificantScale)
	assert.False(t, scaleIncident.SignificantHarm)
	assert.Equal(t, "COMMISSIONER_ONLY", scaleIncident.NotificationRecommendation)
	assert.Equal(t, []string{"SIGNIFICANT_SCALE"}, scaleIncident.NotificationReasons)
	require.NotNil(t, scaleIncident.CommissionerNotificationDueAt)
	assert.Equal(t, awarenessAt.Add(72*time.Hour), *scaleIncident.CommissionerNotificationDueAt)
	assert.Nil(t, scaleIncident.DataSubjectsNotificationDueAt)

	commissionerNotifiedAt := awarenessAt.Add(24 * time.Hour)
	harmIncident := createMalaysiaPDPABreach(t, owner, map[string]any{
		"organizationId":                    owner.GetOrganizationID().String(),
		"title":                             "Sensitive account-data exposure",
		"discoveredAt":                      awarenessAt.Add(-30 * time.Minute).Format(time.RFC3339),
		"awarenessAt":                       awarenessAt.Format(time.RFC3339),
		"affectedDataSubjects":              25,
		"affectedDataRecords":               75,
		"personalDataTypes":                 "Identity and financial account data",
		"potentialPhysicalHarm":             false,
		"potentialFinancialLoss":            true,
		"potentialCreditOrPropertyDamage":   false,
		"potentialIllegalUse":               false,
		"sensitivePersonalData":             true,
		"potentialIdentityFraud":            true,
		"notificationDecision":              "COMMISSIONER_AND_DATA_SUBJECTS",
		"decisionRationale":                 "Significant harm is reasonably likely.",
		"commissionerNotifiedAt":            commissionerNotifiedAt.Format(time.RFC3339),
		"commissionerNotificationReference": "DBN-2026-0001",
	})

	assert.True(t, harmIncident.SignificantHarm)
	assert.False(t, harmIncident.SignificantScale)
	assert.Equal(t, "COMMISSIONER_AND_DATA_SUBJECTS", harmIncident.NotificationRecommendation)
	require.NotNil(t, harmIncident.DataSubjectsNotificationDueAt)
	assert.Equal(t, commissionerNotifiedAt.Add(7*24*time.Hour), *harmIncident.DataSubjectsNotificationDueAt)
	require.NotNil(t, harmIncident.PhasedInformationDueAt)
	assert.Equal(t, commissionerNotifiedAt.Add(30*24*time.Hour), *harmIncident.PhasedInformationDueAt)
}

func TestMalaysiaPDPABreach_StatusHistoryIsAppended(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	incident := createMalaysiaPDPABreach(t, owner, baseMalaysiaPDPABreachInput(owner, "Workflow incident"))

	var transitionResult struct {
		Transition struct {
			Incident    malaysiaPDPABreachResult `json:"incident"`
			HistoryEdge struct {
				Node malaysiaPDPABreachHistoryResult `json:"node"`
			} `json:"historyEdge"`
		} `json:"transitionMalaysiaPDPABreachStatus"`
	}
	err := owner.Execute(transitionMalaysiaPDPABreachMutation, map[string]any{
		"input": map[string]any{
			"id":       incident.ID,
			"toStatus": "ASSESSING",
			"reason":   "Assessment started",
		},
	}, &transitionResult)
	require.NoError(t, err)
	assert.Equal(t, "ASSESSING", transitionResult.Transition.Incident.Status)
	assert.Equal(t, "OPEN", *transitionResult.Transition.HistoryEdge.Node.FromStatus)
	assert.Equal(t, "ASSESSING", transitionResult.Transition.HistoryEdge.Node.ToStatus)

	var historyResult struct {
		Node *struct {
			History struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node malaysiaPDPABreachHistoryResult `json:"node"`
				} `json:"edges"`
			} `json:"statusHistory"`
		} `json:"node"`
	}
	err = owner.Execute(malaysiaPDPABreachHistoryQuery, map[string]any{"id": incident.ID}, &historyResult)
	require.NoError(t, err)
	require.NotNil(t, historyResult.Node)
	assert.Equal(t, 2, historyResult.Node.History.TotalCount)
	require.Len(t, historyResult.Node.History.Edges, 2)
	assert.Nil(t, historyResult.Node.History.Edges[0].Node.FromStatus)
	assert.Equal(t, "OPEN", historyResult.Node.History.Edges[0].Node.ToStatus)
	assert.Equal(t, "ASSESSING", historyResult.Node.History.Edges[1].Node.ToStatus)

	_, err = owner.Do(transitionMalaysiaPDPABreachMutation, map[string]any{
		"input": map[string]any{"id": incident.ID, "toStatus": "CLOSED"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "to_status")
}

func TestMalaysiaPDPABreach_RBACAndTenantIsolation(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
	admin := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)
	otherOwner := testutil.NewClient(t, testutil.RoleOwner)
	incident := createMalaysiaPDPABreach(t, owner, baseMalaysiaPDPABreachInput(owner, "RBAC incident"))

	var viewerList struct {
		Node *struct {
			Incidents struct {
				TotalCount int `json:"totalCount"`
			} `json:"malaysiaPDPABreachIncidents"`
		} `json:"node"`
	}
	err := viewer.Execute(listMalaysiaPDPABreachQuery, map[string]any{"id": owner.GetOrganizationID().String()}, &viewerList)
	require.NoError(t, err)
	require.NotNil(t, viewerList.Node)
	assert.GreaterOrEqual(t, viewerList.Node.Incidents.TotalCount, 1)

	_, err = viewer.Do(createMalaysiaPDPABreachMutation, map[string]any{
		"input": baseMalaysiaPDPABreachInput(owner, "Viewer cannot create"),
	})
	testutil.RequireForbiddenError(t, err, "viewer cannot create Malaysia PDPA breach incident")

	adminIncident := createMalaysiaPDPABreach(t, admin, baseMalaysiaPDPABreachInput(owner, "Admin incident"))
	assert.NotEmpty(t, adminIncident.ID)

	var inaccessible struct {
		Node *malaysiaPDPABreachResult `json:"node"`
	}
	err = otherOwner.Execute(`query($id: ID!) { node(id: $id) { ... on MalaysiaPDPABreachIncident { id } } }`, map[string]any{"id": incident.ID}, &inaccessible)
	testutil.AssertNodeNotAccessible(t, err, inaccessible.Node == nil, "Malaysia PDPA breach incident")
}

func TestMalaysiaPDPABreach_LateNotificationRequiresReasonAndEvidence(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	input := baseMalaysiaPDPABreachInput(owner, "Late notification")
	awarenessAt, err := time.Parse(time.RFC3339, input["awarenessAt"].(string))
	require.NoError(t, err)
	input["affectedDataSubjects"] = 1_001
	input["notificationDecision"] = "COMMISSIONER_ONLY"
	input["decisionRationale"] = "Significant scale requires Commissioner notification."
	input["commissionerNotifiedAt"] = awarenessAt.Add(73 * time.Hour).Format(time.RFC3339)
	input["commissionerNotificationReference"] = "DBN-LATE-0001"

	resp, err := owner.Do(createMalaysiaPDPABreachMutation, map[string]any{"input": input})
	require.Error(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Errors, 2)

	errorFields := make([]string, 0, len(resp.Errors))
	for _, gqlErr := range resp.Errors {
		if field, ok := gqlErr.Extensions["field"].(string); ok {
			errorFields = append(errorFields, field)
		}
	}
	assert.ElementsMatch(t, []string{
		"delayed_notification_reason",
		"delayed_notification_evidence",
	}, errorFields)
}

func createMalaysiaPDPABreach(t *testing.T, client *testutil.Client, input map[string]any) malaysiaPDPABreachResult {
	t.Helper()
	var result struct {
		Create struct {
			IncidentEdge struct {
				Node malaysiaPDPABreachResult `json:"node"`
			} `json:"incidentEdge"`
		} `json:"createMalaysiaPDPABreachIncident"`
	}
	err := client.Execute(createMalaysiaPDPABreachMutation, map[string]any{"input": input}, &result)
	require.NoError(t, err)
	return result.Create.IncidentEdge.Node
}

func baseMalaysiaPDPABreachInput(client *testutil.Client, title string) map[string]any {
	awarenessAt := time.Now().UTC().Truncate(time.Second)
	return map[string]any{
		"organizationId":                  client.GetOrganizationID().String(),
		"title":                           title,
		"discoveredAt":                    awarenessAt.Add(-time.Hour).Format(time.RFC3339),
		"awarenessAt":                     awarenessAt.Format(time.RFC3339),
		"affectedDataSubjects":            10,
		"affectedDataRecords":             20,
		"personalDataTypes":               "Names and email addresses",
		"potentialPhysicalHarm":           false,
		"potentialFinancialLoss":          false,
		"potentialCreditOrPropertyDamage": false,
		"potentialIllegalUse":             false,
		"sensitivePersonalData":           false,
		"potentialIdentityFraud":          false,
		"notificationDecision":            "PENDING",
	}
}
