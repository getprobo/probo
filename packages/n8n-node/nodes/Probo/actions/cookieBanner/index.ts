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
import * as createOp from './create.operation';
import * as getOp from './get.operation';
import * as getAllOp from './getAll.operation';
import * as updateOp from './update.operation';
import * as deleteOp from './delete.operation';
import * as activateOp from './activate.operation';
import * as deactivateOp from './deactivate.operation';
import * as publishOp from './publish.operation';
import * as regeneratePolicyOp from './regeneratePolicy.operation';
import * as translateOp from './translate.operation';

export const description: INodeProperties[] = [
	{
		displayName: 'Operation',
		name: 'operation',
		type: 'options',
		noDataExpression: true,
		displayOptions: {
			show: {
				resource: ['cookieBanner'],
			},
		},
		options: [
			{
				name: 'Activate',
				value: 'activate',
				description: 'Activate a cookie banner',
				action: 'Activate a cookie banner',
			},
			{
				name: 'Create',
				value: 'create',
				description: 'Create a new cookie banner',
				action: 'Create a cookie banner',
			},
			{
				name: 'Deactivate',
				value: 'deactivate',
				description: 'Deactivate a cookie banner',
				action: 'Deactivate a cookie banner',
			},
			{
				name: 'Delete',
				value: 'delete',
				description: 'Delete a cookie banner',
				action: 'Delete a cookie banner',
			},
			{
				name: 'Get',
				value: 'get',
				description: 'Get a cookie banner',
				action: 'Get a cookie banner',
			},
			{
				name: 'Get Many',
				value: 'getAll',
				description: 'Get many cookie banners',
				action: 'Get many cookie banners',
			},
			{
				name: 'Publish',
				value: 'publish',
				description: 'Publish a new cookie banner version',
				action: 'Publish a cookie banner version',
			},
			{
				name: 'Regenerate Policy',
				value: 'regeneratePolicy',
				description: 'Re-arm tracker policy generation for a published cookie banner',
				action: 'Regenerate a cookie banner tracker policy',
			},
			{
				name: 'Translate',
				value: 'translate',
				description: 'Upsert a cookie banner translation',
				action: 'Translate a cookie banner',
			},
			{
				name: 'Update',
				value: 'update',
				description: 'Update an existing cookie banner',
				action: 'Update a cookie banner',
			},
		],
		default: 'create',
	},
	...createOp.description,
	...getOp.description,
	...getAllOp.description,
	...updateOp.description,
	...deleteOp.description,
	...activateOp.description,
	...deactivateOp.description,
	...publishOp.description,
	...regeneratePolicyOp.description,
	...translateOp.description,
];

export {
	createOp as create,
	getOp as get,
	getAllOp as getAll,
	updateOp as update,
	deleteOp as delete,
	activateOp as activate,
	deactivateOp as deactivate,
	publishOp as publish,
	regeneratePolicyOp as regeneratePolicy,
	translateOp as translate,
};
