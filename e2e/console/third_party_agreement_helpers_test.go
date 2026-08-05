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
	actionThirdPartyBusinessAssociateAgreementUpdate = "core:thirdParty-business-associate-agreement:update"
	actionThirdPartyBusinessAssociateAgreementDelete = "core:thirdParty-business-associate-agreement:delete"
	actionThirdPartyDataPrivacyAgreementUpdate       = "core:thirdParty-data-privacy-agreement:update"
	actionThirdPartyDataPrivacyAgreementDelete       = "core:thirdParty-data-privacy-agreement:delete"
	actionThirdPartyRiskAssessmentList               = "core:thirdParty-risk-assessment:list"

	thirdPartyRiskAssessmentExpiresAt2030 = "2030-12-31T00:00:00Z"
)

type (
	thirdPartyAgreementKind string

	thirdPartyAgreementWireNode struct {
		ID         string     `json:"id"`
		ValidFrom  *time.Time `json:"validFrom"`
		ValidUntil *time.Time `json:"validUntil"`
		CreatedAt  time.Time  `json:"createdAt"`
		UpdatedAt  time.Time  `json:"updatedAt"`
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

const (
	thirdPartyAgreementKindBAA thirdPartyAgreementKind = "baa"
	thirdPartyAgreementKindDPA thirdPartyAgreementKind = "dpa"
)

func requireParseRFC3339Time(t *testing.T, value string) time.Time {
	t.Helper()

	parsed, err := time.Parse(time.RFC3339, value)
	require.NoError(t, err)

	return parsed
}

func thirdPartyAgreementUploadPDF(fileName string) testutil.UploadFile {
	return testutil.UploadFile{
		Filename:    fileName,
		ContentType: "application/pdf",
		Content:     []byte(thirdPartyAgreementTestPDF),
	}
}

func uploadThirdPartyAgreement(
	t *testing.T,
	client *testutil.Client,
	kind thirdPartyAgreementKind,
	thirdPartyID string,
	validFrom string,
	validUntil string,
	fileName string,
) thirdPartyAgreementWireNode {
	t.Helper()

	updateAction, deleteAction := thirdPartyAgreementPermissionActions(kind)

	const nodeSelection = `
		id
		validFrom
		validUntil
		createdAt
		updatedAt
		thirdParty { id }
		file { id fileName }
		canUpdate: permission(action: "UPDATE")
		canDelete: permission(action: "DELETE")
	`

	mutation := thirdPartyAgreementUploadMutation(kind, nodeSelection)
	mutation = strings.ReplaceAll(mutation, "UPDATE", updateAction)
	mutation = strings.ReplaceAll(mutation, "DELETE", deleteAction)

	var result thirdPartyAgreementUploadResult

	err := client.ExecuteWithFile(
		mutation,
		map[string]any{
			"input": map[string]any{
				"thirdPartyId": thirdPartyID,
				"fileName":     fileName,
				"validFrom":    validFrom,
				"validUntil":   validUntil,
				"file":         nil,
			},
		},
		"input.file",
		thirdPartyAgreementUploadPDF(fileName),
		result.ptrForKind(kind),
	)
	require.NoError(t, err)

	return result.node(kind)
}

func queryThirdPartyAgreementField(
	t *testing.T,
	client *testutil.Client,
	kind thirdPartyAgreementKind,
	thirdPartyID string,
) *thirdPartyAgreementWireNode {
	t.Helper()

	fieldName := thirdPartyAgreementThirdPartyField(kind)
	updateAction, deleteAction := thirdPartyAgreementPermissionActions(kind)

	query := `
		query($id: ID!) {
			node(id: $id) {
				... on ThirdParty {
					id
					FIELD {
						id
						validFrom
						validUntil
						createdAt
						updatedAt
						thirdParty { id }
						file { id fileName }
						canUpdate: permission(action: "UPDATE")
						canDelete: permission(action: "DELETE")
					}
				}
			}
		}
	`
	query = strings.ReplaceAll(query, "FIELD", fieldName)
	query = strings.ReplaceAll(query, "UPDATE", updateAction)
	query = strings.ReplaceAll(query, "DELETE", deleteAction)

	var result thirdPartyAgreementOnThirdPartyQueryResult

	err := client.Execute(query, map[string]any{"id": thirdPartyID}, result.ptrForKind(kind))
	require.NoError(t, err)

	return result.node(kind)
}

func updateThirdPartyAgreement(
	t *testing.T,
	client *testutil.Client,
	kind thirdPartyAgreementKind,
	thirdPartyID string,
	validFrom *string,
	validUntil *string,
) thirdPartyAgreementWireNode {
	t.Helper()

	updateAction, deleteAction := thirdPartyAgreementPermissionActions(kind)

	const nodeSelection = `
		id
		validFrom
		validUntil
		createdAt
		updatedAt
		thirdParty { id }
		file { id fileName }
		canUpdate: permission(action: "UPDATE")
		canDelete: permission(action: "DELETE")
	`

	mutation := thirdPartyAgreementUpdateMutation(kind, nodeSelection)
	mutation = strings.ReplaceAll(mutation, "UPDATE", updateAction)
	mutation = strings.ReplaceAll(mutation, "DELETE", deleteAction)

	input := map[string]any{
		"thirdPartyId": thirdPartyID,
	}
	if validFrom != nil {
		input["validFrom"] = *validFrom
	}

	if validUntil != nil {
		input["validUntil"] = *validUntil
	}

	var result thirdPartyAgreementUpdateResult

	err := client.Execute(
		mutation,
		map[string]any{"input": input},
		result.ptrForKind(kind),
	)
	require.NoError(t, err)

	return result.node(kind)
}

func updateThirdPartyAgreementExpectError(
	t *testing.T,
	client *testutil.Client,
	kind thirdPartyAgreementKind,
	thirdPartyID string,
	validFrom *string,
	validUntil *string,
) error {
	t.Helper()

	updateAction, deleteAction := thirdPartyAgreementPermissionActions(kind)

	const nodeSelection = `id`

	mutation := thirdPartyAgreementUpdateMutation(kind, nodeSelection)
	mutation = strings.ReplaceAll(mutation, "UPDATE", updateAction)
	mutation = strings.ReplaceAll(mutation, "DELETE", deleteAction)

	input := map[string]any{
		"thirdPartyId": thirdPartyID,
	}
	if validFrom != nil {
		input["validFrom"] = *validFrom
	}

	if validUntil != nil {
		input["validUntil"] = *validUntil
	}

	var result thirdPartyAgreementUpdateResult

	return client.Execute(
		mutation,
		map[string]any{"input": input},
		result.ptrForKind(kind),
	)
}

func deleteThirdPartyAgreement(
	t *testing.T,
	client *testutil.Client,
	kind thirdPartyAgreementKind,
	thirdPartyID string,
) string {
	t.Helper()

	var result thirdPartyAgreementDeleteResult

	err := client.Execute(
		thirdPartyAgreementDeleteMutation(kind),
		map[string]any{
			"input": map[string]any{
				"thirdPartyId": thirdPartyID,
			},
		},
		result.ptrForKind(kind),
	)
	require.NoError(t, err)

	return result.deletedThirdPartyID(kind)
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

func uploadThirdPartyAgreementExpectError(
	t *testing.T,
	client *testutil.Client,
	kind thirdPartyAgreementKind,
	thirdPartyID string,
	validFrom string,
	validUntil string,
	fileName string,
) error {
	t.Helper()

	updateAction, deleteAction := thirdPartyAgreementPermissionActions(kind)

	const nodeSelection = `id`

	mutation := thirdPartyAgreementUploadMutation(kind, nodeSelection)
	mutation = strings.ReplaceAll(mutation, "UPDATE", updateAction)
	mutation = strings.ReplaceAll(mutation, "DELETE", deleteAction)

	var result thirdPartyAgreementUploadResult

	return client.ExecuteWithFile(
		mutation,
		map[string]any{
			"input": map[string]any{
				"thirdPartyId": thirdPartyID,
				"fileName":     fileName,
				"validFrom":    validFrom,
				"validUntil":   validUntil,
				"file":         nil,
			},
		},
		"input.file",
		thirdPartyAgreementUploadPDF(fileName),
		result.ptrForKind(kind),
	)
}

type thirdPartyAgreementUploadResult struct {
	baa struct {
		UploadThirdPartyBusinessAssociateAgreement struct {
			ThirdPartyBusinessAssociateAgreement thirdPartyAgreementWireNode `json:"thirdPartyBusinessAssociateAgreement"`
		} `json:"uploadThirdPartyBusinessAssociateAgreement"`
	}
	dpa struct {
		UploadThirdPartyDataPrivacyAgreement struct {
			ThirdPartyDataPrivacyAgreement thirdPartyAgreementWireNode `json:"thirdPartyDataPrivacyAgreement"`
		} `json:"uploadThirdPartyDataPrivacyAgreement"`
	}
}

func (r *thirdPartyAgreementUploadResult) ptrForKind(kind thirdPartyAgreementKind) any {
	switch kind {
	case thirdPartyAgreementKindBAA:
		return &r.baa
	case thirdPartyAgreementKindDPA:
		return &r.dpa
	default:
		panic("unknown agreement kind")
	}
}

func (r *thirdPartyAgreementUploadResult) node(kind thirdPartyAgreementKind) thirdPartyAgreementWireNode {
	switch kind {
	case thirdPartyAgreementKindBAA:
		return r.baa.UploadThirdPartyBusinessAssociateAgreement.ThirdPartyBusinessAssociateAgreement
	case thirdPartyAgreementKindDPA:
		return r.dpa.UploadThirdPartyDataPrivacyAgreement.ThirdPartyDataPrivacyAgreement
	default:
		panic("unknown agreement kind")
	}
}

type thirdPartyAgreementUpdateResult struct {
	baa struct {
		UpdateThirdPartyBusinessAssociateAgreement struct {
			ThirdPartyBusinessAssociateAgreement thirdPartyAgreementWireNode `json:"thirdPartyBusinessAssociateAgreement"`
		} `json:"updateThirdPartyBusinessAssociateAgreement"`
	}
	dpa struct {
		UpdateThirdPartyDataPrivacyAgreement struct {
			ThirdPartyDataPrivacyAgreement thirdPartyAgreementWireNode `json:"thirdPartyDataPrivacyAgreement"`
		} `json:"updateThirdPartyDataPrivacyAgreement"`
	}
}

func (r *thirdPartyAgreementUpdateResult) ptrForKind(kind thirdPartyAgreementKind) any {
	switch kind {
	case thirdPartyAgreementKindBAA:
		return &r.baa
	case thirdPartyAgreementKindDPA:
		return &r.dpa
	default:
		panic("unknown agreement kind")
	}
}

func (r *thirdPartyAgreementUpdateResult) node(kind thirdPartyAgreementKind) thirdPartyAgreementWireNode {
	switch kind {
	case thirdPartyAgreementKindBAA:
		return r.baa.UpdateThirdPartyBusinessAssociateAgreement.ThirdPartyBusinessAssociateAgreement
	case thirdPartyAgreementKindDPA:
		return r.dpa.UpdateThirdPartyDataPrivacyAgreement.ThirdPartyDataPrivacyAgreement
	default:
		panic("unknown agreement kind")
	}
}

type thirdPartyAgreementDeleteResult struct {
	baa struct {
		DeleteThirdPartyBusinessAssociateAgreement struct {
			DeletedThirdPartyID string `json:"deletedThirdPartyId"`
		} `json:"deleteThirdPartyBusinessAssociateAgreement"`
	}
	dpa struct {
		DeleteThirdPartyDataPrivacyAgreement struct {
			DeletedThirdPartyID string `json:"deletedThirdPartyId"`
		} `json:"deleteThirdPartyDataPrivacyAgreement"`
	}
}

func (r *thirdPartyAgreementDeleteResult) ptrForKind(kind thirdPartyAgreementKind) any {
	switch kind {
	case thirdPartyAgreementKindBAA:
		return &r.baa
	case thirdPartyAgreementKindDPA:
		return &r.dpa
	default:
		panic("unknown agreement kind")
	}
}

func (r *thirdPartyAgreementDeleteResult) deletedThirdPartyID(kind thirdPartyAgreementKind) string {
	switch kind {
	case thirdPartyAgreementKindBAA:
		return r.baa.DeleteThirdPartyBusinessAssociateAgreement.DeletedThirdPartyID
	case thirdPartyAgreementKindDPA:
		return r.dpa.DeleteThirdPartyDataPrivacyAgreement.DeletedThirdPartyID
	default:
		panic("unknown agreement kind")
	}
}

type thirdPartyAgreementOnThirdPartyQueryResult struct {
	baa struct {
		Node struct {
			ID                         string                       `json:"id"`
			BusinessAssociateAgreement *thirdPartyAgreementWireNode `json:"businessAssociateAgreement"`
		} `json:"node"`
	}
	dpa struct {
		Node struct {
			ID                   string                       `json:"id"`
			DataPrivacyAgreement *thirdPartyAgreementWireNode `json:"dataPrivacyAgreement"`
		} `json:"node"`
	}
}

func (r *thirdPartyAgreementOnThirdPartyQueryResult) ptrForKind(kind thirdPartyAgreementKind) any {
	switch kind {
	case thirdPartyAgreementKindBAA:
		return &r.baa
	case thirdPartyAgreementKindDPA:
		return &r.dpa
	default:
		panic("unknown agreement kind")
	}
}

func (r *thirdPartyAgreementOnThirdPartyQueryResult) node(kind thirdPartyAgreementKind) *thirdPartyAgreementWireNode {
	switch kind {
	case thirdPartyAgreementKindBAA:
		return r.baa.Node.BusinessAssociateAgreement
	case thirdPartyAgreementKindDPA:
		return r.dpa.Node.DataPrivacyAgreement
	default:
		panic("unknown agreement kind")
	}
}

func thirdPartyAgreementUploadMutation(kind thirdPartyAgreementKind, nodeSelection string) string {
	switch kind {
	case thirdPartyAgreementKindBAA:
		return `
			mutation($input: UploadThirdPartyBusinessAssociateAgreementInput!) {
				uploadThirdPartyBusinessAssociateAgreement(input: $input) {
					thirdPartyBusinessAssociateAgreement {
						` + nodeSelection + `
					}
				}
			}
		`
	case thirdPartyAgreementKindDPA:
		return `
			mutation($input: UploadThirdPartyDataPrivacyAgreementInput!) {
				uploadThirdPartyDataPrivacyAgreement(input: $input) {
					thirdPartyDataPrivacyAgreement {
						` + nodeSelection + `
					}
				}
			}
		`
	default:
		panic("unknown agreement kind")
	}
}

func thirdPartyAgreementUpdateMutation(kind thirdPartyAgreementKind, nodeSelection string) string {
	switch kind {
	case thirdPartyAgreementKindBAA:
		return `
			mutation($input: UpdateThirdPartyBusinessAssociateAgreementInput!) {
				updateThirdPartyBusinessAssociateAgreement(input: $input) {
					thirdPartyBusinessAssociateAgreement {
						` + nodeSelection + `
					}
				}
			}
		`
	case thirdPartyAgreementKindDPA:
		return `
			mutation($input: UpdateThirdPartyDataPrivacyAgreementInput!) {
				updateThirdPartyDataPrivacyAgreement(input: $input) {
					thirdPartyDataPrivacyAgreement {
						` + nodeSelection + `
					}
				}
			}
		`
	default:
		panic("unknown agreement kind")
	}
}

func thirdPartyAgreementDeleteMutation(kind thirdPartyAgreementKind) string {
	switch kind {
	case thirdPartyAgreementKindBAA:
		return `
			mutation($input: DeleteThirdPartyBusinessAssociateAgreementInput!) {
				deleteThirdPartyBusinessAssociateAgreement(input: $input) {
					deletedThirdPartyId
				}
			}
		`
	case thirdPartyAgreementKindDPA:
		return `
			mutation($input: DeleteThirdPartyDataPrivacyAgreementInput!) {
				deleteThirdPartyDataPrivacyAgreement(input: $input) {
					deletedThirdPartyId
				}
			}
		`
	default:
		panic("unknown agreement kind")
	}
}

func thirdPartyAgreementThirdPartyField(kind thirdPartyAgreementKind) string {
	switch kind {
	case thirdPartyAgreementKindBAA:
		return "businessAssociateAgreement"
	case thirdPartyAgreementKindDPA:
		return "dataPrivacyAgreement"
	default:
		panic("unknown agreement kind")
	}
}

func thirdPartyAgreementPermissionActions(kind thirdPartyAgreementKind) (updateAction, deleteAction string) {
	switch kind {
	case thirdPartyAgreementKindBAA:
		return actionThirdPartyBusinessAssociateAgreementUpdate, actionThirdPartyBusinessAssociateAgreementDelete
	case thirdPartyAgreementKindDPA:
		return actionThirdPartyDataPrivacyAgreementUpdate, actionThirdPartyDataPrivacyAgreementDelete
	default:
		panic("unknown agreement kind")
	}
}
