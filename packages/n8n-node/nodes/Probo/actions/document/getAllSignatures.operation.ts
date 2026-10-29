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
import { proboApiRequestAllItems, executionItems, readResourceId } from '../../GenericFunctions';
import { signatureListFilter, withParentReference } from '../../output';
import { documentVersionLocator, organizationLocator, documentLocator } from '../../locators';

export const description: INodeProperties[] = [
	organizationLocator({
		resource: ['document'],
		operation: ['getAllSignatures'],
		required: false,
		description: 'Organization to search in. Leave this empty when entering an ID.',
	}),
	documentLocator({
		resource: ['document'],
		operation: ['getAllSignatures'],
		required: false,
		description: 'Document to search versions in. Leave this empty when entering a version ID.',
	}),
		documentVersionLocator({
		resource: ['document'],
		operation: ['getAllSignatures'],
	}),
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
				displayName: 'Active Contract',
				name: 'activeContract',
				type: 'boolean',
				default: false,
				description: 'Whether to filter by active contract status',
			},
			{
				displayName: 'Signatory Status',
				name: 'profileStates',
				type: 'multiOptions',
				default: [],
				description: 'Filter by the signatory account status',
				options: [
					{ name: 'Active', value: 'ACTIVE' },
					{ name: 'Deactivated', value: 'DEACTIVATED' },
					{ name: 'Pending', value: 'PENDING' },
				],
			},
			{
				displayName: 'Signature Status',
				name: 'states',
				type: 'multiOptions',
				default: [],
				description: 'Filter by signature status',
				options: [
					{ name: 'Requested', value: 'REQUESTED' },
					{ name: 'Signed', value: 'SIGNED' },
				],
			},
		],
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData[]> {
	const documentVersionId = readResourceId(this.getNodeParameter('documentVersionId', itemIndex));
	const returnAll = this.getNodeParameter('returnAll', itemIndex) as boolean;
	const limit = this.getNodeParameter('limit', itemIndex, 50) as number;
	const filters = this.getNodeParameter('filters', itemIndex, {}) as IDataObject;
	const filter = signatureListFilter(filters);
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
								documentVersion {
									id
									document {
										id
									}
								}
								signedBy {
									id
									fullName
									emailAddress
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

	return executionItems(
		signatures.map((signature) => withParentReference(signature, { versionId: documentVersionId })),
		itemIndex,
	);
}
