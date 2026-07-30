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
import * as revokeOp from './revoke.operation';
import * as setOwnerOp from './setOwner.operation';

export const description: INodeProperties[] = [
	{
		displayName: 'Operation',
		name: 'operation',
		type: 'options',
		noDataExpression: true,
		displayOptions: {
			show: {
				resource: ['device'],
			},
		},
		options: [
			{
				name: 'Create',
				value: 'create',
				description: 'Create a PENDING ITAM device and enrollment token',
				action: 'Create a device',
			},
			{
				name: 'Delete',
				value: 'delete',
				description: 'Delete a revoked ITAM device',
				action: 'Delete a device',
			},
			{
				name: 'Get',
				value: 'get',
				description: 'Get an ITAM device',
				action: 'Get a device',
			},
			{
				name: 'Get Many',
				value: 'getAll',
				description: 'Get many ITAM devices',
				action: 'Get many devices',
			},
			{
				name: 'Revoke',
				value: 'revoke',
				description: 'Revoke an ITAM device enrollment',
				action: 'Revoke a device',
			},
			{
				name: 'Set Owner',
				value: 'setOwner',
				description: 'Set or clear the owner of an ITAM device',
				action: 'Set device owner',
			},
		],
		default: 'getAll',
	},
	...createOp.description,
	...deleteOp.description,
	...getOp.description,
	...getAllOp.description,
	...revokeOp.description,
	...setOwnerOp.description,
];

export {
	createOp as create,
	deleteOp as delete,
	getOp as get,
	getAllOp as getAll,
	revokeOp as revoke,
	setOwnerOp as setOwner,
};
