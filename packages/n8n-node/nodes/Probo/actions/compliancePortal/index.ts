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
import * as deleteOp from './delete.operation';
import * as getOp from './get.operation';
import * as getAllOp from './getAll.operation';
import * as updateOp from './update.operation';
import * as getAllReferencesOp from './getAllReferences.operation';
import * as createReferenceOp from './createReference.operation';
import * as deleteReferenceOp from './deleteReference.operation';
import * as getAllFilesOp from './getAllFiles.operation';
import * as deleteFileOp from './deleteFile.operation';
import * as createCustomLinkOp from './createCustomLink.operation';
import * as deleteCustomLinkOp from './deleteCustomLink.operation';
import * as getAllCommitmentGroupsOp from './getAllCommitmentGroups.operation';
import * as createCommitmentGroupOp from './createCommitmentGroup.operation';
import * as updateCommitmentGroupOp from './updateCommitmentGroup.operation';
import * as deleteCommitmentGroupOp from './deleteCommitmentGroup.operation';
import * as getAllCommitmentsOp from './getAllCommitments.operation';
import * as createCommitmentOp from './createCommitment.operation';
import * as updateCommitmentOp from './updateCommitment.operation';
import * as deleteCommitmentOp from './deleteCommitment.operation';
import * as updateDocumentVisibilityOp from './updateDocumentVisibility.operation';
import * as updateAuditVisibilityOp from './updateAuditVisibility.operation';
import * as updateThirdPartyPublishedOp from './updateThirdPartyPublished.operation';
import * as deleteDocumentOp from './deleteDocument.operation';
import * as deleteAuditOp from './deleteAudit.operation';
import * as deleteThirdPartyOp from './deleteThirdParty.operation';

