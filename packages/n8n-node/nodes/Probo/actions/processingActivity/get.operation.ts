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
		displayName: 'Processing Activity ID',
		name: 'processingActivityId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['processingActivity'],
				operation: ['get'],
			},
		},
		default: '',
		description: 'The ID of the processing activity',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const processingActivityId = this.getNodeParameter('processingActivityId', itemIndex) as string;

	const query = `
		query GetProcessingActivity($processingActivityId: ID!) {
			node(id: $processingActivityId) {
				... on ProcessingActivity {
					id
					name
					purpose
					role
					lawfulBasis
					createdAt
					updatedAt
				}
			}
		}
	`;

	const variables = {
		processingActivityId,
	};

	const responseData = await proboApiRequest.call(this, query, variables);

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
