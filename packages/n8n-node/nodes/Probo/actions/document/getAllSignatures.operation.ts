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
		displayName: 'Document Version ID',
		name: 'documentVersionId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['document'],
				operation: ['getAllSignatures'],
			},
		},
		default: '',
		description: 'The ID of the document version',
		required: true,
	},
	{
		displayName: 'Return All',
		name: 'returnAll',
		type: 'boolean',
		displayOptions: {
			show: {
				resource: ['document'],
				operation: ['getAllSignatures'],
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
				operation: ['getAllSignatures'],
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
				operation: ['getAllSignatures'],
			},
		},
		options: [
			{
				displayName: 'States',
				name: 'states',
				type: 'multiOptions',
				default: [],
				description: 'Filter by signature state',
				options: [
					{ name: 'Requested', value: 'REQUESTED' },
					{ name: 'Signed', value: 'SIGNED' },
				],
			},
			{
				displayName: 'Active Contract',
				name: 'activeContract',
				type: 'boolean',
				default: false,
				description: 'Whether to filter by active contract status',
			},
		],
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const documentVersionId = this.getNodeParameter('documentVersionId', itemIndex) as string;
	const returnAll = this.getNodeParameter('returnAll', itemIndex) as boolean;
	const limit = this.getNodeParameter('limit', itemIndex, 50) as number;
	const filters = this.getNodeParameter('filters', itemIndex, {}) as IDataObject;

	const filter: IDataObject = {};
	if ((filters.states as string[])?.length) filter.states = filters.states;
	if (filters.activeContract !== undefined) filter.activeContract = filters.activeContract;

	const hasFilter = Object.keys(filter).length > 0;

	const query = `
		query GetDocumentVersionSignatures($documentVersionId: ID!, $first: Int, $after: CursorKey${hasFilter ? ', $filter: DocumentVersionSignatureFilter' : ''}) {
			node(id: $documentVersionId) {
				... on DocumentVersion {
					signatures(first: $first, after: $after${hasFilter ? ', filter: $filter' : ''}) {
						edges {
							node {
								id
								state
								signedAt
								requestedAt
								createdAt
								updatedAt
								signedBy {
									id
									fullName
									primaryEmailAddress
								}
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

	const variables: IDataObject = { documentVersionId };
	if (hasFilter) variables.filter = filter;

	const signatures = await proboApiRequestAllItems.call(
		this,
		query,
		variables,
		(response) => {
			const data = response?.data as IDataObject | undefined;
			const node = data?.node as IDataObject | undefined;
			return node?.signatures as IDataObject | undefined;
		},
		returnAll,
		limit,
	);

	return {
		json: { signatures },
		pairedItem: { item: itemIndex },
	};
}
