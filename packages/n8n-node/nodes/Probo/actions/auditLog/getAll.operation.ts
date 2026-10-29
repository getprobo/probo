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
import { organizationLocator } from '../../locators';
import { proboApiRequestAllItems, readResourceId, executionItems } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	organizationLocator({
		resource: ['auditLog'],
		operation: ['getAll'],
	}),
	{
		displayName: 'Return All',
		name: 'returnAll',
		type: 'boolean',
		displayOptions: {
			show: {
				resource: ['auditLog'],
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
				resource: ['auditLog'],
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
				resource: ['auditLog'],
				operation: ['getAll'],
			},
		},
		options: [
			{
				displayName: 'Action',
				name: 'action',
				type: 'string',
				default: '',
				description: 'Filter by action, for example iam:invitation:create',
			},
			{
				displayName: 'Actor ID',
				name: 'actorId',
				type: 'string',
				default: '',
				description: 'Filter by the identity or API key that performed the action',
			},
			{
				displayName: 'Resource ID',
				name: 'resourceId',
				type: 'string',
				default: '',
				description: 'Filter by the ID of the affected resource',
			},
			{
				displayName: 'Resource Type',
				name: 'resourceType',
				type: 'string',
				default: '',
				description: 'Filter by resource type, for example Invitation',
			},
		],
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData[]> {
	const organizationId = readResourceId(this.getNodeParameter('organizationId', itemIndex));
	const returnAll = this.getNodeParameter('returnAll', itemIndex) as boolean;
	const limit = this.getNodeParameter('limit', itemIndex, 50) as number;
	const filters = this.getNodeParameter('filters', itemIndex, {}) as {
		action?: string;
		actorId?: string;
		resourceType?: string;
		resourceId?: string;
	};

	const query = `
		query GetAuditLogEntries($organizationId: ID!, $first: Int, $after: CursorKey, $filter: AuditLogEntryFilter) {
			node(id: $organizationId) {
				... on Organization {
					auditLogEntries(first: $first, after: $after, filter: $filter) {
						edges {
							node {
								id
								actorId
								actorType
								action
								resourceType
								resourceId
								metadata
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

	const variables: IDataObject = { organizationId };
	const filter: IDataObject = {};

	if (filters.action) {
		filter.action = filters.action;
	}

	if (filters.actorId) {
		filter.actorId = filters.actorId;
	}

	if (filters.resourceType) {
		filter.resourceType = filters.resourceType;
	}

	if (filters.resourceId) {
		filter.resourceId = filters.resourceId;
	}

	if (Object.keys(filter).length > 0) {
		variables.filter = filter;
	}

	const auditLogEntries = await proboApiRequestAllItems.call(
		this,
		query,
		variables,
		(response) => {
			const data = response?.data as IDataObject | undefined;
			const node = data?.node as IDataObject | undefined;
			return node?.auditLogEntries as IDataObject | undefined;
		},
		returnAll,
		limit,
	);

	return executionItems(auditLogEntries, itemIndex);
}
