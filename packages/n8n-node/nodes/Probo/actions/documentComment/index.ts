// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
import * as deleteOp from './delete.operation';
import * as getOp from './get.operation';
import * as getAllOp from './getAll.operation';
import * as reopenOp from './reopen.operation';
import * as replyOp from './reply.operation';
import * as resolveOp from './resolve.operation';
import * as updateOp from './update.operation';

export const description: INodeProperties[] = [
	{
		displayName: 'Operation',
		name: 'operation',
		type: 'options',
		noDataExpression: true,
		displayOptions: {
			show: {
				resource: ['documentComment'],
			},
		},
		options: [
			{ name: 'Create', value: 'create', description: 'Start a comment thread from a quote or suggestion', action: 'Create a document comment thread' },
			{ name: 'Delete', value: 'delete', description: 'Delete your own comment', action: 'Delete a document comment' },
			{ name: 'Get', value: 'get', description: 'Get a comment thread', action: 'Get a document comment thread' },
			{ name: 'Get Many', value: 'getAll', description: 'List comment threads on a document', action: 'List document comment threads' },
			{ name: 'Reopen', value: 'reopen', description: 'Reopen a comment thread', action: 'Reopen a document comment thread' },
			{ name: 'Reply', value: 'reply', description: 'Reply on a comment thread', action: 'Reply on a document comment thread' },
			{ name: 'Resolve', value: 'resolve', description: 'Resolve a comment thread', action: 'Resolve a document comment thread' },
			{ name: 'Update', value: 'update', description: 'Edit your own comment', action: 'Update a document comment' },
		],
		default: 'getAll',
	},
	...createOp.description,
	...deleteOp.description,
	...getOp.description,
	...getAllOp.description,
	...reopenOp.description,
	...replyOp.description,
	...resolveOp.description,
	...updateOp.description,
];

export {
	createOp as create,
	deleteOp as delete,
	getOp as get,
	getAllOp as getAll,
	reopenOp as reopen,
	replyOp as reply,
	resolveOp as resolve,
	updateOp as update,
};
