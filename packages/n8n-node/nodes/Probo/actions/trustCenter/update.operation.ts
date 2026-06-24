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
