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
import { organizationLocator, documentLocator } from '../../locators';
import { proboApiRequestAllItems, executionItems } from '../../GenericFunctions';
import { documentVersionSelection, withParentReference } from '../../output';

export const description: INodeProperties[] = [
	organizationLocator({
		resource: ['document'],
		operation: ['getAllVersions'],
		required: false,
		description: 'Organization to search documents in. Leave this empty when entering a document ID.',
	}),
	documentLocator({
		resource: ['document'],
		operation: ['getAllVersions'],
	}),
	{
		displayName: 'Return All',
		name: 'returnAll',
		type: 'boolean',
		displayOptions: {
			show: {
				resource: ['document'],
				operation: ['getAllVersions'],
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
				operation: ['getAllVersions'],
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
		displayName: 'Simplify',
		name: 'simple',
		type: 'boolean',
		displayOptions: {
			show: {
				resource: ['document'],
				operation: ['getAllVersions'],
			},
		},
		default: true,
		description: 'Whether to return a simplified version of the response instead of the raw data',
		hint: 'Omits the version body and keeps titles, statuses, and identifiers',
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
				operation: ['getAllVersions'],
			},
		},
		options: [
			{
				displayName: 'Statuses',
				name: 'statuses',
				type: 'multiOptions',
				default: [],
				description: 'Filter by version status',
				options: [
					{ name: 'Pending Approval', value: 'PENDING_APPROVAL' },
					{ name: 'Published', value: 'PUBLISHED' },
					{ name: 'Rejected', value: 'REJECTED' },
					{ name: 'Voided', value: 'VOIDED' },
				],
			},
		],
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData[]> {
	const documentId = this.getNodeParameter('documentId', itemIndex, '', { extractValue: true }) as string;
	const returnAll = this.getNodeParameter('returnAll', itemIndex) as boolean;
	const limit = this.getNodeParameter('limit', itemIndex, 50) as number;
	const simplify = this.getNodeParameter('simple', itemIndex, true) as boolean;
	const filters = this.getNodeParameter('filters', itemIndex, {}) as IDataObject;

	const filter: IDataObject = {};
	if ((filters.statuses as string[])?.length) filter.statuses = filters.statuses;

	const hasFilter = Object.keys(filter).length > 0;

	const query = `
		query GetDocumentVersions($documentId: ID!, $first: Int, $after: CursorKey${hasFilter ? ', $filter: DocumentVersionFilter' : ''}) {
			node(id: $documentId) {
				... on Document {
					versions(first: $first, after: $after${hasFilter ? ', filter: $filter' : ''}) {
						edges {
							node {
								${documentVersionSelection(simplify)}
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

	const variables: IDataObject = { documentId };
	if (hasFilter) variables.filter = filter;

	const versions = await proboApiRequestAllItems.call(
		this,
		query,
		variables,
		(response) => {
			const data = response?.data as IDataObject | undefined;
			const node = data?.node as IDataObject | undefined;
			return node?.versions as IDataObject | undefined;
		},
		returnAll,
		limit,
	);

	return executionItems(
		versions.map((version) => withParentReference(version, { documentId })),
		itemIndex,
	);
}
