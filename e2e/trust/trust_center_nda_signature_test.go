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

package trust_test

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/testutil"
)

// minimalPDFBase64 is a minimal but structurally valid one-page PDF (no
// visible content). Unlike a hand-rolled "%PDF-1.4 ... /Catalog" stub, it
// has a real Pages tree, so it survives the watermark stamping that
// ProvisionMember runs when creating a pending NDA signature.
const minimalPDFBase64 = "JVBERi0xLjcKJeLjz9MKMSAwIG9iago8PC9QYWdlcyAyIDAgUi9UeXBlL0NhdGFsb2c+PgplbmRv" +
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

// TestTrustCenter_AcceptElectronicSignature_RejectsForeignSignature is a
// regression test for GHSA-22xj-f767-ppw6: any self-provisioned trust
// center visitor could accept another visitor's NDA signature, or inject
// audit-trail events into it, simply by knowing its GID.
func TestTrustCenter_AcceptElectronicSignature_RejectsForeignSignature(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	trustCenterID := lookupTrustCenterID(t, owner)

	uploadTrustCenterNDA(t, owner, trustCenterID)
	trustHost := lookupTrustHost(t, owner, trustCenterID)

	victim := testutil.SelfProvisionTrustCenterVisitor(t, trustHost)
	attacker := testutil.SelfProvisionTrustCenterVisitor(t, trustHost)

	victimSignatureID, victimSignatureStatus := viewerSignature(t, victim, trustHost)
	require.NotEmpty(t, victimSignatureID)
	require.Equal(t, "PENDING", victimSignatureStatus)

	const acceptMutation = `
		mutation($input: AcceptElectronicSignatureInput!) {
			acceptElectronicSignature(input: $input) {
				signature { id status }
			}
		}
	`

	err := attacker.ExecuteTrust(trustHost, acceptMutation, map[string]any{
		"input": map[string]any{"signatureId": victimSignatureID},
	}, nil)
	require.Error(t, err, "attacker must not be able to accept another visitor's signature")
	assertForbidden(t, err)

	const recordEventMutation = `
		mutation($input: RecordSigningEventInput!) {
			recordSigningEvent(input: $input) {
				success
			}
		}
	`

	err = attacker.ExecuteTrust(trustHost, recordEventMutation, map[string]any{
		"input": map[string]any{
			"signatureId": victimSignatureID,
			"eventType":   "DOCUMENT_VIEWED",
		},
	}, nil)
	require.Error(t, err, "attacker must not be able to inject events into another visitor's signature")
	assertForbidden(t, err)

	_, statusAfterAttack := viewerSignature(t, victim, trustHost)
	assert.Equal(t, "PENDING", statusAfterAttack, "attack attempts must not have mutated the victim's signature")

	var acceptResult struct {
		AcceptElectronicSignature struct {
			Signature struct {
				ID     string `json:"id"`
				Status string `json:"status"`
			} `json:"signature"`
		} `json:"acceptElectronicSignature"`
	}

	err = victim.ExecuteTrust(trustHost, acceptMutation, map[string]any{
		"input": map[string]any{"signatureId": victimSignatureID},
	}, &acceptResult)
	require.NoError(t, err, "the legitimate signer must still be able to accept their own signature")
	assert.Equal(t, "ACCEPTED", acceptResult.AcceptElectronicSignature.Signature.Status)
}

func uploadTrustCenterNDA(t *testing.T, owner *testutil.Client, trustCenterID string) {
	t.Helper()

	const query = `
		mutation($input: UploadTrustCenterNDAInput!) {
			uploadTrustCenterNDA(input: $input) {
				trustCenter { id }
			}
		}
	`

	pdfContent, err := base64.StdEncoding.DecodeString(minimalPDFBase64)
	require.NoError(t, err)

	err = owner.ExecuteWithFile(query, map[string]any{
		"input": map[string]any{
			"trustCenterId": trustCenterID,
			"fileName":      "nda.pdf",
			"file":          nil,
		},
	}, "input.file", testutil.UploadFile{
		Filename:    "nda.pdf",
		ContentType: "application/pdf",
		Content:     pdfContent,
	}, nil)
	require.NoError(t, err)
}

func viewerSignature(t *testing.T, visitor *testutil.Client, trustHost string) (string, string) {
	t.Helper()

	const query = `
		query {
			currentTrustCenter {
				nonDisclosureAgreement {
					viewerSignature { id status }
				}
			}
		}
	`

	var result struct {
		CurrentTrustCenter struct {
			NonDisclosureAgreement struct {
				ViewerSignature struct {
					ID     string `json:"id"`
					Status string `json:"status"`
				} `json:"viewerSignature"`
			} `json:"nonDisclosureAgreement"`
		} `json:"currentTrustCenter"`
	}

	err := visitor.ExecuteTrust(trustHost, query, nil, &result)
	require.NoError(t, err)

	sig := result.CurrentTrustCenter.NonDisclosureAgreement.ViewerSignature

	return sig.ID, sig.Status
}

func assertForbidden(t *testing.T, err error) {
	t.Helper()

	gqlErrs, ok := err.(testutil.GraphQLErrors)
	require.True(t, ok, "expected a GraphQL error, got %T: %v", err, err)
	require.NotEmpty(t, gqlErrs)
	assert.Equal(t, "FORBIDDEN", gqlErrs[0].Code())
}
