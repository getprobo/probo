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

export const description: INodeProperties[] = [
	{
		displayName: 'Operation',
		name: 'operation',
		type: 'options',
		noDataExpression: true,
		displayOptions: {
			show: {
				resource: ['rightsRequest'],
			},
		},
		options: [
			{
				name: 'Create',
				value: 'create',
				description: 'Create a new rights request',
				action: 'Create a rights request',
			},
			{
				name: 'Delete',
				value: 'delete',
				description: 'Delete a rights request',
				action: 'Delete a rights request',
			},
			{
				name: 'Get',
				value: 'get',
				description: 'Get a rights request',
				action: 'Get a rights request',
			},
			{
				name: 'Get Many',
				value: 'getAll',
				description: 'Get many rights requests',
				action: 'Get many rights requests',
			},
			{
				name: 'Update',
				value: 'update',
				description: 'Update an existing rights request',
				action: 'Update a rights request',
			},
		],
		default: 'create',
	},
	...createOp.description,
	...updateOp.description,
	...deleteOp.description,
	...getOp.description,
	...getAllOp.description,
];

export { createOp as create, updateOp as update, deleteOp as delete, getOp as get, getAllOp as getAll };
