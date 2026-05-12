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
import * as getOp from './get.operation';
import * as getAllOp from './getAll.operation';
import * as updateOp from './update.operation';
import * as deleteOp from './delete.operation';
import * as activateOp from './activate.operation';
import * as deactivateOp from './deactivate.operation';
import * as publishOp from './publish.operation';
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
	translateOp as translate,
};
