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
