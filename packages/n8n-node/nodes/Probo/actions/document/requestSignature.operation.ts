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

import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Document Version ID',
		name: 'documentVersionId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['document'],
				operation: ['requestSignature'],
			},
		},
		default: '',
		description: 'The ID of the document version',
		required: true,
	},
	{
		displayName: 'Signatory ID',
		name: 'signatoryId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['document'],
				operation: ['requestSignature'],
			},
		},
		default: '',
		description: 'The profile ID of the signatory',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const documentVersionId = this.getNodeParameter('documentVersionId', itemIndex) as string;
	const signatoryId = this.getNodeParameter('signatoryId', itemIndex) as string;

	const query = `
		mutation RequestSignature($input: RequestSignatureInput!) {
			requestSignature(input: $input) {
				documentVersionSignatureEdge {
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
			}
		}
	`;

	const responseData = await proboApiRequest.call(this, query, {
		input: { documentVersionId, signatoryId },
	});

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
