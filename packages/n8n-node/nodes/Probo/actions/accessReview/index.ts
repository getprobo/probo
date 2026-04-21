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
	createOp as create,
	deleteOp as delete,
	getOp as get,
	getAllOp as getAll,
	updateOp as update,
	startOp as start,
	closeOp as close,
	cancelOp as cancel,
};
