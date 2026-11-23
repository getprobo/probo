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
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/testutil"
)

const thirdPartyAgreementTestPDF = "%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\ntrailer\n<< /Root 1 0 R >>\n%%EOF"

const (
	actionThirdPartyAgreementUpdate    = "core:thirdParty-agreement:update"
	actionThirdPartyAgreementDelete    = "core:thirdParty-agreement:delete"
	actionThirdPartyRiskAssessmentList = "core:thirdParty-risk-assessment:list"

	thirdPartyRiskAssessmentExpiresAt2030 = "2030-12-31T00:00:00Z"

	thirdPartyAgreementNodeSelection = `
		id
		title
		type
		validity {
			start
			end
		}
		createdAt
		updatedAt
		thirdParty { id }
		file { id fileName }
		canUpdate: permission(action: "UPDATE")
		canDelete: permission(action: "DELETE")
	`
)

type (
	thirdPartyAgreementWireNode struct {
		ID         string      `json:"id"`
		Title      string      `json:"title"`
		Type       string      `json:"type"`
		Validity   *periodWire `json:"validity"`
		CreatedAt  time.Time   `json:"createdAt"`
		UpdatedAt  time.Time   `json:"updatedAt"`
		ThirdParty struct {
			ID string `json:"id"`
		} `json:"thirdParty"`
		File struct {
			ID       string `json:"id"`
			FileName string `json:"fileName"`
		} `json:"file"`
		CanUpdate bool `json:"canUpdate"`
		CanDelete bool `json:"canDelete"`
	}

	periodWire struct {
		Start *time.Time `json:"start"`
		End   *time.Time `json:"end"`
	}

	thirdPartyRiskAssessmentWireNode struct {
		ID              string    `json:"id"`
		ExpiresAt       time.Time `json:"expiresAt"`
		DataSensitivity string    `json:"dataSensitivity"`
		BusinessImpact  string    `json:"businessImpact"`
		Notes           *string   `json:"notes"`
		CreatedAt       time.Time `json:"createdAt"`
		UpdatedAt       time.Time `json:"updatedAt"`
		ThirdParty      struct {
			ID string `json:"id"`
		} `json:"thirdParty"`
		CanList bool `json:"canList"`
	}
)

func requireParseRFC3339Time(t *testing.T, value string) time.Time {
	t.Helper()

	parsed, err := time.Parse(time.RFC3339, value)
	require.NoError(t, err)

	return parsed
}

func (n thirdPartyAgreementWireNode) ValidFrom() *time.Time {
	if n.Validity == nil {
		return nil
	}

	return n.Validity.Start
}

func (n thirdPartyAgreementWireNode) ValidUntil() *time.Time {
	if n.Validity == nil {
		return nil
	}

	return n.Validity.End
}

func thirdPartyAgreementUploadPDF(fileName string) testutil.UploadFile {
	return testutil.UploadFile{
		Filename:    fileName,
		ContentType: "application/pdf",
		Content:     []byte(thirdPartyAgreementTestPDF),
	}
}

func thirdPartyAgreementSelection(selection string) string {
	query := strings.ReplaceAll(selection, "UPDATE", actionThirdPartyAgreementUpdate)

	return strings.ReplaceAll(query, "DELETE", actionThirdPartyAgreementDelete)
}

func uploadThirdPartyAgreement(
	t *testing.T,
	client *testutil.Client,
	thirdPartyID string,
	agreementType string,
	validFrom string,
	validUntil string,
	fileName string,
) thirdPartyAgreementWireNode {
	t.Helper()

	node, err := uploadThirdPartyAgreementResult(
		t,
		client,
		thirdPartyID,
		agreementType,
		validFrom,
		validUntil,
		thirdPartyAgreementUploadPDF(fileName),
		thirdPartyAgreementNodeSelection,
	)
	require.NoError(t, err)

	return node
}

func uploadThirdPartyAgreementExpectError(
	t *testing.T,
	client *testutil.Client,
	thirdPartyID string,
	agreementType string,
	validFrom string,
	validUntil string,
	fileName string,
) error {
	t.Helper()

	_, err := uploadThirdPartyAgreementResult(
		t,
		client,
		thirdPartyID,
		agreementType,
		validFrom,
		validUntil,
		thirdPartyAgreementUploadPDF(fileName),
		"id",
	)

	return err
}

