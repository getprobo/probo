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

package mcp_test

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

const thirdPartyAgreementPDF = "%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\ntrailer\n<< /Root 1 0 R >>\n%%EOF"

type thirdPartyAgreement struct {
	ID           string `json:"id"`
	ThirdPartyID string `json:"third_party_id"`
	Title        string `json:"title"`
	Type         string `json:"type"`
	Validity     *struct {
		Start *string `json:"start"`
		End   *string `json:"end"`
	} `json:"validity"`
	FileID string `json:"file_id"`
}

func mcpThirdPartyAgreementInput(thirdPartyID, agreementType, fileName string) map[string]any {
	return map[string]any{
		"third_party_id": thirdPartyID,
		"title":          "Agreement " + fileName,
		"type":           agreementType,
		"file_name":      fileName,
		"content_base64": base64.StdEncoding.EncodeToString([]byte(thirdPartyAgreementPDF)),
	}
}

func TestMCP_ThirdPartyAgreement_Lifecycle(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	thirdPartyID := factory.CreateThirdParty(owner)

	input := mcpThirdPartyAgreementInput(thirdPartyID, "MASTER_SERVICE", "msa.pdf")
	input["validity"] = map[string]any{
		"start": "2024-01-01T00:00:00Z",
		"end":   "2025-01-01T00:00:00Z",
	}

	var uploaded struct {
		ThirdPartyAgreement thirdPartyAgreement `json:"third_party_agreement"`
	}
	mc.CallToolInto("uploadThirdPartyAgreement", input, &uploaded)

	agreement := uploaded.ThirdPartyAgreement
	require.NotEmpty(t, agreement.ID)
	require.NotEmpty(t, agreement.FileID)
	assert.Equal(t, thirdPartyID, agreement.ThirdPartyID)
	assert.Equal(t, "Agreement msa.pdf", agreement.Title)
	assert.Equal(t, "MASTER_SERVICE", agreement.Type)
	require.NotNil(t, agreement.Validity)
	require.NotNil(t, agreement.Validity.Start)
	require.NotNil(t, agreement.Validity.End)
	assert.Contains(t, *agreement.Validity.Start, "2024-01-01")
	assert.Contains(t, *agreement.Validity.End, "2025-01-01")

	var second struct {
		ThirdPartyAgreement thirdPartyAgreement `json:"third_party_agreement"`
	}
	mc.CallToolInto(
		"uploadThirdPartyAgreement",
		mcpThirdPartyAgreementInput(thirdPartyID, "MASTER_SERVICE", "msa-renewal.pdf"),
		&second,
	)
	require.NotEmpty(t, second.ThirdPartyAgreement.ID)
	assert.NotEqual(t, agreement.ID, second.ThirdPartyAgreement.ID)

	var listed struct {
		ThirdPartyAgreements []thirdPartyAgreement `json:"third_party_agreements"`
	}
	mc.CallToolInto(
		"listThirdPartyAgreements",
		map[string]any{"third_party_id": thirdPartyID},
		&listed,
	)
	assert.Len(t, listed.ThirdPartyAgreements, 2)

	var got struct {
		ThirdPartyAgreement thirdPartyAgreement `json:"third_party_agreement"`
	}
	mc.CallToolInto("getThirdPartyAgreement", map[string]any{"id": agreement.ID}, &got)
	assert.Equal(t, agreement.ID, got.ThirdPartyAgreement.ID)
	assert.Equal(t, agreement.FileID, got.ThirdPartyAgreement.FileID)

	var updated struct {
		ThirdPartyAgreement thirdPartyAgreement `json:"third_party_agreement"`
	}
	mc.CallToolInto(
		"updateThirdPartyAgreement",
		map[string]any{"id": agreement.ID, "type": "TERMS", "title": "Terms 2026"},
		&updated,
	)
	assert.Equal(t, "Terms 2026", updated.ThirdPartyAgreement.Title)
	assert.Equal(t, "TERMS", updated.ThirdPartyAgreement.Type)
	assert.Equal(t, agreement.FileID, updated.ThirdPartyAgreement.FileID)

	var deleted struct {
		DeletedThirdPartyAgreementID string `json:"deleted_third_party_agreement_id"`
	}
	mc.CallToolInto("deleteThirdPartyAgreement", map[string]any{"id": agreement.ID}, &deleted)
	assert.Equal(t, agreement.ID, deleted.DeletedThirdPartyAgreementID)

	mc.CallToolInto(
		"listThirdPartyAgreements",
		map[string]any{"third_party_id": thirdPartyID},
		&listed,
	)
	require.Len(t, listed.ThirdPartyAgreements, 1)
	assert.Equal(t, second.ThirdPartyAgreement.ID, listed.ThirdPartyAgreements[0].ID)
}

func TestMCP_ThirdPartyAgreement_UploadRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	thirdPartyID := factory.CreateThirdParty(owner)

	t.Run("rejects a file that is not a PDF", func(t *testing.T) {
		t.Parallel()

		input := mcpThirdPartyAgreementInput(thirdPartyID, "OTHER", "dot.png")
		input["content_base64"] = filePNG

		msg := mc.CallToolExpectToolError("uploadThirdPartyAgreement", input)
		assert.Contains(t, msg, `file extension ".png" is not allowed`)
	})

	t.Run("rejects content that is not base64", func(t *testing.T) {
		t.Parallel()

		input := mcpThirdPartyAgreementInput(thirdPartyID, "OTHER", "bad.pdf")
		input["content_base64"] = "not base64 !!!"

		msg := mc.CallToolExpectToolError("uploadThirdPartyAgreement", input)
		assert.Contains(t, msg, "content_base64 is not valid base64")
	})

	t.Run("cannot upload to another organization's third party", func(t *testing.T) {
		t.Parallel()

		other := testutil.NewClient(t, testutil.RoleOwner)
		otherThirdPartyID := factory.CreateThirdParty(other)

		msg := mc.CallToolExpectToolError(
			"uploadThirdPartyAgreement",
			mcpThirdPartyAgreementInput(otherThirdPartyID, "OTHER", "other.pdf"),
		)
		assert.NotEmpty(t, msg)
	})
}

func TestMCP_ThirdPartyAgreement_UploadRequiresPermission(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	thirdPartyID := factory.CreateThirdParty(owner)

	auditor := testutil.NewClientInOrg(t, testutil.RoleAuditor, owner)
	auditorMCP := testutil.NewMCPClient(t, auditor)

	msg := auditorMCP.CallToolExpectToolError(
		"uploadThirdPartyAgreement",
		mcpThirdPartyAgreementInput(thirdPartyID, "OTHER", "denied.pdf"),
	)
	assert.Contains(t, msg, "permission denied")
}
