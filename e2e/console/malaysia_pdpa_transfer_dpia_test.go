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
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

const (
	updateMalaysiaPDPADPIAScreeningMutation = `
		mutation UpdateMalaysiaPDPADPIAScreening($input: UpdateProcessingActivityInput!) {
			updateProcessingActivity(input: $input) {
				processingActivity {
					id
					malaysiaPDPADPIARecommendation
					malaysiaPDPADPIAReasons
					malaysiaPDPADPIAAssessedByProfileId
					malaysiaPDPADPIAAssessedAt
					malaysiaPDPADPIARuleVersion
					malaysiaPDPADPIARuleSource
				}
			}
		}
	`

	createMalaysiaPDPATransferMutation = `
		mutation CreateMalaysiaPDPATransfer($input: CreateTransferImpactAssessmentInput!) {
			createTransferImpactAssessment(input: $input) {
				transferImpactAssessment {
					id
					malaysiaTransferBasis
					malaysiaDestinationCountry
					malaysiaRecipientThirdPartyId
					malaysiaApprovalStatus
					malaysiaApprovedByProfileId
					malaysiaReviewedAt
					malaysiaNextReviewAt
					malaysiaReviewEvidence
					malaysiaRuleVersion
					malaysiaRuleSource
				}
			}
		}
	`
)

type malaysiaPDPADPIAScreeningResult struct {
	ID                  string     `json:"id"`
	Recommendation      string     `json:"malaysiaPDPADPIARecommendation"`
	Reasons             []string   `json:"malaysiaPDPADPIAReasons"`
	AssessedByProfileID *string    `json:"malaysiaPDPADPIAAssessedByProfileId"`
	AssessedAt          *time.Time `json:"malaysiaPDPADPIAAssessedAt"`
	RuleVersion         *string    `json:"malaysiaPDPADPIARuleVersion"`
	RuleSource          *string    `json:"malaysiaPDPADPIARuleSource"`
}

type malaysiaPDPATransferResult struct {
	ID                    string     `json:"id"`
	Basis                 *string    `json:"malaysiaTransferBasis"`
	DestinationCountry    *string    `json:"malaysiaDestinationCountry"`
	RecipientThirdPartyID *string    `json:"malaysiaRecipientThirdPartyId"`
	ApprovalStatus        *string    `json:"malaysiaApprovalStatus"`
	ApprovedByProfileID   *string    `json:"malaysiaApprovedByProfileId"`
	ReviewedAt            *time.Time `json:"malaysiaReviewedAt"`
	NextReviewAt          *time.Time `json:"malaysiaNextReviewAt"`
	ReviewEvidence        *string    `json:"malaysiaReviewEvidence"`
	RuleVersion           *string    `json:"malaysiaRuleVersion"`
	RuleSource            *string    `json:"malaysiaRuleSource"`
}

func TestMalaysiaPDPADPIAScreening_QuantitativeAndQualitativeCriteria(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	processingActivityID := factory.NewProcessingActivity(owner).WithName("Malaysia DPIA screening").Create()

	quantitative := updateMalaysiaPDPADPIAScreening(t, owner, processingActivityID, map[string]any{
		"totalDataSubjects":                20_001,
		"sensitiveDataSubjects":            10_001,
		"legalOrSignificantEffects":        false,
		"systematicMonitoring":             false,
		"innovativeTechnology":             false,
		"denialOrRestrictionOfRights":      false,
		"locationOrBehaviourTracking":      false,
		"childrenOrVulnerableDataSubjects": false,
		"highRiskAutomatedDecisionMaking":  false,
	})
	assert.Equal(t, "REQUIRED", quantitative.Recommendation)
	assert.ElementsMatch(t, []string{"PERSONAL_DATA_VOLUME", "SENSITIVE_OR_FINANCIAL_DATA_VOLUME"}, quantitative.Reasons)
	require.NotNil(t, quantitative.AssessedByProfileID)
	assert.Equal(t, owner.GetProfileID().String(), *quantitative.AssessedByProfileID)
	require.NotNil(t, quantitative.AssessedAt)
	require.NotNil(t, quantitative.RuleVersion)
	assert.Equal(t, "MY-PDPA-DPIA-2026-04", *quantitative.RuleVersion)
	require.NotNil(t, quantitative.RuleSource)
	assert.Equal(t, "https://www.pdp.gov.my/ppdpv1/wp-content/uploads/2026/04/Data-Protection-Impact-Assessment-Guideline-DPIA.pdf", *quantitative.RuleSource)

	qualitative := updateMalaysiaPDPADPIAScreening(t, owner, processingActivityID, map[string]any{
		"totalDataSubjects":                20_000,
		"sensitiveDataSubjects":            10_000,
		"legalOrSignificantEffects":        false,
		"systematicMonitoring":             true,
		"innovativeTechnology":             false,
		"denialOrRestrictionOfRights":      false,
		"locationOrBehaviourTracking":      false,
		"childrenOrVulnerableDataSubjects": false,
		"highRiskAutomatedDecisionMaking":  true,
	})
	assert.Equal(t, "DPO_REVIEW_REQUIRED", qualitative.Recommendation)
	assert.ElementsMatch(t, []string{"SYSTEMATIC_MONITORING", "HIGH_RISK_AUTOMATED_DECISIONS"}, qualitative.Reasons)
}

