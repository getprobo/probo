// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import type { INodePropertyOptions } from 'n8n-workflow';

export const WEBHOOK_EVENT_OPTIONS: INodePropertyOptions[] = [
	{ name: 'Document Archived', value: 'DOCUMENT_ARCHIVED' },
	{ name: 'Document Created', value: 'DOCUMENT_CREATED' },
	{ name: 'Document Deleted', value: 'DOCUMENT_DELETED' },
	{ name: 'Document Unarchived', value: 'DOCUMENT_UNARCHIVED' },
	{ name: 'Document Updated', value: 'DOCUMENT_UPDATED' },
	{ name: 'Document Version Approval Quorum Approved', value: 'DOCUMENT_VERSION_APPROVAL_QUORUM_APPROVED' },
	{ name: 'Document Version Approval Quorum Rejected', value: 'DOCUMENT_VERSION_APPROVAL_QUORUM_REJECTED' },
	{ name: 'Document Version Approval Quorum Requested', value: 'DOCUMENT_VERSION_APPROVAL_QUORUM_REQUESTED' },
	{ name: 'Document Version Approval Quorum Updated', value: 'DOCUMENT_VERSION_APPROVAL_QUORUM_UPDATED' },
	{ name: 'Document Version Approval Quorum Voided', value: 'DOCUMENT_VERSION_APPROVAL_QUORUM_VOIDED' },
	{ name: 'Document Version Created', value: 'DOCUMENT_VERSION_CREATED' },
	{ name: 'Document Version Deleted', value: 'DOCUMENT_VERSION_DELETED' },
	{ name: 'Document Version Published', value: 'DOCUMENT_VERSION_PUBLISHED' },
	{ name: 'Document Version Rejected', value: 'DOCUMENT_VERSION_REJECTED' },
	{ name: 'Document Version Signature Cancelled', value: 'DOCUMENT_VERSION_SIGNATURE_CANCELLED' },
	{ name: 'Document Version Signature Requested', value: 'DOCUMENT_VERSION_SIGNATURE_REQUESTED' },
	{ name: 'Document Version Signature Signed', value: 'DOCUMENT_VERSION_SIGNATURE_SIGNED' },
	{ name: 'Document Version Updated', value: 'DOCUMENT_VERSION_UPDATED' },
	{ name: 'Obligation Created', value: 'OBLIGATION_CREATED' },
	{ name: 'Obligation Deleted', value: 'OBLIGATION_DELETED' },
	{ name: 'Obligation Updated', value: 'OBLIGATION_UPDATED' },
	{ name: 'Right Request Created', value: 'RIGHT_REQUEST_CREATED' },
	{ name: 'Right Request Deleted', value: 'RIGHT_REQUEST_DELETED' },
	{ name: 'Right Request Updated', value: 'RIGHT_REQUEST_UPDATED' },
	{ name: 'Third Party Created', value: 'THIRD_PARTY_CREATED' },
	{ name: 'Third Party Deleted', value: 'THIRD_PARTY_DELETED' },
	{ name: 'Third Party Updated', value: 'THIRD_PARTY_UPDATED' },
	{ name: 'User Created', value: 'USER_CREATED' },
	{ name: 'User Deleted', value: 'USER_DELETED' },
	{ name: 'User Updated', value: 'USER_UPDATED' },
];
