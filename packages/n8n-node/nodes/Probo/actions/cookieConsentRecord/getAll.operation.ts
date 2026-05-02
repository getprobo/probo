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
