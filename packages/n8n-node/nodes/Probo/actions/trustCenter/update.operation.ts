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

import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Trust Center ID',
		name: 'trustCenterId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['trustCenter'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The ID of the trust center to update',
		required: true,
	},
	{
		displayName: 'Active',
		name: 'active',
		type: 'boolean',
		displayOptions: {
			show: {
				resource: ['trustCenter'],
				operation: ['update'],
			},
		},
		default: false,
		description: 'Whether the trust center is active',
	},
	{
		displayName: 'Search Engine Indexing',
		name: 'searchEngineIndexing',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['trustCenter'],
				operation: ['update'],
			},
		},
		options: [
			{
				name: '(Unchanged)',
				value: '',
			},
			{
				name: 'Indexable',
				value: 'INDEXABLE',
			},
			{
				name: 'Not Indexable',
				value: 'NOT_INDEXABLE',
			},
		],
		default: '',
		description: 'Whether search engines should index the trust center',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const trustCenterId = this.getNodeParameter('trustCenterId', itemIndex) as string;
	const active = this.getNodeParameter('active', itemIndex) as boolean | undefined;
	const searchEngineIndexing = this.getNodeParameter('searchEngineIndexing', itemIndex, '') as string;

	const query = `
		mutation UpdateTrustCenter($input: UpdateTrustCenterInput!) {
			updateTrustCenter(input: $input) {
				trustCenter {
					id
					active
					searchEngineIndexing
					createdAt
					updatedAt
				}
			}
		}
	`;

	const input: Record<string, unknown> = { trustCenterId };
	if (active !== undefined) input.active = active;
	if (searchEngineIndexing) input.searchEngineIndexing = searchEngineIndexing;

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
