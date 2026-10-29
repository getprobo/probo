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
import { proboApiRequestAllItems, executionItems } from '../../GenericFunctions';
import { withParentReference } from '../../output';
import { documentVersionLocator, organizationLocator, documentLocator } from '../../locators';

export const description: INodeProperties[] = [
	organizationLocator({
		resource: ['document'],
		operation: ['getAllApprovalQuorums'],
		required: false,
		description: 'Organization to search in. Leave this empty when entering an ID.',
	}),
	documentLocator({
		resource: ['document'],
		operation: ['getAllApprovalQuorums'],
		required: false,
		description: 'Document to search versions in. Leave this empty when entering a version ID.',
	}),
		documentVersionLocator({
		resource: ['document'],
		operation: ['getAllApprovalQuorums'],
	}),
	{
		displayName: 'Return All',
		name: 'returnAll',
		type: 'boolean',
		displayOptions: {
			show: {
				resource: ['document'],
				operation: ['getAllApprovalQuorums'],
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
				operation: ['getAllApprovalQuorums'],
				returnAll: [false],
			},
		},
		typeOptions: {
			minValue: 1,
		},
		default: 50,
		description: 'Max number of results to return',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData[]> {
	const documentVersionId = this.getNodeParameter('documentVersionId', itemIndex, '', { extractValue: true }) as string;
	const returnAll = this.getNodeParameter('returnAll', itemIndex) as boolean;
	const limit = this.getNodeParameter('limit', itemIndex, 50) as number;

	const query = `
		query GetDocumentVersionApprovalQuorums($documentVersionId: ID!, $first: Int, $after: CursorKey) {
			node(id: $documentVersionId) {
				... on DocumentVersion {
					approvalQuorums(first: $first, after: $after) {
						edges {
							node {
								id
								status
								createdAt
								updatedAt
								documentVersion {
									id
									document {
										id
									}
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

	const approvalQuorums = await proboApiRequestAllItems.call(
		this,
		query,
		{ documentVersionId },
		(response) => {
			const data = response?.data as IDataObject | undefined;
			const node = data?.node as IDataObject | undefined;
			return node?.approvalQuorums as IDataObject | undefined;
		},
		returnAll,
		limit,
	);

	return executionItems(
		approvalQuorums.map((quorum) => withParentReference(quorum, { versionId: documentVersionId })),
		itemIndex,
	);
}
