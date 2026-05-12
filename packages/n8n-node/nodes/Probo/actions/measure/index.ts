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
import * as updateOp from './update.operation';
import * as deleteOp from './delete.operation';
import * as getOp from './get.operation';
import * as getAllOp from './getAll.operation';
import * as linkDocumentOp from './linkDocument.operation';
import * as unlinkDocumentOp from './unlinkDocument.operation';

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
				name: 'Unlink Document',
				value: 'unlinkDocument',
				description: 'Unlink a document from a measure',
				action: 'Unlink a document from a measure',
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
];

export {
	createOp as create,
	updateOp as update,
	deleteOp as delete,
	getOp as get,
	getAllOp as getAll,
	linkDocumentOp as linkDocument,
	unlinkDocumentOp as unlinkDocument,
};
