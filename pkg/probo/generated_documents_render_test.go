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

package probo_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/docgen"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/prosemirror"
)

// hostileText exercises the template escaping every generated document relies
// on: a double quote, a backslash, a control character and an HTML tag all
// break the emitted JSON if a field is interpolated unescaped.
const hostileText = "quote \" backslash \\ newline \n tag <script> ampersand &"

func TestBuildDocuments_ProduceParseableProseMirror(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		build func() (string, error)
	}{
		{
			name: "data list",
			build: func() (string, error) {
				return probo.BuildDataListDocument(
					docgen.DataListData{
						Title:            hostileText,
						OrganizationName: hostileText,
						TotalData:        1,
						Rows: []docgen.DataListRow{
							{
								Name:           hostileText,
								Classification: hostileText,
								Owner:          hostileText,
								ThirdParties:   hostileText,
							},
						},
					},
				)
			},
		},
		{
			name: "asset list",
			build: func() (string, error) {
				return probo.BuildAssetListDocument(
					docgen.AssetListData{
						Title:            hostileText,
						OrganizationName: hostileText,
						TotalAssets:      1,
						Rows: []docgen.AssetListRow{
							{
								Name:            hostileText,
								AssetType:       hostileText,
								Amount:          1,
								DataTypesStored: hostileText,
								Owner:           hostileText,
								ThirdParties:    hostileText,
							},
						},
					},
				)
			},
		},
		{
			name: "statement of applicability",
			build: func() (string, error) {
				return probo.BuildStatementOfApplicabilityDocument(
					docgen.StatementOfApplicabilityData{
						Title:            hostileText,
						OrganizationName: hostileText,
						TotalControls:    1,
						Rows: []docgen.SOARow{
							{
								FrameworkName:  hostileText,
								ControlSection: hostileText,
								ControlName:    hostileText,
								Applicability:  hostileText,
								Justification:  hostileText,
								MaturityLevel:  hostileText,
							},
						},
					},
				)
			},
		},
		{
			name: "finding list",
			build: func() (string, error) {
				return probo.BuildFindingListDocument(
					docgen.FindingListData{
						Title:            hostileText,
						OrganizationName: hostileText,
						TotalFindings:    1,
						Rows: []docgen.FindingListRow{
							{
								ReferenceID:        hostileText,
								Kind:               hostileText,
								Description:        hostileText,
								Source:             hostileText,
								IdentifiedOn:       hostileText,
								RootCause:          hostileText,
								CorrectiveAction:   hostileText,
								EffectivenessCheck: hostileText,
								Status:             hostileText,
								Priority:           hostileText,
								Owner:              hostileText,
								DueDate:            hostileText,
							},
						},
					},
				)
			},
		},
		{
			name: "business function list",
			build: func() (string, error) {
				return probo.BuildBusinessFunctionListDocument(
					docgen.BusinessFunctionListData{
						Title:                  hostileText,
						OrganizationName:       hostileText,
						TotalBusinessFunctions: 1,
						Rows: []docgen.BusinessFunctionListRow{
							{
								ReferenceID:    hostileText,
								Name:           hostileText,
								Classification: hostileText,
								Notes:          hostileText,
								Owner:          hostileText,
							},
						},
					},
				)
			},
		},
		{
			name: "ai system list",
			build: func() (string, error) {
				return probo.BuildAiSystemListDocument(
					docgen.AiSystemListData{
						Title:            hostileText,
						OrganizationName: hostileText,
						TotalAiSystems:   1,
						Rows: []docgen.AiSystemListRow{
							{
								Name:               hostileText,
								Version:            hostileText,
								CompanyRoles:       hostileText,
								Status:             hostileText,
								Owner:              hostileText,
								Purpose:            hostileText,
								RiskClassification: hostileText,
								Notes:              hostileText,
							},
						},
					},
				)
			},
		},
		{
			name: "obligation list",
			build: func() (string, error) {
				return probo.BuildObligationListDocument(
					docgen.ObligationListData{
						Title:            hostileText,
						OrganizationName: hostileText,
						TotalObligations: 1,
						Rows: []docgen.ObligationListRow{
							{
								Area:                   hostileText,
								Source:                 hostileText,
								Requirement:            hostileText,
								ActionsToBeImplemented: hostileText,
								Status:                 hostileText,
								Type:                   hostileText,
								Regulator:              hostileText,
								Owner:                  hostileText,
								DueDate:                hostileText,
							},
						},
					},
				)
			},
		},
		{
			name: "processing activity list",
			build: func() (string, error) {
				return probo.BuildProcessingActivityListDocument(
					docgen.ProcessingActivityListData{
						Title:                     hostileText,
						OrganizationName:          hostileText,
						TotalProcessingActivities: 1,
						Rows: []docgen.ProcessingActivityListRow{
							{
								Name:                 hostileText,
								Purpose:              hostileText,
								Role:                 hostileText,
								DataSubjectCategory:  hostileText,
								PersonalDataCategory: hostileText,
								LawfulBasis:          hostileText,
								Recipients:           hostileText,
								Location:             hostileText,
								RetentionPeriod:      hostileText,
								SecurityMeasures:     hostileText,
							},
						},
					},
				)
			},
		},
		{
			name: "data protection impact assessment list",
			build: func() (string, error) {
				return probo.BuildDataProtectionImpactAssessmentListDocument(
					docgen.DataProtectionImpactAssessmentListData{
						Title:                                hostileText,
						OrganizationName:                     hostileText,
						TotalDataProtectionImpactAssessments: 1,
						Rows: []docgen.DataProtectionImpactAssessmentListRow{
							{
								ProcessingActivityName:      hostileText,
								Description:                 hostileText,
								NecessityAndProportionality: hostileText,
								PotentialRisk:               hostileText,
								Mitigations:                 hostileText,
								ResidualRisk:                hostileText,
							},
						},
					},
				)
			},
		},
		{
			name: "transfer impact assessment list",
			build: func() (string, error) {
				return probo.BuildTransferImpactAssessmentListDocument(
					docgen.TransferImpactAssessmentListData{
						Title:                          hostileText,
						OrganizationName:               hostileText,
						TotalTransferImpactAssessments: 1,
						Rows: []docgen.TransferImpactAssessmentListRow{
							{
								ProcessingActivityName: hostileText,
								DataSubjects:           hostileText,
								Transfer:               hostileText,
								LegalMechanism:         hostileText,
								LocalLawRisk:           hostileText,
								SupplementaryMeasures:  hostileText,
							},
						},
					},
				)
			},
		},
		{
			name: "third party list",
			build: func() (string, error) {
				return probo.BuildThirdPartyListDocument(
					docgen.ThirdPartyListData{
						Title:             hostileText,
						OrganizationName:  hostileText,
						TotalThirdParties: 1,
						Rows: []docgen.ThirdPartyListRow{
							{
								Name:        hostileText,
								LegalName:   hostileText,
								Description: hostileText,
								Category:    hostileText,
								Services: []docgen.ThirdPartyListService{
									{Name: hostileText, Description: hostileText},
								},
								Contacts: []docgen.ThirdPartyListContact{
									{FullName: hostileText, Email: hostileText, Phone: hostileText, Role: hostileText},
								},
								RiskAnalyses: []docgen.ThirdPartyListRiskAssessment{
									{
										AssessedAt:      hostileText,
										ExpiresAt:       hostileText,
										DataSensitivity: hostileText,
										BusinessImpact:  hostileText,
										Notes:           "# Heading\n\nBody with a \"quote\" and <tag>.\n",
									},
								},
								ComplianceReports: []docgen.ThirdPartyListComplianceReport{
									{ReportName: hostileText, ReportDate: hostileText, ValidUntil: hostileText},
								},
							},
						},
					},
				)
			},
		},
		{
			name: "risk list",
			build: func() (string, error) {
				return probo.BuildRiskListDocument(
					docgen.RiskListData{
						Title:            hostileText,
						OrganizationName: hostileText,
						TotalRisks:       1,
						Rows: []docgen.RiskListRow{
							{
								Name:               hostileText,
								Description:        hostileText,
								Category:           hostileText,
								Treatment:          hostileText,
								Owner:              hostileText,
								InherentLikelihood: hostileText,
								InherentImpact:     hostileText,
								InherentRiskScore:  hostileText,
								ResidualLikelihood: hostileText,
								ResidualImpact:     hostileText,
								ResidualRiskScore:  hostileText,
								Note:               hostileText,
							},
						},
					},
				)
			},
		},
		{
			name: "tracker policy",
			build: func() (string, error) {
				return probo.BuildTrackerPolicyDocument(
					docgen.TrackerPolicyData{
						OrganizationName:  hostileText,
						WebsiteOrigin:     hostileText,
						PrivacyPolicyURL:  hostileText,
						ConsentExpiryDays: 180,
						Categories: []docgen.TrackerPolicyCategory{
							{
								Name:        hostileText,
								Description: hostileText,
								Necessary:   true,
								Trackers: []docgen.TrackerPolicyTracker{
									{Name: hostileText, Type: hostileText, Purpose: hostileText, Duration: hostileText},
								},
							},
						},
						ThirdParties: []docgen.TrackerPolicyThirdParty{
							{Name: hostileText, Description: hostileText, PrivacyPolicyURL: hostileText},
						},
					},
				)
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				content, err := tt.build()
				require.NoError(t, err)

				node, err := prosemirror.Parse(content)
				require.NoError(t, err)
				require.Equal(t, prosemirror.NodeDoc, node.Type)

				require.NoError(t, prosemirror.ValidateDocumentContentJSON(content))
			},
		)
	}
}

