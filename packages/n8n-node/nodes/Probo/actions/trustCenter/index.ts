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
import * as getOp from './get.operation';
import * as updateOp from './update.operation';
import * as getAllReferencesOp from './getAllReferences.operation';
import * as createReferenceOp from './createReference.operation';
import * as deleteReferenceOp from './deleteReference.operation';
import * as getAllFilesOp from './getAllFiles.operation';
import * as deleteFileOp from './deleteFile.operation';
import * as createExternalUrlOp from './createExternalUrl.operation';
import * as deleteExternalUrlOp from './deleteExternalUrl.operation';

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
				name: 'Create External URL',
				value: 'createExternalUrl',
				description: 'Create a new compliance external URL',
				action: 'Create a compliance external URL',
			},
			{
				name: 'Create Reference',
				value: 'createReference',
				description: 'Create a new trust center reference',
				action: 'Create a trust center reference',
			},
			{
				name: 'Delete External URL',
				value: 'deleteExternalUrl',
				description: 'Delete a compliance external URL',
				action: 'Delete a compliance external URL',
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
	...createExternalUrlOp.description,
	...deleteExternalUrlOp.description,
];

export {
	getOp as get,
	updateOp as update,
	getAllReferencesOp as getAllReferences,
	createReferenceOp as createReference,
	deleteReferenceOp as deleteReference,
	getAllFilesOp as getAllFiles,
	deleteFileOp as deleteFile,
	createExternalUrlOp as createExternalUrl,
	deleteExternalUrlOp as deleteExternalUrl,
};
