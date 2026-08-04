// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

const (
	createFrameworkMutation = `
		mutation CreateFramework($input: CreateFrameworkInput!) {
			createFramework(input: $input) {
				frameworkEdge { node { id } }
			}
		}`

	updateFrameworkMutation = `
		mutation UpdateFramework($input: UpdateFrameworkInput!) {
			updateFramework(input: $input) {
				framework { id }
			}
		}`

	deleteFrameworkMutation = `
		mutation DeleteFramework($input: DeleteFrameworkInput!) {
			deleteFramework(input: $input) {
				deletedFrameworkId
			}
		}`

	listFrameworksQuery = `
		query GetFrameworks($id: ID!) {
			node(id: $id) {
				... on Organization {
					frameworks(first: 10) { totalCount }
				}
			}
		}`

	createControlMutation = `
		mutation CreateControl($input: CreateControlInput!) {
			createControl(input: $input) {
				controlEdge { node { id } }
			}
		}`

	updateControlMutation = `
		mutation UpdateControl($input: UpdateControlInput!) {
			updateControl(input: $input) {
				control { id }
			}
		}`

	deleteControlMutation = `
		mutation DeleteControl($input: DeleteControlInput!) {
			deleteControl(input: $input) {
				deletedControlId
			}
		}`

	listControlsQuery = `
		query GetControls($id: ID!) {
			node(id: $id) {
				... on Framework {
					controls(first: 10) { edges { node { id } } }
				}
			}
		}`

	createMeasureMutation = `
		mutation CreateMeasure($input: CreateMeasureInput!) {
			createMeasure(input: $input) {
				measureEdge { node { id } }
			}
		}`

	updateMeasureMutation = `
		mutation UpdateMeasure($input: UpdateMeasureInput!) {
			updateMeasure(input: $input) {
				measure { id }
			}
		}`

	deleteMeasureMutation = `
		mutation DeleteMeasure($input: DeleteMeasureInput!) {
			deleteMeasure(input: $input) {
				deletedMeasureId
			}
		}`

	listMeasuresQuery = `
		query GetMeasures($id: ID!) {
			node(id: $id) {
				... on Organization {
					measures(first: 10) { totalCount }
				}
			}
		}`

	createTaskMutation = `
		mutation CreateTask($input: CreateTaskInput!) {
			createTask(input: $input) {
				taskEdge { node { id } }
			}
		}`

	updateTaskMutation = `
		mutation UpdateTask($input: UpdateTaskInput!) {
			updateTask(input: $input) {
				task { id }
			}
		}`

	deleteTaskMutation = `
		mutation DeleteTask($input: DeleteTaskInput!) {
			deleteTask(input: $input) {
				deletedTaskId
			}
		}`

	listTasksQuery = `
		query GetTasks($id: ID!) {
			node(id: $id) {
				... on Measure {
					tasks(first: 10) { totalCount }
				}
			}
		}`

	createRiskMutation = `
		mutation CreateRisk($input: CreateRiskInput!) {
			createRisk(input: $input) {
				riskEdge { node { id } }
			}
		}`

	updateRiskMutation = `
		mutation UpdateRisk($input: UpdateRiskInput!) {
			updateRisk(input: $input) {
				risk { id }
			}
		}`

	deleteRiskMutation = `
		mutation DeleteRisk($input: DeleteRiskInput!) {
			deleteRisk(input: $input) {
				deletedRiskId
			}
		}`

	listRisksQuery = `
		query GetRisks($id: ID!) {
			node(id: $id) {
				... on Organization {
					risks(first: 10) { totalCount }
				}
			}
		}`

	createThirdPartyMutation = `
		mutation CreateThirdParty($input: CreateThirdPartyInput!) {
			createThirdParty(input: $input) {
				thirdPartyEdge { node { id } }
			}
		}`

	updateThirdPartyMutation = `
		mutation UpdateThirdParty($input: UpdateThirdPartyInput!) {
			updateThirdParty(input: $input) {
				thirdParty { id }
			}
		}`

	deleteThirdPartyMutation = `
		mutation DeleteThirdParty($input: DeleteThirdPartyInput!) {
			deleteThirdParty(input: $input) {
				deletedThirdPartyId
			}
		}`

	listThirdPartiesQuery = `
		query GetThirdParties($id: ID!) {
			node(id: $id) {
				... on Organization {
					thirdParties(first: 10) { totalCount }
				}
			}
		}`

	createAccessReviewSourceMutation = `
		mutation CreateAccessReviewSource($input: CreateAccessReviewSourceInput!) {
			createAccessReviewSource(input: $input) {
				accessReviewSourceEdge { node { id } }
			}
		}`

	updateAccessReviewSourceMutation = `
		mutation UpdateAccessReviewSource($input: UpdateAccessReviewSourceInput!) {
			updateAccessReviewSource(input: $input) {
				accessReviewSource { id }
			}
		}`

	deleteAccessReviewSourceMutation = `
		mutation DeleteAccessReviewSource($input: DeleteAccessReviewSourceInput!) {
			deleteAccessReviewSource(input: $input) {
				deletedAccessReviewSourceId
			}
		}`

	listAccessReviewSourcesQuery = `
		query GetAccessReviewSources($id: ID!) {
			node(id: $id) {
				... on Organization {
					accessReviewSources(first: 10) { totalCount }
				}
			}
		}`

	createAccessReviewCampaignMutation = `
		mutation CreateCampaign($input: CreateAccessReviewCampaignInput!) {
			createAccessReviewCampaign(input: $input) {
				accessReviewCampaignEdge { node { id } }
			}
		}`

	updateAccessReviewCampaignMutation = `
		mutation UpdateCampaign($input: UpdateAccessReviewCampaignInput!) {
			updateAccessReviewCampaign(input: $input) {
				accessReviewCampaign { id }
			}
		}`

	deleteAccessReviewCampaignMutation = `
		mutation DeleteCampaign($input: DeleteAccessReviewCampaignInput!) {
			deleteAccessReviewCampaign(input: $input) {
				deletedAccessReviewCampaignId
			}
		}`

	listAccessReviewCampaignsQuery = `
		query GetCampaigns($id: ID!) {
			node(id: $id) {
				... on Organization {
					accessReviewCampaigns(first: 10) { totalCount }
				}
			}
		}`

	updateOrganizationMutation = `
		mutation UpdateOrganization($input: UpdateOrganizationInput!) {
			updateOrganization(input: $input) {
				organization { id }
			}
		}`

	getOrganizationQuery = `
		query GetOrganization($id: ID!) {
			node(id: $id) {
				... on Organization {
					id
					name
				}
			}
		}`

	listUsersQuery = `
		query GetProfiles($id: ID!) {
			node(id: $id) {
				... on Organization {
					profiles(first: 10) { totalCount }
				}
			}
		}`
)

