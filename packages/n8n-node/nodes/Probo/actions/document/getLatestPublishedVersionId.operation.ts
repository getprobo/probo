// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
