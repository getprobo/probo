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
import { withParentReference } from '../../output';

export const description: INodeProperties[] = [
	organizationLocator({
		resource: ['document'],
		operation: ['getHistory'],
		required: false,
		description: 'Organization to search documents in. Leave this empty when entering a document ID.',
	}),
	documentLocator({
		resource: ['document'],
		operation: ['getHistory'],
	}),
	{
		displayName: 'Return All',
		name: 'returnAll',
		type: 'boolean',
		displayOptions: {
			show: {
				resource: ['document'],
				operation: ['getHistory'],
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
				operation: ['getHistory'],
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
		displayName: 'Only Versioned',
		name: 'versioned',
		type: 'boolean',
		displayOptions: {
			show: {
				resource: ['document'],
				operation: ['getHistory'],
			},
		},
		default: false,
		description: 'Whether to return only the revisions tagged by a version',
		hint: 'Read a version with Get Version, using the ID listed under its revision',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData[]> {
	const documentId = this.getNodeParameter('documentId', itemIndex, '', { extractValue: true }) as string;
	const returnAll = this.getNodeParameter('returnAll', itemIndex) as boolean;
	const limit = this.getNodeParameter('limit', itemIndex, 50) as number;
	const versioned = this.getNodeParameter('versioned', itemIndex, false) as boolean;

	const query = `
		query GetDocumentHistory($documentId: ID!, $first: Int, $after: CursorKey, $filter: DocumentRevisionFilter) {
			node(id: $documentId) {
				... on Document {
					history(first: $first, after: $after, filter: $filter) {
						edges {
							node {
								id
								heads
								authorName
								startedAt
								endedAt
								changeCount
								synthetic
								versions {
									id
									title
									major
									minor
									status
									publishedAt
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

	const revisions = await proboApiRequestAllItems.call(
		this,
		query,
		{ documentId, filter: versioned ? { versioned: true } : {} },
		(response) => {
			const data = response?.data as IDataObject | undefined;
			const node = data?.node as IDataObject | undefined;
			return node?.history as IDataObject | undefined;
		},
		returnAll,
		limit,
	);

	return executionItems(
		revisions.map((revision) => withParentReference(revision, { documentId })),
		itemIndex,
	);
}
