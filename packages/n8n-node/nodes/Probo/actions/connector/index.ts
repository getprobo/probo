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
import * as createOrgOp from './createOrg.operation';
import * as discoverOp from './discover.operation';
import * as enableOp from './enable.operation';
import * as getOp from './get.operation';
import * as getAllOp from './getAll.operation';

export const description: INodeProperties[] = [
	{
		displayName: 'Operation',
		name: 'operation',
		type: 'options',
		noDataExpression: true,
		displayOptions: {
			show: {
				resource: ['connector'],
			},
		},
		options: [
			{
				name: 'Create Organization Connector',
				value: 'createOrg',
				description: 'Create an organization-scoped workload-identity connector',
				action: 'Create an organization connector',
			},
			{
				name: 'Discover Accounts',
				value: 'discover',
				description: 'Discover accounts under a connector',
				action: 'Discover connector accounts',
			},
			{
				name: 'Enable Accounts',
				value: 'enable',
				description: 'Enable discovered connector accounts',
				action: 'Enable connector accounts',
			},
			{
				name: 'Get',
				value: 'get',
				description: 'Get a connector',
				action: 'Get a connector',
			},
			{
				name: 'Get Many',
				value: 'getAll',
				description: 'Get many connectors',
				action: 'Get many connectors',
			},
		],
		default: 'getAll',
	},
	...createOrgOp.description,
	...discoverOp.description,
	...enableOp.description,
	...getOp.description,
	...getAllOp.description,
];

export {
	createOrgOp as createOrg,
	discoverOp as discover,
	enableOp as enable,
	getOp as get,
	getAllOp as getAll,
};
