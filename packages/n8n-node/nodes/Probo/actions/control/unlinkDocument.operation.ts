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
		displayName: 'Control ID',
		name: 'controlId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['control'],
				operation: ['unlinkDocument'],
			},
		},
		default: '',
		description: 'The ID of the control',
		required: true,
	},
	{
		displayName: 'Document ID',
		name: 'documentId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['control'],
				operation: ['unlinkDocument'],
			},
		},
		default: '',
		description: 'The ID of the document to unlink',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const controlId = this.getNodeParameter('controlId', itemIndex) as string;
	const documentId = this.getNodeParameter('documentId', itemIndex) as string;

	const query = `
		mutation DeleteControlDocumentMapping($input: DeleteControlDocumentMappingInput!) {
			deleteControlDocumentMapping(input: $input) {
				deletedControlId
				deletedDocumentId
			}
		}
	`;

	const responseData = await proboApiRequest.call(this, query, { input: { controlId, documentId } });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
