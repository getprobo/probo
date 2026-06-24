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

import type { INodeProperties } from 'n8n-workflow';
import * as createOp from './create.operation';
import * as getOp from './get.operation';
import * as getAllOp from './getAll.operation';
import * as updateOp from './update.operation';
import * as deleteOp from './delete.operation';
import * as archiveOp from './archive.operation';
import * as unarchiveOp from './unarchive.operation';
import * as getVersionOp from './getVersion.operation';
import * as getLatestPublishedVersionIdOp from './getLatestPublishedVersionId.operation';
import * as getAllVersionsOp from './getAllVersions.operation';
import * as updateVersionOp from './updateVersion.operation';
import * as deleteDraftVersionOp from './deleteDraftVersion.operation';
import * as publishOp from './publish.operation';
import * as voidApprovalOp from './voidApproval.operation';
import * as getSignatureOp from './getSignature.operation';
import * as getAllSignaturesOp from './getAllSignatures.operation';
import * as requestSignatureOp from './requestSignature.operation';
import * as cancelSignatureOp from './cancelSignature.operation';
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
				name: 'Get Latest Published Version ID',
				value: 'getLatestPublishedVersionId',
				description: 'Get the latest published document version ID',
				action: 'Get the latest published document version ID',
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
				name: 'Publish',
				value: 'publish',
				description: 'Publish a draft document, request approval, or publish as minor',
				action: 'Publish a document',
			},
			{
				name: 'Request Signature',
				value: 'requestSignature',
				description: 'Request a signature for a document version',
				action: 'Request a document version signature',
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
	...getLatestPublishedVersionIdOp.description,
	...getAllVersionsOp.description,
	...updateVersionOp.description,
	...deleteDraftVersionOp.description,
	...publishOp.description,
	...voidApprovalOp.description,
	...getSignatureOp.description,
	...getAllSignaturesOp.description,
	...requestSignatureOp.description,
	...cancelSignatureOp.description,
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
	getLatestPublishedVersionIdOp as getLatestPublishedVersionId,
	getAllVersionsOp as getAllVersions,
	updateVersionOp as updateVersion,
	deleteDraftVersionOp as deleteDraftVersion,
	publishOp as publish,
	voidApprovalOp as voidApproval,
	getSignatureOp as getSignature,
	getAllSignaturesOp as getAllSignatures,
	requestSignatureOp as requestSignature,
	cancelSignatureOp as cancelSignature,
	getApprovalQuorumOp as getApprovalQuorum,
	getAllApprovalQuorumsOp as getAllApprovalQuorums,
	getApprovalDecisionOp as getApprovalDecision,
	getAllApprovalDecisionsOp as getAllApprovalDecisions,
};
