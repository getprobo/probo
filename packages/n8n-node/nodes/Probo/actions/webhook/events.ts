// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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
	{ name: 'Third Party Created', value: 'THIRD_PARTY_CREATED' },
	{ name: 'Third Party Deleted', value: 'THIRD_PARTY_DELETED' },
	{ name: 'Third Party Updated', value: 'THIRD_PARTY_UPDATED' },
	{ name: 'User Created', value: 'USER_CREATED' },
	{ name: 'User Deleted', value: 'USER_DELETED' },
	{ name: 'User Updated', value: 'USER_UPDATED' },
];
