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
		displayName: 'Organization ID',
		name: 'organizationId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['document'],
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
				displayName: 'Query',
				name: 'query',
				type: 'string',
				default: '',
				description: 'Search query to filter documents',
			},
			{
				displayName: 'Status',
				name: 'status',
				type: 'multiOptions',
				default: [],
				description: 'Filter by document status',
				options: [
					{ name: 'Active', value: 'ACTIVE' },
					{ name: 'Archived', value: 'ARCHIVED' },
				],
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
): Promise<INodeExecutionData> {
	const organizationId = this.getNodeParameter('organizationId', itemIndex) as string;
	const returnAll = this.getNodeParameter('returnAll', itemIndex) as boolean;
	const limit = this.getNodeParameter('limit', itemIndex, 50) as number;
	const filters = this.getNodeParameter('filters', itemIndex, {}) as IDataObject;

	const filter: IDataObject = {};
	if (filters.query) filter.query = filters.query;
	if ((filters.writeModes as string[])?.length) filter.writeModes = filters.writeModes;
	if ((filters.documentTypes as string[])?.length) filter.documentTypes = filters.documentTypes;
	if ((filters.classifications as string[])?.length) filter.classifications = filters.classifications;
	filter.status = (filters.status as string[])?.length ? filters.status : ['ACTIVE'];

	const query = `
		query GetDocuments($organizationId: ID!, $first: Int, $after: CursorKey, $filter: DocumentFilter) {
			node(id: $organizationId) {
				... on Organization {
					documents(first: $first, after: $after, filter: $filter) {
						edges {
							node {
								id
								status
								trustCenterVisibility
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

	return {
		json: { documents },
		pairedItem: { item: itemIndex },
	};
}
