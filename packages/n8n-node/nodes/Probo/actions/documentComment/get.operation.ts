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
		displayName: 'Thread ID',
		name: 'threadId',
		type: 'string',
		required: true,
		default: '',
		displayOptions: { show: { resource: ['documentComment'], operation: ['get'] } },
	},
];

export async function execute(this: IExecuteFunctions, itemIndex: number): Promise<INodeExecutionData> {
	const threadId = this.getNodeParameter('threadId', itemIndex) as string;
	const query = `
		query GetDocumentCommentThread($threadId: ID!, $first: Int, $after: CursorKey) {
			node(id: $threadId) {
				__typename
				... on DocumentCommentThread {
					id
					anchorQuote
					resolvedAt
					suggestion { id status }
					comments(first: $first, after: $after, orderBy: { field: CREATED_AT, direction: ASC }) {
						edges { node { id content createdAt author { id fullName } } }
						pageInfo { hasNextPage endCursor }
					}
				}
			}
		}
	`;

	let thread: IDataObject | undefined;
	const comments = await proboApiRequestAllItems.call(
		this,
		query,
		{ threadId },
		(response) => {
			const data = response?.data as IDataObject | undefined;
			const node = data?.node as IDataObject | undefined;
			if (!node || node.__typename !== 'DocumentCommentThread') {
				throw new NodeOperationError(this.getNode(), `Document comment thread ${threadId} not found`);
			}
			if (thread === undefined) {
				thread = { ...node };
				delete thread.comments;
				delete thread.__typename;
			}
			return node.comments as IDataObject | undefined;
		},
	);

	return { json: { ...thread, comments }, pairedItem: { item: itemIndex } };
}
