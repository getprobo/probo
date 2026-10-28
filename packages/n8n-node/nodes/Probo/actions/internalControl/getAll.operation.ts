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
				resource: ['internalControl'],
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
				resource: ['internalControl'],
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
				resource: ['internalControl'],
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
				resource: ['internalControl'],
				operation: ['getAll'],
			},
		},
		options: [
			{
				displayName: 'Category',
				name: 'category',
				type: 'string',
				default: '',
				description: 'Only controls in this category',
			},
			{
				displayName: 'Implementation Statuses',
				name: 'implementationStatuses',
				type: 'multiOptions',
				options: [
					{ name: 'Not Implemented', value: 'NOT_IMPLEMENTED' },
					{ name: 'In Progress', value: 'IN_PROGRESS' },
					{ name: 'Implemented', value: 'IMPLEMENTED' },
					{ name: 'Operating', value: 'OPERATING' },
				],
				default: [],
				description: 'Only controls in one of these statuses',
			},
			{
				displayName: 'Needs Attention',
				name: 'needsAttention',
				type: 'boolean',
				default: false,
				description: 'Whether to keep only controls that are not implemented or have an overdue evidence or test date',
			},
			{
				displayName: 'Order By',
				name: 'orderBy',
				type: 'options',
				options: [
					{ name: 'Created At', value: 'CREATED_AT' },
					{ name: 'Name', value: 'NAME' },
					{ name: 'Next Due', value: 'NEXT_DUE' },
				],
				default: 'CREATED_AT',
				description: 'Field used to order the list',
			},
			{
				displayName: 'Owner ID',
				name: 'ownerId',
				type: 'string',
				default: '',
				description: 'Only controls owned by this profile',
			},
			{
				displayName: 'Query',
				name: 'query',
				type: 'string',
				default: '',
				description: 'Search internal controls by name or code',
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
		query?: string;
		implementationStatuses?: string[];
		category?: string;
		ownerId?: string;
		needsAttention?: boolean;
		orderBy?: string;
	};

	const variables: IDataObject = { organizationId };
	const filter: IDataObject = {};
	if (options.query) {
		filter.query = options.query;
	}
	if (options.implementationStatuses?.length) {
		filter.implementationStatuses = options.implementationStatuses;
	}
	if (options.category) {
		filter.category = options.category;
	}
	if (options.ownerId) {
		filter.ownerId = options.ownerId;
	}
	if (options.needsAttention) {
		filter.needsAttention = true;
	}
	if (Object.keys(filter).length > 0) {
		variables.filter = filter;
	}
	if (options.orderBy) {
		variables.orderBy = { field: options.orderBy, direction: 'ASC' };
	}

	const query = `
		query GetInternalControls(
			$organizationId: ID!
			$first: Int
			$after: CursorKey
			$filter: InternalControlFilter
			$orderBy: InternalControlOrder
		) {
			node(id: $organizationId) {
				... on Organization {
					internalControls(first: $first, after: $after, filter: $filter, orderBy: $orderBy) {
						edges {
							node {
								id
								name
								description
								category
								code
								controlType
								nature
								operatingFrequency {
									mode
									interval
									event
								}
								evidenceCadence
								testingCadence
								nextEvidenceDue
								nextTestDue
								implementationStatus
								state
								owner {
									id
									fullName
								}
								reviewer {
									id
									fullName
								}
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

	const internalControls = await proboApiRequestAllItems.call(
		this,
		query,
		variables,
		(response) => {
			const data = response?.data as IDataObject | undefined;
			const node = data?.node as IDataObject | undefined;
			return node?.internalControls as IDataObject | undefined;
		},
		returnAll,
		limit,
	);

	return {
		json: { internalControls },
		pairedItem: { item: itemIndex },
	};
}
