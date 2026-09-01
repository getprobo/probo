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
import * as getAllOp from './getAll.operation';
import * as setupAwsOp from './setupAws.operation';

export const description: INodeProperties[] = [
	{
		displayName: 'Operation',
		name: 'operation',
		type: 'options',
		noDataExpression: true,
		displayOptions: {
			show: {
				resource: ['accessReviewSource'],
			},
		},
		options: [
			{
				name: 'Create',
				value: 'create',
				description: 'Create an AWS workload-identity access source',
				action: 'Create an access review source',
			},
			{
				name: 'Get Many',
				value: 'getAll',
				description: 'Get many access review sources',
				action: 'Get many access review sources',
			},
			{
				name: 'Setup AWS',
				value: 'setupAws',
				description: 'Get AWS access source setup values',
				action: 'Get AWS access source setup',
			},
		],
		default: 'getAll',
	},
	...createOp.description,
	...getAllOp.description,
	...setupAwsOp.description,
];

export { createOp as create, getAllOp as getAll, setupAwsOp as setupAws };
