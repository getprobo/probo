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
import * as reorderOp from './reorder.operation';

export const description: INodeProperties[] = [
	{
		displayName: 'Operation',
		name: 'operation',
		type: 'options',
		noDataExpression: true,
		displayOptions: {
			show: {
				resource: ['cookieCategory'],
			},
		},
		options: [
			{
				name: 'Create',
				value: 'create',
				description: 'Create a new cookie category',
				action: 'Create a cookie category',
			},
			{
				name: 'Delete',
				value: 'delete',
				description: 'Delete a cookie category',
				action: 'Delete a cookie category',
			},
			{
				name: 'Get',
				value: 'get',
				description: 'Get a cookie category',
				action: 'Get a cookie category',
			},
			{
				name: 'Get Many',
				value: 'getAll',
				description: 'Get many cookie categories',
				action: 'Get many cookie categories',
			},
			{
				name: 'Reorder',
				value: 'reorder',
				description: 'Change the rank of a cookie category',
				action: 'Reorder a cookie category',
			},
			{
				name: 'Update',
				value: 'update',
				description: 'Update an existing cookie category',
				action: 'Update a cookie category',
			},
		],
		default: 'create',
	},
	...createOp.description,
	...getOp.description,
	...getAllOp.description,
	...updateOp.description,
	...deleteOp.description,
	...reorderOp.description,
];

export {
	createOp as create,
	getOp as get,
	getAllOp as getAll,
	updateOp as update,
	deleteOp as delete,
	reorderOp as reorder,
};