export const description: INodeProperties[] = [
	{
		displayName: 'Operation',
		name: 'operation',
		type: 'options',
		noDataExpression: true,
		displayOptions: {
			show: {
				resource: ['compliancePortal'],
			},
		},
		options: [
			{
				name: 'Create',
				value: 'create',
				description: 'Create a new compliance portal',
				action: 'Create a compliance portal',
			},
			{
				name: 'Create Commitment',
				value: 'createCommitment',
				description: 'Create a new compliance portal commitment',
				action: 'Create a compliance portal commitment',
			},
			{
				name: 'Create Commitment Group',
				value: 'createCommitmentGroup',
				description: 'Create a new compliance portal commitment group',
				action: 'Create a compliance portal commitment group',
			},
			{
				name: 'Create Custom Link',
				value: 'createCustomLink',
				description: 'Create a new compliance custom link',
				action: 'Create a compliance custom link',
			},
			{
				name: 'Create Reference',
				value: 'createReference',
				description: 'Create a new compliance portal reference',
				action: 'Create a compliance portal reference',
			},
			{
				name: 'Delete',
				value: 'delete',
				description: 'Delete a compliance portal',
				action: 'Delete a compliance portal',
			},
			{
				name: 'Delete Audit',
				value: 'deleteAudit',
				description: 'Remove an audit from a compliance portal',
				action: 'Remove an audit from a compliance portal',
			},
			{
				name: 'Delete Commitment',
				value: 'deleteCommitment',
				description: 'Delete a compliance portal commitment',
				action: 'Delete a compliance portal commitment',
			},
			{
				name: 'Delete Commitment Group',
				value: 'deleteCommitmentGroup',
				description: 'Delete a compliance portal commitment group',
				action: 'Delete a compliance portal commitment group',
			},
			{
				name: 'Delete Custom Link',
				value: 'deleteCustomLink',
				description: 'Delete a compliance custom link',
				action: 'Delete a compliance custom link',
			},
			{
				name: 'Delete Document',
				value: 'deleteDocument',
				description: 'Remove a document from a compliance portal',
				action: 'Remove a document from a compliance portal',
			},
			{
				name: 'Delete File',
				value: 'deleteFile',
				description: 'Delete a compliance portal file',
				action: 'Delete a compliance portal file',
			},
			{
				name: 'Delete Reference',
				value: 'deleteReference',
				description: 'Delete a compliance portal reference',
				action: 'Delete a compliance portal reference',
			},
			{
				name: 'Delete Third Party',
				value: 'deleteThirdParty',
				description: 'Remove a third party from a compliance portal',
				action: 'Remove a third party from a compliance portal',
			},
			{
				name: 'Get',
				value: 'get',
				description: 'Get compliance portal settings',
				action: 'Get compliance portal settings',
			},
			{
				name: 'Get Many',
				value: 'getAll',
				description: 'Get many compliance portals',
				action: 'Get many compliance portals',
			},
			{
				name: 'Get Many Commitment Groups',
				value: 'getAllCommitmentGroups',
				description: 'Get many compliance portal commitment groups',
				action: 'Get many compliance portal commitment groups',
			},
			{
				name: 'Get Many Commitments',
				value: 'getAllCommitments',
				description: 'Get many compliance portal commitments',
				action: 'Get many compliance portal commitments',
			},
			{
				name: 'Get Many Files',
				value: 'getAllFiles',
				description: 'Get many compliance portal files',
				action: 'Get many compliance portal files',
			},
			{
				name: 'Get Many References',
				value: 'getAllReferences',
				description: 'Get many compliance portal references',
				action: 'Get many compliance portal references',
			},
			{
				name: 'Update',
				value: 'update',
				description: 'Update compliance portal settings',
				action: 'Update compliance portal settings',
			},
			{
				name: 'Update Audit Visibility',
				value: 'updateAuditVisibility',
				description: 'Update the visibility of an audit on a compliance portal',
				action: 'Update an audit visibility on a compliance portal',
			},
			{
				name: 'Update Commitment',
				value: 'updateCommitment',
				description: 'Update a compliance portal commitment',
				action: 'Update a compliance portal commitment',
			},
			{
				name: 'Update Commitment Group',
				value: 'updateCommitmentGroup',
				description: 'Update a compliance portal commitment group',
				action: 'Update a compliance portal commitment group',
			},
			{
				name: 'Update Document Visibility',
				value: 'updateDocumentVisibility',
				description: 'Update the visibility of a document on a compliance portal',
				action: 'Update a document visibility on a compliance portal',
			},
			{
				name: 'Update Third Party Published',
				value: 'updateThirdPartyPublished',
				description: 'Publish or unpublish a third party on a compliance portal',
				action: 'Update a third party published state on a compliance portal',
			},
		],
		default: 'get',
	},
	...createOp.description,
	...deleteOp.description,
	...getOp.description,
	...getAllOp.description,
	...updateOp.description,
	...getAllReferencesOp.description,
	...createReferenceOp.description,
	...deleteReferenceOp.description,
	...getAllFilesOp.description,
	...deleteFileOp.description,
	...createCustomLinkOp.description,
	...deleteCustomLinkOp.description,
	...getAllCommitmentGroupsOp.description,
	...createCommitmentGroupOp.description,
	...updateCommitmentGroupOp.description,
	...deleteCommitmentGroupOp.description,
	...getAllCommitmentsOp.description,
	...createCommitmentOp.description,
	...updateCommitmentOp.description,
	...deleteCommitmentOp.description,
	...updateDocumentVisibilityOp.description,
	...updateAuditVisibilityOp.description,
	...updateThirdPartyPublishedOp.description,
	...deleteDocumentOp.description,
	...deleteAuditOp.description,
	...deleteThirdPartyOp.description,
];

export {
	createOp as create,
	deleteOp as delete,
	getOp as get,
	getAllOp as getAll,
	updateOp as update,
	getAllReferencesOp as getAllReferences,
	createReferenceOp as createReference,
	deleteReferenceOp as deleteReference,
	getAllFilesOp as getAllFiles,
	deleteFileOp as deleteFile,
	createCustomLinkOp as createCustomLink,
	deleteCustomLinkOp as deleteCustomLink,
	getAllCommitmentGroupsOp as getAllCommitmentGroups,
	createCommitmentGroupOp as createCommitmentGroup,
	updateCommitmentGroupOp as updateCommitmentGroup,
	deleteCommitmentGroupOp as deleteCommitmentGroup,
	getAllCommitmentsOp as getAllCommitments,
	createCommitmentOp as createCommitment,
	updateCommitmentOp as updateCommitment,
	deleteCommitmentOp as deleteCommitment,
	updateDocumentVisibilityOp as updateDocumentVisibility,
	updateAuditVisibilityOp as updateAuditVisibility,
	updateThirdPartyPublishedOp as updateThirdPartyPublished,
	deleteDocumentOp as deleteDocument,
	deleteAuditOp as deleteAudit,
	deleteThirdPartyOp as deleteThirdParty,
};
