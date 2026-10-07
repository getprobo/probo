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
import * as linkInternalControlOp from './linkInternalControl.operation';
import * as unlinkInternalControlOp from './unlinkInternalControl.operation';
import * as linkDocumentOp from './linkDocument.operation';
import * as unlinkDocumentOp from './unlinkDocument.operation';
import * as linkObligationOp from './linkObligation.operation';
import * as unlinkObligationOp from './unlinkObligation.operation';
import * as publishOp from './publish.operation';

export const description: INodeProperties[] = [
	{
		displayName: 'Operation',
		name: 'operation',
		type: 'options',
		noDataExpression: true,
		displayOptions: {
			show: {
				resource: ['risk'],
			},
		},
		options: [
			{
				name: 'Create',
				value: 'create',
				description: 'Create a new risk',
				action: 'Create a risk',
			},
			{
				name: 'Delete',
				value: 'delete',
				description: 'Delete a risk. Deleting a risk also deletes its history in risk analyses.',
				action: 'Delete a risk',
			},
			{
				name: 'Get',
				value: 'get',
				description: 'Get a risk',
				action: 'Get a risk',
			},
			{
				name: 'Get Many',
				value: 'getAll',
				description: 'Get many risks',
				action: 'Get many risks',
			},
			{
				name: 'Link Document',
				value: 'linkDocument',
				description: 'Link a document to a risk',
				action: 'Link a document to a risk',
			},
			{
				name: 'Link Internal Control',
				value: 'linkInternalControl',
				description: 'Link an internal control to a risk',
				action: 'Link an internal control to a risk',
			},
			{
				name: 'Link Obligation',
				value: 'linkObligation',
				description: 'Link an obligation to a risk',
				action: 'Link an obligation to a risk',
			},
			{
				name: 'Publish List',
				value: 'publish',
				description: 'Publish the risk register as a document version',
				action: 'Publish the risk register',
			},
			{
				name: 'Unlink Document',
				value: 'unlinkDocument',
				description: 'Unlink a document from a risk',
				action: 'Unlink a document from a risk',
			},
			{
				name: 'Unlink Internal Control',
				value: 'unlinkInternalControl',
				description: 'Unlink an internal control from a risk',
				action: 'Unlink an internal control from a risk',
			},
			{
				name: 'Unlink Obligation',
				value: 'unlinkObligation',
				description: 'Unlink an obligation from a risk',
				action: 'Unlink an obligation from a risk',
			},
			{
				name: 'Update',
				value: 'update',
				description: 'Update an existing risk',
				action: 'Update a risk',
			},
		],
		default: 'create',
	},
	...createOp.description,
	...updateOp.description,
	...deleteOp.description,
	...getOp.description,
	...getAllOp.description,
	...linkInternalControlOp.description,
	...unlinkInternalControlOp.description,
	...linkDocumentOp.description,
	...unlinkDocumentOp.description,
	...linkObligationOp.description,
	...unlinkObligationOp.description,
	...publishOp.description,
];

export {
	createOp as create,
	updateOp as update,
	deleteOp as delete,
	getOp as get,
	getAllOp as getAll,
	linkInternalControlOp as linkInternalControl,
	unlinkInternalControlOp as unlinkInternalControl,
	linkDocumentOp as linkDocument,
	unlinkDocumentOp as unlinkDocument,
	linkObligationOp as linkObligation,
	unlinkObligationOp as unlinkObligation,
	publishOp as publish,
};