type (
	rbacShared struct {
		orgID       string
		frameworkID string
		measureID   string
	}

	rbacTestCase struct {
		resource    string
		operation   string
		name        string
		role        testutil.TestRole
		shouldAllow bool
		useConnect  bool
	}
)

func TestRBAC(t *testing.T) {
	t.Parallel()

	org := testutil.NewOrganizationRoles(t)
	ownerClient := org.Client(t, testutil.RoleOwner)
	shared := rbacShared{
		orgID:       ownerClient.GetOrganizationID().String(),
		frameworkID: factory.NewFramework(ownerClient).WithName("RBAC Test Framework").Create(),
		measureID:   factory.NewMeasure(ownerClient).WithName("RBAC Test Measure").Create(),
	}
	_ = factory.NewControl(ownerClient, shared.frameworkID).WithName("RBAC Test Control").Create()
	_ = factory.NewTask(ownerClient, shared.measureID).WithName("RBAC Test Task").Create()

	tests := []rbacTestCase{
		{resource: "framework", operation: "create", name: "owner can create framework", role: testutil.RoleOwner, shouldAllow: true},
		{resource: "framework", operation: "create", name: "admin can create framework", role: testutil.RoleAdmin, shouldAllow: true},
		{resource: "framework", operation: "create", name: "viewer cannot create framework", role: testutil.RoleViewer, shouldAllow: false},
		{resource: "framework", operation: "update", name: "owner can update framework", role: testutil.RoleOwner, shouldAllow: true},
		{resource: "framework", operation: "update", name: "admin can update framework", role: testutil.RoleAdmin, shouldAllow: true},
		{resource: "framework", operation: "update", name: "viewer cannot update framework", role: testutil.RoleViewer, shouldAllow: false},
		{resource: "framework", operation: "delete", name: "owner can delete framework", role: testutil.RoleOwner, shouldAllow: true},
		{resource: "framework", operation: "delete", name: "admin can delete framework", role: testutil.RoleAdmin, shouldAllow: true},
		{resource: "framework", operation: "delete", name: "viewer cannot delete framework", role: testutil.RoleViewer, shouldAllow: false},
		{resource: "framework", operation: "list", name: "owner can list frameworks", role: testutil.RoleOwner, shouldAllow: true},
		{resource: "framework", operation: "list", name: "admin can list frameworks", role: testutil.RoleAdmin, shouldAllow: true},
		{resource: "framework", operation: "list", name: "viewer can list frameworks", role: testutil.RoleViewer, shouldAllow: true},
		{resource: "control", operation: "create", name: "owner can create control", role: testutil.RoleOwner, shouldAllow: true},
		{resource: "control", operation: "create", name: "admin can create control", role: testutil.RoleAdmin, shouldAllow: true},
		{resource: "control", operation: "create", name: "viewer cannot create control", role: testutil.RoleViewer, shouldAllow: false},
		{resource: "control", operation: "update", name: "owner can update control", role: testutil.RoleOwner, shouldAllow: true},
		{resource: "control", operation: "update", name: "admin can update control", role: testutil.RoleAdmin, shouldAllow: true},
		{resource: "control", operation: "update", name: "viewer cannot update control", role: testutil.RoleViewer, shouldAllow: false},
		{resource: "control", operation: "delete", name: "owner can delete control", role: testutil.RoleOwner, shouldAllow: true},
		{resource: "control", operation: "delete", name: "admin can delete control", role: testutil.RoleAdmin, shouldAllow: true},
		{resource: "control", operation: "list", name: "owner can list controls", role: testutil.RoleOwner, shouldAllow: true},
		{resource: "control", operation: "list", name: "admin can list controls", role: testutil.RoleAdmin, shouldAllow: true},
		{resource: "control", operation: "list", name: "viewer can list controls", role: testutil.RoleViewer, shouldAllow: true},
		{resource: "measure", operation: "create", name: "owner can create measure", role: testutil.RoleOwner, shouldAllow: true},
		{resource: "measure", operation: "create", name: "admin can create measure", role: testutil.RoleAdmin, shouldAllow: true},
		{resource: "measure", operation: "create", name: "viewer cannot create measure", role: testutil.RoleViewer, shouldAllow: false},
		{resource: "measure", operation: "update", name: "owner can update measure", role: testutil.RoleOwner, shouldAllow: true},
		{resource: "measure", operation: "update", name: "admin can update measure", role: testutil.RoleAdmin, shouldAllow: true},
		{resource: "measure", operation: "update", name: "viewer cannot update measure", role: testutil.RoleViewer, shouldAllow: false},
		{resource: "measure", operation: "delete", name: "owner can delete measure", role: testutil.RoleOwner, shouldAllow: true},
		{resource: "measure", operation: "delete", name: "admin can delete measure", role: testutil.RoleAdmin, shouldAllow: true},
		{resource: "measure", operation: "delete", name: "viewer cannot delete measure", role: testutil.RoleViewer, shouldAllow: false},
		{resource: "measure", operation: "list", name: "owner can list measures", role: testutil.RoleOwner, shouldAllow: true},
		{resource: "measure", operation: "list", name: "admin can list measures", role: testutil.RoleAdmin, shouldAllow: true},
		{resource: "measure", operation: "list", name: "viewer can list measures", role: testutil.RoleViewer, shouldAllow: true},
		{resource: "task", operation: "create", name: "owner can create task", role: testutil.RoleOwner, shouldAllow: true},
		{resource: "task", operation: "create", name: "admin can create task", role: testutil.RoleAdmin, shouldAllow: true},
		{resource: "task", operation: "create", name: "viewer cannot create task", role: testutil.RoleViewer, shouldAllow: false},
		{resource: "task", operation: "update", name: "owner can update task", role: testutil.RoleOwner, shouldAllow: true},
		{resource: "task", operation: "update", name: "admin can update task", role: testutil.RoleAdmin, shouldAllow: true},
		{resource: "task", operation: "update", name: "viewer cannot update task", role: testutil.RoleViewer, shouldAllow: false},
		{resource: "task", operation: "delete", name: "owner can delete task", role: testutil.RoleOwner, shouldAllow: true},
		{resource: "task", operation: "delete", name: "admin can delete task", role: testutil.RoleAdmin, shouldAllow: true},
		{resource: "task", operation: "delete", name: "viewer cannot delete task", role: testutil.RoleViewer, shouldAllow: false},
		{resource: "task", operation: "list", name: "owner can list tasks", role: testutil.RoleOwner, shouldAllow: true},
		{resource: "task", operation: "list", name: "admin can list tasks", role: testutil.RoleAdmin, shouldAllow: true},
		{resource: "task", operation: "list", name: "viewer can list tasks", role: testutil.RoleViewer, shouldAllow: true},
		{resource: "risk", operation: "create", name: "owner can create risk", role: testutil.RoleOwner, shouldAllow: true},
		{resource: "risk", operation: "create", name: "admin can create risk", role: testutil.RoleAdmin, shouldAllow: true},
		{resource: "risk", operation: "create", name: "viewer cannot create risk", role: testutil.RoleViewer, shouldAllow: false},
		{resource: "risk", operation: "update", name: "owner can update risk", role: testutil.RoleOwner, shouldAllow: true},
		{resource: "risk", operation: "update", name: "admin can update risk", role: testutil.RoleAdmin, shouldAllow: true},
		{resource: "risk", operation: "update", name: "viewer cannot update risk", role: testutil.RoleViewer, shouldAllow: false},
		{resource: "risk", operation: "delete", name: "owner can delete risk", role: testutil.RoleOwner, shouldAllow: true},
		{resource: "risk", operation: "delete", name: "admin can delete risk", role: testutil.RoleAdmin, shouldAllow: true},
		{resource: "risk", operation: "delete", name: "viewer cannot delete risk", role: testutil.RoleViewer, shouldAllow: false},
		{resource: "risk", operation: "list", name: "owner can list risks", role: testutil.RoleOwner, shouldAllow: true},
		{resource: "risk", operation: "list", name: "admin can list risks", role: testutil.RoleAdmin, shouldAllow: true},
		{resource: "risk", operation: "list", name: "viewer can list risks", role: testutil.RoleViewer, shouldAllow: true},
		{resource: "third_party", operation: "create", name: "owner can create thirdParty", role: testutil.RoleOwner, shouldAllow: true},
		{resource: "third_party", operation: "create", name: "admin can create thirdParty", role: testutil.RoleAdmin, shouldAllow: true},
		{resource: "third_party", operation: "create", name: "viewer cannot create thirdParty", role: testutil.RoleViewer, shouldAllow: false},
		{resource: "third_party", operation: "update", name: "owner can update thirdParty", role: testutil.RoleOwner, shouldAllow: true},
		{resource: "third_party", operation: "update", name: "admin can update thirdParty", role: testutil.RoleAdmin, shouldAllow: true},
		{resource: "third_party", operation: "update", name: "viewer cannot update thirdParty", role: testutil.RoleViewer, shouldAllow: false},
		{resource: "third_party", operation: "delete", name: "owner can delete thirdParty", role: testutil.RoleOwner, shouldAllow: true},
		{resource: "third_party", operation: "delete", name: "admin can delete thirdParty", role: testutil.RoleAdmin, shouldAllow: true},
		{resource: "third_party", operation: "delete", name: "viewer cannot delete thirdParty", role: testutil.RoleViewer, shouldAllow: false},
		{resource: "third_party", operation: "list", name: "owner can list third parties", role: testutil.RoleOwner, shouldAllow: true},
		{resource: "third_party", operation: "list", name: "admin can list third parties", role: testutil.RoleAdmin, shouldAllow: true},
		{resource: "third_party", operation: "list", name: "viewer can list third parties", role: testutil.RoleViewer, shouldAllow: true},
		{resource: "access_review_source", operation: "create", name: "owner can create access source", role: testutil.RoleOwner, shouldAllow: true},
		{resource: "access_review_source", operation: "create", name: "admin can create access source", role: testutil.RoleAdmin, shouldAllow: true},
		{resource: "access_review_source", operation: "create", name: "viewer cannot create access source", role: testutil.RoleViewer, shouldAllow: false},
		{resource: "access_review_source", operation: "update", name: "owner can update access source", role: testutil.RoleOwner, shouldAllow: true},
		{resource: "access_review_source", operation: "update", name: "admin can update access source", role: testutil.RoleAdmin, shouldAllow: true},
		{resource: "access_review_source", operation: "update", name: "viewer cannot update access source", role: testutil.RoleViewer, shouldAllow: false},
		{resource: "access_review_source", operation: "delete", name: "owner can delete access source", role: testutil.RoleOwner, shouldAllow: true},
		{resource: "access_review_source", operation: "delete", name: "admin can delete access source", role: testutil.RoleAdmin, shouldAllow: true},
		{resource: "access_review_source", operation: "delete", name: "viewer cannot delete access source", role: testutil.RoleViewer, shouldAllow: false},
		{resource: "access_review_source", operation: "list", name: "owner can list access sources", role: testutil.RoleOwner, shouldAllow: true},
		{resource: "access_review_source", operation: "list", name: "admin can list access sources", role: testutil.RoleAdmin, shouldAllow: true},
		{resource: "access_review_source", operation: "list", name: "viewer can list access sources", role: testutil.RoleViewer, shouldAllow: true},
		{resource: "access_review_campaign", operation: "create", name: "owner can create access review campaign", role: testutil.RoleOwner, shouldAllow: true},
		{resource: "access_review_campaign", operation: "create", name: "admin can create access review campaign", role: testutil.RoleAdmin, shouldAllow: true},
		{resource: "access_review_campaign", operation: "create", name: "viewer cannot create access review campaign", role: testutil.RoleViewer, shouldAllow: false},
		{resource: "access_review_campaign", operation: "update", name: "owner can update access review campaign", role: testutil.RoleOwner, shouldAllow: true},
		{resource: "access_review_campaign", operation: "update", name: "admin can update access review campaign", role: testutil.RoleAdmin, shouldAllow: true},
		{resource: "access_review_campaign", operation: "update", name: "viewer cannot update access review campaign", role: testutil.RoleViewer, shouldAllow: false},
		{resource: "access_review_campaign", operation: "delete", name: "owner can delete access review campaign", role: testutil.RoleOwner, shouldAllow: true},
		{resource: "access_review_campaign", operation: "delete", name: "admin can delete access review campaign", role: testutil.RoleAdmin, shouldAllow: true},
		{resource: "access_review_campaign", operation: "delete", name: "viewer cannot delete access review campaign", role: testutil.RoleViewer, shouldAllow: false},
		{resource: "access_review_campaign", operation: "list", name: "owner can list access review campaigns", role: testutil.RoleOwner, shouldAllow: true},
		{resource: "access_review_campaign", operation: "list", name: "admin can list access review campaigns", role: testutil.RoleAdmin, shouldAllow: true},
		{resource: "access_review_campaign", operation: "list", name: "viewer can list access review campaigns", role: testutil.RoleViewer, shouldAllow: true},
		{resource: "organization", operation: "update", name: "owner can update organization", role: testutil.RoleOwner, shouldAllow: true, useConnect: true},
		{resource: "organization", operation: "update", name: "admin can update organization", role: testutil.RoleAdmin, shouldAllow: true, useConnect: true},
		{resource: "organization", operation: "update", name: "viewer cannot update organization", role: testutil.RoleViewer, shouldAllow: false, useConnect: true},
		{resource: "organization", operation: "get", name: "owner can get organization", role: testutil.RoleOwner, shouldAllow: true, useConnect: true},
		{resource: "organization", operation: "get", name: "admin can get organization", role: testutil.RoleAdmin, shouldAllow: true, useConnect: true},
		{resource: "organization", operation: "get", name: "viewer can get organization", role: testutil.RoleViewer, shouldAllow: true, useConnect: true},
		{resource: "organization", operation: "list_users", name: "owner can list users", role: testutil.RoleOwner, shouldAllow: true, useConnect: true},
		{resource: "organization", operation: "list_users", name: "admin can list users", role: testutil.RoleAdmin, shouldAllow: true, useConnect: true},
		{resource: "organization", operation: "list_users", name: "viewer can list users", role: testutil.RoleViewer, shouldAllow: true, useConnect: true},
	}

	// TODO: Fix permission bug - viewer should not be able to delete controls
	// {resource: "control", operation: "delete", name: "viewer cannot delete control", role: testutil.RoleViewer, shouldAllow: false},

	for _, tt := range tests {
		t.Run(
			tt.resource+"/"+tt.operation+"/"+tt.name,
			func(t *testing.T) {
				client := org.Client(t, tt.role)
				vars := rbacVariables(t, tt.resource, tt.operation, tt.role, org, shared)

				t.Parallel()

				err := executeRBACRequest(
					client,
					rbacQuery(tt.resource, tt.operation),
					vars,
					tt.useConnect,
				)
				assertRBACRequestResult(t, err, tt.shouldAllow)
			},
		)
	}
}

