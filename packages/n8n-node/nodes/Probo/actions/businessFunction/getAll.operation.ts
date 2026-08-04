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

import type { INodeProperties, IExecuteFunctions, INodeExecutionData, IDataObject } from 'n8n-workflow';
import { proboApiRequestAllItems } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Organization ID',
		name: 'organizationId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['businessFunction'],
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
				resource: ['businessFunction'],
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
				resource: ['businessFunction'],
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
		displayName: 'Filters',
		name: 'filters',
		type: 'collection',
		placeholder: 'Add Filter',
		default: {},
		displayOptions: {
			show: {
				resource: ['businessFunction'],
				operation: ['getAll'],
			},
		},
		options: [
			{
				displayName: 'Classification',
				name: 'classification',
				type: 'options',
				options: [
					{ name: 'All', value: '' },
					{ name: 'Critical', value: 'CRITICAL' },
					{ name: 'Important', value: 'IMPORTANT' },
					{ name: 'Secondary', value: 'SECONDARY' },
					{ name: 'Standard', value: 'STANDARD' },
				],
				default: '',
				description: 'Filter by classification',
			},
		],
	},
	{
		displayName: 'Options',
		name: 'options',
		type: 'collection',
		placeholder: 'Add Option',
		default: {},
		displayOptions: {
			show: {
				resource: ['businessFunction'],
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
				displayName: 'Include Assets',
				name: 'includeAssets',
				type: 'boolean',
				default: false,
				description: 'Whether to include asset dependencies in the response',
			},
			{
				displayName: 'Include Third Parties',
				name: 'includeThirdParties',
				type: 'boolean',
				default: false,
				description: 'Whether to include third-party dependencies in the response',
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
	const filters = this.getNodeParameter('filters', itemIndex, {}) as {
		classification?: string;
	};
	const options = this.getNodeParameter('options', itemIndex, {}) as {
		includeOwner?: boolean;
		includeAssets?: boolean;
		includeThirdParties?: boolean;
	};

	const ownerFragment = options.includeOwner
		? `owner {
			id
			fullName
			emailAddress
		}`
		: '';

	const assetsFragment = options.includeAssets
		? `assets(first: 100) {
			edges {
				node {
					id
					name
				}
			}
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
		query GetBusinessFunctions($organizationId: ID!, $first: Int, $after: CursorKey, $filter: BusinessFunctionFilter) {
			node(id: $organizationId) {
				... on Organization {
					businessFunctions(first: $first, after: $after, filter: $filter) {
						edges {
							node {
								id
								referenceId
								name
								classification
								mtdMinutes
								rtoMinutes
								rpoMinutes
								impactTolerance
								notes
								${ownerFragment}
								${assetsFragment}
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

	const variables: IDataObject = { organizationId };
	const filter: IDataObject = {};

	if (filters.classification) {
		filter.classification = filters.classification;
	}

	if (Object.keys(filter).length > 0) {
		variables.filter = filter;
	}

	const businessFunctions = await proboApiRequestAllItems.call(
		this,
		query,
		variables,
		(response) => {
			const data = response?.data as IDataObject | undefined;
			const node = data?.node as IDataObject | undefined;
			return node?.businessFunctions as IDataObject | undefined;
		},
		returnAll,
		limit,
	);

	return {
		json: { businessFunctions },
		pairedItem: { item: itemIndex },
	};
}
