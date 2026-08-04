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
import * as getOp from './get.operation';
import * as getAllOp from './getAll.operation';
import * as getHistoryOp from './getHistory.operation';
import * as transitionOp from './transition.operation';
import * as updateOp from './update.operation';

export const description: INodeProperties[] = [
	{
		displayName: 'Operation',
		name: 'operation',
		type: 'options',
		noDataExpression: true,
		displayOptions: { show: { resource: ['malaysiaPDPABreach'] } },
		options: [
			{
				name: 'Create',
				value: 'create',
				description: 'Record and assess a personal data breach',
				action: 'Create a malaysia pdpa breach incident',
			},
			{
				name: 'Get',
				value: 'get',
				description: 'Get a breach incident and its deadlines',
				action: 'Get a malaysia pdpa breach incident',
			},
			{
				name: 'Get Many',
				value: 'getAll',
				description: 'Get breach incidents for an organization',
				action: 'Get many malaysia pdpa breach incidents',
			},
			{
				name: 'Get Status History',
				value: 'getHistory',
				description: 'Get immutable status history for an incident',
				action: 'Get malaysia pdpa breach status history',
			},
			{
				name: 'Transition Status',
				value: 'transition',
				description: 'Move an incident to an allowed workflow status',
				action: 'Transition a malaysia pdpa breach status',
			},
			{
				name: 'Update',
				value: 'update',
				description: 'Update notification decisions and evidence',
				action: 'Update a malaysia pdpa breach incident',
			},
		],
		default: 'create',
	},
	...createOp.description,
	...getOp.description,
	...getAllOp.description,
	...getHistoryOp.description,
	...transitionOp.description,
	...updateOp.description,
];

export {
	createOp as create,
	getOp as get,
	getAllOp as getAll,
	getHistoryOp as getHistory,
	transitionOp as transition,
	updateOp as update,
};
