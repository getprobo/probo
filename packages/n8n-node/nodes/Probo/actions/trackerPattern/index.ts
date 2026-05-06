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
import * as getOp from './get.operation';
import * as getAllOp from './getAll.operation';
import * as updateOp from './update.operation';
import * as deleteOp from './delete.operation';
import * as moveOp from './move.operation';

export const description: INodeProperties[] = [
	{
		displayName: 'Operation',
		name: 'operation',
		type: 'options',
		noDataExpression: true,
		displayOptions: {
			show: {
				resource: ['trackerPattern'],
			},
		},
		options: [
			{
				name: 'Create',
				value: 'create',
				description: 'Create a new tracker pattern',
				action: 'Create a tracker pattern',
			},
			{
				name: 'Delete',
				value: 'delete',
				description: 'Delete a tracker pattern',
				action: 'Delete a tracker pattern',
			},
			{
				name: 'Get',
				value: 'get',
				description: 'Get a tracker pattern',
				action: 'Get a tracker pattern',
			},
			{
				name: 'Get Many',
				value: 'getAll',
				description: 'Get many tracker patterns',
				action: 'Get many tracker patterns',
			},
			{
				name: 'Move',
				value: 'move',
				description: 'Move a tracker pattern to a different category',
				action: 'Move a tracker pattern',
			},
			{
				name: 'Update',
				value: 'update',
				description: 'Update an existing tracker pattern',
				action: 'Update a tracker pattern',
			},
		],
		default: 'create',
	},
	...createOp.description,
	...getOp.description,
	...getAllOp.description,
	...updateOp.description,
	...deleteOp.description,
	...moveOp.description,
];

export {
	createOp as create,
	getOp as get,
	getAllOp as getAll,
	updateOp as update,
	deleteOp as delete,
	moveOp as move,
};
