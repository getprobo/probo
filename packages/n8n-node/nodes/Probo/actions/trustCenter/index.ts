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
import * as getOp from './get.operation';
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

export const description: INodeProperties[] = [
	{
		displayName: 'Operation',
		name: 'operation',
		type: 'options',
		noDataExpression: true,
		displayOptions: {
			show: {
				resource: ['trustCenter'],
			},
		},
		options: [
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
				description: 'Create a new trust center reference',
				action: 'Create a trust center reference',
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
				name: 'Delete File',
				value: 'deleteFile',
				description: 'Delete a trust center file',
				action: 'Delete a trust center file',
			},
			{
				name: 'Delete Reference',
				value: 'deleteReference',
				description: 'Delete a trust center reference',
				action: 'Delete a trust center reference',
			},
			{
				name: 'Get',
				value: 'get',
				description: 'Get trust center settings',
				action: 'Get trust center settings',
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
				description: 'Get many trust center files',
				action: 'Get many trust center files',
			},
			{
				name: 'Get Many References',
				value: 'getAllReferences',
				description: 'Get many trust center references',
				action: 'Get many trust center references',
			},
			{
				name: 'Update',
				value: 'update',
				description: 'Update trust center settings',
				action: 'Update trust center settings',
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
		],
		default: 'get',
	},
	...getOp.description,
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
];

export {
	getOp as get,
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
};
