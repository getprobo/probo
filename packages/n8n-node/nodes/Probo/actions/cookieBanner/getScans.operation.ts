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
import { cookieBannerLocator, organizationLocator } from '../../locators';

const scanStates = [
	{ name: 'Any', value: '' },
	{ name: 'Pending', value: 'PENDING' },
	{ name: 'Running', value: 'RUNNING' },
	{ name: 'Paused', value: 'PAUSED' },
	{ name: 'Completed', value: 'COMPLETED' },
	{ name: 'Failed', value: 'FAILED' },
];

export const description: INodeProperties[] = [
	organizationLocator({
		resource: ['cookieBanner'],
		operation: ['getScans'],
		required: false,
		description: 'Organization to search in. Leave this empty when entering an ID.',
	}),
	cookieBannerLocator({
		resource: ['cookieBanner'],
		operation: ['getScans'],
	}),
	{
		displayName: 'State',
		name: 'state',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['cookieBanner'],
				operation: ['getScans'],
			},
		},
		options: scanStates,
		default: '',
		description: 'Only scans in this state. Any lists every state.',
	},
	{
		displayName: 'Order',
		name: 'direction',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['cookieBanner'],
				operation: ['getScans'],
			},
		},
		options: [
			{ name: 'Newest First', value: 'DESC' },
			{ name: 'Oldest First', value: 'ASC' },
		],
		default: 'DESC',
		description: 'Sort by the time the scan was requested',
	},
	{
		displayName: 'Return All',
		name: 'returnAll',
		type: 'boolean',
		displayOptions: {
			show: {
				resource: ['cookieBanner'],
				operation: ['getScans'],
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
				operation: ['getScans'],
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
	const cookieBannerId = this.getNodeParameter('cookieBannerId', itemIndex, '', { extractValue: true }) as string;
	const state = this.getNodeParameter('state', itemIndex, '') as string;
	const direction = this.getNodeParameter('direction', itemIndex, 'DESC') as string;
	const returnAll = this.getNodeParameter('returnAll', itemIndex) as boolean;
	const limit = this.getNodeParameter('limit', itemIndex, 50) as number;

	const query = `
		query GetCookieBannerScans(
			$cookieBannerId: ID!
			$first: Int
			$after: CursorKey
			$direction: OrderDirection!
			$state: CookieBannerScanState
		) {
			node(id: $cookieBannerId) {
				... on CookieBanner {
					scans(
						first: $first
						after: $after
						orderBy: { field: CREATED_AT, direction: $direction }
						filter: { state: $state }
					) {
						edges {
							node {
								id
								state
								failureReason
								truncated
								pagesScanned
								cookiesFound
								storageFound
								resourcesFound
								startedAt
								completedAt
								createdAt
								updatedAt
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

	const scans = await proboApiRequestAllItems.call(
		this,
		query,
		{
			cookieBannerId,
			direction,
			state: state === '' ? null : state,
		},
		(response) => {
			const data = response?.data as IDataObject | undefined;
			const node = data?.node as IDataObject | undefined;
			return node?.scans as IDataObject | undefined;
		},
		returnAll,
		limit,
	);

	return executionItems(scans, itemIndex);
}
