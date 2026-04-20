// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

import type { INodeProperties } from 'n8n-workflow';
import * as createOp from './create.operation';
import * as getOp from './get.operation';
import * as getAllOp from './getAll.operation';
import * as updateOp from './update.operation';
import * as deleteOp from './delete.operation';
import * as archiveOp from './archive.operation';
import * as unarchiveOp from './unarchive.operation';
import * as getVersionOp from './getVersion.operation';
import * as getAllVersionsOp from './getAllVersions.operation';
import * as createDraftVersionOp from './createDraftVersion.operation';
import * as updateVersionOp from './updateVersion.operation';
import * as deleteDraftVersionOp from './deleteDraftVersion.operation';
import * as publishMajorVersionOp from './publishMajorVersion.operation';
import * as publishMinorVersionOp from './publishMinorVersion.operation';
import * as requestApprovalOp from './requestApproval.operation';
import * as voidApprovalOp from './voidApproval.operation';
import * as getSignatureOp from './getSignature.operation';
import * as getAllSignaturesOp from './getAllSignatures.operation';
import * as requestSignatureOp from './requestSignature.operation';
import * as cancelSignatureOp from './cancelSignature.operation';
import * as sendSigningNotificationsOp from './sendSigningNotifications.operation';
import * as getApprovalQuorumOp from './getApprovalQuorum.operation';
import * as getAllApprovalQuorumsOp from './getAllApprovalQuorums.operation';
import * as getApprovalDecisionOp from './getApprovalDecision.operation';
import * as getAllApprovalDecisionsOp from './getAllApprovalDecisions.operation';

export const description: INodeProperties[] = [
	{
		displayName: 'Operation',
		name: 'operation',
		type: 'options',
		noDataExpression: true,
		displayOptions: {
			show: {
				resource: ['document'],
			},
		},
		options: [
			{
				name: 'Archive',
				value: 'archive',
				description: 'Archive a document',
				action: 'Archive a document',
			},
			{
				name: 'Cancel Signature',
				value: 'cancelSignature',
				description: 'Cancel a signature request',
				action: 'Cancel a document signature request',
			},
			{
				name: 'Create',
				value: 'create',
				description: 'Create a new document',
				action: 'Create a document',
			},
			{
				name: 'Create Draft Version',
				value: 'createDraftVersion',
				description: 'Create a new draft version from the latest published version',
				action: 'Create a draft document version',
			},
			{
				name: 'Delete',
				value: 'delete',
				description: 'Delete a document',
				action: 'Delete a document',
			},
			{
				name: 'Delete Draft Version',
				value: 'deleteDraftVersion',
				description: 'Delete a draft document version',
				action: 'Delete a draft document version',
			},
			{
				name: 'Get',
				value: 'get',
				description: 'Get a document',
				action: 'Get a document',
			},
			{
				name: 'Get Approval Decision',
				value: 'getApprovalDecision',
				description: 'Get an approval decision',
				action: 'Get an approval decision',
			},
			{
				name: 'Get Approval Quorum',
				value: 'getApprovalQuorum',
				description: 'Get an approval quorum',
				action: 'Get an approval quorum',
			},
			{
				name: 'Get Many',
				value: 'getAll',
				description: 'Get many documents',
				action: 'Get many documents',
			},
			{
				name: 'Get Many Approval Decisions',
				value: 'getAllApprovalDecisions',
				description: 'Get many approval decisions for an approval quorum',
				action: 'Get many approval decisions',
			},
			{
				name: 'Get Many Approval Quorums',
				value: 'getAllApprovalQuorums',
				description: 'Get many approval quorums for a document version',
				action: 'Get many approval quorums',
			},
			{
				name: 'Get Many Signatures',
				value: 'getAllSignatures',
				description: 'Get many signatures for a document version',
				action: 'Get many document version signatures',
			},
			{
				name: 'Get Many Versions',
				value: 'getAllVersions',
				description: 'Get many versions of a document',
				action: 'Get many document versions',
			},
			{
				name: 'Get Signature',
				value: 'getSignature',
				description: 'Get a document version signature',
				action: 'Get a document version signature',
			},
			{
				name: 'Get Version',
				value: 'getVersion',
				description: 'Get a document version',
				action: 'Get a document version',
			},
			{
				name: 'Publish Major Version',
				value: 'publishMajorVersion',
				description: 'Publish a draft as a new major version',
				action: 'Publish a major document version',
			},
			{
				name: 'Publish Minor Version',
				value: 'publishMinorVersion',
				description: 'Publish a draft as a minor version',
				action: 'Publish a minor document version',
			},
			{
				name: 'Request Approval',
				value: 'requestApproval',
				description: 'Request approval for a document version',
				action: 'Request document version approval',
			},
			{
				name: 'Request Signature',
				value: 'requestSignature',
				description: 'Request a signature for a document version',
				action: 'Request a document version signature',
			},
			{
				name: 'Send Signing Notifications',
				value: 'sendSigningNotifications',
				description: 'Send signing notifications to all pending signatories',
				action: 'Send signing notifications',
			},
			{
				name: 'Unarchive',
				value: 'unarchive',
				description: 'Unarchive a document',
				action: 'Unarchive a document',
			},
			{
				name: 'Update',
				value: 'update',
				description: 'Update an existing document',
				action: 'Update a document',
			},
			{
				name: 'Update Version',
				value: 'updateVersion',
				description: 'Update a draft document version',
				action: 'Update a document version',
			},
			{
				name: 'Void Approval',
				value: 'voidApproval',
				description: 'Void a pending approval request',
				action: 'Void a document version approval',
			},
		],
		default: 'create',
	},
	...createOp.description,
	...getOp.description,
	...getAllOp.description,
	...updateOp.description,
	...deleteOp.description,
	...archiveOp.description,
	...unarchiveOp.description,
	...getVersionOp.description,
	...getAllVersionsOp.description,
	...createDraftVersionOp.description,
	...updateVersionOp.description,
	...deleteDraftVersionOp.description,
	...publishMajorVersionOp.description,
	...publishMinorVersionOp.description,
	...requestApprovalOp.description,
	...voidApprovalOp.description,
	...getSignatureOp.description,
	...getAllSignaturesOp.description,
	...requestSignatureOp.description,
	...cancelSignatureOp.description,
	...sendSigningNotificationsOp.description,
	...getApprovalQuorumOp.description,
	...getAllApprovalQuorumsOp.description,
	...getApprovalDecisionOp.description,
	...getAllApprovalDecisionsOp.description,
];

export {
	createOp as create,
	getOp as get,
	getAllOp as getAll,
	updateOp as update,
	deleteOp as delete,
	archiveOp as archive,
	unarchiveOp as unarchive,
	getVersionOp as getVersion,
	getAllVersionsOp as getAllVersions,
	createDraftVersionOp as createDraftVersion,
	updateVersionOp as updateVersion,
	deleteDraftVersionOp as deleteDraftVersion,
	publishMajorVersionOp as publishMajorVersion,
	publishMinorVersionOp as publishMinorVersion,
	requestApprovalOp as requestApproval,
	voidApprovalOp as voidApproval,
	getSignatureOp as getSignature,
	getAllSignaturesOp as getAllSignatures,
	requestSignatureOp as requestSignature,
	cancelSignatureOp as cancelSignature,
	sendSigningNotificationsOp as sendSigningNotifications,
	getApprovalQuorumOp as getApprovalQuorum,
	getAllApprovalQuorumsOp as getAllApprovalQuorums,
	getApprovalDecisionOp as getApprovalDecision,
	getAllApprovalDecisionsOp as getAllApprovalDecisions,
};
