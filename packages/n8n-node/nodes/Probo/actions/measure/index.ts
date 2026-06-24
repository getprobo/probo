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
import * as updateOp from './update.operation';
import * as deleteOp from './delete.operation';
import * as getOp from './get.operation';
import * as getAllOp from './getAll.operation';
import * as linkDocumentOp from './linkDocument.operation';
import * as linkThirdPartyOp from './linkThirdParty.operation';
import * as unlinkDocumentOp from './unlinkDocument.operation';
import * as unlinkThirdPartyOp from './unlinkThirdParty.operation';

export const description: INodeProperties[] = [
	{
		displayName: 'Operation',
		name: 'operation',
		type: 'options',
		noDataExpression: true,
		displayOptions: {
			show: {
				resource: ['measure'],
			},
		},
		options: [
			{
				name: 'Create',
				value: 'create',
				description: 'Create a new measure',
				action: 'Create a measure',
			},
			{
				name: 'Delete',
				value: 'delete',
				description: 'Delete a measure',
				action: 'Delete a measure',
			},
			{
				name: 'Get',
				value: 'get',
				description: 'Get a measure',
				action: 'Get a measure',
			},
			{
				name: 'Get Many',
				value: 'getAll',
				description: 'Get many measures',
				action: 'Get many measures',
			},
			{
				name: 'Link Document',
				value: 'linkDocument',
				description: 'Link a document to a measure',
				action: 'Link a document to a measure',
			},
			{
				name: 'Link Third Party',
				value: 'linkThirdParty',
				description: 'Link a third party to a measure',
				action: 'Link a third party to a measure',
			},
			{
				name: 'Unlink Document',
				value: 'unlinkDocument',
				description: 'Unlink a document from a measure',
				action: 'Unlink a document from a measure',
			},
			{
				name: 'Unlink Third Party',
				value: 'unlinkThirdParty',
				description: 'Unlink a third party from a measure',
				action: 'Unlink a third party from a measure',
			},
			{
				name: 'Update',
				value: 'update',
				description: 'Update an existing measure',
				action: 'Update a measure',
			},
		],
		default: 'create',
	},
	...createOp.description,
	...updateOp.description,
	...deleteOp.description,
	...getOp.description,
	...getAllOp.description,
	...linkDocumentOp.description,
	...unlinkDocumentOp.description,
	...linkThirdPartyOp.description,
	...unlinkThirdPartyOp.description,
];

export {
	createOp as create,
	updateOp as update,
	deleteOp as delete,
	getOp as get,
	getAllOp as getAll,
	linkDocumentOp as linkDocument,
	unlinkDocumentOp as unlinkDocument,
	linkThirdPartyOp as linkThirdParty,
	unlinkThirdPartyOp as unlinkThirdParty,
};
