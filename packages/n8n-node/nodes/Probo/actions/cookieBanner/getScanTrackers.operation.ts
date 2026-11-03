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

const trackerTypes = [
	{ name: 'Cookie', value: 'COOKIE' },
	{ name: 'Local Storage', value: 'LOCAL_STORAGE' },
	{ name: 'Session Storage', value: 'SESSION_STORAGE' },
	{ name: 'IndexedDB', value: 'INDEXED_DB' },
	{ name: 'Cache Storage', value: 'CACHE_STORAGE' },
];

export const description: INodeProperties[] = [
	{
		displayName: 'Scan ID',
		name: 'scanId',
		type: 'string',
		required: true,
		displayOptions: {
			show: {
				resource: ['cookieBanner'],
				operation: ['getScanTrackers'],
			},
		},
		default: '',
		description: 'Cookie banner scan ID',
	},
	{
		displayName: 'Tracker Types',
		name: 'trackerTypes',
		type: 'multiOptions',
		displayOptions: {
			show: {
				resource: ['cookieBanner'],
				operation: ['getScanTrackers'],
			},
		},
		options: trackerTypes,
		default: [],
		description: 'Only these tracker types. Leave empty for every scanner detection on the scan.',
	},
	{
		displayName: 'Order',
		name: 'direction',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['cookieBanner'],
				operation: ['getScanTrackers'],
			},
		},
		options: [
			{ name: 'Newest First', value: 'DESC' },
			{ name: 'Oldest First', value: 'ASC' },
		],
		default: 'DESC',
		description: 'Sort by the time the detection was first seen',
	},
	{
		displayName: 'Return All',
		name: 'returnAll',
		type: 'boolean',
		displayOptions: {
			show: {
				resource: ['cookieBanner'],
				operation: ['getScanTrackers'],
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
				operation: ['getScanTrackers'],
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
	const trackerTypes = this.getNodeParameter('trackerTypes', itemIndex, []) as string[];
	const direction = this.getNodeParameter('direction', itemIndex, 'DESC') as string;
	const returnAll = this.getNodeParameter('returnAll', itemIndex) as boolean;
	const limit = this.getNodeParameter('limit', itemIndex, 50) as number;

	const query = `
		query GetScanTrackers(
			$scanId: ID!
			$first: Int
			$after: CursorKey
			$direction: OrderDirection!
			$trackerTypes: [TrackerType!]
		) {
			node(id: $scanId) {
				... on CookieBannerScan {
					detectedTrackers(
						first: $first
						after: $after
						orderBy: { field: CREATED_AT, direction: $direction }
						filter: { trackerTypes: $trackerTypes }
					) {
						edges {
							node {
								id
								identifier
								observedBy
								initiatorUrl
								cookieDomain
								source
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

	const trackers = await proboApiRequestAllItems.call(
		this,
		query,
		{
			scanId,
			direction,
			trackerTypes: trackerTypes.length === 0 ? null : trackerTypes,
		},
		(response) => {
			const data = response?.data as IDataObject | undefined;
			const node = data?.node as IDataObject | undefined;
			return node?.detectedTrackers as IDataObject | undefined;
		},
		returnAll,
		limit,
	);

	return executionItems(trackers, itemIndex);
}
