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
		displayName: 'Organization ID',
		name: 'organizationId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['asset'],
				operation: ['getAll'],
			},
		},
		default: '',
		description: 'The ID of the organization',
		required: true,
	},
	{
		displayName: 'Return All',
		name: 'returnAll',
		type: 'boolean',
		displayOptions: {
			show: {
				resource: ['asset'],
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
				resource: ['asset'],
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
		displayName: 'Options',
		name: 'options',
		type: 'collection',
		placeholder: 'Add Option',
		default: {},
		displayOptions: {
			show: {
				resource: ['asset'],
				operation: ['getAll'],
			},
		},
		options: [
			{
				displayName: 'Include Owner',
				name: 'includeOwner',
				type: 'boolean',
				default: false,
				description: 'Whether to include owner details in the response',
			},
			{
				displayName: 'Include ThirdParties',
				name: 'includeThirdParties',
				type: 'boolean',
				default: false,
				description: 'Whether to include thirdParties in the response',
			},
		],
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const organizationId = this.getNodeParameter('organizationId', itemIndex) as string;
	const returnAll = this.getNodeParameter('returnAll', itemIndex) as boolean;
	const limit = this.getNodeParameter('limit', itemIndex, 50) as number;
	const options = this.getNodeParameter('options', itemIndex, {}) as {
		includeOwner?: boolean;
		includeThirdParties?: boolean;
	};

	const ownerFragment = options.includeOwner
		? `owner {
			id
			fullName
			emailAddress
		}`
		: '';

	const thirdPartiesFragment = options.includeThirdParties
		? `thirdParties(first: 100) {
			edges {
				node {
					id
					name
				}
			}
		}`
		: '';

	const query = `
		query GetAssets($organizationId: ID!, $first: Int, $after: CursorKey) {
			node(id: $organizationId) {
				... on Organization {
					assets(first: $first, after: $after) {
						edges {
							node {
								id
								name
								amount
								assetType
								dataTypesStored
								${ownerFragment}
								${thirdPartiesFragment}
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

	const assets = await proboApiRequestAllItems.call(
		this,
		query,
		{ organizationId },
		(response) => {
			const data = response?.data as IDataObject | undefined;
			const node = data?.node as IDataObject | undefined;
			return node?.assets as IDataObject | undefined;
		},
		returnAll,
		limit,
	);

	return {
		json: { assets },
		pairedItem: { item: itemIndex },
	};
}
