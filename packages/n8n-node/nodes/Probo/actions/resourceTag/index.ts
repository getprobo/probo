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
import * as attachOp from './attach.operation';
import * as createOp from './create.operation';
import * as deleteOp from './delete.operation';
import * as detachOp from './detach.operation';
import * as listOp from './list.operation';
import * as updateOp from './update.operation';

export const description: INodeProperties[] = [
	{
		displayName: 'Operation',
		name: 'operation',
		type: 'options',
		noDataExpression: true,
		displayOptions: {
			show: {
				resource: ['resourceTag'],
			},
		},
		options: [
			{
				name: 'Attach',
				value: 'attach',
				description: 'Attach a resource tag to a resource',
				action: 'Attach a resource tag',
			},
			{
				name: 'Create',
				value: 'create',
				description: 'Create a new resource tag',
				action: 'Create a resource tag',
			},
			{
				name: 'Delete',
				value: 'delete',
				description: 'Delete a resource tag',
				action: 'Delete a resource tag',
			},
			{
				name: 'Detach',
				value: 'detach',
				description: 'Detach a resource tag from a resource',
				action: 'Detach a resource tag',
			},
			{
				name: 'List',
				value: 'list',
				description: 'List resource tags in an organization',
				action: 'List resource tags',
			},
			{
				name: 'Update',
				value: 'update',
				description: 'Update a resource tag',
				action: 'Update a resource tag',
			},
		],
		default: 'create',
	},
	...attachOp.description,
	...createOp.description,
	...deleteOp.description,
	...detachOp.description,
	...listOp.description,
	...updateOp.description,
];

export {
	attachOp as attach,
	createOp as create,
	deleteOp as delete,
	detachOp as detach,
	listOp as list,
	updateOp as update,
};
