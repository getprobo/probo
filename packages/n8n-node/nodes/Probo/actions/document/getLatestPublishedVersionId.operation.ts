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
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Document ID',
		name: 'documentId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['document'],
				operation: ['getLatestPublishedVersionId'],
			},
		},
		default: '',
		description: 'The ID of the document whose latest published version ID should be returned',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const documentId = this.getNodeParameter('documentId', itemIndex) as string;

	const query = `
		query GetLatestPublishedDocumentVersionId($documentId: ID!) {
			node(id: $documentId) {
				... on Document {
					versions(first: 1, orderBy: { field: CREATED_AT, direction: DESC }, filter: { statuses: [PUBLISHED] }) {
						edges {
							node {
								id
							}
						}
					}
				}
			}
		}
	`;

	const responseData = await proboApiRequest.call(this, query, { documentId });
	const data = responseData.data as IDataObject | undefined;
	const node = data?.node as IDataObject | undefined;
	const versions = node?.versions as IDataObject | undefined;
	const edges = versions?.edges as Array<{ node?: IDataObject }> | undefined;

	return {
		json: {
			documentVersionId: edges?.[0]?.node?.id ?? null,
		},
		pairedItem: { item: itemIndex },
	};
}
