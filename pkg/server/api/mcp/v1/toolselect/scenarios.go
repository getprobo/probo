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

package toolselect

// DocumentScenarios are realistic document tasks and the tool an agent should
// call first. Names that share a stem (get/read/delete/publish) are the cases
// where a name-only choice goes wrong.
var DocumentScenarios = []Scenario{
	{Task: "show me the markdown body of the access control policy", Want: "readDocument"},
	{Task: "find policies by title in this organization", Want: "listDocuments"},
	{Task: "get the document metadata and published version numbers without the body", Want: "getDocument"},
	{Task: "change the draft wording of this policy", Want: "updateDocument"},
	{Task: "create a new policy document from this markdown", Want: "addDocument"},
	{Task: "discard the unpublished draft and revert to the last published version", Want: "deleteDocumentDraft"},
	{Task: "permanently delete the policy and every version", Want: "deleteDocument"},
	{Task: "archive the policy so nobody can edit it", Want: "archiveDocument"},
	{Task: "restore an archived policy so it can be edited again", Want: "unarchiveDocument"},
	{Task: "list the revisions of this document", Want: "listDocumentVersions"},
	{Task: "I already have the version id, get that revision markdown", Want: "getDocumentVersion"},
	{Task: "publish the draft as a minor version", Want: "publishDocument"},
	{Task: "email the default approvers for a new major version", Want: "publishDocument"},
	{Task: "who still needs to sign this published revision", Want: "listDocumentVersionSignatures"},
	{Task: "ask this profile to sign the published revision", Want: "requestDocumentVersionSignature"},
	{Task: "cancel an outstanding signature request", Want: "cancelSignatureRequest"},
	{Task: "sign the published revision as me", Want: "signDocument"},
	{Task: "I am an approver, approve the pending revision", Want: "approveDocumentVersion"},
	{Task: "I am an approver, reject the pending revision", Want: "rejectDocumentVersion"},
	{Task: "void the pending approval quorum", Want: "voidDocumentVersionApproval"},
	{Task: "remove the policy from the compliance portal without deleting it", Want: "deleteCompliancePortalDocument"},
}

type Scenario struct {
	Task string
	Want string
}
