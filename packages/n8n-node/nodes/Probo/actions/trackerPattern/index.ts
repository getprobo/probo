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
