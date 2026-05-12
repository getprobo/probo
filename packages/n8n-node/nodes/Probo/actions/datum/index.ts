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
import * as publishOp from './publish.operation';

export const description: INodeProperties[] = [
	{
		displayName: 'Operation',
		name: 'operation',
		type: 'options',
		noDataExpression: true,
		displayOptions: {
			show: {
				resource: ['datum'],
			},
		},
		options: [
			{
				name: 'Create',
				value: 'create',
				description: 'Create a new datum',
				action: 'Create a datum',
			},
			{
				name: 'Delete',
				value: 'delete',
				description: 'Delete a datum',
				action: 'Delete a datum',
			},
			{
				name: 'Get',
				value: 'get',
				description: 'Get a datum',
				action: 'Get a datum',
			},
			{
				name: 'Get Many',
				value: 'getAll',
				description: 'Get many data',
				action: 'Get many data',
			},
			{
				name: 'Publish',
				value: 'publish',
				description: 'Publish the data list as a document',
				action: 'Publish the data list',
			},
			{
				name: 'Update',
				value: 'update',
				description: 'Update an existing datum',
				action: 'Update a datum',
			},
		],
		default: 'create',
	},
	...createOp.description,
	...updateOp.description,
	...deleteOp.description,
	...getOp.description,
	...getAllOp.description,
	...publishOp.description,
];

export { createOp as create, updateOp as update, deleteOp as delete, getOp as get, getAllOp as getAll, publishOp as publish };
