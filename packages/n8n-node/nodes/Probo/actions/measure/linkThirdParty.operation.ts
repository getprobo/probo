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

import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Measure ID',
		name: 'measureId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['measure'],
				operation: ['linkThirdParty'],
			},
		},
		default: '',
		description: 'The ID of the measure',
		required: true,
	},
	{
		displayName: 'Third Party ID',
		name: 'thirdPartyId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['measure'],
				operation: ['linkThirdParty'],
			},
		},
		default: '',
		description: 'The ID of the third party to link',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const measureId = this.getNodeParameter('measureId', itemIndex) as string;
	const thirdPartyId = this.getNodeParameter('thirdPartyId', itemIndex) as string;

	const query = `
		mutation CreateMeasureThirdPartyMapping($input: CreateMeasureThirdPartyMappingInput!) {
			createMeasureThirdPartyMapping(input: $input) {
				measureEdge {
					node {
						id
						name
					}
				}
				thirdPartyEdge {
					node {
						id
						name
					}
				}
			}
		}
	`;

	const responseData = await proboApiRequest.call(this, query, { input: { measureId, thirdPartyId } });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
