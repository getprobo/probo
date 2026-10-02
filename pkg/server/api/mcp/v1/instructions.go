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

package mcp_v1

const documentPlaybook = `Document ids and document version ids are different.

A document id identifies the policy or other document. listDocuments and getDocument return metadata for that id: title (from the latest version), document_type, status, and current_published_major/minor. They do not return markdown. Pass that id as id to getDocument, updateDocument, deleteDocumentDraft, archiveDocument, and unarchiveDocument. Pass it as document_id to readDocument, listDocumentVersions, publishDocument, and deleteDocument.

A document version id identifies one revision. It has major, minor, status (DRAFT, PENDING_APPROVAL, PUBLISHED), and markdown content. getDocumentVersion takes that id in id. listDocumentVersionSignatures, requestDocumentVersionSignature, signDocument, voidDocumentVersionApproval, approveDocumentVersion, and rejectDocumentVersion take it in document_version_id. cancelSignatureRequest takes a signature id in document_version_signature_id.

To read markdown, call readDocument with the document id. version=LATEST is the tip that updateDocument edits. version=PUBLISHED is the current published revision. version=DRAFT fails unless the tip is still a draft.

To change wording, call updateDocument with the document id. To publish that draft, call publishDocument with document_id, minor, and changelog. minor=true publishes a minor version and needs an existing published major. minor=false publishes a major version, or emails the document's default approvers when they are set. publishDocument has no approver_ids field; set default_approver_ids with updateDocument. approveDocumentVersion is the approver's vote, not the publish step.

deleteDocumentDraft discards the unpublished draft and keeps the document. deleteDocument deletes the document and every version. deleteCompliancePortalDocument only removes it from a compliance portal.`

const (
	fullServerInstructions = documentPlaybook + `

For document work, connect to this same MCP URL with ?toolset=documents so only document tools are listed. The OAuth resource stays /api/mcp/v1.`

	documentToolsetInstructions = documentPlaybook + `

This connection lists document tools only.`
)