func uploadThirdPartyAgreementFileExpectError(
	t *testing.T,
	client *testutil.Client,
	thirdPartyID string,
	file testutil.UploadFile,
) error {
	t.Helper()

	_, err := uploadThirdPartyAgreementResult(
		t,
		client,
		thirdPartyID,
		"OTHER",
		"2024-01-01T00:00:00Z",
		"2025-01-01T00:00:00Z",
		file,
		"id",
	)

	return err
}

func uploadThirdPartyAgreementResult(
	t *testing.T,
	client *testutil.Client,
	thirdPartyID string,
	agreementType string,
	validFrom string,
	validUntil string,
	file testutil.UploadFile,
	nodeSelection string,
) (thirdPartyAgreementWireNode, error) {
	t.Helper()

	mutation := `
		mutation UploadThirdPartyAgreement($input: UploadThirdPartyAgreementInput!) {
			uploadThirdPartyAgreement(input: $input) {
				thirdPartyAgreementEdge {
					node {
						SELECTION
					}
				}
			}
		}
	`
	mutation = strings.ReplaceAll(mutation, "SELECTION", thirdPartyAgreementSelection(nodeSelection))

	var result struct {
		UploadThirdPartyAgreement struct {
			ThirdPartyAgreementEdge struct {
				Node thirdPartyAgreementWireNode `json:"node"`
			} `json:"thirdPartyAgreementEdge"`
		} `json:"uploadThirdPartyAgreement"`
	}

	err := client.ExecuteWithFile(
		mutation,
		map[string]any{
			"input": map[string]any{
				"thirdPartyId": thirdPartyID,
				"title":        strings.TrimSuffix(file.Filename, ".pdf"),
				"type":         agreementType,
				"validity": map[string]any{
					"start": validFrom,
					"end":   validUntil,
				},
				"file": nil,
			},
		},
		"input.file",
		file,
		&result,
	)
	if err != nil {
		return thirdPartyAgreementWireNode{}, err
	}

	return result.UploadThirdPartyAgreement.ThirdPartyAgreementEdge.Node, nil
}

func listThirdPartyAgreements(
	t *testing.T,
	client *testutil.Client,
	thirdPartyID string,
) []thirdPartyAgreementWireNode {
	t.Helper()

	query := `
		query($id: ID!) {
			node(id: $id) {
				... on ThirdParty {
					agreements(first: 20) {
						edges {
							node {
								SELECTION
							}
						}
					}
				}
			}
		}
	`
	query = strings.ReplaceAll(query, "SELECTION", thirdPartyAgreementSelection(thirdPartyAgreementNodeSelection))

	var result struct {
		Node struct {
			Agreements struct {
				Edges []struct {
					Node thirdPartyAgreementWireNode `json:"node"`
				} `json:"edges"`
			} `json:"agreements"`
		} `json:"node"`
	}

	err := client.Execute(query, map[string]any{"id": thirdPartyID}, &result)
	require.NoError(t, err)

	agreements := make([]thirdPartyAgreementWireNode, 0, len(result.Node.Agreements.Edges))
	for _, edge := range result.Node.Agreements.Edges {
		agreements = append(agreements, edge.Node)
	}

	return agreements
}

func updateThirdPartyAgreement(
	t *testing.T,
	client *testutil.Client,
	agreementID string,
	agreementType *string,
	validFrom *string,
	validUntil *string,
) thirdPartyAgreementWireNode {
	t.Helper()

	node, err := updateThirdPartyAgreementResult(
		t,
		client,
		agreementID,
		agreementType,
		validFrom,
		validUntil,
		thirdPartyAgreementNodeSelection,
	)
	require.NoError(t, err)

	return node
}

func updateThirdPartyAgreementExpectError(
	t *testing.T,
	client *testutil.Client,
	agreementID string,
	agreementType *string,
	validFrom *string,
	validUntil *string,
) error {
	t.Helper()

	_, err := updateThirdPartyAgreementResult(
		t,
		client,
		agreementID,
		agreementType,
		validFrom,
		validUntil,
		"id",
	)

	return err
}