func TestBuildDocuments_EmptyDataProducesParseableProseMirror(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		build func() (string, error)
	}{
		{"data list", func() (string, error) {
			return probo.BuildDataListDocument(docgen.DataListData{})
		}},
		{"asset list", func() (string, error) {
			return probo.BuildAssetListDocument(docgen.AssetListData{})
		}},
		{"statement of applicability", func() (string, error) {
			return probo.BuildStatementOfApplicabilityDocument(docgen.StatementOfApplicabilityData{})
		}},
		{"finding list", func() (string, error) {
			return probo.BuildFindingListDocument(docgen.FindingListData{})
		}},
		{"business function list", func() (string, error) {
			return probo.BuildBusinessFunctionListDocument(docgen.BusinessFunctionListData{})
		}},
		{"ai system list", func() (string, error) {
			return probo.BuildAiSystemListDocument(docgen.AiSystemListData{})
		}},
		{"obligation list", func() (string, error) {
			return probo.BuildObligationListDocument(docgen.ObligationListData{})
		}},
		{"processing activity list", func() (string, error) {
			return probo.BuildProcessingActivityListDocument(docgen.ProcessingActivityListData{})
		}},
		{"data protection impact assessment list", func() (string, error) {
			return probo.BuildDataProtectionImpactAssessmentListDocument(docgen.DataProtectionImpactAssessmentListData{})
		}},
		{"transfer impact assessment list", func() (string, error) {
			return probo.BuildTransferImpactAssessmentListDocument(docgen.TransferImpactAssessmentListData{})
		}},
		{"third party list", func() (string, error) {
			return probo.BuildThirdPartyListDocument(docgen.ThirdPartyListData{})
		}},
		{"risk list", func() (string, error) {
			return probo.BuildRiskListDocument(docgen.RiskListData{})
		}},
		{"tracker policy", func() (string, error) {
			return probo.BuildTrackerPolicyDocument(docgen.TrackerPolicyData{})
		}},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				content, err := tt.build()
				require.NoError(t, err)

				node, err := prosemirror.Parse(content)
				require.NoError(t, err)
				require.Equal(t, prosemirror.NodeDoc, node.Type)

				require.NoError(t, prosemirror.ValidateDocumentContentJSON(content))
			},
		)
	}
}
