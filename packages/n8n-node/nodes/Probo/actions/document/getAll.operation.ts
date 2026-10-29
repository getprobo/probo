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
import { documentListFilter } from '../../output';

export const description: INodeProperties[] = [
	organizationLocator({
		resource: ['document'],
		operation: ['getAll'],
	}),
	{
		displayName: 'Return All',
		name: 'returnAll',
		type: 'boolean',
		displayOptions: {
			show: {
				resource: ['document'],
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
				resource: ['document'],
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
		displayName: 'Status',
		name: 'status',
		type: 'multiOptions',
		displayOptions: {
			show: {
				resource: ['document'],
				operation: ['getAll'],
			},
		},
		options: [
			{ name: 'Active', value: 'ACTIVE' },
			{ name: 'Archived', value: 'ARCHIVED' },
		],
		default: ['ACTIVE'],
		description: 'Statuses to include. Active is used when nothing is selected.',
	},
	{
		displayName: 'Filters',
		name: 'filters',
		type: 'collection',
		placeholder: 'Add Filter',
		default: {},
		displayOptions: {
			show: {
				resource: ['document'],
				operation: ['getAll'],
			},
		},
		options: [
			{
				displayName: 'Classifications',
				name: 'classifications',
				type: 'multiOptions',
				default: [],
				description: 'Filter by document classification',
				options: [
					{ name: 'Confidential', value: 'CONFIDENTIAL' },
					{ name: 'Internal', value: 'INTERNAL' },
					{ name: 'Public', value: 'PUBLIC' },
					{ name: 'Secret', value: 'SECRET' },
				],
			},
			{
				displayName: 'Document Types',
				name: 'documentTypes',
				type: 'multiOptions',
				default: [],
				description: 'Filter by document type',
				options: [
					{ name: 'Governance', value: 'GOVERNANCE' },
					{ name: 'Other', value: 'OTHER' },
					{ name: 'Plan', value: 'PLAN' },
					{ name: 'Policy', value: 'POLICY' },
					{ name: 'Procedure', value: 'PROCEDURE' },
					{ name: 'Record', value: 'RECORD' },
					{ name: 'Register', value: 'REGISTER' },
					{ name: 'Report', value: 'REPORT' },
					{ name: 'Statement of Applicability', value: 'STATEMENT_OF_APPLICABILITY' },
					{ name: 'Template', value: 'TEMPLATE' },
				],
			},
			{
				displayName: 'Published',
				name: 'published',
				type: 'options',
				default: 'any',
				description: 'Limit to documents that have or have not been published. Any does not filter.',
				options: [
					{ name: 'Any', value: 'any' },
					{ name: 'Not Published', value: 'false' },
					{ name: 'Published', value: 'true' },
				],
			},
			{
				displayName: 'Query',
				name: 'query',
				type: 'string',
				default: '',
				description: 'Search query to filter documents',
			},
			{
				displayName: 'Write Modes',
				name: 'writeModes',
				type: 'multiOptions',
				default: [],
				description: 'Filter by write mode',
				options: [
					{ name: 'Authored', value: 'AUTHORED' },
					{ name: 'Generated', value: 'GENERATED' },
				],
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
	const status = this.getNodeParameter('status', itemIndex, ['ACTIVE']) as string[];
	const filters = this.getNodeParameter('filters', itemIndex, {}) as IDataObject;
	const filter = documentListFilter(status, filters);

	const query = `
		query GetDocuments($organizationId: ID!, $first: Int, $after: CursorKey, $filter: DocumentFilter) {
			node(id: $organizationId) {
				... on Organization {
					documents(first: $first, after: $after, filter: $filter) {
						edges {
							node {
								id
								status
								currentPublishedMajor
								currentPublishedMinor
								archivedAt
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

	const variables: IDataObject = { organizationId, filter };

	const documents = await proboApiRequestAllItems.call(
		this,
		query,
		variables,
		(response) => {
			const data = response?.data as IDataObject | undefined;
			const node = data?.node as IDataObject | undefined;
			return node?.documents as IDataObject | undefined;
		},
		returnAll,
		limit,
	);

	return executionItems(documents, itemIndex);
}
