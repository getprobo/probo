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
				resource: ['aiSystem'],
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
				resource: ['aiSystem'],
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
				resource: ['aiSystem'],
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
				resource: ['aiSystem'],
				operation: ['getAll'],
			},
		},
		options: [
			{
				displayName: 'Risk Classification',
				name: 'riskClassification',
				type: 'options',
				options: [
					{ name: 'All', value: '' },
					{ name: 'GPAI', value: 'GPAI' },
					{ name: 'High Risk', value: 'HIGH_RISK' },
					{ name: 'Limited', value: 'LIMITED' },
					{ name: 'Minimal', value: 'MINIMAL' },
				],
				default: '',
				description: 'Filter by risk classification',
			},
			{
				displayName: 'Status',
				name: 'status',
				type: 'options',
				options: [
					{ name: 'All', value: '' },
					{ name: 'Active', value: 'ACTIVE' },
					{ name: 'In Development', value: 'IN_DEVELOPMENT' },
					{ name: 'Decommissioned', value: 'DECOMMISSIONED' },
				],
				default: '',
				description: 'Filter by status',
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
				resource: ['aiSystem'],
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
		status?: string;
		riskClassification?: string;
	};
	const options = this.getNodeParameter('options', itemIndex, {}) as {
		includeOwner?: boolean;
	};

	const ownerFragment = options.includeOwner
		? `owner {
			id
			fullName
			emailAddress
		}`
		: '';

	const query = `
		query GetAiSystems($organizationId: ID!, $first: Int, $after: CursorKey, $filter: AiSystemFilter) {
			node(id: $organizationId) {
				... on Organization {
					aiSystems(first: $first, after: $after, filter: $filter) {
						edges {
							node {
								id
								name
								version
								companyRoles
								status
								source
								purpose
								intendedUseCases
								autonomyLevel
								humanOversightMechanism
								riskClassification
								keyStakeholders
								dataSourcesAndType
								deploymentDate
								lastReviewDate
								nextReviewDate
								notes
								${ownerFragment}
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

	if (filters.status) {
		filter.status = filters.status;
	}

	if (filters.riskClassification) {
		filter.riskClassification = filters.riskClassification;
	}

	if (Object.keys(filter).length > 0) {
		variables.filter = filter;
	}

	const aiSystems = await proboApiRequestAllItems.call(
		this,
		query,
		variables,
		(response) => {
			const data = response?.data as IDataObject | undefined;
			const node = data?.node as IDataObject | undefined;
			return node?.aiSystems as IDataObject | undefined;
		},
		returnAll,
		limit,
	);

	return {
		json: { aiSystems },
		pairedItem: { item: itemIndex },
	};
}
