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
	"encoding/base64"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

const compliancePortalAccessNDAQuery = `
	query($id: ID!) {
		node(id: $id) {
			... on CompliancePortalAccess {
				id
				ndaBypassed
				recordedNdaSignedAt
				recordedNda {
					fileName
					downloadUrl
				}
			}
		}
	}
`

const setCompliancePortalAccessNDABypassMutation = `
	mutation($input: SetCompliancePortalAccessNDABypassInput!) {
		setCompliancePortalAccessNDABypass(input: $input) {
			compliancePortalAccess {
				id
				ndaBypassed
			}
		}
	}
`

const uploadCompliancePortalAccessRecordedNDAMutation = `
	mutation($input: UploadCompliancePortalAccessRecordedNDAInput!) {
		uploadCompliancePortalAccessRecordedNDA(input: $input) {
			compliancePortalAccess {
				id
				ndaBypassed
				recordedNdaSignedAt
				recordedNda {
					fileName
					downloadUrl
				}
			}
		}
	}
`

const deleteCompliancePortalAccessRecordedNDAMutation = `
	mutation($input: DeleteCompliancePortalAccessRecordedNDAInput!) {
		deleteCompliancePortalAccessRecordedNDA(input: $input) {
			compliancePortalAccess {
				id
				recordedNdaSignedAt
				recordedNda { id }
			}
		}
	}
`

type recordedNDAFile struct {
	FileName    string `json:"fileName"`
	DownloadURL string `json:"downloadUrl"`
}

type compliancePortalAccessNDANode struct {
	ID                  string           `json:"id"`
	NDABypassed         bool             `json:"ndaBypassed"`
	RecordedNDASignedAt *string          `json:"recordedNdaSignedAt"`
	RecordedNDA         *recordedNDAFile `json:"recordedNda"`
}

func loadCompliancePortalAccessNDA(
	t *testing.T,
	client *testutil.Client,
	accessID string,
) compliancePortalAccessNDANode {
	t.Helper()

	var result struct {
		Node compliancePortalAccessNDANode `json:"node"`
	}

	err := client.Execute(compliancePortalAccessNDAQuery, map[string]any{"id": accessID}, &result)
	require.NoError(t, err)
	require.Equal(t, accessID, result.Node.ID)

	return result.Node
}

func setCompliancePortalAccessNDABypass(
	t *testing.T,
	client *testutil.Client,
	accessID string,
	bypassed bool,
) bool {
	t.Helper()

	var result struct {
		SetCompliancePortalAccessNDABypass struct {
			CompliancePortalAccess struct {
				NDABypassed bool `json:"ndaBypassed"`
			} `json:"compliancePortalAccess"`
		} `json:"setCompliancePortalAccessNDABypass"`
	}

	err := client.Execute(setCompliancePortalAccessNDABypassMutation, map[string]any{
		"input": map[string]any{
			"id":       accessID,
			"bypassed": bypassed,
		},
	}, &result)
	require.NoError(t, err)

	return result.SetCompliancePortalAccessNDABypass.CompliancePortalAccess.NDABypassed
}

func TestCompliancePortalAccess_NDABypassAndRecordedFile(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	compliancePortalID := compliancePortalID(t, owner)
	email := factory.SafeEmail()
	node := createCompliancePortalAccess(t, owner, map[string]any{
		"compliancePortalId": compliancePortalID,
		"email":              email,
	})

	loaded := loadCompliancePortalAccessNDA(t, owner, node.ID)
	assert.False(t, loaded.NDABypassed)
	assert.Nil(t, loaded.RecordedNDA)
	assert.Nil(t, loaded.RecordedNDASignedAt)

	assert.True(t, setCompliancePortalAccessNDABypass(t, owner, node.ID, true))
	assert.False(t, setCompliancePortalAccessNDABypass(t, owner, node.ID, false))

	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
	err := viewer.ExecuteShouldFail(setCompliancePortalAccessNDABypassMutation, map[string]any{
		"input": map[string]any{
			"id":       node.ID,
			"bypassed": true,
		},
	})
	require.Error(t, err)

	signedAt := time.Now().Add(-24 * time.Hour).UTC().Truncate(time.Second)
	pdfContent, err := base64.StdEncoding.DecodeString(recordedNDAPDFBase64)
	require.NoError(t, err)

	var uploaded struct {
		UploadCompliancePortalAccessRecordedNDA struct {
			CompliancePortalAccess compliancePortalAccessNDANode `json:"compliancePortalAccess"`
		} `json:"uploadCompliancePortalAccessRecordedNDA"`
	}

	err = owner.ExecuteWithFile(uploadCompliancePortalAccessRecordedNDAMutation, map[string]any{
		"input": map[string]any{
			"compliancePortalAccessId": node.ID,
			"fileName":                 "external-nda.pdf",
			"file":                     nil,
			"signedAt":                 signedAt.Format(time.RFC3339),
		},
	}, "input.file", testutil.UploadFile{
		Filename:    "external-nda.pdf",
		ContentType: "application/pdf",
		Content:     pdfContent,
	}, &uploaded)
	require.NoError(t, err)

	recorded := uploaded.UploadCompliancePortalAccessRecordedNDA.CompliancePortalAccess
	assert.False(t, recorded.NDABypassed)
	require.NotNil(t, recorded.RecordedNDA)
	assert.Equal(t, "external-nda.pdf", recorded.RecordedNDA.FileName)
	assert.NotEmpty(t, recorded.RecordedNDA.DownloadURL)
	require.NotNil(t, recorded.RecordedNDASignedAt)
	parsedSignedAt, err := time.Parse(time.RFC3339, *recorded.RecordedNDASignedAt)
	require.NoError(t, err)
	assert.WithinDuration(t, signedAt, parsedSignedAt, time.Second)

	err = owner.ExecuteWithFile(uploadCompliancePortalAccessRecordedNDAMutation, map[string]any{
		"input": map[string]any{
			"compliancePortalAccessId": node.ID,
			"fileName":                 "notes.txt",
			"file":                     nil,
			"signedAt":                 signedAt.Format(time.RFC3339),
		},
	}, "input.file", testutil.UploadFile{
		Filename:    "notes.txt",
		ContentType: "text/plain",
		Content:     []byte("not a pdf"),
	}, nil)
	require.Error(t, err)
	assertGraphQLCode(t, err, "INVALID")

	stillRecorded := loadCompliancePortalAccessNDA(t, owner, node.ID)
	require.NotNil(t, stillRecorded.RecordedNDA)
	assert.Equal(t, "external-nda.pdf", stillRecorded.RecordedNDA.FileName)

	var deleted struct {
		DeleteCompliancePortalAccessRecordedNDA struct {
			CompliancePortalAccess struct {
				RecordedNDASignedAt *string `json:"recordedNdaSignedAt"`
				RecordedNDA         *struct {
					ID string `json:"id"`
				} `json:"recordedNda"`
			} `json:"compliancePortalAccess"`
		} `json:"deleteCompliancePortalAccessRecordedNDA"`
	}

	err = owner.Execute(deleteCompliancePortalAccessRecordedNDAMutation, map[string]any{
		"input": map[string]any{
			"compliancePortalAccessId": node.ID,
		},
	}, &deleted)
	require.NoError(t, err)
	assert.Nil(t, deleted.DeleteCompliancePortalAccessRecordedNDA.CompliancePortalAccess.RecordedNDA)
	assert.Nil(t, deleted.DeleteCompliancePortalAccessRecordedNDA.CompliancePortalAccess.RecordedNDASignedAt)

	cleared := loadCompliancePortalAccessNDA(t, owner, node.ID)
	assert.Nil(t, cleared.RecordedNDA)
	assert.Nil(t, cleared.RecordedNDASignedAt)
	assert.False(t, cleared.NDABypassed)
}

