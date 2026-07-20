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
import * as addSourceOp from './addSource.operation';
import * as createOp from './create.operation';
import * as deleteOp from './delete.operation';
import * as getOp from './get.operation';
import * as getAllOp from './getAll.operation';
import * as updateOp from './update.operation';
import * as startOp from './start.operation';
import * as closeOp from './close.operation';
import * as cancelOp from './cancel.operation';

export const description: INodeProperties[] = [
	{
		displayName: 'Operation',
		name: 'operation',
		type: 'options',
		noDataExpression: true,
		displayOptions: {
			show: {
				resource: ['accessReview'],
			},
		},
		options: [
			{
				name: 'Add Source',
				value: 'addSource',
				description: 'Add a scope source to an access review campaign',
				action: 'Add a scope source to an access review campaign',
			},
			{
				name: 'Cancel',
				value: 'cancel',
				description: 'Cancel an access review campaign',
				action: 'Cancel an access review campaign',
			},
			{
				name: 'Close',
				value: 'close',
				description: 'Close an access review campaign',
				action: 'Close an access review campaign',
			},
			{
				name: 'Create',
				value: 'create',
				description: 'Create a new access review campaign',
				action: 'Create an access review campaign',
			},
			{
				name: 'Delete',
				value: 'delete',
				description: 'Delete an access review campaign',
				action: 'Delete an access review campaign',
			},
			{
				name: 'Get',
				value: 'get',
				description: 'Get an access review campaign',
				action: 'Get an access review campaign',
			},
			{
				name: 'Get Many',
				value: 'getAll',
				description: 'Get many access review campaigns',
				action: 'Get many access review campaigns',
			},
			{
				name: 'Start',
				value: 'start',
				description: 'Start an access review campaign',
				action: 'Start an access review campaign',
			},
			{
				name: 'Update',
				value: 'update',
				description: 'Update an existing access review campaign',
				action: 'Update an access review campaign',
			},
		],
		default: 'create',
	},
	...addSourceOp.description,
	...createOp.description,
	...deleteOp.description,
	...getOp.description,
	...getAllOp.description,
	...updateOp.description,
	...startOp.description,
	...closeOp.description,
	...cancelOp.description,
];

export {
	addSourceOp as addSource,
	createOp as create,
	deleteOp as delete,
	getOp as get,
	getAllOp as getAll,
	updateOp as update,
	startOp as start,
	closeOp as close,
	cancelOp as cancel,
};
