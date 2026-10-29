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

import type { IDataObject, IExecuteFunctions, INodeExecutionData, INodeProperties } from 'n8n-workflow';
import { NodeOperationError } from 'n8n-workflow';
import { proboApiRequestAllItems } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Document ID',
		name: 'documentId',
		type: 'string',
		required: true,
		default: '',
		displayOptions: { show: { resource: ['documentSuggestion'], operation: ['getAll'] } },
	},
	{
		displayName: 'Status',
		name: 'status',
		type: 'options',
		default: 'OPEN',
		options: [
			{ name: 'Open', value: 'OPEN' },
			{ name: 'Accepted', value: 'ACCEPTED' },
			{ name: 'Rejected', value: 'REJECTED' },
			{ name: 'All', value: 'ALL' },
		],
		displayOptions: { show: { resource: ['documentSuggestion'], operation: ['getAll'] } },
	},
	{
		displayName: 'Return All',
		name: 'returnAll',
		type: 'boolean',
		default: false,
		description: 'Whether to return all results or only up to a given limit',
		displayOptions: { show: { resource: ['documentSuggestion'], operation: ['getAll'] } },
	},
	{
		displayName: 'Limit',
		name: 'limit',
		type: 'number',
		typeOptions: { minValue: 1 },
		default: 50,
		description: 'Max number of results to return',
		displayOptions: { show: { resource: ['documentSuggestion'], operation: ['getAll'], returnAll: [false] } },
	},
];

export async function execute(this: IExecuteFunctions, itemIndex: number): Promise<INodeExecutionData> {
	const documentId = this.getNodeParameter('documentId', itemIndex) as string;
	const status = this.getNodeParameter('status', itemIndex) as string;
	const returnAll = this.getNodeParameter('returnAll', itemIndex) as boolean;
	const limit = this.getNodeParameter('limit', itemIndex, 50) as number;
	const filter = status === 'ALL' ? null : { status };

	const query = `
		query ListDocumentSuggestions($documentId: ID!, $first: Int, $after: CursorKey, $filter: DocumentSuggestionFilter) {
			node(id: $documentId) {
				__typename
				... on Document {
					suggestions(first: $first, after: $after, filter: $filter, orderBy: { field: CREATED_AT, direction: ASC }) {
						edges { node { id status author { id fullName } createdAt } }
						pageInfo { hasNextPage endCursor }
					}
				}
			}
		}
	`;

	const suggestions = await proboApiRequestAllItems.call(
		this,
		query,
		{ documentId, filter },
		(response) => {
			const data = response?.data as IDataObject | undefined;
			const node = data?.node as IDataObject | undefined;
			if (!node || node.__typename !== 'Document') {
				throw new NodeOperationError(this.getNode(), `Document ${documentId} not found`);
			}
			return node.suggestions as IDataObject | undefined;
		},
		returnAll,
		limit,
	);

	return { json: { suggestions }, pairedItem: { item: itemIndex } };
}