func rbacQuery(resource, operation string) string {
	switch resource {
	case "framework":
		switch operation {
		case "create":
			return createFrameworkMutation
		case "update":
			return updateFrameworkMutation
		case "delete":
			return deleteFrameworkMutation
		case "list":
			return listFrameworksQuery
		}
	case "control":
		switch operation {
		case "create":
			return createControlMutation
		case "update":
			return updateControlMutation
		case "delete":
			return deleteControlMutation
		case "list":
			return listControlsQuery
		}
	case "measure":
		switch operation {
		case "create":
			return createMeasureMutation
		case "update":
			return updateMeasureMutation
		case "delete":
			return deleteMeasureMutation
		case "list":
			return listMeasuresQuery
		}
	case "task":
		switch operation {
		case "create":
			return createTaskMutation
		case "update":
			return updateTaskMutation
		case "delete":
			return deleteTaskMutation
		case "list":
			return listTasksQuery
		}
	case "risk":
		switch operation {
		case "create":
			return createRiskMutation
		case "update":
			return updateRiskMutation
		case "delete":
			return deleteRiskMutation
		case "list":
			return listRisksQuery
		}
	case "third_party":
		switch operation {
		case "create":
			return createThirdPartyMutation
		case "update":
			return updateThirdPartyMutation
		case "delete":
			return deleteThirdPartyMutation
		case "list":
			return listThirdPartiesQuery
		}
	case "access_review_source":
		switch operation {
		case "create":
			return createAccessReviewSourceMutation
		case "update":
			return updateAccessReviewSourceMutation
		case "delete":
			return deleteAccessReviewSourceMutation
		case "list":
			return listAccessReviewSourcesQuery
		}
	case "access_review_campaign":
		switch operation {
		case "create":
			return createAccessReviewCampaignMutation
		case "update":
			return updateAccessReviewCampaignMutation
		case "delete":
			return deleteAccessReviewCampaignMutation
		case "list":
			return listAccessReviewCampaignsQuery
		}
	case "organization":
		switch operation {
		case "update":
			return updateOrganizationMutation
		case "get":
			return getOrganizationQuery
		case "list_users":
			return listUsersQuery
		}
	}

	panic("unknown rbac query: " + resource + "/" + operation)
}

