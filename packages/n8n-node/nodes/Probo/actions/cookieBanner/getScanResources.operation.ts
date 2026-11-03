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

import type { IDataObject, IExecuteFunctions, INodeExecutionData, INodeProperties } from 'n8n-workflow';
import { executionItems, proboApiRequestAllItems } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Scan ID',
		name: 'scanId',
		type: 'string',
		required: true,
		displayOptions: {
			show: {
				resource: ['cookieBanner'],
				operation: ['getScanResources'],
			},
		},
		default: '',
		description: 'Cookie banner scan ID',
	},
	{
		displayName: 'Order',
		name: 'direction',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['cookieBanner'],
				operation: ['getScanResources'],
			},
		},
		options: [
			{ name: 'Newest First', value: 'DESC' },
			{ name: 'Oldest First', value: 'ASC' },
		],
		default: 'DESC',
		description: 'Sort by the time the resource was recorded',
	},
	{
		displayName: 'Return All',
		name: 'returnAll',
		type: 'boolean',
		displayOptions: {
			show: {
				resource: ['cookieBanner'],
				operation: ['getScanResources'],
			},
		},
		default: false,
		description: 'Whether to return all results or only up to a given limit',
	},
	{
		displayName: 'Limit',
		name: 'limit',
		type: 'number',
		displayOptions: {
			show: {
				resource: ['cookieBanner'],
				operation: ['getScanResources'],
				returnAll: [false],
			},
		},
		typeOptions: {
			minValue: 1,
		},
		default: 50,
		description: 'Max number of results to return',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData[]> {
	const scanId = this.getNodeParameter('scanId', itemIndex) as string;
	const direction = this.getNodeParameter('direction', itemIndex, 'DESC') as string;
	const returnAll = this.getNodeParameter('returnAll', itemIndex) as boolean;
	const limit = this.getNodeParameter('limit', itemIndex, 50) as number;

	const query = `
		query GetScanResources(
			$scanId: ID!
			$first: Int
			$after: CursorKey
			$direction: OrderDirection!
		) {
			node(id: $scanId) {
				... on CookieBannerScan {
					trackerResources(
						first: $first
						after: $after
						orderBy: { field: CREATED_AT, direction: $direction }
					) {
						edges {
							node {
								id
								type
								origin
								path
								displayName
								createdAt
							}
						}
						pageInfo {
							hasNextPage
							endCursor
						}
					}
				}
			}
		}
	`;

	const resources = await proboApiRequestAllItems.call(
		this,
		query,
		{ scanId, direction },
		(response) => {
			const data = response?.data as IDataObject | undefined;
			const node = data?.node as IDataObject | undefined;
			return node?.trackerResources as IDataObject | undefined;
		},
		returnAll,
		limit,
	);

	return executionItems(resources, itemIndex);
}
