// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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
		displayName: 'ThirdParty Service ID',
		name: 'thirdPartyServiceId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['thirdParty'],
				operation: ['getService'],
			},
		},
		default: '',
		description: 'The ID of the thirdParty service',
		required: true,
	},
	{
		displayName: 'Options',
		name: 'options',
		type: 'collection',
		placeholder: 'Add Option',
		default: {},
		displayOptions: {
			show: {
				resource: ['thirdParty'],
				operation: ['getService'],
			},
		},
		options: [
			{
				displayName: 'Include ThirdParty',
				name: 'includeThirdParty',
				type: 'boolean',
				default: false,
				description: 'Whether to include thirdParty in the response',
			},
		],
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const thirdPartyServiceId = this.getNodeParameter('thirdPartyServiceId', itemIndex) as string;
	const options = this.getNodeParameter('options', itemIndex, {}) as {
		includeThirdParty?: boolean;
	};

	const thirdPartyFragment = options.includeThirdParty
		? `thirdParty {
			id
			name
		}`
		: '';

	const query = `
		query GetThirdPartyService($thirdPartyServiceId: ID!) {
			node(id: $thirdPartyServiceId) {
				... on ThirdPartyService {
					id
					name
					description
					${thirdPartyFragment}
					createdAt
					updatedAt
				}
			}
		}
	`;

	const variables = {
		thirdPartyServiceId,
	};

	const responseData = await proboApiRequest.call(this, query, variables);

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