func TestMalaysiaPDPATransfer_ApprovalAndThreeYearReview(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	processingActivityID := factory.NewProcessingActivity(owner).WithName("Malaysia cross-border transfer").WithInternationalTransfers(true).Create()
	recipientID := factory.NewThirdParty(owner).WithName("Singapore recipient").Create()

	transfer := createMalaysiaPDPATransfer(t, owner, processingActivityID, recipientID, "APPROVED", "TIA and contract record")
	require.NotNil(t, transfer.Basis)
	assert.Equal(t, "SUBSTANTIALLY_SIMILAR_LAW", *transfer.Basis)
	require.NotNil(t, transfer.DestinationCountry)
	assert.Equal(t, "SG", *transfer.DestinationCountry)
	require.NotNil(t, transfer.ApprovalStatus)
	assert.Equal(t, "APPROVED", *transfer.ApprovalStatus)
	require.NotNil(t, transfer.ApprovedByProfileID)
	assert.Equal(t, owner.GetProfileID().String(), *transfer.ApprovedByProfileID)
	require.NotNil(t, transfer.ReviewedAt)
	require.NotNil(t, transfer.NextReviewAt)
	assert.Equal(t, transfer.ReviewedAt.AddDate(3, 0, 0), *transfer.NextReviewAt)
	require.NotNil(t, transfer.RuleVersion)
	assert.Equal(t, "MY-PDPA-CBPDT-2025-04-29", *transfer.RuleVersion)
	require.NotNil(t, transfer.RuleSource)
	assert.Equal(t, "https://www.pdp.gov.my/ppdpv1/wp-content/uploads/2025/08/GP_CBPDT_EN-1.pdf", *transfer.RuleSource)
}

func TestMalaysiaPDPATransfer_ValidationAndRBAC(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
	processingActivityID := factory.NewProcessingActivity(owner).WithName("Malaysia transfer validation").WithInternationalTransfers(true).Create()
	recipientID := factory.NewThirdParty(owner).WithName("Foreign recipient").Create()

	_, err := owner.Do(createMalaysiaPDPATransferMutation, map[string]any{"input": malaysiaPDPATransferInput(processingActivityID, recipientID, "MY", "APPROVED", "evidence")})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "destination_country")

	_, err = owner.Do(createMalaysiaPDPATransferMutation, map[string]any{"input": malaysiaPDPATransferInput(processingActivityID, recipientID, "SG", "APPROVED", "")})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "review_evidence")

	nonTransferActivityID := factory.NewProcessingActivity(owner).WithName("Domestic processing").Create()
	_, err = owner.Do(createMalaysiaPDPATransferMutation, map[string]any{"input": malaysiaPDPATransferInput(nonTransferActivityID, recipientID, "SG", "PENDING", "")})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "international_transfers")

	otherOwner := testutil.NewClient(t, testutil.RoleOwner)
	otherRecipientID := factory.NewThirdParty(otherOwner).WithName("Other tenant recipient").Create()
	_, err = owner.Do(createMalaysiaPDPATransferMutation, map[string]any{"input": malaysiaPDPATransferInput(processingActivityID, otherRecipientID, "SG", "PENDING", "")})
	require.Error(t, err, "should not link a transfer recipient from another tenant")

	_, err = viewer.Do(updateMalaysiaPDPADPIAScreeningMutation, map[string]any{"input": map[string]any{
		"id": processingActivityID,
		"malaysiaPDPADPIAScreening": map[string]any{
			"totalDataSubjects": 0, "sensitiveDataSubjects": 0,
			"legalOrSignificantEffects": false, "systematicMonitoring": false,
			"innovativeTechnology": false, "denialOrRestrictionOfRights": false,
			"locationOrBehaviourTracking": false, "childrenOrVulnerableDataSubjects": false,
			"highRiskAutomatedDecisionMaking": false,
		},
	}})
	testutil.RequireForbiddenError(t, err, "viewer cannot update Malaysia PDPA DPIA screening")
}

func updateMalaysiaPDPADPIAScreening(t *testing.T, client *testutil.Client, processingActivityID string, screening map[string]any) malaysiaPDPADPIAScreeningResult {
	t.Helper()
	var result struct {
		Update struct {
			ProcessingActivity malaysiaPDPADPIAScreeningResult `json:"processingActivity"`
		} `json:"updateProcessingActivity"`
	}
	err := client.Execute(updateMalaysiaPDPADPIAScreeningMutation, map[string]any{"input": map[string]any{
		"id": processingActivityID, "malaysiaPDPADPIAScreening": screening,
	}}, &result)
	require.NoError(t, err)
	return result.Update.ProcessingActivity
}

func createMalaysiaPDPATransfer(t *testing.T, client *testutil.Client, processingActivityID, recipientID, status, evidence string) malaysiaPDPATransferResult {
	t.Helper()
	var result struct {
		Create struct {
			TransferImpactAssessment malaysiaPDPATransferResult `json:"transferImpactAssessment"`
		} `json:"createTransferImpactAssessment"`
	}
	err := client.Execute(createMalaysiaPDPATransferMutation, map[string]any{"input": malaysiaPDPATransferInput(processingActivityID, recipientID, "SG", status, evidence)}, &result)
	require.NoError(t, err)
	return result.Create.TransferImpactAssessment
}

func malaysiaPDPATransferInput(processingActivityID, recipientID, destination, status, evidence string) map[string]any {
	return map[string]any{
		"processingActivityId": processingActivityID,
		"malaysiaPDPA": map[string]any{
			"basis": "SUBSTANTIALLY_SIMILAR_LAW", "destinationCountry": destination,
			"recipientThirdPartyId": recipientID, "receiverRegistrationNumber": "202601234567",
			"receiverContact": "dpo@example.sg", "transferPurpose": "Regional service delivery",
			"personalDataCategories": "Customer identifiers and account data",
			"safeguards":             "TIA, data-processing clauses, encryption, and access controls",
			"approvalStatus":         status, "reviewEvidence": evidence,
		},
	}
}