func updateThirdPartyAgreementResult(
	t *testing.T,
	client *testutil.Client,
	agreementID string,
	agreementType *string,
	validFrom *string,
	validUntil *string,
	nodeSelection string,
) (thirdPartyAgreementWireNode, error) {
	t.Helper()

	mutation := `
		mutation UpdateThirdPartyAgreement($input: UpdateThirdPartyAgreementInput!) {
			updateThirdPartyAgreement(input: $input) {
				thirdPartyAgreement {
					SELECTION
				}
			}
		}
	`
	mutation = strings.ReplaceAll(mutation, "SELECTION", thirdPartyAgreementSelection(nodeSelection))

	input := map[string]any{
		"id": agreementID,
	}
	if agreementType != nil {
		input["type"] = *agreementType
	}

	period := map[string]any{}
	if validFrom != nil {
		period["start"] = *validFrom
	}

	if validUntil != nil {
		period["end"] = *validUntil
	}

	if len(period) > 0 {
		input["validity"] = period
	}

	var result struct {
		UpdateThirdPartyAgreement struct {
			ThirdPartyAgreement thirdPartyAgreementWireNode `json:"thirdPartyAgreement"`
		} `json:"updateThirdPartyAgreement"`
	}

	err := client.Execute(mutation, map[string]any{"input": input}, &result)
	if err != nil {
		return thirdPartyAgreementWireNode{}, err
	}

	return result.UpdateThirdPartyAgreement.ThirdPartyAgreement, nil
}

func deleteThirdPartyAgreement(
	t *testing.T,
	client *testutil.Client,
	agreementID string,
) string {
	t.Helper()

	const mutation = `
		mutation DeleteThirdPartyAgreement($input: DeleteThirdPartyAgreementInput!) {
			deleteThirdPartyAgreement(input: $input) {
				deletedThirdPartyAgreementId
			}
		}
	`

	var result struct {
		DeleteThirdPartyAgreement struct {
			DeletedThirdPartyAgreementID string `json:"deletedThirdPartyAgreementId"`
		} `json:"deleteThirdPartyAgreement"`
	}

	err := client.Execute(
		mutation,
		map[string]any{
			"input": map[string]any{
				"agreementId": agreementID,
			},
		},
		&result,
	)
	require.NoError(t, err)

	return result.DeleteThirdPartyAgreement.DeletedThirdPartyAgreementID
}

func createThirdPartyRiskAssessment(
	t *testing.T,
	client *testutil.Client,
	thirdPartyID string,
	expiresAt string,
	dataSensitivity string,
	businessImpact string,
	notes string,
) thirdPartyRiskAssessmentWireNode {
	t.Helper()

	const mutation = `
		mutation CreateThirdPartyRiskAssessment($input: CreateThirdPartyRiskAssessmentInput!) {
			createThirdPartyRiskAssessment(input: $input) {
				thirdPartyRiskAssessmentEdge {
					node {
						id
						expiresAt
						dataSensitivity
						businessImpact
						notes
						createdAt
						updatedAt
						thirdParty { id }
						canList: permission(action: "LIST_ACTION")
					}
				}
			}
		}
	`

	input := map[string]any{
		"thirdPartyId":    thirdPartyID,
		"expiresAt":       expiresAt,
		"dataSensitivity": dataSensitivity,
		"businessImpact":  businessImpact,
	}
	if notes != "" {
		input["notes"] = notes
	}

	var result struct {
		CreateThirdPartyRiskAssessment struct {
			ThirdPartyRiskAssessmentEdge struct {
				Node thirdPartyRiskAssessmentWireNode `json:"node"`
			} `json:"thirdPartyRiskAssessmentEdge"`
		} `json:"createThirdPartyRiskAssessment"`
	}

	mutationQuery := strings.ReplaceAll(mutation, "LIST_ACTION", actionThirdPartyRiskAssessmentList)

	err := client.Execute(mutationQuery, map[string]any{"input": input}, &result)
	require.NoError(t, err)

	return result.CreateThirdPartyRiskAssessment.ThirdPartyRiskAssessmentEdge.Node
}

func updateThirdPartyAgreementTitle(
	client *testutil.Client,
	agreementID string,
	title string,
) (thirdPartyAgreementWireNode, error) {
	const mutation = `
		mutation UpdateThirdPartyAgreement($input: UpdateThirdPartyAgreementInput!) {
			updateThirdPartyAgreement(input: $input) {
				thirdPartyAgreement {
					id
					title
					type
				}
			}
		}
	`

	var result struct {
		UpdateThirdPartyAgreement struct {
			ThirdPartyAgreement thirdPartyAgreementWireNode `json:"thirdPartyAgreement"`
		} `json:"updateThirdPartyAgreement"`
	}

	err := client.Execute(
		mutation,
		map[string]any{"input": map[string]any{"id": agreementID, "title": title}},
		&result,
	)
	if err != nil {
		return thirdPartyAgreementWireNode{}, err
	}

	return result.UpdateThirdPartyAgreement.ThirdPartyAgreement, nil
}
