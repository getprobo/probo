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
import * as getOp from './get.operation';
import * as getAllOp from './getAll.operation';
import * as listPrincipalsOp from './listPrincipals.operation';
import * as listResourcesOp from './listResources.operation';

export const description: INodeProperties[] = [
	{
		displayName: 'Operation',
		name: 'operation',
		type: 'options',
		noDataExpression: true,
		displayOptions: {
			show: {
				resource: ['accessInventory'],
			},
		},
		options: [
			{
				name: 'Get',
				value: 'get',
				description: 'Get an access inventory grant',
				action: 'Get an access inventory grant',
			},
			{
				name: 'Get Many',
				value: 'getAll',
				description: 'Get many access inventory grants',
				action: 'Get many access inventory grants',
			},
			{
				name: 'List Principals',
				value: 'listPrincipals',
				description: 'List access inventory principals',
				action: 'List access inventory principals',
			},
			{
				name: 'List Resources',
				value: 'listResources',
				description: 'List access inventory resources',
				action: 'List access inventory resources',
			},
		],
		default: 'getAll',
	},
	...getOp.description,
	...getAllOp.description,
	...listPrincipalsOp.description,
	...listResourcesOp.description,
];

export {
	getOp as get,
	getAllOp as getAll,
	listPrincipalsOp as listPrincipals,
	listResourcesOp as listResources,
};
