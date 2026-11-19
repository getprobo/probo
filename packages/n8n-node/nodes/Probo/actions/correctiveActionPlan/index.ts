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
import * as linkInternalControlOp from './linkInternalControl.operation';
import * as unlinkInternalControlOp from './unlinkInternalControl.operation';
import * as updateOp from './update.operation';

export const description: INodeProperties[] = [
	{
		displayName: 'Operation',
		name: 'operation',
		type: 'options',
		noDataExpression: true,
		displayOptions: {
			show: {
				resource: ['correctiveActionPlan'],
			},
		},
		options: [
			{
				name: 'Create',
				value: 'create',
				description: 'Create a corrective action plan for a finding',
				action: 'Create a corrective action plan',
			},
			{
				name: 'Delete',
				value: 'delete',
				description: 'Delete a corrective action plan',
				action: 'Delete a corrective action plan',
			},
			{
				name: 'Get',
				value: 'get',
				description: 'Get a corrective action plan',
				action: 'Get a corrective action plan',
			},
			{
				name: 'Get Many',
				value: 'getAll',
				description: 'Get many corrective action plans of a finding',
				action: 'Get many corrective action plans',
			},
			{
				name: 'Link Internal Control',
				value: 'linkInternalControl',
				description: 'Link an internal control to a corrective action plan',
				action: 'Link an internal control to a corrective action plan',
			},
			{
				name: 'Unlink Internal Control',
				value: 'unlinkInternalControl',
				description: 'Unlink an internal control from a corrective action plan',
				action: 'Unlink an internal control from a corrective action plan',
			},
			{
				name: 'Update',
				value: 'update',
				description: 'Update a corrective action plan',
				action: 'Update a corrective action plan',
			},
		],
		default: 'create',
	},
	...createOp.description,
	...deleteOp.description,
	...getOp.description,
	...getAllOp.description,
	...linkInternalControlOp.description,
	...unlinkInternalControlOp.description,
	...updateOp.description,
];

export {
	createOp as create,
	deleteOp as delete,
	getOp as get,
	getAllOp as getAll,
	linkInternalControlOp as linkInternalControl,
	unlinkInternalControlOp as unlinkInternalControl,
	updateOp as update,
};
