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
import * as deleteOp from './delete.operation';
import * as getOp from './get.operation';
import * as getAllOp from './getAll.operation';
import * as uploadOp from './upload.operation';

export const description: INodeProperties[] = [
	{
		displayName: 'Operation',
		name: 'operation',
		type: 'options',
		noDataExpression: true,
		displayOptions: {
			show: {
				resource: ['evidence'],
			},
		},
		options: [
			{
				name: 'Delete',
				value: 'delete',
				description: 'Delete an evidence',
				action: 'Delete an evidence',
			},
			{
				name: 'Get',
				value: 'get',
				description: 'Get an evidence',
				action: 'Get an evidence',
			},
			{
				name: 'Get Many',
				value: 'getAll',
				description: 'Get many evidences for a measure',
				action: 'Get many evidences',
			},
			{
				name: 'Upload',
				value: 'upload',
				description: 'Upload evidence for a measure',
				action: 'Upload evidence',
			},
		],
		default: 'getAll',
	},
	...deleteOp.description,
	...getOp.description,
	...getAllOp.description,
	...uploadOp.description,
];

export { deleteOp as delete, getOp as get, getAllOp as getAll, uploadOp as upload };
