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

import type { INodeProperties, IExecuteFunctions, INodeExecutionData, IDataObject } from 'n8n-workflow';
import { proboApiRequestAllItems } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Cookie Banner ID',
		name: 'cookieBannerId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['cookieConsentRecord'],
				operation: ['getAll'],
			},
		},
		default: '',
		description: 'The ID of the cookie banner',
		required: true,
	},
	{
		displayName: 'Return All',
		name: 'returnAll',
		type: 'boolean',
		displayOptions: {
			show: {
				resource: ['cookieConsentRecord'],
				operation: ['getAll'],
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
				resource: ['cookieConsentRecord'],
				operation: ['getAll'],
				returnAll: [false],
			},
		},
		typeOptions: {
			minValue: 1,
		},
		default: 50,
		description: 'Max number of results to return',
	},
	{
		displayName: 'Filter by Action',
		name: 'filterAction',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['cookieConsentRecord'],
				operation: ['getAll'],
			},
		},
		options: [
			{
				name: '(No Filter)',
				value: '',
			},
			{
				name: 'Accept All',
				value: 'ACCEPT_ALL',
			},
			{
				name: 'Customize',
				value: 'CUSTOMIZE',
			},
			{
				name: 'GPC',
				value: 'GPC',
			},
			{
				name: 'Reject All',
				value: 'REJECT_ALL',
			},
		],
		default: '',
		description: 'Filter consent records by action',
	},
	{
		displayName: 'Filter by Visitor ID',
		name: 'filterVisitorId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['cookieConsentRecord'],
				operation: ['getAll'],
			},
		},
		default: '',
		description: 'Filter consent records by visitor ID',
	},
	{
		displayName: 'Filter by Version',
		name: 'filterVersion',
		type: 'number',
		displayOptions: {
			show: {
				resource: ['cookieConsentRecord'],
				operation: ['getAll'],
			},
		},
		default: 0,
		description: 'Filter consent records by banner version number (0 to skip)',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const cookieBannerId = this.getNodeParameter('cookieBannerId', itemIndex) as string;
	const returnAll = this.getNodeParameter('returnAll', itemIndex) as boolean;
	const limit = this.getNodeParameter('limit', itemIndex, 50) as number;
	const filterAction = this.getNodeParameter('filterAction', itemIndex, '') as string;
	const filterVisitorId = this.getNodeParameter('filterVisitorId', itemIndex, '') as string;
	const filterVersion = this.getNodeParameter('filterVersion', itemIndex, 0) as number;

	const hasFilter = filterAction || filterVisitorId || filterVersion;
	const filterClause = hasFilter ? ', $filter: CookieConsentRecordFilter' : '';
	const filterArg = hasFilter ? ', filter: $filter' : '';

	const query = `
		query GetCookieConsentRecords($cookieBannerId: ID!, $first: Int, $after: CursorKey${filterClause}) {
			node(id: $cookieBannerId) {
				... on CookieBanner {
					consentRecords(first: $first, after: $after${filterArg}) {
						edges {
							node {
								id
								visitorId
								ipAddress
								userAgent
								consentData
								action
								sdkVersion
								regulation
								regulationSource
								countryCode
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

	const variables: IDataObject = { cookieBannerId };
	if (hasFilter) {
		const filter: IDataObject = {};
		if (filterAction) filter.action = filterAction;
		if (filterVisitorId) filter.visitorId = filterVisitorId;
		if (filterVersion) filter.version = filterVersion;
		variables.filter = filter;
	}

	const cookieConsentRecords = await proboApiRequestAllItems.call(
		this,
		query,
		variables,
		(response) => {
			const data = response?.data as IDataObject | undefined;
			const node = data?.node as IDataObject | undefined;
			return node?.consentRecords as IDataObject | undefined;
		},
		returnAll,
		limit,
	);

	return {
		json: { cookieConsentRecords },
		pairedItem: { item: itemIndex },
	};
}
