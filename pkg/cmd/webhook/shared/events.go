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

package shared

// ValidEvents is the set of event types accepted by the webhook
// create and update commands. Keep in sync with
// pkg/coredata/webhook_event_type.go.
var ValidEvents = []string{
	"THIRD_PARTY_CREATED",
	"THIRD_PARTY_UPDATED",
	"THIRD_PARTY_DELETED",
	"USER_CREATED",
	"USER_UPDATED",
	"USER_DELETED",
	"OBLIGATION_CREATED",
	"OBLIGATION_UPDATED",
	"OBLIGATION_DELETED",
	"DOCUMENT_CREATED",
	"DOCUMENT_UPDATED",
	"DOCUMENT_ARCHIVED",
	"DOCUMENT_UNARCHIVED",
	"DOCUMENT_DELETED",
	"DOCUMENT_VERSION_CREATED",
	"DOCUMENT_VERSION_UPDATED",
	"DOCUMENT_VERSION_PUBLISHED",
	"DOCUMENT_VERSION_REJECTED",
	"DOCUMENT_VERSION_DELETED",
	"DOCUMENT_VERSION_SIGNATURE_REQUESTED",
	"DOCUMENT_VERSION_SIGNATURE_SIGNED",
	"DOCUMENT_VERSION_SIGNATURE_CANCELLED",
	"DOCUMENT_VERSION_APPROVAL_QUORUM_REQUESTED",
	"DOCUMENT_VERSION_APPROVAL_QUORUM_UPDATED",
	"DOCUMENT_VERSION_APPROVAL_QUORUM_APPROVED",
	"DOCUMENT_VERSION_APPROVAL_QUORUM_REJECTED",
	"DOCUMENT_VERSION_APPROVAL_QUORUM_VOIDED",
}
