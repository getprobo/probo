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
import * as acceptOp from './accept.operation';
import * as getOp from './get.operation';
import * as getAllOp from './getAll.operation';
import * as rejectOp from './reject.operation';
import * as replyOp from './reply.operation';
import * as startOp from './start.operation';

export const description: INodeProperties[] = [
	{
		displayName: 'Operation',
		name: 'operation',
		type: 'options',
		noDataExpression: true,
		displayOptions: { show: { resource: ['documentSuggestion'] } },
		options: [
			{ name: 'Accept', value: 'accept', description: 'Accept a redline', action: 'Accept a document suggestion' },
			{ name: 'Get', value: 'get', description: 'Get a suggestion', action: 'Get a document suggestion' },
			{ name: 'Get Many', value: 'getAll', description: 'List suggestions on a document', action: 'List document suggestions' },
			{ name: 'Reject', value: 'reject', description: 'Reject a redline', action: 'Reject a document suggestion' },
			{ name: 'Reply', value: 'reply', description: 'Comment on a redline', action: 'Reply on a document suggestion' },
			{ name: 'Start', value: 'start', description: 'Allocate a suggestion ID', action: 'Start a document suggestion' },
		],
		default: 'getAll',
	},
	...acceptOp.description,
	...getOp.description,
	...getAllOp.description,
	...rejectOp.description,
	...replyOp.description,
	...startOp.description,
];

export {
	acceptOp as accept,
	getOp as get,
	getAllOp as getAll,
	rejectOp as reject,
	replyOp as reply,
	startOp as start,
};