func rbacVariables(
	t *testing.T,
	resource string,
	operation string,
	role testutil.TestRole,
	org testutil.OrganizationRoles,
	shared rbacShared,
) map[string]any {
	owner := org.Client(t, testutil.RoleOwner)

	switch resource {
	case "framework":
		switch operation {
		case "create":
			return map[string]any{
				"input": map[string]any{
					"organizationId": shared.orgID,
					"name":           factory.SafeName("Framework"),
				},
			}
		case "update":
			frameworkID := factory.NewFramework(owner).
				WithName(factory.SafeName("RBAC Test Framework")).
				Create()

			return map[string]any{
				"input": map[string]any{
					"id":   frameworkID,
					"name": factory.SafeName("Updated Framework"),
				},
			}
		case "delete":
			id := factory.NewFramework(owner).WithName(factory.SafeName("ToDelete")).Create()

			return map[string]any{"input": map[string]any{"frameworkId": id}}
		case "list":
			return map[string]any{"id": shared.orgID}
		}
	case "control":
		switch operation {
		case "create":
			frameworkID := factory.NewFramework(owner).
				WithName(factory.SafeName("RBAC control framework")).
				Create()

			var sectionTitle string

			switch role {
			case testutil.RoleOwner:
				sectionTitle = factory.SafeName("Section Owner")
			case testutil.RoleAdmin:
				sectionTitle = factory.SafeName("Section Admin")
			case testutil.RoleViewer:
				sectionTitle = factory.SafeName("Section Viewer")
			}

			return map[string]any{
				"input": map[string]any{
					"frameworkId":   frameworkID,
					"name":          factory.SafeName("Control"),
					"description":   "Test",
					"sectionTitle":  sectionTitle,
					"bestPractice":  true,
					"maturityLevel": "INITIAL",
				},
			}
		case "update":
			frameworkID := factory.NewFramework(owner).
				WithName(factory.SafeName("RBAC control framework")).
				Create()
			controlID := factory.NewControl(owner, frameworkID).
				WithName(factory.SafeName("RBAC Test Control")).
				Create()

			return map[string]any{
				"input": map[string]any{
					"id":   controlID,
					"name": factory.SafeName("Updated Control"),
				},
			}
		case "delete":
			frameworkID := factory.NewFramework(owner).
				WithName(factory.SafeName("RBAC control framework")).
				Create()
			id := factory.NewControl(owner, frameworkID).WithName(factory.SafeName("ToDelete")).Create()

			return map[string]any{"input": map[string]any{"controlId": id}}
		case "list":
			return map[string]any{"id": shared.frameworkID}
		}
	case "measure":
		switch operation {
		case "create":
			return map[string]any{
				"input": map[string]any{
					"organizationId": shared.orgID,
					"name":           factory.SafeName("Measure"),
					"category":       "POLICY",
				},
			}
		case "update":
			measureID := factory.NewMeasure(owner).
				WithName(factory.SafeName("RBAC Test Measure")).
				Create()

			return map[string]any{
				"input": map[string]any{
					"id":   measureID,
					"name": factory.SafeName("Updated Measure"),
				},
			}
		case "delete":
			id := factory.NewMeasure(owner).WithName(factory.SafeName("ToDelete")).Create()

			return map[string]any{"input": map[string]any{"measureId": id}}
		case "list":
			return map[string]any{"id": shared.orgID}
		}
	case "task":
		switch operation {
		case "create":
			measureID := factory.NewMeasure(owner).
				WithName(factory.SafeName("RBAC task measure")).
				Create()

			return map[string]any{
				"input": map[string]any{
					"organizationId": shared.orgID,
					"measureId":      measureID,
					"name":           factory.SafeName("Task"),
					"priority":       "MEDIUM",
				},
			}
		case "update":
			measureID := factory.NewMeasure(owner).
				WithName(factory.SafeName("RBAC task measure")).
				Create()
			taskID := factory.NewTask(owner, measureID).
				WithName(factory.SafeName("RBAC Test Task")).
				Create()

			return map[string]any{
				"input": map[string]any{
					"taskId": taskID,
					"name":   factory.SafeName("Updated Task"),
				},
			}
		case "delete":
			measureID := factory.NewMeasure(owner).
				WithName(factory.SafeName("RBAC task measure")).
				Create()
			id := factory.NewTask(owner, measureID).WithName(factory.SafeName("ToDelete")).Create()

			return map[string]any{"input": map[string]any{"taskId": id}}
		case "list":
			return map[string]any{"id": shared.measureID}
		}
	case "risk":
		switch operation {
		case "create":
			return map[string]any{
				"input": map[string]any{
					"organizationId":     shared.orgID,
					"name":               factory.SafeName("Risk"),
					"category":           "SECURITY",
					"treatment":          "MITIGATED",
					"inherentLikelihood": 2,
					"inherentImpact":     2,
				},
			}
		case "update":
			riskID := factory.NewRisk(owner).
				WithName(factory.SafeName("RBAC Test Risk")).
				Create()

			return map[string]any{
				"input": map[string]any{
					"id":   riskID,
					"name": factory.SafeName("Updated Risk"),
				},
			}
		case "delete":
			id := factory.NewRisk(owner).WithName(factory.SafeName("ToDelete")).Create()

			return map[string]any{"input": map[string]any{"riskId": id}}
		case "list":
			return map[string]any{"id": shared.orgID}
		}
	case "third_party":
		switch operation {
		case "create":
			return map[string]any{
				"input": map[string]any{
					"organizationId": shared.orgID,
					"name":           factory.SafeName("ThirdParty"),
				},
			}
		case "update":
			thirdPartyID := factory.NewThirdParty(owner).
				WithName(factory.SafeName("RBAC Test ThirdParty")).
				Create()

			return map[string]any{
				"input": map[string]any{
					"id":   thirdPartyID,
					"name": factory.SafeName("Updated ThirdParty"),
				},
			}
		case "delete":
			id := factory.NewThirdParty(owner).WithName(factory.SafeName("ToDelete")).Create()

			return map[string]any{"input": map[string]any{"thirdPartyId": id}}
		case "list":
			return map[string]any{"id": shared.orgID}
		}
	case "access_review_source":
		switch operation {
		case "create":
			return map[string]any{
				"input": map[string]any{
					"organizationId": shared.orgID,
					"name":           factory.SafeName("AccessReviewSource"),
				},
			}
		case "update":
			accessReviewSourceID := factory.NewAccessReviewSource(owner, shared.orgID).
				WithName(factory.SafeName("RBAC Test Source")).
				Create()

			return map[string]any{
				"input": map[string]any{
					"accessReviewSourceId": accessReviewSourceID,
					"name":                 factory.SafeName("Updated Source"),
				},
			}
		case "delete":
			id := factory.NewAccessReviewSource(owner, shared.orgID).WithName(factory.SafeName("ToDelete")).Create()

			return map[string]any{"input": map[string]any{"accessReviewSourceId": id}}
		case "list":
			return map[string]any{"id": shared.orgID}
		}
	case "access_review_campaign":
		switch operation {
		case "create":
			return map[string]any{
				"input": map[string]any{
					"organizationId": shared.orgID,
					"name":           factory.SafeName("Campaign"),
				},
			}
		case "update":
			accessReviewCampaignID := factory.NewAccessReviewCampaign(owner, shared.orgID).
				WithName(factory.SafeName("RBAC Test Campaign")).
				Create()

			return map[string]any{
				"input": map[string]any{
					"accessReviewCampaignId": accessReviewCampaignID,
					"name":                   factory.SafeName("Updated Campaign"),
				},
			}
		case "delete":
			id := factory.NewAccessReviewCampaign(owner, shared.orgID).WithName(factory.SafeName("ToDelete")).Create()

			return map[string]any{"input": map[string]any{"accessReviewCampaignId": id}}
		case "list":
			return map[string]any{"id": shared.orgID}
		}
	case "organization":
		switch operation {
		case "update":
			return map[string]any{
				"input": map[string]any{
					"organizationId": shared.orgID,
					"name":           factory.SafeName("Updated Org"),
				},
			}
		case "get", "list_users":
			return map[string]any{"id": shared.orgID}
		}
	}

	t.Fatalf("unknown rbac variables: %s/%s", resource, operation)

	return nil
}

func executeRBACRequest(
	client *testutil.Client,
	query string,
	variables map[string]any,
	useConnect bool,
) error {
	if useConnect {
		_, err := client.DoConnect(query, variables)

		return err
	}

	_, err := client.Do(query, variables)

	return err
}

func assertRBACRequestResult(t *testing.T, err error, shouldAllow bool) {
	t.Helper()

	if shouldAllow {
		require.NoError(t, err, "expected request to be allowed")

		return
	}

	var gqlErrors testutil.GraphQLErrors

	require.ErrorAs(t, err, &gqlErrors, "expected GraphQL error, got: %T", err)
	require.Len(t, gqlErrors, 1, "expected exactly one GraphQL error, got %d errors: %v", len(gqlErrors), gqlErrors)
	// Connect API uses a different error format - check either code or message
	code := gqlErrors[0].Code()
	msg := gqlErrors[0].Message
	isForbidden := code == "FORBIDDEN" || (code == "" && (strings.Contains(msg, "does not have sufficient permissions") || strings.Contains(msg, "insufficient permissions")))
	require.True(t, isForbidden, "expected FORBIDDEN error, got code=%q message=%q", code, msg)
}
