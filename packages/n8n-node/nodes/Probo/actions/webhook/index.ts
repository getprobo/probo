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
import * as updateOp from './update.operation';
import * as deleteOp from './delete.operation';
import * as getOp from './get.operation';
import * as getAllOp from './getAll.operation';
import * as getEventsOp from './getEvents.operation';

export const description: INodeProperties[] = [
	{
		displayName: 'Operation',
		name: 'operation',
		type: 'options',
		noDataExpression: true,
		displayOptions: {
			show: {
				resource: ['webhook'],
			},
		},
		options: [
			{
				name: 'Create',
				value: 'create',
				description: 'Create a new webhook subscription',
				action: 'Create a webhook subscription',
			},
			{
				name: 'Delete',
				value: 'delete',
				description: 'Delete a webhook subscription',
				action: 'Delete a webhook subscription',
			},
			{
				name: 'Get',
				value: 'get',
				description: 'Get a webhook subscription',
				action: 'Get a webhook subscription',
			},
			{
				name: 'Get Events',
				value: 'getEvents',
				description: 'Get delivery events for a webhook subscription',
				action: 'Get webhook delivery events',
			},
			{
				name: 'Get Many',
				value: 'getAll',
				description: 'Get many webhook subscriptions',
				action: 'Get many webhook subscriptions',
			},
			{
				name: 'Update',
				value: 'update',
				description: 'Update a webhook subscription',
				action: 'Update a webhook subscription',
			},
		],
		default: 'create',
	},
	...createOp.description,
	...updateOp.description,
	...deleteOp.description,
	...getOp.description,
	...getAllOp.description,
	...getEventsOp.description,
];

export { createOp as create, updateOp as update, deleteOp as delete, getOp as get, getAllOp as getAll, getEventsOp as getEvents };
