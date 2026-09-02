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
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Query',
		name: 'query',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['cookieBanner'],
				operation: ['listGVLCatalog'],
			},
		},
		default: '',
		description: 'Search by vendor name or IAB vendor ID',
	},
	{
		displayName: 'Cookie Banner ID',
		name: 'cookieBannerId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['cookieBanner'],
				operation: ['listGVLCatalog'],
			},
		},
		default: '',
		description: 'Cookie banner ID (required when Membership is set)',
	},
	{
		displayName: 'Membership',
		name: 'membership',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['cookieBanner'],
				operation: ['listGVLCatalog'],
			},
		},
		options: [
			{ name: 'All', value: '' },
			{ name: 'On Banner', value: 'ON_BANNER' },
			{ name: 'Not On Banner', value: 'NOT_ON_BANNER' },
		],
		default: '',
		description: 'Restrict to vendors already on the banner, or not on it',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const query = this.getNodeParameter('query', itemIndex, '') as string;
	const cookieBannerId = this.getNodeParameter('cookieBannerId', itemIndex, '') as string;
	const membership = this.getNodeParameter('membership', itemIndex, '') as string;

	const gql = `
		query ListCommonGVLVendors($first: Int, $filter: CommonGVLVendorFilter) {
			commonGVLVendors(first: $first, filter: $filter) {
				totalCount
				edges {
					node {
						id
						iabVendorId
						name
						policyUrl
					}
				}
			}
		}
	`;

	const variables: IDataObject = { first: 50 };
	const filter: IDataObject = {};
	if (query) {
		filter.query = query;
	}
	if (membership) {
		if (!cookieBannerId) {
			throw new Error("Cookie Banner ID is required when Membership is set");
		}
		filter.cookieBannerId = cookieBannerId;
		filter.membership = membership;
	}
	if (Object.keys(filter).length > 0) {
		variables.filter = filter;
	}

	const responseData = await proboApiRequest.call(this, gql, variables);

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