func assertGraphQLCode(t *testing.T, err error, code string) {
	t.Helper()

	gqlErrs, ok := err.(testutil.GraphQLErrors)
	require.True(t, ok, "expected a GraphQL error, got %T: %v", err, err)
	require.NotEmpty(t, gqlErrs)
	assert.Equal(t, code, gqlErrs[0].Code())
}

// recordedNDAPDFBase64 is the same one-page PDF the portal NDA upload uses.
const recordedNDAPDFBase64 = "JVBERi0xLjcKJeLjz9MKMSAwIG9iago8PC9QYWdlcyAyIDAgUi9UeXBlL0NhdGFsb2c+PgplbmRv" +
	"YmoKNCAwIG9iago8PC9GaWx0ZXIvRmxhdGVEZWNvZGUvTGVuZ3RoIDExPj4Kc3RyZWFtCnicAQAA" +
	"//8AAAABZW5kc3RyZWFtCmVuZG9iagoyMyAwIG9iago8PC9GaWx0ZXIvRmxhdGVEZWNvZGUvRmly" +
	"c3QgMTQvTGVuZ3RoIDE2Ni9OIDMvVHlwZS9PYmpTdG0+PgpzdHJlYW0KeJxczkHKwjAQBeCrzAn+" +
	"Sdr+ugmzaEEEEUp1V7qI7SAFSaSZit5epi6UZhPem2/x/sFADtsCMrCZdQ6rGISDJCjAQIO1nzgI" +
	"ZEtoOMV56jk5h7sYRD8Lud6IiPD8ujPW/spEzmHpE6vCPd8eLGPv8TRfRI1C++EqFl7FOQhYPIxD" +
	"anVW0+GRh9GX8dmaP/PzYBULsyo2q6L7TktE7wAAAP//bYpDg2VuZHN0cmVhbQplbmRvYmoKNiAw" +
	"IG9iago8PC9DcmVhdGlvbkRhdGUoRDoyMDE5MDcwNDEwMjcyOCswMicwMCcpL01vZERhdGUoRDoy" +
	"MDE5MDcwNDEwMjcyOCswMicwMCcpL1Byb2R1Y2VyKHBkZmNwdSB2MC4xLjI1KT4+CmVuZG9iagoy" +
	"MiAwIG9iago8PC9GaWx0ZXIvRmxhdGVEZWNvZGUvSURbPDEyNjNDMjQ4RDcyOUI5MTNGQzM5MkYy" +
	"NjQ2MTk1NDJBPiA8MGNiOTUzNjAyNzk4NWQ1NTQ5YzNlY2ZlNmE5MjM5ZTc+XS9JbmRleFswIDIy" +
	"IDIzIDFdL0luZm8gNiAwIFIvTGVuZ3RoIDcyL1Jvb3QgMSAwIFIvU2l6ZSAyNC9UeXBlL1hSZWYv" +
	"V1sxIDIgMl0+PgpzdHJlYW0KeJwkzEkKgDAUBNH62TgbR7yMFxS8c6RMLx70puAsJciQuEgSwV0v" +
	"ES//AhpppZNeBhllklmyLLLKJrsclh/4AgAA//9tzQSTZW5kc3RyZWFtCmVuZG9iagoKc3RhcnR4" +
	"cmVmCjUxMgolJUVPRg=="
